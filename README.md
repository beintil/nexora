# Nexora

Платформа для учёта и управления телефонией: мультитенантные компании, пользователи, звонки, интеграция с Twilio и веб-интерфейс для дашбордов и аналитики.

## Что это

**Nexora** — это монорепозиторий с бэкендом (Go) и фронтендом (React). Сервис позволяет:

- **Регистрировать компании** и пользователей (email/пароль, верификация по ссылке).
- **Управлять звонками**: приём вебхуков от Twilio (Voice Status Callback), сохранение событий и деталей звонков.
- **Работать с телефонией**: привязка телефонии к компаниям, пайплайн приёма и обработки входящих данных.
- **Просматривать данные в веб-интерфейсе**: дашборды, список звонков, аналитика, профиль и настройки.

Архитектура модульная (transport → service → repository), контракт API описан в едином Swagger и используется и бэкендом, и фронтом.

## Структура репозитория

```
nexora/
├── backend/          # Go-сервис 
├── front/
│   └── nexora/       # React + TypeScript + Vite
├── swagger.yaml      # Общий контракт API (NEXORA API)
├── LICENSE
└── README.md
```

## Требования

- **Backend:** Go 1.24+, PostgreSQL, Redis. Опционально: Yandex Object Storage (S3-совместимый) для аватаров, SMTP для писем.
- **Frontend:** Node.js 18+, npm/yarn. Типы для API генерируются из `swagger.yaml`.

## Запуск

### Backend

```bash
cd backend
cp .env.example .env   # заполнить переменные
make run               # или go run ./cmd/app
```

Конфиги окружения: `configs/local.json`, `configs/dev.json`, `configs/prod.json`.

### Frontend

```bash
cd front/nexora
cp .env.example .env
npm install
npm run dev
```

Генерация типов из Swagger:

```bash
npm run openapi:generate
```

### Swagger

Схема API в корне: `swagger.yaml`. Бэкенд и фронт используют её как единый контракт (генерация моделей/типов).

## Лицензия

Проприетарная лицензия — все права защищены. См. [LICENSE](LICENSE).
