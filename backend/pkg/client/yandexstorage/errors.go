package yandexstorage

import "errors"

var (
	// ErrUpload — ошибка при загрузке объекта в бакет.
	ErrUpload = errors.New("yandexstorage upload failed")
	// ErrDelete — ошибка при удалении объекта из бакета.
	ErrDelete = errors.New("yandexstorage delete failed")
	// ErrPing — ошибка при проверочном запросе (ping) к бакету.
	ErrPing = errors.New("yandexstorage ping failed")
)
