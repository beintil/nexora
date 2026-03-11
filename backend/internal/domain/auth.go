package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserFromAccess struct {
	ID        uuid.UUID
	CompanyID uuid.UUID
	Role      Role
}

// User — пользователь системы.
type User struct {
	ID                   uuid.UUID
	CompanyID            uuid.UUID
	RoleID               Role
	Email                *string
	PasswordHash         string
	FullName             *string
	AvatarURL            *string
	AvatarID             *string // идентификатор файла аватара в хранилище (UUID)
	VerifiedRegistration bool
	IsDelete             bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Profile — данные профиля для личного кабинета (пользователь + название компании).
type Profile struct {
	ID          uuid.UUID
	CompanyID   uuid.UUID
	CompanyName string
	RoleID      Role
	Email       *string
	FullName    *string
	AvatarURL   *string
	AvatarID    *string // идентификатор файла аватара в хранилище (UUID)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UpdateProfileInput — входные данные обновления профиля (только имя).
type UpdateProfileInput struct {
	FullName *string
}

// AuthRegisterInput — входные данные регистрации.
// Обязательно: CompanyName, Email, Password.
type AuthRegisterInput struct {
	CompanyName string
	Email       string
	Password    string
	FullName    string
}

// AuthLoginInput — входные данные входа (по email).
type AuthLoginInput struct {
	Email    string
	Password string
}

// AuthTokens — пара токенов после успешной авторизации.
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}

// SendCodeInput — запрос на отправку кода подтверждения (6-значный код) на email.
type SendCodeInput struct {
	Email string
}

// VerifyLinkInput — подтверждение регистрации по коду (token = 6-значный код из ссылки или формы) и email.
type VerifyLinkInput struct {
	Token string // 6-значный код из query или body
	Email string // email пользователя, для которого подтверждаем регистрацию
}
