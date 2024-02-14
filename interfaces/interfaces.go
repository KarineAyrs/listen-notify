package interfaces

import (
	"ListenNotifyArticle/domain"
	"context"
)

type UsersRepo interface {
	ListAll(ctx context.Context) ([]domain.User, error)
	ListenNotifications(ctx context.Context) error
}
