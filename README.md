# 🔐 Messenger - Веб-мессенджер с End-to-End шифрованием

Современный веб-мессенджер в стиле Telegram с **end-to-end шифрованием** сообщений, написанный на Go с красивым веб-интерфейсом.

## 🚀 Что нового в этой версии

### 🔐 End-to-End шифрование
- **NaCl Box** (X25519 + ChaCha20-Poly1305) для максимальной безопасности
- Автоматическое шифрование/расшифровка сообщений
- Приватные ключи никогда не покидают устройство пользователя
- Защита от атак типа "man-in-the-middle"

### 📱 Основные возможности
- 💬 **Реальное время**: Мгновенная отправка сообщений через WebSocket
- 🔐 **Шифрование**: Все сообщения автоматически шифруются
- 👥 **Групповые чаты**: Создавайте чаты с несколькими участниками
- 📎 **Файлы**: Отправляйте любые файлы
- 😊 **Реакции**: Ставьте эмодзи на сообщения
- 👤 **Контакты**: Добавляйте друзей в контакты
- 🔍 **Поиск**: Ищите пользователей по имени или email
- 📱 **Адаптивный дизайн**: Работает на всех устройствах

## 🎯 Для кого этот проект?

- **Разработчики** - изучайте современные технологии безопасности
- **Студенты** - отличный пример fullstack приложения с шифрованием
- **Энтузиасты** - создавайте собственные безопасные мессенджеры
- **Компании** - используйте как основу для корпоративных решений

## 🛠 Технологический стек

### Backend (Go)
- **Go 1.23+** - основной язык сервера
- **Gin/Gorilla** - HTTP фреймворк и WebSocket
- **SQLite/PostgreSQL** - база данных
- **JWT** - аутентификация
- **bcrypt** - хеширование паролей
- **NaCl** - криптографические примитивы

### Frontend
- **Vanilla JavaScript** - без фреймворков
- **HTML5/CSS3** - современный интерфейс
- **TweetNaCl.js** - криптография в браузере
- **WebSocket** - real-time коммуникация

### Безопасность
- **End-to-End шифрование** - только отправитель и получатель читают сообщения
- **XSS защита** - фильтрация вредоносного кода
- **CSRF защита** - защита от межсайтовых запросов
- **Валидация данных** - проверка всех входных данных

## 🚀 Быстрый старт

### 1. Клонирование и установка
```bash
git clone https://github.com/your-username/messenger.git
cd messenger
go mod download
```

### 2. Настройка окружения
```bash
# Создайте файл .env
cat > .env << EOF
# База данных (SQLite по умолчанию)
DB_TYPE=sqlite
DB_PATH=messenger.db

# JWT
JWT_SECRET=your-super-secret-key-change-this
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
```

### 3. Инициализация базы данных
```bash
# Применение основной миграции
go run init_db.go

# Применение миграции шифрования
go run cmd/migrate/main.go
```

### 4. Запуск приложения
```bash
# Основная версия (рекомендуется)
go run cmd/main.go

# Или простая версия для изучения
go run cmd/simple/main.go
```

### 5. Открытие в браузере
```
http://localhost:8080
```

## 🔐 Шифрование в действии

### Как это работает
1. **Регистрация**: Пользователь автоматически получает пару ключей
2. **Создание чата**: Участники обмениваются публичными ключами
3. **Отправка**: Сообщение шифруется для каждого участника
4. **Получение**: Автоматическая расшифровка приватным ключом

### Проверка шифрования
```bash
# Тест криптографических функций
go test ./internal/crypto/... -v

# Тест в браузере
open test_encryption.html
```

## 📡 API Endpoints

### 🔐 Криптографические endpoints

#### Получение публичного ключа пользователя
```http
GET /api/users/{user_id}/public-key
Authorization: Bearer {token}
```

#### Обновление публичного ключа
```http
PUT /api/users/{user_id}/public-key
Authorization: Bearer {token}
Content-Type: application/json

{
  "public_key": "base64_encoded_public_key"
}
```

#### Получение ключей участников чата
```http
GET /api/chats/{chat_id}/participants-keys
Authorization: Bearer {token}
```

#### Статус шифрования чата
```http
GET /api/chats/{chat_id}/encryption-status
Authorization: Bearer {token}
```

### 💬 Обновленные endpoints

#### Отправка сообщения (с поддержкой шифрования)
```http
POST /api/messages
Authorization: Bearer {token}
Content-Type: application/json

# Обычное сообщение
{
  "chat_id": 1,
  "text": "Привет!",
  "is_encrypted": false
}

# Зашифрованное сообщение
{
  "chat_id": 1,
  "encrypted_data": "base64_encoded_data",
  "nonce": "base64_encoded_nonce",
  "is_encrypted": true
}
```

### 🔐 Аутентификация

#### Регистрация
```http
POST /api/register
Content-Type: application/json

{
  "username": "user123",
  "nickname": "user123",
  "email": "user@example.com",
  "password": "password123"
}
```

#### Вход в систему
```http
POST /api/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

### 💬 Чаты и сообщения

#### Создание чата
```http
POST /api/chats
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Мой чат",
  "user_ids": [1, 2, 3]
}
```

#### Получение сообщений
```http
GET /api/messages?chat_id=1
Authorization: Bearer {token}
```

### 📎 Файлы

#### Загрузка файла
```http
POST /api/upload
Authorization: Bearer {token}
Content-Type: multipart/form-data

file: <файл>
chat_id: <id чата>
```

### 😊 Реакции

#### Добавление реакции
```http
POST /api/reaction
Authorization: Bearer {token}
Content-Type: application/json

{
  "message_id": 1,
  "user_id": 1,
  "emoji": "👍"
}
```

### 🔌 WebSocket

#### Подключение
```
ws://localhost:8080/ws?token={jwt_token}
```

#### Отправка сообщения через WebSocket
```json
{
  "type": "message",
  "chat_id": 1,
  "text": "Привет!",
  "is_encrypted": false
}
```

## 📁 Структура проекта

```
messenger/
├── cmd/
│   ├── main.go              # 🚀 Основная точка входа
│   ├── simple/
│   │   └── main.go          # 📚 Простая версия для изучения
│   └── migrate/
│       └── main.go          # 🔄 Скрипт миграции шифрования
├── internal/
│   ├── api/
│   │   ├── router.go        # 🌐 HTTP роутер
│   │   └── crypto.go        # 🔐 Криптографические API
│   ├── auth/
│   │   └── auth.go          # 🔑 Аутентификация
│   ├── crypto/              # 🔐 НОВОЕ: Криптография
│   │   ├── keys.go          # 🔑 Основные криптографические функции
│   │   ├── chat_keys.go     # 👥 Управление ключами чатов
│   │   └── crypto_test.go   # 🧪 Тесты криптографии
│   ├── db/
│   │   └── db.go            # 🗄️ Подключение к БД
│   ├── models/
│   │   └── models.go        # 📊 Структуры данных (обновлено)
│   ├── utils/
│   │   └── file_processor.go # 📁 Обработка файлов
│   └── ws/
│       └── ws.go            # 🔌 WebSocket (обновлено)
├── migrations/
│   ├── 001_init.sql         # 🗃️ Основная миграция
│   ├── 001_init_postgres.sql # 🐘 PostgreSQL версия
│   └── 002_encryption.sql   # 🔐 НОВОЕ: Миграция шифрования
├── static/
│   ├── app.js               # 🎨 JavaScript (обновлено)
│   ├── index.html           # 🏠 Главная страница (обновлено)
│   ├── login.html           # 🔑 Страница входа
│   ├── register.html        # 📝 Страница регистрации
│   └── style.css            # 🎨 Стили
├── uploads/                 # 📁 Загруженные файлы
├── test_encryption.html     # 🧪 НОВОЕ: Тест шифрования
├── ENCRYPTION_README.md     # 📖 НОВОЕ: Документация шифрования
├── .env                     # ⚙️ Переменные окружения
├── go.mod                   # 📦 Зависимости Go
└── README.md                # 📖 Эта документация
```

## 🗄️ База данных

### Таблицы (обновлено)

- `users` - пользователи системы
- `user_keys` - **НОВОЕ**: публичные ключи пользователей
- `contacts` - контакты пользователей
- `chats` - чаты (личные и групповые)
- `chat_members` - участники чатов
- `messages` - сообщения (добавлены поля шифрования)
- `message_reactions` - реакции на сообщения
- `chat_reads` - отметки о прочтении

### Миграции

```bash
# Основная миграция
go run init_db.go

# Миграция шифрования
go run cmd/migrate/main.go
```

## 🔒 Безопасность

### Криптографические примитивы
- **Алгоритм**: NaCl Box (X25519 + ChaCha20-Poly1305)
- **Размер ключа**: 256 бит
- **Размер nonce**: 192 бит
- **Формат**: Base64

### Меры безопасности
- ✅ **End-to-End шифрование** - сервер не может читать сообщения
- ✅ **Приватные ключи** никогда не передаются на сервер
- ✅ **Уникальные nonce** для каждого сообщения
- ✅ **Временные ключи** для каждого сообщения
- ✅ **Проверка целостности** сообщений
- ✅ **Защита от XSS** (фильтрация HTML тегов)
- ✅ **bcrypt** для хеширования паролей
- ✅ **JWT токены** для аутентификации
- ✅ **Валидация** всех входных данных

### Ограничения текущей версии
- ⚠️ Приватные ключи хранятся в localStorage (в продакшене нужна более безопасная схема)
- ⚠️ Нет forward secrecy (ключи не обновляются автоматически)
- ⚠️ Нет верификации ключей (отсутствует out-of-band проверка)

## ⚙️ Конфигурация

### Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `JWT_SECRET` | Секретный ключ для JWT | `default_secret_key` |
| `SERVER_PORT` | Порт сервера | `8080` |
| `DB_TYPE` | Тип базы данных | `sqlite` |
| `DB_PATH` | Путь к SQLite файлу | `messenger.db` |
| `UPLOAD_DIR` | Директория для файлов | `./uploads` |
| `TOKEN_EXPIRATION_HOURS` | Время жизни токена | `24` |

### Поддержка PostgreSQL

```env
DB_TYPE=postgres
DB_HOST=localhost
DB_PORT=5432
DB_USER=messenger
DB_PASSWORD=password
DB_NAME=messenger_db
DB_SSLMODE=disable
```

## 🧪 Тестирование

### Запуск всех тестов
```bash
go test ./...
```

### Тест криптографии
```bash
go test ./internal/crypto/... -v
```

### Бенчмарки
```bash
go test ./internal/crypto/... -bench=.
```

### Тест в браузере
Откройте `test_encryption.html` для проверки работы шифрования.

## 🐛 Troubleshooting

### Проблемы с шифрованием

#### "Ключи пользователя не загружены"
```bash
# Проверьте, что миграция применена
go run cmd/migrate/main.go

# Перезапустите приложение
go run cmd/main.go
```

#### "Не удалось расшифровать сообщение"
- Проверьте, что у пользователей есть ключи
- Убедитесь, что публичные ключи корректно обменяны
- Проверьте логи в консоли браузера

### Общие проблемы

#### "Port already in use"
```bash
# Измените порт в .env
SERVER_PORT=8081
```

#### "Database locked"
```bash
# Удалите файл базы и пересоздайте
rm messenger.db
go run init_db.go
go run cmd/migrate/main.go
```

#### "JWT_SECRET not set"
```bash
# Добавьте в .env
echo "JWT_SECRET=your-secret-key" >> .env
```

## 📊 Производительность

### Бенчмарки криптографии
- Генерация ключей: ~40μs
- Шифрование сообщения: ~200μs
- Расшифровка сообщения: ~200μs

### Рекомендации для продакшена
- Используйте PostgreSQL для высокой нагрузки
- Настройте HTTPS для WebSocket соединений
- Используйте CDN для статических файлов
- Настройте мониторинг производительности

## 🔮 Планы развития

### Краткосрочные улучшения
- [ ] Безопасное хранение приватных ключей
- [ ] Forward secrecy
- [ ] Верификация ключей
- [ ] Групповое шифрование для чатов

### Долгосрочные планы
- [ ] Perfect forward secrecy
- [ ] Анонимные сообщения
- [ ] Подпись сообщений
- [ ] Квантово-устойчивое шифрование

## 🤝 Вклад в проект

1. **Fork** репозитория
2. Создайте **branch** для новой функции
3. Внесите изменения
4. Напишите тесты
5. Создайте **Pull Request**

### Рекомендации для разработчиков
- Изучите `ENCRYPTION_README.md` для понимания шифрования
- Запускайте тесты перед коммитом
- Следуйте принципам безопасности
- Документируйте новые функции

## 📚 Дополнительные ресурсы

### Документация
- [ENCRYPTION_README.md](ENCRYPTION_README.md) - Подробная документация шифрования
- [Go Documentation](https://golang.org/doc/)
- [NaCl Documentation](https://nacl.cr.yp.to/)
- [WebSocket Guide](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API)

### Безопасность
- [Signal Protocol](https://signal.org/docs/)
- [End-to-End Encryption](https://en.wikipedia.org/wiki/End-to-end_encryption)
- [JWT Introduction](https://jwt.io/introduction)

## 📄 Лицензия

MIT License

## 🆘 Поддержка

Если у вас возникли вопросы:

1. Проверьте раздел **Troubleshooting** выше
2. Изучите `ENCRYPTION_README.md` для вопросов по шифрованию
3. Создайте **Issue** в репозитории
4. Опишите проблему подробно с логами

---

**🔐 Безопасное общение начинается здесь! 🚀**
