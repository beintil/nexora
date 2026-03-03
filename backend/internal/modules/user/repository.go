package user

import (
	"context"
	"errors"
	"fmt"
	"telephony/internal/domain"
	"telephony/internal/shared/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	errRepoUserNotFound  = errors.New("user not found")
	errRepoUserDuplicate = errors.New("user duplicate")
)

type repository interface {
	setUserVerified(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error
	createUser(ctx context.Context, tx pgx.Tx, account *domain.User) error
	getUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*domain.User, error)
	getUserByEmail(ctx context.Context, tx pgx.Tx, email string) (*domain.User, error)
	updateUserProfile(ctx context.Context, tx pgx.Tx, userID uuid.UUID, fullName, avatarURL, avatarID *string) error
}

type repositoryImpl struct{}

func NewRepository() repository {
	return &repositoryImpl{}
}

func (r *repositoryImpl) setUserVerified(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	const query = `UPDATE users SET verified_registration = true WHERE id = $1`
	_, err := tx.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("user.SetUserVerified: %w", err)
	}
	return nil
}

func (r *repositoryImpl) createUser(ctx context.Context, tx pgx.Tx, account *domain.User) error {
	const query = `
		INSERT INTO users (id, company_id, role_id, email, password_hash, full_name, avatar_url, avatar_id, verified_registration)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	var email, fullName, avatarURL, avatarID *string
	if account.Email != nil {
		email = account.Email
	}
	if account.FullName != nil {
		fullName = account.FullName
	}
	if account.AvatarURL != nil {
		avatarURL = account.AvatarURL
	}
	if account.AvatarID != nil {
		avatarID = account.AvatarID
	}
	err := tx.QueryRow(ctx, query,
		account.ID, account.CompanyID, int16(account.RoleID), email, account.PasswordHash, fullName, avatarURL, avatarID, account.VerifiedRegistration,
	).Scan(&account.CreatedAt, &account.UpdatedAt)
	if err != nil {
		if postgres.ErrorIs(err, postgres.DuplicateKeyValueViolatesUniqueConstraint) {
			return errRepoUserDuplicate
		}
		return err
	}
	return nil
}

func (r *repositoryImpl) getUserByEmail(ctx context.Context, tx pgx.Tx, email string) (*domain.User, error) {
	const query = `SELECT id, company_id, role_id, email, password_hash, full_name, avatar_url, avatar_id, verified_registration, created_at, updated_at
		FROM users WHERE email = $1`
	row := tx.QueryRow(ctx, query, email)
	var u domain.User
	var roleID int16
	var emailPtr, fullNamePtr, avatarURLPtr, avatarIDPtr *string
	if err := row.Scan(&u.ID, &u.CompanyID, &roleID, &emailPtr, &u.PasswordHash, &fullNamePtr, &avatarURLPtr, &avatarIDPtr, &u.VerifiedRegistration, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoUserNotFound
		}
		return nil, err
	}
	u.RoleID = domain.Role(roleID)
	u.Email = emailPtr
	u.FullName = fullNamePtr
	u.AvatarURL = avatarURLPtr
	u.AvatarID = avatarIDPtr
	return &u, nil
}

func (r *repositoryImpl) getUserByID(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (*domain.User, error) {
	const query = `SELECT id, company_id, role_id, email, password_hash, full_name, avatar_url, avatar_id, verified_registration, created_at, updated_at
		FROM users WHERE id = $1`
	row := tx.QueryRow(ctx, query, userID)
	var u domain.User
	var roleID int16
	var emailPtr, fullNamePtr, avatarURLPtr, avatarIDPtr *string
	if err := row.Scan(&u.ID, &u.CompanyID, &roleID, &emailPtr, &u.PasswordHash, &fullNamePtr, &avatarURLPtr, &avatarIDPtr, &u.VerifiedRegistration, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if postgres.ErrorIs(err, postgres.ErrNoRows) {
			return nil, errRepoUserNotFound
		}
		return nil, err
	}
	u.RoleID = domain.Role(roleID)
	u.Email = emailPtr
	u.FullName = fullNamePtr
	u.AvatarURL = avatarURLPtr
	u.AvatarID = avatarIDPtr
	return &u, nil
}

func (r *repositoryImpl) updateUserProfile(ctx context.Context, tx pgx.Tx, userID uuid.UUID, fullName, avatarURL, avatarID *string) error {
	const query = `UPDATE users SET full_name = COALESCE($2, full_name), avatar_url = COALESCE($3, avatar_url), avatar_id = COALESCE($4, avatar_id), updated_at = NOW() WHERE id = $1`
	_, err := tx.Exec(ctx, query, userID, fullName, avatarURL, avatarID)
	if err != nil {
		return err
	}
	return nil
}
