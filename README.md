# Messenger

Веб-мессенджер в стиле Telegram, написанный на Go с современным веб-интерфейсом.

## 📱 Что это такое?

Представьте себе Telegram, но в веб-версии! Это приложение позволяет:
- 💬 Общаться с друзьями в реальном времени
- 👥 Создавать групповые чаты
- 📎 Отправлять файлы
- 😊 Ставить реакции на сообщения
- 👤 Добавлять контакты
- 🔍 Искать пользователей

## 🎯 Для кого этот проект?

- **Начинающие программисты** - отличный пример современного веб-приложения
- **Студенты** - изучайте Go, WebSocket, JWT и другие технологии
- **Разработчики** - используйте как основу для своих проектов

## 🛠 Технологии (простыми словами)

- **Go** - язык программирования для сервера (как PHP, но быстрее)
- **HTML/CSS/JavaScript** - создание красивого интерфейса
- **SQLite** - база данных (как Excel файл, но для программ)
- **JWT** - "пропуск" для входа в систему (как билет в кино)
- **WebSocket** - технология для мгновенной отправки сообщений
- **bcrypt** - защита паролей (превращает "123456" в "a1b2c3...")

## 📚 Глоссарий для начинающих

| Термин | Простое объяснение |
|--------|-------------------|
| **Backend** | "Мозг" приложения, работает на сервере |
| **Frontend** | "Лицо" приложения, что видит пользователь |
| **API** | Способ для программ общаться друг с другом |
| **JWT токен** | Электронный пропуск в систему |
| **WebSocket** | Телефонная линия для мгновенных сообщений |
| **База данных** | Большая таблица для хранения информации |
| **Хеширование** | Превращение пароля в секретный код |
| **XSS атака** | Попытка взлома через вредоносный код |

## 🚀 Быстрый старт (5 минут)

### Шаг 1: Установка Go
```bash
# Для Ubuntu/Debian
sudo apt update
sudo apt install golang-go

# Для Windows: скачайте с golang.org
# Для Mac: brew install go
```

### Шаг 2: Скачивание проекта
```bash
git clone https://github.com/post-kserks/messenger.git
cd messenger
```

### Шаг 3: Настройка
```bash
# Создаем файл настроек
cat > .env << EOF
# База данных
DB_HOST=localhost
DB_PORT=5432
DB_USER=messenger
DB_PASSWORD=securepassword
DB_NAME=messenger_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=my_super_secret_key
TOKEN_EXPIRATION_HOURS=24

# Сервер
SERVER_PORT=8080

# Хранилище файлов
UPLOAD_DIR=./uploads

# WebSocket
WS_READ_BUFFER=1024
WS_WRITE_BUFFER=1024

# Режим
ENV=development
EOF

# Создаем папку для файлов
mkdir uploads
```

### Шаг 4: Запуск
```bash
go mod download
go run cmd/main.go
```

### Шаг 5: Открываем в браузере
```
http://localhost:8080
```

## 🎮 Как использовать

### Регистрация и вход
1. Откройте `http://localhost:8080`
2. Нажмите "Зарегистрироваться"
3. Введите email, имя пользователя и пароль
4. Войдите в систему

### Создание чата
1. Нажмите "+" рядом с "Чаты"
2. Введите название чата
3. Выберите участников
4. Нажмите "Создать"

### Отправка сообщений
1. Выберите чат из списка
2. Введите сообщение в поле внизу
3. Нажмите Enter или кнопку отправки

### Загрузка файлов
1. В чате нажмите на скрепку 📎
2. Выберите файл
3. Дождитесь загрузки

## 🔧 Установка и запуск (подробно)

### Предварительные требования

1. **Go** версии 1.23 или выше
   ```bash
   go version  # Проверяем версию
   ```
2. **Git** для клонирования репозитория
   ```bash
   git --version  # Проверяем версию
   ```

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
# База данных
DB_HOST=localhost
DB_PORT=5432
DB_USER=messenger
DB_PASSWORD=securepassword
DB_NAME=messenger_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=my_super_secret_key
TOKEN_EXPIRATION_HOURS=24

# Сервер
SERVER_PORT=8080

# Хранилище файлов
UPLOAD_DIR=./uploads

# WebSocket
WS_READ_BUFFER=1024
WS_WRITE_BUFFER=1024

# Режим
ENV=development
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

## ⚙️ Переменные окружения

### Обязательные переменные

| Переменная | Описание | Пример значения |
|------------|----------|-----------------|
| `JWT_SECRET` | Секретный ключ для подписи JWT токенов | `my_super_secret_key` |
| `DB_HOST` | Хост базы данных | `localhost` |
| `DB_PORT` | Порт базы данных | `5432` |
| `DB_USER` | Пользователь базы данных | `messenger` |
| `DB_PASSWORD` | Пароль базы данных | `securepassword` |
| `DB_NAME` | Имя базы данных | `messenger_db` |

### Опциональные переменные

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `SERVER_PORT` | Порт для запуска сервера | `8080` |
| `DB_SSLMODE` | Режим SSL для базы данных | `disable` |
| `TOKEN_EXPIRATION_HOURS` | Время жизни JWT токена в часах | `24` |
| `UPLOAD_DIR` | Директория для загрузки файлов | `./uploads` |
| `WS_READ_BUFFER` | Размер буфера чтения WebSocket | `1024` |
| `WS_WRITE_BUFFER` | Размер буфера записи WebSocket | `1024` |
| `ENV` | Режим работы приложения | `development` |

## 🔌 API Эндпоинты

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

## 📁 Структура проекта

```
messenger/
├── cmd/
│   └── main.go              # 🚀 Точка входа в приложение
├── internal/
│   ├── api/
│   │   ├── router.go        # 🌐 HTTP роутер и обработчики
│   │   └── api_test.go      # 🧪 Тесты API
│   ├── auth/
│   │   ├── auth.go          # 🔐 Аутентификация и регистрация
│   │   └── auth_test.go     # 🧪 Тесты аутентификации
│   ├── chat/
│   │   ├── chat.go          # 💬 Логика чатов
│   │   └── chat_test.go     # 🧪 Тесты чатов
│   ├── db/
│   │   ├── db.go            # 🗄️ Подключение к базе данных
│   │   └── db_test.go       # 🧪 Тесты базы данных
│   ├── models/
│   │   ├── models.go        # 📊 Структуры данных
│   │   └── models_test.go   # 🧪 Тесты моделей
│   ├── user/
│   │   ├── user.go          # 👤 Логика пользователей
│   │   └── user_test.go     # 🧪 Тесты пользователей
│   ├── utils/
│   │   ├── utils.go         # 🛠️ Утилиты
│   │   └── utils_test.go    # 🧪 Тесты утилит
│   └── ws/
│       ├── ws.go            # 🔌 WebSocket обработчики
│       └── ws_test.go       # 🧪 Тесты WebSocket
├── migrations/
│   └── 001_init.sql         # 🗃️ Инициализация базы данных
├── static/
│   ├── app.js               # 🎨 Основной JavaScript
│   ├── index.html           # 🏠 Главная страница
│   ├── login.html           # 🔑 Страница входа
│   ├── register.html        # 📝 Страница регистрации
│   └── style.css            # 🎨 Стили
├── uploads/                 # 📁 Загруженные файлы
├── .env                     # ⚙️ Переменные окружения
├── .gitignore              # 🚫 Исключения Git
├── go.mod                  # 📦 Зависимости Go
├── go.sum                  # 🔒 Хеши зависимостей
└── README.md               # 📖 Документация
```

## 🗄️ База данных

Приложение использует SQLite3 с следующими таблицами:

- `users` - пользователи системы
- `contacts` - контакты пользователей
- `chats` - чаты (личные и групповые)
- `chat_members` - участники чатов
- `messages` - сообщения
- `chat_reads` - отметки о прочтении

## 🔒 Безопасность

- Пароли хешируются с помощью bcrypt
- JWT токены для аутентификации
- Защита от XSS атак (фильтрация HTML тегов)
- Валидация входных данных
- Проверка прав доступа к чатам

## 🐛 Troubleshooting (Решение проблем)

### Проблема: "Port already in use"
```bash
# Решение: измените порт в .env файле
SERVER_PORT=8081
```

### Проблема: "Permission denied"
```bash
# Решение: дайте права на папку uploads
chmod 755 uploads
```

### Проблема: "Database locked"
```bash
# Решение: удалите файл базы и перезапустите
rm messenger.db
go run cmd/main.go
```

### Проблема: "JWT_SECRET not set"
```bash
# Решение: добавьте JWT_SECRET в .env файл
echo "JWT_SECRET=my-secret-key-123" >> .env
```

### Проблема: "Go version too old"
```bash
# Решение: обновите Go
# Ubuntu/Debian:
sudo apt update && sudo apt upgrade golang-go

# Или скачайте с golang.org
```

## 🧪 Разработка

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

## 🤝 Как внести свой вклад

1. **Fork** репозитория
2. Создайте **branch** для новой функции
3. Внесите изменения
4. Напишите тесты
5. Создайте **Pull Request**

## 📝 Примеры использования

### Создание чата через API
```bash
curl -X POST http://localhost:8080/api/chats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Мой первый чат",
    "user_ids": [1, 2]
  }'
```

### Отправка сообщения
```bash
curl -X POST http://localhost:8080/api/messages \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "chat_id": 1,
    "text": "Привет всем!"
  }'
```

## 🎯 Что изучать дальше

После изучения этого проекта рекомендую:

1. **Go**: Изучите goroutines и channels
2. **WebSocket**: Понимание real-time коммуникации
3. **JWT**: Токены и безопасность
4. **SQLite**: Работа с базами данных
5. **Frontend**: JavaScript, HTML, CSS

## 📚 Полезные ресурсы

- [Go Documentation](https://golang.org/doc/)
- [WebSocket Guide](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)
- [JWT Introduction](https://jwt.io/introduction)
- [SQLite Tutorial](https://www.sqlitetutorial.net/)

## 📄 Лицензия

MIT License

## 🆘 Поддержка

Если у вас возникли вопросы или проблемы:

1. Проверьте раздел **Troubleshooting** выше
2. Создайте **Issue** в репозитории
3. Опишите проблему подробно с логами

---

**Удачи в изучении! 🚀**
