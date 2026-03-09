package notification

import (
	"context"
	"fmt"
	"telephony/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repository interface {
	getNotifications(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]*domain.Notification, error)
	markAsRead(ctx context.Context, tx pgx.Tx, notificationID uuid.UUID) error
	createNotification(ctx context.Context, tx pgx.Tx, n *domain.Notification) error
}

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) getNotifications(ctx context.Context, tx pgx.Tx, userID uuid.UUID) ([]*domain.Notification, error) {
	const query = `
		SELECT id, user_id, company_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50`
	rows, err := tx.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("notification.GetNotifications: %w", err)
	}
	defer rows.Close()

	var res []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.CompanyID, &n.Type, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		res = append(res, &n)
	}
	return res, nil
}

func (r *repositoryImpl) markAsRead(ctx context.Context, tx pgx.Tx, notificationID uuid.UUID) error {
	const query = `UPDATE notifications SET is_read = true WHERE id = $1`
	_, err := tx.Exec(ctx, query, notificationID)
	return err
}

func (r *repositoryImpl) createNotification(ctx context.Context, tx pgx.Tx, n *domain.Notification) error {
	const query = `
		INSERT INTO notifications (id, user_id, company_id, type, title, message, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := tx.Exec(ctx, query, n.ID, n.UserID, n.CompanyID, string(n.Type), n.Title, n.Message, n.IsRead, n.CreatedAt)
	return err
}
