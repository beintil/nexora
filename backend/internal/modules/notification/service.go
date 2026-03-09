package notification

import (
	"context"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"
	srverr "telephony/internal/shared/server_error"

	"github.com/google/uuid"
)

type service struct {
	repo        repository
	transaction postgres.Transaction
}

func NewService(repo repository, transaction postgres.Transaction) Service {
	return &service{
		repo:        repo,
		transaction: transaction,
	}
}

func (s *service) GetNotifications(ctx context.Context, userID uuid.UUID) ([]*domain.Notification, srverr.ServerError) {
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "notification.GetNotifications/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	res, err := s.repo.getNotifications(ctx, tx, userID)
	if err != nil {
		return nil, srverr.NewServerError(srverr.ErrInternalServerError, "notification.GetNotifications/repo").SetError(err.Error())
	}
	return res, nil
}

func (s *service) MarkAsRead(ctx context.Context, notificationID uuid.UUID) srverr.ServerError {
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "notification.MarkAsRead/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	if err := s.repo.markAsRead(ctx, tx, notificationID); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "notification.MarkAsRead/repo").SetError(err.Error())
	}
	if err := s.transaction.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "notification.MarkAsRead/commit").SetError(err.Error())
	}
	return nil
}

func (s *service) CreateNotification(ctx context.Context, n *domain.Notification) srverr.ServerError {
	tx, err := s.transaction.BeginTransaction(ctx)
	if err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "notification.CreateNotification/begin").SetError(err.Error())
	}
	defer func() { _ = s.transaction.Rollback(ctx, tx) }()

	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}

	if err := s.repo.createNotification(ctx, tx, n); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "notification.CreateNotification/repo").SetError(err.Error())
	}
	if err := s.transaction.Commit(ctx, tx); err != nil {
		return srverr.NewServerError(srverr.ErrInternalServerError, "notification.CreateNotification/commit").SetError(err.Error())
	}
	return nil
}
