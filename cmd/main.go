package main

import (
	"ListenNotifyArticle/adapters/postgres"
	"ListenNotifyArticle/pkg/pg"
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	dsn := "user=postgres password=postgres host=localhost port=5432 dbname=postgres connect_timeout=3 sslmode=disable"

	client := pg.NewClient()
	err := client.Open(dsn)
	if err != nil {
		log.Fatalf("can't establish connection with postgres, err: %v", err)
	}

	err = client.AddListener(dsn)
	if err != nil {
		log.Fatalf("can't establish connection for listener, err: %v", err)
	}

	usersRepo := postgres.NewUsersRepo(client, "users")

	ctx := context.Background()
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s := <-sig:
			return fmt.Errorf("signal received: %s", s)
		}
	})

	group.Go(func() error {
		go func() {
			<-ctx.Done()
			log.Println("app was interrupted")
			_, cancel := context.WithCancel(ctx)
			defer cancel()
		}()

		err := usersRepo.ListenNotifications(ctx)
		if err != nil {
			return fmt.Errorf("listening to channel error: %w", err)
		}
		return nil
	})

	if err = group.Wait(); err != nil {
		log.Printf("app stopped with err: %v", err)
	}
}
