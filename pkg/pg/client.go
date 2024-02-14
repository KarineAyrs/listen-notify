package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"time"
)

type Client struct {
	DB         *sql.DB
	listener   *pq.Listener
	channels   map[string]chan *pq.Notification
	errc       map[string]chan error
	waitCalled bool
}

func NewClient() *Client {
	return &Client{
		channels: make(map[string]chan *pq.Notification),
		errc:     make(map[string]chan error),
	}
}

func (c *Client) Open(dsn string) error {
	var err error

	if c.DB, err = sql.Open("postgres", dsn); err != nil {
		return fmt.Errorf("sql open failed, %w", err)
	}

	if err = c.DB.Ping(); err != nil {
		return fmt.Errorf("sql ping failed, %w", err)
	}

	c.DB.SetMaxOpenConns(8)
	c.DB.SetMaxIdleConns(8)
	c.DB.SetConnMaxLifetime(time.Minute)

	fmt.Println("connected to postgres")

	return nil
}

func (c *Client) Close() {
	if err := c.DB.Close(); err != nil {
		fmt.Printf("postgres db closing failed, err: %v", err)
	}
}

func (c *Client) AddListener(dataSourceName string) error {
	c.listener = pq.NewListener(dataSourceName, time.Second*10, time.Minute, func(event pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Printf("listener error: %s", err.Error())
		}
	})
	return nil
}

func (c *Client) ListenChannel(ctx context.Context, channelName string) (chan *pq.Notification, chan error, error) {
	channel, ok := c.channels[channelName]
	if !ok {
		err := c.listener.Listen(channelName)
		if err != nil {
			if !errors.Is(err, pq.ErrChannelAlreadyOpen) {
				return nil, nil, fmt.Errorf("can't listen to channel %s, err: %w", channelName, err)
			}
		}

		if err = c.listener.Ping(); err != nil {
			return nil, nil, fmt.Errorf("listener ping failed, %w", err)
		}

		c.channels[channelName] = make(chan *pq.Notification)
		channel = c.channels[channelName]
	}

	errc, ok := c.errc[channelName]
	if !ok {
		c.errc[channelName] = make(chan error)
		errc = c.errc[channelName]
	}

	if !c.waitCalled {
		go c.waitForNotification(ctx)
		c.waitCalled = true
	}
	return channel, errc, nil
}

func (c *Client) waitForNotification(ctx context.Context) {
	for {
		select {
		case notification := <-c.listener.Notify:
			fmt.Printf("got notification from DB, channel: %s", notification.Channel)
			c.channels[notification.Channel] <- notification
		case <-time.After(90 * time.Second):
			go c.listener.Ping()
			// Check if there's more work available, just in case it takes
			// a while for the Listener to notice connection loss and
			// reconnect.
			fmt.Println("received no notifications from DB for 90 seconds, checking for new messages")
		case <-ctx.Done():
			for _, errc := range c.errc {
				errc <- ctx.Err()
			}
			return
		}
	}
}

func (c *Client) CloseListener() error {
	return c.listener.Close()
}
