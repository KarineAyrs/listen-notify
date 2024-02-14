package postgres

import (
	"ListenNotifyArticle/adapters/postgres/sqlgen"
	"ListenNotifyArticle/domain"
	"ListenNotifyArticle/interfaces"
	"ListenNotifyArticle/pkg/pg"
	"context"
	"fmt"
	"log"
)

//go:generate sqlc -x -f=sqlc.yaml generate
type pgUsers struct {
	*pg.Client
	queries *sqlgen.Queries
	channel string
}

func NewUsersRepo(client *pg.Client, channelName string) interfaces.UsersRepo {
	return &pgUsers{
		client,
		sqlgen.New(client.DB),
		channelName,
	}
}

func (p *pgUsers) ListAll(ctx context.Context) ([]domain.User, error) {
	q := p.queries
	rows, err := q.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("can't list users, err: %v", err)
	}

	users := make([]domain.User, 0, len(rows))

	for _, row := range rows {
		users = append(users, domain.User{
			ID:        row.User.ID,
			FirstName: row.User.FirstName,
			LastName:  row.User.LastName,
		})
	}

	return users, nil
}

func (p *pgUsers) ListenNotifications(ctx context.Context) error {
	log.Println("listening for notifications in repo")

	notificationc, errc, err := p.ListenChannel(ctx, p.channel)
	if err != nil {
		return fmt.Errorf("can't connect to channel %s, err: %w", p.channel, err)
	}

	for {
		select {
		case res := <-notificationc:
			// do smth with notification
			fmt.Printf("notification in repo %s\n", res.Extra)
		case err := <-errc:
			fmt.Println("got error")
			return err
		}
	}
}
