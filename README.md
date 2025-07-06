# Messenger

Веб-мессенджер в стиле Telegram, написанный на Go с современным веб-интерфейсом.

## Описание

Это полнофункциональное веб-приложение для обмена сообщениями с поддержкой:
- Регистрации и аутентификации пользователей
- Создания личных и групповых чатов
- Отправки текстовых сообщений в реальном времени
- Загрузки файлов
- Реакций на сообщения (эмодзи)
- Поиска пользователей
- Системы контактов
- Отметок о прочтении сообщений

## Технологии

- **Backend**: Go 1.23+
- **Frontend**: HTML5, CSS3, JavaScript (ES6+)
- **База данных**: SQLite3
- **Аутентификация**: JWT токены
- **Реальное время**: WebSocket соединения
- **Безопасность**: bcrypt хеширование паролей, защита от XSS

## Установка и запуск

### Предварительные требования

1. **Go** версии 1.23 или выше
2. **Git** для клонирования репозитория

### Пошаговая установка

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/post-kserks/messenger.git
cd messenger
```

2. **Установите зависимости:**
```bash
go mod download
```

3. **Создайте файл `.env` в корне проекта:**
```bash
# Основные настройки сервера
SERVER_PORT=8080

# База данных
DB_PATH=messenger.db

# JWT секрет для аутентификации (обязательно!)
JWT_SECRET=your-super-secret-jwt-key-here

# Директория для загрузки файлов
UPLOAD_DIR=uploads
```

4. **Создайте директорию для загрузок:**
```bash
mkdir uploads
```

5. **Запустите приложение:**
```bash
go run cmd/main.go
```

6. **Откройте браузер и перейдите по адресу:**
```
http://localhost:8080
```

## Переменные окружения

### Обязательные переменные

| Переменная | Описание | Пример значения |
|------------|----------|-----------------|
| `JWT_SECRET` | Секретный ключ для подписи JWT токенов | `my-super-secret-key-123` |

### Опциональные переменные

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `SERVER_PORT` | Порт для запуска сервера | `8080` |
| `DB_PATH` | Путь к файлу базы данных SQLite | `messenger.db` |
| `UPLOAD_DIR` | Директория для загрузки файлов | `uploads` |

## API Эндпоинты

### Аутентификация

#### Регистрация
```
POST /api/register
Content-Type: application/json

{
  "username": "user123",
  "email": "user@example.com",
  "password": "password123"
}
```

#### Вход в систему
```
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

### Чаты

#### Создание чата
```
POST /api/chats
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "name": "Название чата",
  "user_ids": [1, 2, 3]
}
```

#### Получение списка чатов
```
GET /api/chats
Authorization: Bearer <jwt_token>
```

### Сообщения

#### Отправка сообщения
```
POST /api/messages
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "chat_id": 1,
  "text": "Привет, как дела?"
}
```

#### Получение истории сообщений
```
GET /api/messages?chat_id=1
Authorization: Bearer <jwt_token>
```

### Контакты

#### Добавление контакта
```
POST /api/contacts
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "user_id": 1,
  "contact_id": 2
}
```

#### Получение списка контактов
```
GET /api/contacts
Authorization: Bearer <jwt_token>
```

#### Удаление контакта
```
DELETE /api/contacts
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "user_id": 1,
  "contact_id": 2
}
```

### Пользователи

#### Получение информации о пользователе
```
GET /api/user
Authorization: Bearer <jwt_token>
```

#### Обновление профиля
```
PUT /api/user
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "id": 1,
  "username": "new_username",
  "email": "new@example.com",
  "password": "new_password"
}
```

#### Поиск пользователей
```
GET /api/search_users?q=username
Authorization: Bearer <jwt_token>
```

### Файлы

#### Загрузка файла
```
POST /api/upload
Authorization: Bearer <jwt_token>
Content-Type: multipart/form-data

file: <файл>
```

### Реакции

#### Добавление реакции
```
POST /api/reaction
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "message_id": 1,
  "user_id": 1,
  "emoji": "👍"
}
```

#### Получение реакций
```
GET /api/reactions?message_id=1
Authorization: Bearer <jwt_token>
```

### WebSocket

#### Подключение к WebSocket
```
ws://localhost:8080/ws?token=<jwt_token>
```

## Структура проекта

```
messenger/
├── cmd/
│   └── main.go              # Точка входа в приложение
├── internal/
│   ├── api/
│   │   ├── router.go        # HTTP роутер и обработчики
│   │   └── api_test.go      # Тесты API
│   ├── auth/
│   │   ├── auth.go          # Аутентификация и регистрация
│   │   └── auth_test.go     # Тесты аутентификации
│   ├── chat/
│   │   ├── chat.go          # Логика чатов
│   │   └── chat_test.go     # Тесты чатов
│   ├── db/
│   │   ├── db.go            # Подключение к базе данных
│   │   └── db_test.go       # Тесты базы данных
│   ├── models/
│   │   ├── models.go        # Структуры данных
│   │   └── models_test.go   # Тесты моделей
│   ├── user/
│   │   ├── user.go          # Логика пользователей
│   │   └── user_test.go     # Тесты пользователей
│   ├── utils/
│   │   ├── utils.go         # Утилиты
│   │   └── utils_test.go    # Тесты утилит
│   └── ws/
│       ├── ws.go            # WebSocket обработчики
│       └── ws_test.go       # Тесты WebSocket
├── migrations/
│   └── 001_init.sql         # Инициализация базы данных
├── static/
│   ├── app.js               # Основной JavaScript
│   ├── index.html           # Главная страница
│   ├── login.html           # Страница входа
│   ├── register.html        # Страница регистрации
│   └── style.css            # Стили
├── uploads/                 # Загруженные файлы
├── .env                     # Переменные окружения
├── .gitignore              # Исключения Git
├── go.mod                  # Зависимости Go
├── go.sum                  # Хеши зависимостей
└── README.md               # Документация
```

## База данных

Приложение использует SQLite3 с следующими таблицами:

- `users` - пользователи системы
- `contacts` - контакты пользователей
- `chats` - чаты (личные и групповые)
- `chat_members` - участники чатов
- `messages` - сообщения
- `chat_reads` - отметки о прочтении

## Безопасность

- Пароли хешируются с помощью bcrypt
- JWT токены для аутентификации
- Защита от XSS атак (фильтрация HTML тегов)
- Валидация входных данных
- Проверка прав доступа к чатам

## Разработка

### Запуск тестов
```bash
go test ./...
```

### Сборка для продакшена
```bash
go build -o messenger cmd/main.go
```

### Запуск собранного приложения
```bash
./messenger
```

## Лицензия

MIT License

## Поддержка

Если у вас возникли вопросы или проблемы, создайте issue в репозитории.
