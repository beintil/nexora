package config

// StorageConfig — настройки объектного хранилища (Yandex Object Storage / S3). Все поля из env.
type StorageConfig struct {
	AccessKeyID     string `env:"STORAGE_ACCESS_KEY_ID" env-required:"true"`
	SecretAccessKey string `env:"STORAGE_SECRET_ACCESS_KEY" env-required:"true"`
	Bucket          string `env:"STORAGE_BUCKET" env-required:"true"`
	Region          string `env:"STORAGE_REGION" env-required:"true"`          // для Yandex: ru-central1
	Endpoint        string `env:"STORAGE_ENDPOINT" env-required:"true"`        // для Yandex: https://storage.yandexcloud.net
	PublicBaseURL   string `env:"STORAGE_PUBLIC_BASE_URL" env-required:"true"` // публичный URL бакета для ссылок

	AvatarPrefix string `env:"STORAGE_AVATAR_PREFIX" env-required:"true"` // префикс ключа, например avatars
}
