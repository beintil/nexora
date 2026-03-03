package config

// Константы модуля user (профиль пользователя, аватар).
const (
	UserMaxFullNameLength = 256
	UserMaxAvatarSize     = 5 * 1024 * 1024 // 5 MiB
)

// UserAllowedAvatarContentTypes — допустимые MIME-типы для аватара.
var UserAllowedAvatarContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}
