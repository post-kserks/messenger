// Глобальные переменные
let currentUser = null;
let currentChat = null;
let ws = null;
let currentMessages = [];
let chatsList = [];

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
});

// Инициализация обработчиков событий
function initEventHandlers() {
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
        msgForm.addEventListener('submit', sendMessage);
    }

    const fileForm = document.getElementById('fileForm');
    if (fileForm) {
        fileForm.addEventListener('submit', sendFile);
    }

    const createChatBtn = document.getElementById('createChatBtn');
    if (createChatBtn) {
        createChatBtn.addEventListener('click', showNewChatModal);
    }

    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }

    const profileBtn = document.getElementById('profileBtn');
    if (profileBtn) {
        profileBtn.addEventListener('click', showProfileModal);
    }

    const contactsBtn = document.getElementById('contactsBtn');
    if (contactsBtn) {
        contactsBtn.addEventListener('click', showContactsModal);
    }

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

        // Если мы на странице логина, перенаправляем на главную
        if (window.location.pathname === '/static/login.html' || window.location.pathname === '/static/register.html') {
            window.location.href = '/';
        }

        // Загружаем данные для главной страницы
        if (window.location.pathname === '/' || window.location.pathname === '/static/index.html') {
            loadChats();
            connectWebSocket();
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

// Отображение чатов
function displayChats(chats) {
    const chatList = document.getElementById('chatList');
    if (!chatList) return;
    chatList.innerHTML = '';
    chats.forEach(chat => {
        const li = document.createElement('li');
        li.textContent = chat.name;
        if (chat.unread_count && chat.unread_count > 0) {
            const badge = document.createElement('span');
            badge.textContent = ` (${chat.unread_count})`;
            badge.style.color = '#4f8cff';
            badge.style.fontWeight = 'bold';
            li.appendChild(badge);
        }
        li.onclick = () => selectChat(chat);
        chatList.appendChild(li);
    });
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
    const fileForm = document.getElementById('fileForm');
    if (msgForm) msgForm.style.display = 'flex';
    if (fileForm) fileForm.style.display = 'flex';
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
    // Обновляем активный чат в списке
    const chatItems = document.querySelectorAll('#chatList li');
    chatItems.forEach(item => item.classList.remove('active'));
    event.target.classList.add('active');
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
            currentMessages = messages;
            renderMessages();
        }
    } catch (error) {
        console.error('Ошибка загрузки сообщений:', error);
    }
}

// Отображение сообщений
function renderMessages() {
    const messagesDiv = document.getElementById('messages');
    if (!messagesDiv) return;
    messagesDiv.innerHTML = '';
    // Проверяем, что currentMessages - это массив
    if (!currentMessages || !Array.isArray(currentMessages)) {
        currentMessages = [];
    }
    // Сортируем по времени
    currentMessages.sort((a, b) => new Date(a.sent_at) - new Date(b.sent_at));
    currentMessages.forEach(msg => {
        const msgDiv = document.createElement('div');
        let className = `msg ${Number(msg.sender_id) === Number(currentUser.id) ? 'me' : ''}`;

        // Добавляем стили для временных и неудачных сообщений
        if (msg.isTemp) {
            className += ' temp-message';
        }
        if (msg.failed) {
            className += ' failed-message';
        }

        msgDiv.className = className;
        const bubble = document.createElement('div');
        bubble.className = 'bubble';

        if (msg.type === 'file' || (msg.text && msg.text.startsWith('[file]'))) {
            const fileUrl = msg.file_url || (msg.text ? msg.text.replace('[file]', '') : '');
            bubble.innerHTML = `<a href="${fileUrl}" target="_blank">📎 ${msg.file_name || 'Файл'}</a>`;
        } else {
            bubble.textContent = msg.text;
        }

        // Добавляем индикатор статуса для временных сообщений
        if (msg.isTemp && !msg.failed) {
            const status = document.createElement('span');
            status.className = 'status-indicator';
            status.innerHTML = '⏳';
            status.title = 'Отправляется...';
            bubble.appendChild(status);
        }

        // Добавляем индикатор ошибки для неудачных сообщений
        if (msg.failed) {
            const error = document.createElement('span');
            error.className = 'error-indicator';
            error.innerHTML = '❌';
            error.title = 'Ошибка отправки';
            bubble.appendChild(error);
        }

        const meta = document.createElement('div');
        meta.className = 'meta';
        let timeText = new Date(msg.sent_at).toLocaleString();
        if (msg.isTemp && !msg.failed) {
            timeText += ' (отправляется...)';
        } else if (msg.failed) {
            timeText += ' (ошибка)';
        }
        meta.innerHTML = `<span class="username">${msg.username}</span> • ${timeText}`;
        msgDiv.appendChild(bubble);
        msgDiv.appendChild(meta);
        messagesDiv.appendChild(msgDiv);
    });
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

// Отправка сообщения
async function sendMessage(e) {
    e.preventDefault();

    if (!currentChat) return;

    const input = document.getElementById('msgInput');
    const text = input.value.trim();

    if (!text) return;

    // Создаем временное сообщение для мгновенного отображения
    const tempMessage = {
        id: 'temp_' + Date.now(), // Временный ID с префиксом
        chat_id: currentChat.id,
        sender_id: currentUser.id,
        username: currentUser.username,
        text: text,
        sent_at: new Date().toISOString(),
        type: 'text',
        isTemp: true // Флаг временного сообщения
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

// Функция отправки сообщения на сервер
async function sendMessageToServer(chatId, text, tempId) {
    try {
        const response = await fetch('/api/messages', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            },
            body: JSON.stringify({
                chat_id: chatId,
                text: text
            })
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

// Отправка файла
async function sendFile(e) {
    e.preventDefault();

    if (!currentChat) return;

    const fileInput = document.getElementById('fileInput');
    const file = fileInput.files[0];

    if (!file) return;

    // Создаем временное сообщение для мгновенного отображения
    const tempMessage = {
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
    fileInput.value = '';
    currentMessages.push(tempMessage);
    renderMessages();

    // Обновляем чат в списке локально
    updateChatLocally(currentChat.id, tempMessage);

    // Отправляем файл на сервер в фоне
    sendFileToServer(currentChat.id, file, tempMessage.id);
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

// Обработка WebSocket сообщений
function handleWebSocketMessage(data) {
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
        file_name: data.file_name
    };
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

// Выход из системы
function logout() {
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

// Показать модальное окно профиля
async function showProfileModal() {
    const profileModal = document.getElementById('profileModal');
    const profileInfo = document.getElementById('profileInfo');
    const profileUsername = document.getElementById('profileUsername');
    const profileNickname = document.getElementById('profileNickname');
    const profileEmail = document.getElementById('profileEmail');
    const profilePassword = document.getElementById('profilePassword');
    const profileMsg = document.getElementById('profileMsg');
    if (currentUser) {
        profileInfo.innerHTML = `
            <p><strong>ID:</strong> ${currentUser.id}</p>
            <p><strong>Имя:</strong> ${currentUser.username}</p>
            <p><strong>Никнейм:</strong> ${currentUser.nickname || 'Не задан'}</p>
            <p><strong>Email:</strong> ${currentUser.email}</p>
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
