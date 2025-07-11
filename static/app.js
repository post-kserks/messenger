// Глобальные переменные
let currentUser = null;
let currentChat = null;
let ws = null;
let currentMessages = [];
let chatsList = [];

// Криптографические переменные
let currentUserKeys = null;
let chatParticipantsKeys = {};

// Криптографические функции
// Генерация пары ключей
async function generateKeyPair() {
    const keyPair = nacl.box.keyPair();
    return {
        publicKey: nacl.util.encodeBase64(keyPair.publicKey),
        privateKey: nacl.util.encodeBase64(keyPair.secretKey)
    };
}

// Шифрование сообщения для конкретного чата
async function encryptMessageForChat(message, chatId) {
    const participants = chatParticipantsKeys[chatId];
    if (!participants || participants.length === 0) {
        throw new Error('Нет ключей участников чата');
    }

    const encryptedMessages = [];

    for (const participant of participants) {
        if (participant.user_id === currentUser.id) continue; // Пропускаем себя

        const publicKey = nacl.util.decodeBase64(participant.public_key);
        const ephemeralKeyPair = nacl.box.keyPair();
        const nonce = nacl.randomBytes(24);

        const encrypted = nacl.box(
            nacl.util.decodeUTF8(message),
            nonce,
            publicKey,
            ephemeralKeyPair.secretKey
        );

        encryptedMessages.push({
            user_id: participant.user_id,
            encrypted_data: nacl.util.encodeBase64(encrypted),
            nonce: nacl.util.encodeBase64(nonce),
            ephemeral_public_key: nacl.util.encodeBase64(ephemeralKeyPair.publicKey)
        });
    }

    return encryptedMessages;
}

// Расшифровка сообщения
async function decryptMessage(encryptedData, nonce, senderPublicKey) {
    if (!currentUserKeys) {
        throw new Error('Ключи пользователя не загружены');
    }

    const privateKey = nacl.util.decodeBase64(currentUserKeys.privateKey);
    const publicKey = nacl.util.decodeBase64(senderPublicKey);
    const nonceBytes = nacl.util.decodeBase64(nonce);
    const encryptedBytes = nacl.util.decodeBase64(encryptedData);

    const decrypted = nacl.box.open(encryptedBytes, nonceBytes, publicKey, privateKey);
    if (!decrypted) {
        throw new Error('Не удалось расшифровать сообщение');
    }

    return nacl.util.encodeUTF8(decrypted);
}

// Функция получения публичного ключа пользователя
async function getUserPublicKey(userId) {
    try {
        const response = await fetch(`/api/users/${userId}/public-key`, {
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        });

        if (response.ok) {
            return await response.json();
        }
    } catch (error) {
        console.error('Ошибка получения ключа:', error);
    }
    return null;
}

// Функция загрузки ключей участников чата
async function loadChatParticipantsKeys(chatId) {
    try {
        const response = await fetch(`/api/chats/${chatId}/participants-keys`, {
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        });

        if (response.ok) {
            const keys = await response.json();
            chatParticipantsKeys[chatId] = keys;
        }
    } catch (error) {
        console.error('Ошибка загрузки ключей участников:', error);
    }
}

// Функция получения статуса шифрования чата
async function getChatEncryptionStatus(chatId) {
    try {
        const response = await fetch(`/api/chats/${chatId}/encryption-status`, {
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        });

        if (response.ok) {
            return await response.json();
        }
    } catch (error) {
        console.error('Ошибка получения статуса шифрования:', error);
    }
    return null;
}

// Проверка авторизации при загрузке страницы
document.addEventListener('DOMContentLoaded', function() {
    const token = localStorage.getItem('token');
    if (token) {
        // Если есть токен, проверяем его валидность
        checkAuth();
    } else {
        // Если нет токена и мы на главной странице, перенаправляем на логин
        if (window.location.pathname === '/' || window.location.pathname === '/static/index.html') {
            window.location.href = '/static/login.html';
        }
    }

    // Инициализация обработчиков событий
    initEventHandlers();

    // Защита: скрыть мобильное меню на десктопе и наоборот
    function fixMenuDisplay() {
        const sidebar = document.getElementById('sidebar');
        const mobileMenu = document.getElementById('mobileMenu');
        if (window.innerWidth > 768) {
            if (sidebar) sidebar.style.display = 'flex';
            if (mobileMenu) mobileMenu.style.display = 'none';
        } else {
            if (sidebar) sidebar.style.display = 'none';
            if (mobileMenu) mobileMenu.style.display = 'block';
        }
    }
    fixMenuDisplay();
    window.addEventListener('resize', fixMenuDisplay);
});

// Инициализация обработчиков событий
function initEventHandlers() {
    // Простая защита от перехвата событий input file
    document.addEventListener('click', function(e) {
        const fileInput = e.target.closest('input[type="file"]');
        const fileBtn = e.target.closest('.file-btn');

        // Если клик не по кнопке файла, но по input file - блокируем
        if (fileInput && !fileBtn) {
            e.stopPropagation();
            e.preventDefault();
            return false;
        }
    }, true);
    // Обработчик формы логина
    const loginForm = document.getElementById('loginForm');
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    // Обработчик формы регистрации
    const registerForm = document.getElementById('registerForm');
    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }

    // Обработчики для главной страницы
        const msgForm = document.getElementById('msgForm');
    if (msgForm) {
        msgForm.addEventListener('submit', (e) => {
            console.log('Отправка формы');
            sendMessage(e);
        });

        // Добавляем обработчик для мобильных устройств
        msgForm.addEventListener('touchstart', (e) => {
            console.log('Touch по форме');
            // Предотвращаем двойное срабатывание
            e.stopPropagation();
        });
    }

    const fileInput = document.getElementById('fileInput');
    if (fileInput) {
        fileInput.addEventListener('change', handleFileSelect);
    }

    const createChatBtn = document.getElementById('createChatBtn');
    if (createChatBtn) {
        createChatBtn.addEventListener('click', showNewChatModal);
    }

    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', showLogoutConfirm);
    }

    const profileBtn = document.getElementById('profileBtn');
    if (profileBtn) {
        profileBtn.addEventListener('click', showProfileModal);
    }

    const contactsBtn = document.getElementById('contactsBtn');
    if (contactsBtn) {
        contactsBtn.addEventListener('click', showContactsModal);
    }

    // Обработчики мобильного меню
    initMobileMenuHandlers();

    // Обработчики модальных окон
    initModalHandlers();


}

// Обработка логина
async function handleLogin(e) {
    e.preventDefault();

    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('loginError');

    try {
        const response = await fetch('/api/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ email, password })
        });

        const data = await response.json();

        if (response.ok && data.success) {
            localStorage.setItem('token', data.token);
            // Создаем объект пользователя из полученных данных
            const user = {
                id: data.user_id,
                username: data.username,
                nickname: data.nickname,
                email: document.getElementById('email').value
            };
            localStorage.setItem('user', JSON.stringify(user));
            currentUser = user;
            window.location.href = '/';
        } else {
            errorDiv.textContent = data.error || 'Ошибка входа';
        }
    } catch (error) {
        console.error('Ошибка при входе:', error);
        errorDiv.textContent = 'Ошибка соединения';
    }
}

// Обработка регистрации
async function handleRegister(e) {
    e.preventDefault();

    const username = document.getElementById('username').value;
    const nickname = document.getElementById('nickname').value;
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;
    const errorDiv = document.getElementById('registerError');

    try {
        const response = await fetch('/api/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, nickname, email, password })
        });

        const data = await response.json();

        if (response.ok && data.success) {
            localStorage.setItem('token', data.token);
            // Создаем объект пользователя из полученных данных
            const user = {
                id: data.user_id,
                username: data.username,
                nickname: data.nickname,
                email: document.getElementById('email').value
            };
            localStorage.setItem('user', JSON.stringify(user));
            currentUser = user;
            window.location.href = '/';
        } else {
            errorDiv.textContent = data.error || 'Ошибка регистрации';
        }
    } catch (error) {
        console.error('Ошибка при регистрации:', error);
        errorDiv.textContent = 'Ошибка соединения';
    }
}

// Проверка авторизации
async function checkAuth() {
    const token = localStorage.getItem('token');
    if (!token) {
        window.location.href = '/static/login.html';
        return;
    }

    try {
        const response = await fetch('/api/user', {
            headers: {
                'Authorization': 'Bearer ' + token
            }
        });

        if (!response.ok) {
            localStorage.removeItem('token');
            localStorage.removeItem('user');
            window.location.href = '/static/login.html';
            return;
        }

        const user = await response.json();
        currentUser = user;
        localStorage.setItem('user', JSON.stringify(user));

        // Инициализируем ключи пользователя
        try {
            const userKey = await getUserPublicKey(user.id);
            if (userKey) {
                // В реальном приложении приватный ключ должен храниться безопасно
                // Для демонстрации генерируем новую пару ключей
                const keyPair = await generateKeyPair();
                currentUserKeys = keyPair;

                // Обновляем публичный ключ на сервере
                await fetch(`/api/users/${user.id}/public-key`, {
                    method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + token
                    },
                    body: JSON.stringify({
                        public_key: keyPair.publicKey
                    })
                });
            }
        } catch (error) {
            console.error('Ошибка инициализации ключей:', error);
        }

        // Если мы на странице логина, перенаправляем на главную
        if (window.location.pathname === '/static/login.html' || window.location.pathname === '/static/register.html') {
            window.location.href = '/';
        }

        // Загружаем данные для главной страницы
        if (window.location.pathname === '/' || window.location.pathname === '/static/index.html') {
            loadChats();
            connectWebSocket();

            // Показываем форму на мобильных устройствах
            const msgForm = document.getElementById('msgForm');
            const inputContainer = document.querySelector('.input-container');
            if (msgForm && inputContainer && window.innerWidth <= 768) {
                msgForm.style.display = 'flex';
                inputContainer.style.display = 'block';
            }
        }
    } catch (error) {
        console.error('Ошибка проверки авторизации:', error);
        localStorage.removeItem('token');
        localStorage.removeItem('user');
        window.location.href = '/static/login.html';
    }
}

// Загрузка чатов
async function loadChats() {
    if (!currentUser) return;
    try {
        const response = await fetch(`/api/chats?user_id=${currentUser.id}`, {
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        });
        if (response.ok) {
            const chats = await response.json();
            // Проверяем, что chats - это массив
            if (chats && Array.isArray(chats)) {
                // Сортируем чаты по last_msg_time (сначала самые новые)
                chats.sort((a, b) => {
                    const tA = a.last_msg_time ? new Date(a.last_msg_time).getTime() : 0;
                    const tB = b.last_msg_time ? new Date(b.last_msg_time).getTime() : 0;
                    return tB - tA;
                });
                chatsList = chats;
                displayChats(chatsList);
            } else {
                console.log('Получены неверные данные чатов:', chats);
                chatsList = [];
                displayChats([]);
            }
        }
    } catch (error) {
        console.error('Ошибка загрузки чатов:', error);
    }
}

// Выбор чата
async function selectChat(chat) {
    currentChat = chat;
    // Обновляем заголовок
    const chatTitle = document.getElementById('chatTitle');
    if (chatTitle) {
        chatTitle.textContent = chat.name;
    }
    // Показываем форму отправки сообщений
    const msgForm = document.getElementById('msgForm');
    const inputContainer = document.querySelector('.input-container');
    if (msgForm) {
        msgForm.style.display = 'flex';
        console.log('Форма показана:', msgForm.style.display);
    }
    if (inputContainer) {
        inputContainer.style.display = 'block';
        console.log('Контейнер показан:', inputContainer.style.display);
    }

    // Загружаем ключи участников чата для шифрования
    await loadChatParticipantsKeys(chat.id);

    // Загружаем сообщения
    await loadMessages(chat.id);

    // Сбрасываем счетчик непрочитанных сообщений
    try {
        await fetch('/api/mark_read', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            },
            body: JSON.stringify({ user_id: currentUser.id, chat_id: chat.id })
        });
        // Обновляем счетчик непрочитанных локально
        const chatItem = chatsList.find(c => c.id === currentChat.id);
        if (chatItem) {
            chatItem.unread_count = 0;
            displayChats(chatsList);
        }
    } catch (error) {
        console.error('Ошибка сброса непрочитанных:', error);
    }
    // Обновляем активный чат в списке (для десктопа и мобильного)
    const chatItems = document.querySelectorAll('#chatList li, #mobileChatList li');
    chatItems.forEach(item => item.classList.remove('active'));

    // Находим и активируем выбранный чат
    const selectedChatItem = Array.from(chatItems).find(item => {
        const span = item.querySelector('span');
        return span && span.textContent === chat.name;
    });
    if (selectedChatItem) {
        selectedChatItem.classList.add('active');
    }
}

// Загрузка сообщений
async function loadMessages(chatId) {
    try {
        const response = await fetch(`/api/messages?chat_id=${chatId}`, {
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        });

        if (response.ok) {
            const messages = await response.json();

            // Обрабатываем зашифрованные сообщения
            for (const message of messages) {
                if (message.is_encrypted && message.encrypted_data) {
                    try {
                        // Получаем публичный ключ отправителя
                        const senderKey = await getUserPublicKey(message.sender_id);
                        if (senderKey) {
                            message.text = await decryptMessage(
                                message.encrypted_data,
                                message.nonce,
                                senderKey.public_key
                            );
                        } else {
                            message.text = '[Зашифрованное сообщение]';
                        }
                    } catch (error) {
                        console.error('Ошибка расшифровки:', error);
                        message.text = '[Зашифрованное сообщение]';
                    }
                }
            }

            currentMessages = messages;
            renderMessages();
        }
    } catch (error) {
        console.error('Ошибка загрузки сообщений:', error);
    }
}

// --- НАБОР ЭМОДЗИ ДЛЯ РЕАКЦИЙ ---
const REACTION_EMOJIS = ['👍', '❤️', '😂', '😮', '👎'];

// --- КОНТЕКСТНОЕ МЕНЮ ДЛЯ РЕАКЦИЙ ---
let reactionMenu = null;
function showReactionMenu(x, y, messageId) {
    if (reactionMenu) reactionMenu.remove();
    reactionMenu = document.createElement('div');
    reactionMenu.className = 'reaction-context-menu';
    REACTION_EMOJIS.forEach(emoji => {
        const btn = document.createElement('button');
        btn.className = 'reaction-menu-btn';
        btn.type = 'button';
        btn.textContent = emoji;
        btn.onclick = (e) => {
            e.stopPropagation();
            const msg = currentMessages.find(m => m.id === messageId);
            if (msg) {
                if (!msg.reactions) msg.reactions = [];
                // Удаляем старую реакцию пользователя, если была
                msg.reactions = msg.reactions.filter(r => r.user_id !== currentUser.id);
                msg.reactions.push({ user_id: currentUser.id, emoji });
                renderMessages();
                // Отправляем реакцию на сервер (всегда, даже если локально обновили)
                sendReaction(messageId, emoji);
            } else {
                sendReaction(messageId, emoji);
            }
            hideReactionMenu();
        };
        reactionMenu.appendChild(btn);
    });
    document.body.appendChild(reactionMenu);
    const menuRect = reactionMenu.getBoundingClientRect();
    let left = x;
    let top = y;
    if (left + menuRect.width > window.innerWidth) {
        left = x - menuRect.width;
        if (left < 0) left = 0;
    }
    if (top + menuRect.height > window.innerHeight) {
        top = window.innerHeight - menuRect.height;
        if (top < 0) top = 0;
    }
    reactionMenu.style.left = left + 'px';
    reactionMenu.style.top = top + 'px';
    setTimeout(() => {
        document.addEventListener('click', hideReactionMenu, { once: true });
    }, 0);
}
function hideReactionMenu() {
    if (reactionMenu) {
        reactionMenu.remove();
        reactionMenu = null;
    }
}

// --- ОТОБРАЖЕНИЕ СООБЩЕНИЙ С РЕАКЦИЯМИ ---
function renderMessages() {
    const messagesDiv = document.getElementById('messages');
    if (!messagesDiv) return;
    messagesDiv.innerHTML = '';
    if (!currentMessages || !Array.isArray(currentMessages)) {
        currentMessages = [];
    }
    currentMessages.sort((a, b) => new Date(a.sent_at) - new Date(b.sent_at));
    currentMessages.forEach(msg => {
        const msgDiv = document.createElement('div');
        let className = `msg ${Number(msg.sender_id) === Number(currentUser.id) ? 'me' : ''}`;
        if (msg.isTemp) className += ' temp-message';
        if (msg.failed) className += ' failed-message';
        msgDiv.className = className;
        const bubble = document.createElement('div');
        bubble.className = 'bubble';
        if (msg.type === 'file' || (msg.text && msg.text.startsWith('[file]'))) {
            const fileUrl = msg.file_url || (msg.text ? msg.text.replace('[file]', '') : '');
            bubble.innerHTML = `<a href="${fileUrl}" target="_blank">📎 ${msg.file_name || 'Файл'}</a>`;
        } else {
            bubble.textContent = msg.text;
        }
        if (msg.isTemp && !msg.failed) {
            const status = document.createElement('span');
            status.className = 'status-indicator';
            status.innerHTML = '⏳';
            status.title = 'Отправляется...';
            bubble.appendChild(status);
        }
        if (msg.failed) {
            const error = document.createElement('span');
            error.className = 'error-indicator';
            error.innerHTML = '❌';
            error.title = 'Ошибка отправки';
            bubble.appendChild(error);
        }
        // --- РЕАКЦИИ (только отображение) ---
        const reactionsDiv = document.createElement('div');
        reactionsDiv.className = 'reactions-bar';
        const reactionsMap = {};
        if (msg.reactions && Array.isArray(msg.reactions)) {
            msg.reactions.forEach(r => {
                if (!reactionsMap[r.emoji]) reactionsMap[r.emoji] = [];
                reactionsMap[r.emoji].push(r.user_id);
            });
        }
        Object.keys(reactionsMap).forEach(emoji => {
            const btn = document.createElement('span');
            btn.className = 'reaction-btn';
            btn.textContent = emoji + (reactionsMap[emoji].length > 1 ? ` ${reactionsMap[emoji].length}` : '');
            if (reactionsMap[emoji].includes(currentUser.id)) {
                btn.classList.add('reacted');
                btn.style.cursor = 'pointer';
                btn.onclick = (e) => {
                    // Снять свою реакцию
                    const msgObj = currentMessages.find(m => m.id === msg.id);
                    if (msgObj) {
                        msgObj.reactions = msgObj.reactions.filter(r => !(r.user_id === currentUser.id && r.emoji === emoji));
                        renderMessages();
                        // Отправить на сервер снятие реакции
                        sendReaction(msg.id, "");
                    }
                };
            }
            reactionsDiv.appendChild(btn);
        });
        if (Object.keys(reactionsMap).length > 0) bubble.appendChild(reactionsDiv);
        // ---
        const meta = document.createElement('div');
        meta.className = 'meta';
        let timeText = new Date(msg.sent_at).toLocaleString();
        if (msg.isTemp && !msg.failed) timeText += ' (отправляется...)';
        else if (msg.failed) timeText += ' (ошибка)';
        meta.innerHTML = `<span class="username">${msg.username}</span> • ${timeText}`;
        msgDiv.appendChild(bubble);
        msgDiv.appendChild(meta);
        // --- обработчик контекстного меню ---
        bubble.oncontextmenu = (e) => {
            e.preventDefault();
            showReactionMenu(e.clientX, e.clientY, msg.id);
        };
        messagesDiv.appendChild(msgDiv);
    });
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

// --- ОТПРАВКА РЕАКЦИИ ---
function sendReaction(messageId, emoji) {
    if (!ws || ws.readyState !== 1) return;
    ws.send(JSON.stringify({
        type: 'reaction',
        message_id: messageId,
        chat_id: currentChat.id,
        emoji: emoji
    }));
}

// --- ОБРАБОТКА ВХОДЯЩИХ РЕАКЦИЙ ---
async function handleWebSocketMessage(data) {
    if (data.type === 'reaction' && data.message_id) {
        // Найти сообщение и обновить его реакции
        const msg = currentMessages.find(m => m.id === data.message_id);
        if (msg) {
            msg.reactions = data.reactions;
            renderMessages();
        }
        return;
    }

    // Унифицируем структуру для файлов и сообщений
    let msg = {
        id: data.id,
        chat_id: data.chat_id,
        sender_id: data.sender_id,
        username: data.username,
        sent_at: data.sent_at,
        type: data.type === 'new_file' || data.type === 'file' ? 'file' : undefined,
        text: data.text,
        file_url: data.file_url,
        file_name: data.file_name,
        reactions: data.reactions, // <-- добавляем реакции, если есть
        is_encrypted: data.is_encrypted || false,
        encrypted_data: data.encrypted_data,
        nonce: data.nonce
    };

    // Обрабатываем зашифрованные сообщения
    if (msg.is_encrypted && msg.encrypted_data && msg.nonce) {
        try {
            // Получаем публичный ключ отправителя
            const senderKey = await getUserPublicKey(msg.sender_id);
            if (senderKey) {
                msg.text = await decryptMessage(
                    msg.encrypted_data,
                    msg.nonce,
                    senderKey.public_key
                );
            } else {
                msg.text = '[Зашифрованное сообщение]';
            }
        } catch (error) {
            console.error('Ошибка расшифровки:', error);
            msg.text = '[Зашифрованное сообщение]';
        }
    }

    if (msg.type === 'file' && !msg.file_url && msg.text && msg.text.startsWith('[file]')) {
        msg.file_url = msg.text.replace('[file]', '');
    }
    if (msg.type === 'file' && !msg.file_name) {
        msg.file_name = 'Файл';
    }
    if (!msg.username && currentUser && Number(msg.sender_id) === Number(currentUser.id)) {
        msg.username = currentUser.username;
    }

    // Если сообщение для текущего открытого чата
    if (currentChat && data.chat_id === currentChat.id) {
        // Проверяем, есть ли временное сообщение с таким же текстом от того же пользователя
        const existingIndex = currentMessages.findIndex(m =>
            m.isTemp &&
            m.sender_id === msg.sender_id &&
            m.text === msg.text &&
            new Date(m.sent_at).getTime() > Date.now() - 10000 // В последние 10 секунд
        );

        if (existingIndex !== -1) {
            // Заменяем временное сообщение на реальное
            currentMessages[existingIndex] = msg;
        } else {
            // Добавляем новое сообщение
            currentMessages.push(msg);
        }
        renderMessages();

        // Обновляем чат локально
        updateChatLocally(data.chat_id, msg);
    } else {
        // Если чат не открыт, увеличиваем счетчик непрочитанных
        const chat = chatsList.find(c => c.id === data.chat_id);
        if (chat) {
            chat.last_msg_time = msg.sent_at;
            chat.last_msg_text = msg.text || msg.file_name || '';
            chat.unread_count = (chat.unread_count || 0) + 1;

            // Сортируем и обновляем отображение
            chatsList.sort((a, b) => {
                const tA = a.last_msg_time ? new Date(a.last_msg_time).getTime() : 0;
                const tB = b.last_msg_time ? new Date(b.last_msg_time).getTime() : 0;
                return tB - tA;
            });
            displayChats(chatsList);
        }
    }
}

// Отправка сообщения
async function sendMessage(e) {
    e.preventDefault();

    if (!currentChat) return;

    const input = document.getElementById('msgInput');
    const fileInput = document.getElementById('fileInput');
    const text = input.value.trim();
    const file = fileInput.files[0];

    // Если нет ни текста, ни файла - не отправляем
    if (!text && !file) return;

    let tempMessage;

    if (file) {
        // Отправляем файл
        tempMessage = {
            id: 'temp_' + Date.now(),
            chat_id: currentChat.id,
            sender_id: currentUser.id,
            username: currentUser.username,
            text: `[file]${file.name}`,
            sent_at: new Date().toISOString(),
            type: 'file',
            file_name: file.name,
            isTemp: true
        };

        // Очищаем поле ввода и добавляем сообщение локально
        input.value = '';
        fileInput.value = '';
        currentMessages.push(tempMessage);
        renderMessages();

        // Обновляем чат в списке локально
        updateChatLocally(currentChat.id, tempMessage);

        // Отправляем файл на сервер в фоне
        sendFileToServer(currentChat.id, file, tempMessage.id);
    } else {
        // Отправляем только текст
        tempMessage = {
            id: 'temp_' + Date.now(),
            chat_id: currentChat.id,
            sender_id: currentUser.id,
            username: currentUser.username,
            text: text,
            sent_at: new Date().toISOString(),
            type: 'text',
            isTemp: true
        };

        // Очищаем поле ввода и добавляем сообщение локально
        input.value = '';
        currentMessages.push(tempMessage);
        renderMessages();

        // Обновляем чат в списке локально
        updateChatLocally(currentChat.id, tempMessage);

        // Отправляем сообщение на сервер в фоне
        sendMessageToServer(currentChat.id, text, tempMessage.id);
    }
}

// Обработка выбора файла
function handleFileSelect(e) {
    const file = e.target.files[0];
    const input = document.getElementById('msgInput');
    if (file) {
        input.value = `📎 ${file.name}`;
        input.placeholder = 'Добавьте сообщение к файлу (необязательно)';
        console.log('Файл выбран:', file.name);
    } else {
        input.placeholder = 'Введите сообщение';
    }
}

// Функция отправки сообщения на сервер
async function sendMessageToServer(chatId, text, tempId) {
    try {
        let messageData = {
            chat_id: chatId,
            text: text,
            is_encrypted: false
        };

        // Проверяем, есть ли ключи участников чата
        if (chatParticipantsKeys[chatId] && chatParticipantsKeys[chatId].length > 0) {
            try {
                // Шифруем сообщение
                const encryptedMessages = await encryptMessageForChat(text, chatId);

                if (encryptedMessages.length > 0) {
                    // Отправляем зашифрованное сообщение (берем первый для простоты)
                    messageData = {
                        chat_id: chatId,
                        encrypted_data: encryptedMessages[0].encrypted_data,
                        nonce: encryptedMessages[0].nonce,
                        is_encrypted: true
                    };
                }
            } catch (error) {
                console.error('Ошибка шифрования:', error);
                // Продолжаем с обычным сообщением
            }
        }

        const response = await fetch('/api/messages', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            },
            body: JSON.stringify(messageData)
        });

        if (!response.ok) {
            // Если отправка не удалась, помечаем сообщение как неудачное
            const tempMessage = currentMessages.find(m => m.id === tempId);
            if (tempMessage) {
                tempMessage.failed = true;
                renderMessages();
            }
        }
    } catch (error) {
        console.error('Ошибка отправки сообщения:', error);
        // Помечаем сообщение как неудачное
        const tempMessage = currentMessages.find(m => m.id === tempId);
        if (tempMessage) {
            tempMessage.failed = true;
            renderMessages();
        }
    }
}

// Функция локального обновления чата
function updateChatLocally(chatId, message) {
    const chat = chatsList.find(c => c.id === chatId);
    if (chat) {
        chat.last_msg_time = message.sent_at;
        chat.last_msg_text = message.text || message.file_name || '';
        chat.unread_count = 0; // Сбросить счетчик, так как чат открыт

        // Сортируем чаты по времени последнего сообщения
        chatsList.sort((a, b) => {
            const tA = a.last_msg_time ? new Date(a.last_msg_time).getTime() : 0;
            const tB = b.last_msg_time ? new Date(b.last_msg_time).getTime() : 0;
            return tB - tA;
        });

        displayChats(chatsList);
    }
}

// Функция отправки файла на сервер
async function sendFileToServer(chatId, file, tempId) {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('chat_id', chatId);

    try {
        const response = await fetch('/api/upload', {
            method: 'POST',
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            },
            body: formData
        });

        if (!response.ok) {
            // Если отправка не удалась, помечаем сообщение как неудачное
            const tempMessage = currentMessages.find(m => m.id === tempId);
            if (tempMessage) {
                tempMessage.failed = true;
                renderMessages();
            }
        }
    } catch (error) {
        console.error('Ошибка отправки файла:', error);
        // Помечаем сообщение как неудачное
        const tempMessage = currentMessages.find(m => m.id === tempId);
        if (tempMessage) {
            tempMessage.failed = true;
            renderMessages();
        }
    }
}

// Подключение к WebSocket
function connectWebSocket() {
    const token = localStorage.getItem('token');
    if (!token) return;

    const wsUrl = `ws://${window.location.host}/ws?token=${token}`;
    ws = new WebSocket(wsUrl);

    ws.onopen = function() {
        console.log('WebSocket подключен');
    };

    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        handleWebSocketMessage(data);
    };

    ws.onclose = function() {
        console.log('WebSocket отключен');
        // Переподключение через 5 секунд
        setTimeout(connectWebSocket, 5000);
    };

    ws.onerror = function(error) {
        console.error('WebSocket ошибка:', error);
    };
}

// Модальное окно подтверждения выхода
function showLogoutConfirm() {
    let modal = document.getElementById('logoutConfirmModal');
    if (modal) {
        modal.innerHTML = `
            <div class="modal-content" style="text-align:center; min-width:320px;">
                <span class="close" id="closeLogoutConfirm">×</span>
                <h3>Выход из аккаунта</h3>
                <p>Вы уверены, что хотите выйти?</p>
                <button id="confirmLogoutBtn" style="background:#ff4444;color:#fff;padding:8px 18px;border:none;border-radius:5px;margin:10px 8px 0 0;">Да, выйти</button>
                <button id="cancelLogoutBtn" style="background:#eee;color:#222;padding:8px 18px;border:none;border-radius:5px;margin:10px 0 0 0;">Отмена</button>
            </div>
        `;
    } else {
        modal = document.createElement('div');
        modal.id = 'logoutConfirmModal';
        modal.className = 'modal';
        modal.innerHTML = `
            <div class="modal-content" style="text-align:center; min-width:320px;">
                <span class="close" id="closeLogoutConfirm">×</span>
                <h3>Выход из аккаунта</h3>
                <p>Вы уверены, что хотите выйти?</p>
                <button id="confirmLogoutBtn" style="background:#ff4444;color:#fff;padding:8px 18px;border:none;border-radius:5px;margin:10px 8px 0 0;">Да, выйти</button>
                <button id="cancelLogoutBtn" style="background:#eee;color:#222;padding:8px 18px;border:none;border-radius:5px;margin:10px 0 0 0;">Отмена</button>
            </div>
        `;
        document.body.appendChild(modal);
    }
    modal.style.display = 'flex';
    document.getElementById('closeLogoutConfirm').onclick = () => modal.style.display = 'none';
    document.getElementById('cancelLogoutBtn').onclick = () => modal.style.display = 'none';
    document.getElementById('confirmLogoutBtn').onclick = () => {
        modal.style.display = 'none';
        doLogout();
    };
}

function doLogout() {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    currentUser = null;
    currentChat = null;
    if (ws) {
        ws.close();
        ws = null;
    }
    window.location.href = '/static/login.html';
}

// Показываем модальное окно создания чата
function showNewChatModal() {
    const newChatModal = document.getElementById('newChatModal');
    const newChatContacts = document.getElementById('newChatContacts');
    if (newChatModal && newChatContacts) {
        // Загружаем контакты пользователя для выбора участников чата
        fetch(`/api/contacts?user_id=${currentUser.id}`, {
            headers: {
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        })
            .then(r => r.json())
            .then(contacts => {
                newChatContacts.innerHTML = '<b>Выберите участников:</b><br>';
            if (contacts && Array.isArray(contacts) && contacts.length > 0) {
                contacts.forEach(c => {
                    const id = 'chk_contact_' + c.id;
                    const cb = document.createElement('input');
                    cb.type = 'checkbox';
                    cb.id = id;
                    cb.value = c.id;
                    cb.name = 'chat_users';
                    newChatContacts.appendChild(cb);
                    const lbl = document.createElement('label');
                    lbl.htmlFor = id;
                    lbl.textContent = ` ${c.username} ${c.nickname ? '(@' + c.nickname + ')' : '(без никнейма)'}`;
                    newChatContacts.appendChild(lbl);
                    newChatContacts.appendChild(document.createElement('br'));
                });
            } else {
                newChatContacts.innerHTML += '<p>У вас пока нет контактов. Создайте чат только для себя.</p>';
            }
            newChatModal.style.display = 'flex';
        })
        .catch(error => {
            console.error('Ошибка загрузки контактов:', error);
            newChatContacts.innerHTML = '<b>Выберите участников:</b><br><p>Ошибка загрузки контактов. Создайте чат только для себя.</p>';
                newChatModal.style.display = 'flex';
        });
    }
}

// Улучшенный профиль
async function showProfileModal() {
    const profileModal = document.getElementById('profileModal');
    const profileInfo = document.getElementById('profileInfo');
    const profileUsername = document.getElementById('profileUsername');
    const profileNickname = document.getElementById('profileNickname');
    const profileEmail = document.getElementById('profileEmail');
    const profilePassword = document.getElementById('profilePassword');
    const profileMsg = document.getElementById('profileMsg');
    if (currentUser) {
        // Генерируем "аватар" (инициалы)
        let initials = (currentUser.username||'')[0] || '?';
        if (currentUser.nickname) initials = currentUser.nickname[0].toUpperCase();
        profileInfo.innerHTML = `
            <div class="profile-avatar">${initials}</div>
            <div class="profile-details">
                <div style="font-size:1.2em;font-weight:600;">${currentUser.username}</div>
                <div style="color:#888;">@${currentUser.nickname||'нет_никнейма'}</div>
                <div style="color:#666;font-size:0.98em;">${currentUser.email}</div>
                <div style="color:#bbb;font-size:0.9em;">ID: ${currentUser.id}</div>
            </div>
        `;
        profileUsername.value = currentUser.username;
        profileNickname.value = currentUser.nickname || '';
        profileEmail.value = currentUser.email;
        if (profilePassword) profilePassword.value = '';
        if (profileMsg) profileMsg.textContent = '';
    }
    if (profileModal) profileModal.style.display = 'flex';
}

// Показать модальное окно контактов
async function showContactsModal() {
    const contactsModal = document.getElementById('contactsModal');
    const contactsList = document.getElementById('contactsList');
    if (contactsModal && contactsList) {
        try {
            const response = await fetch('/api/contacts', {
                headers: {
                    'Authorization': 'Bearer ' + localStorage.getItem('token')
                }
            });

            if (response.ok) {
                const contacts = await response.json();
                contactsList.innerHTML = '';

                if (contacts && Array.isArray(contacts) && contacts.length > 0) {
                    contacts.forEach(contact => {
                        const li = document.createElement('li');
                        li.style.cssText = 'display: flex; justify-content: space-between; align-items: center; padding: 10px; border-bottom: 1px solid #eee;';
                        li.innerHTML = `
                            <div>
                                <strong>${contact.username}</strong><br>
                                <small style="color: #666;">${contact.nickname ? '@' + contact.nickname : 'без никнейма'}</small><br>
                                <small style="color: #999;">${contact.email}</small>
                            </div>
                            <button onclick="removeContact(${contact.id})" style="background: #ff4444; color: white; border: none; padding: 5px 10px; border-radius: 3px; cursor: pointer;">Удалить</button>
                        `;
                        contactsList.appendChild(li);
                    });
                } else {
                    contactsList.innerHTML = '<li style="color: #666; font-style: italic;">У вас пока нет контактов</li>';
                }
            }
        } catch (error) {
            console.error('Ошибка загрузки контактов:', error);
            contactsList.innerHTML = '<li style="color: #ff4444;">Ошибка загрузки контактов</li>';
        }

        contactsModal.style.display = 'flex';
    }
}

// Удалить контакт
async function removeContact(contactId) {
    if (!confirm('Удалить контакт?')) return;

    try {
        const response = await fetch('/api/contacts', {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            },
            body: JSON.stringify({
                contact_id: contactId
            })
        });

        if (response.ok) {
            showContactsModal(); // Перезагружаем список контактов
        } else {
            console.error('Ошибка удаления контакта');
            alert('Ошибка удаления контакта');
        }
    } catch (error) {
        console.error('Ошибка удаления контакта:', error);
        alert('Ошибка удаления контакта');
    }
}

// Доработка обработчиков модальных окон
function initModalHandlers() {
    // Модальное окно создания чата
    const newChatModal = document.getElementById('newChatModal');
    const closeModal = document.getElementById('closeModal');
    const newChatForm = document.getElementById('newChatForm');
    if (closeModal && newChatModal) {
        closeModal.onclick = () => newChatModal.style.display = 'none';
    }
    if (newChatForm) {
        newChatForm.onsubmit = async function(e) {
        e.preventDefault();
            const name = document.getElementById('newChatName').value;
            const checked = Array.from(document.querySelectorAll('#newChatContacts input[type=checkbox]:checked'));
        const users = checked.map(cb => parseInt(cb.value));
            if (!users.includes(currentUser.id)) users.push(currentUser.id);
            try {
                const response = await fetch('/api/chats', {
            method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + localStorage.getItem('token')
                    },
            body: JSON.stringify({ name, user_ids: users })
                });
                if (response.ok) {
            newChatModal.style.display = 'none';
                    newChatForm.reset();
                    await loadChats();
                }
            } catch (error) {
                alert('Ошибка создания чата');
            }
        };
    }
    // Модальное окно профиля
    const profileModal = document.getElementById('profileModal');
    const closeProfile = document.getElementById('closeProfile');
    const profileForm = document.getElementById('profileForm');
    const profileUsername = document.getElementById('profileUsername');
    const profileEmail = document.getElementById('profileEmail');
    const profilePassword = document.getElementById('profilePassword');
    const profileMsg = document.getElementById('profileMsg');
    if (closeProfile && profileModal) {
        closeProfile.onclick = () => profileModal.style.display = 'none';
    }
    if (profileForm) {
        profileForm.onsubmit = async function(e) {
        e.preventDefault();
            try {
                const response = await fetch('/api/user', {
            method: 'PUT',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer ' + localStorage.getItem('token')
                    },
            body: JSON.stringify({
                        id: currentUser.id,
                username: profileUsername.value,
                nickname: profileNickname.value,
                email: profileEmail.value,
                password: profilePassword.value
            })
                });
                if (response.ok) {
                    const result = await response.json();
                    if (profileMsg) profileMsg.textContent = result.message || 'Профиль обновлён';
                    setTimeout(() => { if (profileMsg) profileMsg.textContent = ''; }, 2000);
                    currentUser.username = profileUsername.value;
                    currentUser.nickname = profileNickname.value;
                    currentUser.email = profileEmail.value;
                    localStorage.setItem('user', JSON.stringify(currentUser));
            } else {
                    const error = await response.json();
                    if (profileMsg) profileMsg.textContent = error.error || 'Ошибка обновления';
                }
            } catch (error) {
                if (profileMsg) profileMsg.textContent = 'Ошибка соединения';
            }
        };
    }
    // Закрытие модальных окон по клику вне их
    window.onclick = function(event) {
        if (event.target.classList && event.target.classList.contains('modal')) {
            event.target.style.display = 'none';
        }
    };

    // Модальное окно контактов
    const contactsModal = document.getElementById('contactsModal');
    const closeContacts = document.getElementById('closeContacts');
    const searchUserForm = document.getElementById('searchUserForm');

    if (closeContacts && contactsModal) {
        closeContacts.onclick = () => contactsModal.style.display = 'none';
    }

    if (searchUserForm) {
        searchUserForm.onsubmit = async function(e) {
            e.preventDefault();
            const searchInput = document.getElementById('searchUsername');
            const username = searchInput.value.trim();
            if (!username) return;
            await addContactByUsername(username);
        };
    }
}

// Поиск пользователей по никнейму
async function searchUsers() {
    // Эта функция больше не используется
}

// Добавление контакта по никнейму
async function addContactByUsername(username) {
    try {
        const response = await fetch('/api/contacts', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            },
            body: JSON.stringify({
                username: username
            })
        });

        if (response.ok) {
            const result = await response.json();
            alert('Контакт добавлен!');
            // Очищаем поле поиска и скрываем результаты
            document.getElementById('searchUsername').value = '';
            document.getElementById('searchResults').innerHTML = '';
            // Обновляем список контактов
            await showContactsModal();
        } else {
            const error = await response.json();
            alert(error.error || 'Ошибка добавления контакта');
        }
    } catch (error) {
        console.error('Ошибка добавления контакта:', error);
        alert('Ошибка добавления контакта');
    }
}

// === МОБИЛЬНОЕ МЕНЮ ===

// Инициализация обработчиков мобильного меню
function initMobileMenuHandlers() {
    const burgerMenuBtn = document.getElementById('burgerMenuBtn');
    const mobileMenu = document.getElementById('mobileMenu');
    const mobileMenuOverlay = document.getElementById('mobileMenuOverlay');
    const closeMobileMenu = document.getElementById('closeMobileMenu');

    // Кнопка бургер-меню
    if (burgerMenuBtn) {
        burgerMenuBtn.addEventListener('click', openMobileMenu);
    }

    // Кнопка закрытия
    if (closeMobileMenu) {
        closeMobileMenu.addEventListener('click', closeMobileMenuFunc);
    }

    // Закрытие по клику на оверлей
    if (mobileMenuOverlay) {
        mobileMenuOverlay.addEventListener('click', closeMobileMenuFunc);
    }

    // Закрытие по нажатию Escape
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            closeMobileMenuFunc();
        }
    });

    // Обработчик изменения размера окна
    window.addEventListener('resize', function() {
        const inputContainer = document.querySelector('.input-container');
        const mobileMenu = document.getElementById('mobileMenu');

        // Если меню открыто на мобильном, блок ввода должен быть скрыт
        if (inputContainer && window.innerWidth <= 768) {
            if (mobileMenu && mobileMenu.classList.contains('open')) {
                inputContainer.classList.add('hidden');
            } else {
                inputContainer.classList.remove('hidden');
            }
        }
    });

    // Мобильные кнопки
    const mobileCreateChatBtn = document.getElementById('mobileCreateChatBtn');
    const mobileContactsBtn = document.getElementById('mobileContactsBtn');
    const mobileProfileBtn = document.getElementById('mobileProfileBtn');
    const mobileLogoutBtn = document.getElementById('mobileLogoutBtn');

    if (mobileCreateChatBtn) {
        mobileCreateChatBtn.addEventListener('click', function() {
            closeMobileMenuFunc();
            showNewChatModal();
        });
    }

    if (mobileContactsBtn) {
        mobileContactsBtn.addEventListener('click', function() {
            closeMobileMenuFunc();
            showContactsModal();
        });
    }

    if (mobileProfileBtn) {
        mobileProfileBtn.addEventListener('click', function() {
            closeMobileMenuFunc();
            showProfileModal();
        });
    }

    if (mobileLogoutBtn) {
        mobileLogoutBtn.addEventListener('click', function() {
            closeMobileMenuFunc();
            showLogoutConfirm();
        });
    }
}

// Открытие мобильного меню
function openMobileMenu() {
    const mobileMenu = document.getElementById('mobileMenu');
    const mobileMenuOverlay = document.getElementById('mobileMenuOverlay');
    const inputContainer = document.querySelector('.input-container');
    const msgForm = document.getElementById('msgForm');

    if (mobileMenu && mobileMenuOverlay) {
        mobileMenu.classList.add('open');
        mobileMenuOverlay.classList.add('open');
        document.body.style.overflow = 'hidden'; // Блокируем прокрутку

        // Скрываем блок ввода при открытии меню
        if (inputContainer && window.innerWidth <= 768) {
            inputContainer.classList.add('hidden');
        }
    }
}

// Закрытие мобильного меню
function closeMobileMenuFunc() {
    const mobileMenu = document.getElementById('mobileMenu');
    const mobileMenuOverlay = document.getElementById('mobileMenuOverlay');
    const inputContainer = document.querySelector('.input-container');
    const msgForm = document.getElementById('msgForm');

    if (mobileMenu && mobileMenuOverlay) {
        mobileMenu.classList.remove('open');
        mobileMenuOverlay.classList.remove('open');
        document.body.style.overflow = ''; // Возвращаем прокрутку

        // Показываем блок ввода при закрытии меню (только на мобильных)
        if (inputContainer && window.innerWidth <= 768) {
            inputContainer.classList.remove('hidden');
        }
    }
}

// Обновленная функция отображения чатов (для мобильного меню)
function displayChats(chats) {
    chatsList = chats;

    // Обновляем основной список чатов
    const chatList = document.getElementById('chatList');
    if (chatList) {
        chatList.innerHTML = '';
        chats.forEach(chat => {
            const li = document.createElement('li');
            li.onclick = () => selectChat(chat);
            li.innerHTML = `
                <span>${chat.name}</span>
                ${chat.unread_count > 0 ? `<span class="unread">${chat.unread_count}</span>` : ''}
            `;
            if (currentChat && currentChat.id === chat.id) {
                li.classList.add('active');
            }
            chatList.appendChild(li);
        });
    }

    // Обновляем мобильный список чатов
    const mobileChatList = document.getElementById('mobileChatList');
    if (mobileChatList) {
        mobileChatList.innerHTML = '';
        chats.forEach(chat => {
            const li = document.createElement('li');
            li.onclick = () => {
                selectChat(chat);
                closeMobileMenuFunc(); // Закрываем меню после выбора чата
            };
            li.innerHTML = `
                <span>${chat.name}</span>
                ${chat.unread_count > 0 ? `<span class="unread">${chat.unread_count}</span>` : ''}
            `;
            if (currentChat && currentChat.id === chat.id) {
                li.classList.add('active');
            }
            mobileChatList.appendChild(li);
        });
    }
}
