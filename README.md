# Messenger

Веб-мессенджер в стиле Telegram на Go.

## Структура проекта

- `cmd/` — точка входа (main.go)
- `internal/api/` — HTTP-обработчики
- `internal/auth/` — аутентификация
- `internal/chat/` — чаты и сообщения
- `internal/db/` — работа с базой данных
- `internal/models/` — структуры данных
- `internal/user/` — логика пользователей
- `internal/utils/` — утилиты
- `migrations/` — SQL-миграции
- `static/` — статические файлы (фронтенд)

## MVP-функции
- Регистрация и логин
- JWT авторизация
- Создание и получение чатов
- Отправка и получение сообщений
- Добавление в контакты

## Технологии
- Go
- PostgreSQL или SQLite
- REST API
- JWT
# messenger
