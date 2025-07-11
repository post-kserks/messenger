# 🔐 Техническая документация по шифрованию

> **Примечание**: Основная информация о шифровании находится в [README.md](README.md). Этот файл содержит технические детали для разработчиков.

## 🐛 Отладка и диагностика

### Логирование отладки
Включите отладочные логи в консоли браузера:
```javascript
localStorage.setItem('debug_encryption', 'true');
```

### Проверка ключей в браузере
```javascript
// Проверить наличие ключей пользователя
console.log('Ключи пользователя:', currentUserKeys);

// Проверить ключи участников чата
console.log('Ключи участников:', chatParticipantsKeys[currentChat.id]);

// Проверить статус шифрования чата
const status = await getChatEncryptionStatus(chatId);
console.log(`Шифрование: ${status.encryption_enabled ? 'Включено' : 'Отключено'}`);
console.log(`Участников с ключами: ${status.members_with_keys}/${status.total_members}`);
```

### Диагностика ошибок шифрования

#### "Ключи пользователя не загружены"
```javascript
// Проверить инициализацию ключей
if (!currentUserKeys || !currentUserKeys.publicKey) {
    console.error('Ключи пользователя не инициализированы');
    await initializeUserKeys();
}
```

#### "Нет ключей участников чата"
```javascript
// Проверить загрузку ключей участников
if (!chatParticipantsKeys[chatId]) {
    console.error('Ключи участников не загружены');
    await loadChatParticipantsKeys(chatId);
}
```

#### "Не удалось расшифровать сообщение"
```javascript
// Проверить целостность данных
try {
    const decrypted = decryptMessage(encryptedData, nonce, senderPublicKey, currentUserKeys.privateKey);
    console.log('Сообщение успешно расшифровано:', decrypted);
} catch (error) {
    console.error('Ошибка расшифровки:', error);
    // Проверить ключи и данные
}
```

## 📊 Мониторинг и метрики

### SQL запросы для мониторинга

#### Проверка статуса миграции
```sql
-- Количество пользователей с ключами
SELECT COUNT(*) as users_with_keys FROM user_keys;

-- Общее количество пользователей
SELECT COUNT(*) as total_users FROM users;

-- Процент пользователей с ключами
SELECT
    (COUNT(uk.user_id) * 100.0 / COUNT(u.id)) as percentage_with_keys
FROM users u
LEFT JOIN user_keys uk ON u.id = uk.user_id;
```

#### Статистика зашифрованных сообщений
```sql
-- Количество зашифрованных сообщений
SELECT COUNT(*) as encrypted_messages
FROM messages
WHERE is_encrypted = 1;

-- Процент зашифрованных сообщений
SELECT
    (COUNT(CASE WHEN is_encrypted = 1 THEN 1 END) * 100.0 / COUNT(*)) as encryption_percentage
FROM messages;
```

#### Анализ чатов по шифрованию
```sql
-- Чаты с полным шифрованием
SELECT
    c.id,
    c.name,
    COUNT(m.id) as total_messages,
    COUNT(CASE WHEN m.is_encrypted = 1 THEN 1 END) as encrypted_messages
FROM chats c
LEFT JOIN messages m ON c.id = m.chat_id
GROUP BY c.id, c.name
HAVING encrypted_messages = total_messages AND total_messages > 0;
```

### Метрики для отслеживания

#### Ключевые показатели
- **Процент пользователей с ключами** (должен быть >95%)
- **Процент зашифрованных сообщений** (должен быть >90%)
- **Время шифрования/расшифровки** (<1ms на сообщение)
- **Количество ошибок криптографии** (<1% от всех операций)

#### Алерты для настройки
```javascript
// Пример алерта для мониторинга
if (encryptionErrorRate > 0.01) {
    alert('Высокий процент ошибок шифрования: ' + encryptionErrorRate);
}

if (usersWithoutKeys > 0.05) {
    alert('Много пользователей без ключей: ' + usersWithoutKeys);
}
```

## 🔧 Ручная миграция

### Проверка и исправление данных

#### Генерация ключей для пользователей без ключей
```sql
-- Найти пользователей без ключей
SELECT u.id, u.username, u.email
FROM users u
LEFT JOIN user_keys uk ON u.id = uk.user_id
WHERE uk.user_id IS NULL;
```

#### Исправление поврежденных ключей
```sql
-- Найти ключи с некорректным форматом
SELECT user_id, public_key
FROM user_keys
WHERE LENGTH(public_key) != 44; -- Base64 длина для 32 байт
```

### API для миграции

#### Генерация ключей для всех пользователей
```bash
curl -X POST http://localhost:8080/api/migrate-users \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

#### Проверка статуса миграции
```bash
curl -X GET http://localhost:8080/api/migration-status \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

## 🧪 Расширенное тестирование

### Стресс-тест шифрования
```javascript
// Тест производительности шифрования
async function stressTestEncryption() {
    const iterations = 1000;
    const startTime = performance.now();

    for (let i = 0; i < iterations; i++) {
        const message = `Тестовое сообщение ${i}`;
        const encrypted = await encryptMessage(message, recipientPublicKey);
        const decrypted = await decryptMessage(encrypted.data, encrypted.nonce, senderPublicKey, privateKey);

        if (message !== decrypted) {
            throw new Error(`Ошибка в итерации ${i}`);
        }
    }

    const endTime = performance.now();
    console.log(`Время на ${iterations} операций: ${endTime - startTime}ms`);
    console.log(`Среднее время на операцию: ${(endTime - startTime) / iterations}ms`);
}
```

### Тест совместимости ключей
```javascript
// Проверить совместимость ключей между пользователями
async function testKeyCompatibility() {
    const users = await getUsers();

    for (let i = 0; i < users.length; i++) {
        for (let j = i + 1; j < users.length; j++) {
            const user1 = users[i];
            const user2 = users[j];

            try {
                const message = `Тест от ${user1.username} к ${user2.username}`;
                const encrypted = await encryptMessage(message, user2.publicKey);
                const decrypted = await decryptMessage(encrypted.data, encrypted.nonce, user1.publicKey, user2.privateKey);

                if (message !== decrypted) {
                    console.error(`Несовместимость ключей: ${user1.username} -> ${user2.username}`);
                }
            } catch (error) {
                console.error(`Ошибка тестирования: ${user1.username} -> ${user2.username}:`, error);
            }
        }
    }
}
```

## 📚 Дополнительные ресурсы

- [NaCl Documentation](https://nacl.cr.yp.to/)
- [TweetNaCl.js](https://github.com/dchest/tweetnacl-js)
- [Signal Protocol](https://signal.org/docs/)
- [End-to-End Encryption](https://en.wikipedia.org/wiki/End-to-end_encryption)
- [X25519 Key Exchange](https://tools.ietf.org/html/rfc7748)
- [ChaCha20-Poly1305](https://tools.ietf.org/html/rfc8439)

## 🤝 Поддержка разработчиков

При возникновении проблем:

1. **Проверьте логи** сервера и браузера
2. **Убедитесь**, что миграция применена
3. **Проверьте наличие ключей** у пользователей
4. **Используйте отладочные команды** выше
5. **Создайте Issue** с подробным описанием проблемы

---

**Для основной информации о шифровании см. [README.md](README.md)**
