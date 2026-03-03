package yandexstorage

import (
	"bytes"
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client описывает методы для работы с бакетом Yandex Object Storage (S3-совместимым).
// Реализация инкапсулирует создание S3-клиента и построение публичных URL.
type Client interface {
	// Upload загружает объект в бакет по ключу key и возвращает публичный URL.
	Upload(ctx context.Context, key string, body []byte, contentType string) (string, error)
	// Delete удаляет объект по ключу key.
	Delete(ctx context.Context, key string) error
	// PublicURL возвращает публичный URL для уже загруженного объекта.
	PublicURL(key string) string
	// Ping выполняет дешёвый запрос к бакету (HEAD Bucket), проверяя доступность
	// и корректность кредов. Предназначен для использования на старте приложения.
	Ping(ctx context.Context) error
}

type client struct {
	s3            *s3.Client
	bucket        string
	publicBaseURL string
}

// NewClient создаёт клиента для Yandex Object Storage.
// Все параметры обязательны, кроме region: если он пустой, используется ru-central1.
// endpoint — например https://storage.yandexcloud.net.
// publicBaseURL — базовый публичный URL бакета, например https://storage.yandexcloud.net/bucket-name.
func NewClient(
	accessKeyID,
	secretAccessKey,
	bucket,
	region,
	endpoint,
	publicBaseURL string,
) (Client, error) {
	if accessKeyID == "" || secretAccessKey == "" || bucket == "" || endpoint == "" || publicBaseURL == "" || region == "" {
		return nil, fmt.Errorf("yandexstorage.NewClient: accessKeyID, secretAccessKey, bucket, endpoint and publicBaseURL are required")
	}

	credProvider := credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")
	opts := s3.Options{
		Region:      region,
		Credentials: credProvider,
	}
	// Для Yandex Object Storage требуется path-style и собственный endpoint.
	opts.UsePathStyle = true
	opts.BaseEndpoint = aws.String(endpoint)

	s3Client := s3.New(opts)

	return &client{
		s3:            s3Client,
		bucket:        bucket,
		publicBaseURL: trimTrailingSlash(publicBaseURL),
	}, nil
}

func (c *client) Upload(ctx context.Context, key string, body []byte, contentType string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("%w: empty key", ErrUpload)
	}
	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUpload, err)
	}
	return c.PublicURL(key), nil
}

func (c *client) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("%w: empty key", ErrDelete)
	}
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDelete, err)
	}
	return nil
}

func (c *client) PublicURL(key string) string {
	if key == "" {
		return ""
	}
	return c.publicBaseURL + "/" + c.bucket + "/" + key
}

// Ping выполняет HEAD Bucket для проверки доступа и кредов.
func (c *client) Ping(ctx context.Context) error {
	_, err := c.s3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(c.bucket),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPing, err)
	}
	return nil
}

func trimTrailingSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
