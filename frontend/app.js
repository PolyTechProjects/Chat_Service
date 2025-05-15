// app.js
document.addEventListener('DOMContentLoaded', function() {
    const chatList = document.getElementById('chatList');
    const chatsContainer = document.getElementById('chatsContainer');
    const chatContent = document.getElementById('chatContent');
    const noChatSelected = document.querySelector('.no-chat-selected');
    const chatActive = document.querySelector('.chat-active');
    const chatInput = document.querySelector('.chat-input');
    const chatTitle = document.getElementById('chatTitle');
    const chatAvatar = document.getElementById('chatAvatar');
    const chatMessages = document.getElementById('chatMessages');
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');
    const attachButton = document.getElementById('attachButton');
    const backButton = document.getElementById('backButton');
    const fileInput = document.getElementById('fileInput');
    const currentUsername = document.getElementById('currentUsername');
    const currentUserProfilePic = document.getElementById('currentUserProfilePic');
    const createChatButton = document.getElementById('createChatButton');
    
    let currentChat = null;
    let currentSocket = null;
    let currentUserId = null; // ID текущего пользователя
    let currentProfilePic = null;
    let currentNickname = null;
    let currentRoleId = null;
    let currentPermissions = null;
    let currentRoleName = null;
    let currentFilesToSend = [];
    let isCurrentChatDirect = false;
    
    // Загружаем список чатов
    loadChats();
    
    // Загружаем ID текущего пользователя (нужно для direct чатов)
    loadCurrentUser();
    
    // Обработчики событий
    sendButton.addEventListener('click', sendMessage);
    attachButton.addEventListener('click', () => fileInput.click());
    backButton.addEventListener('click', closeChat);
    fileInput.addEventListener('change', handleFileUpload);
    messageInput.addEventListener('keypress', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    });
    createChatButton.addEventListener('click', function() {
        window.open('/chat_info?mode=create', '_blank');
    });
    
    // Функция загрузки текущего пользователя
    async function loadCurrentUser() {
        try {
            const response = await fetch('http://localhost:4001/api/v1/auth/me', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (response.ok) {
                const data = await response.json();
                currentUserId = data.user_id;
                localStorage.setItem('userId', currentUserId);
            } else if (response.status === 401) {
                const response = await fetch('http://localhost:4001/api/v1/auth/refresh', {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    },
                    method: 'POST'
                });
                if (response.ok) {
                    const data = await response.json();
                    localStorage.setItem('token', data.token);

                    const response = await fetch('http://localhost:4001/api/v1/auth/me', {
                        headers: {
                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                        }
                    });
            
                    if (response.ok) {
                        const data = await response.json();
                        currentUserId = data.user_id;
                        localStorage.setItem('userId', currentUserId);
                    }
                } else {
                    localStorage.removeItem('token');
                    window.location.pathname = '/login';
                }
            }

            const userResponse = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${currentUserId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });

            if (userResponse.ok) {
                const data = await userResponse.json();
                currentProfilePic = data.profile_pic;
                currentUserProfilePic.src = `http://localhost:4004/api/v1/media/uploads?mediaId=${data.profile_pic}`;
                currentUsername.className = "welcome-text";
                currentUsername.innerText = data.name;
                console.log(data.name)
                console.log(data)
            }
        } catch (error) {
            console.error('Ошибка загрузки пользователя:', error);
        }
    }
    
    // Функция загрузки списка чатов
    async function loadChats() {
        try {
            const response = await fetch('http://localhost:4003/api/v1/chats', {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (response.ok) {
                const chats = await response.json();
                renderChatList(chats);
            } else {
                if (response.status === 401) {
                    const response = await fetch('http://localhost:4001/api/v1/auth/refresh', {
                        headers: {
                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                        },
                        method: 'POST'
                    });
                    if (response.ok) {
                        const data = await response.json();
                        localStorage.setItem('token', data.token);

                        const response = await fetch('http://localhost:4003/api/v1/chats', {
                            headers: {
                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                            }
                        });
                        
                        if (response.ok) {
                            const chats = await response.json();
                            renderChatList(chats);
                        }
                    } else {
                        localStorage.removeItem('token');
                        window.location.pathname = '/login';
                    }
                } else {
                    console.error('Ошибка загрузки чатов');
                }
            }
        } catch (error) {
            console.error('Ошибка:', error);
        }
    }
    
    // Функция отрисовки списка чатов
    async function renderChatList(chats) {
        chatsContainer.innerHTML = '';
        
        chats.chats.forEach(chat => {
            const chatItem = document.createElement('div');
            chatItem.className = 'chat-item';
            chatItem.innerHTML = `
                <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${chat.profile_pic}" class="chat-avatar" width="128" height="128">
                <span class="chat-name">${chat.name}</span>
            `;
            
            chatItem.addEventListener('click', () => openChat(chat));
            chatsContainer.appendChild(chatItem);
        });
        chats.direct_chats.forEach(async chat => {
            const chatItem = document.createElement('div');
            chatItem.className = 'chat-item';
            targetUserId = chat.first_user_id === currentUserId ? chat.second_user_id : chat.first_user_id
            const response = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${targetUserId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            const data = await response.json();
            var targetProfilePic = data.profile_pic;
            chatItem.innerHTML = `
                <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${targetProfilePic}" class="chat-avatar" width="128" height="128">
                <span class="chat-name">${data.name}</span>
            `;
            
            chatItem.addEventListener('click', () => openChat(chat));
            chatsContainer.appendChild(chatItem);
        });
    }
    
    // Функция открытия чата
    async function openChat(chat) {
        currentChat = chat;
        noChatSelected.style.display = 'none';
        chatActive.style.display = 'flex';
        
        // Устанавливаем заголовок и аватар
        chatMessages.scrollTop = chatMessages.scrollHeight;
        let chat_id;
        console.log(chat)
        if (chat.first_user_id) {
            isCurrentChatDirect = true;
            const response = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${currentUserId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            const data = await response.json();

            var otherUserId = chat.first_user_id === currentUserId ? chat.second_user_id : chat.first_user_id;
            const otherResponse = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${otherUserId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            const otherData = await otherResponse.json();
            chatTitle.textContent = otherData.name;
            chatAvatar.src=`http://localhost:4004/api/v1/media/uploads?mediaId=${otherData.profile_pic}`;
            currentNickname = data.name;
            currentRoleId = 0;
            currentRoleName = '';
            currentPermissions = [];
            chat_id = chat.chat_id;
            console.log(data);
            console.log(otherData);
        } else {
            isCurrentChatDirect = false;
            chatTitle.textContent = chat.name;
            chatAvatar.src=`http://localhost:4004/api/v1/media/uploads?mediaId=${chat.profile_pic}`;
            const response = await fetch(`http://localhost:4003/api/v1/chats/${chat.id}/users/${currentUserId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            const data = await response.json()

            const roleResponse = await fetch(`http://localhost:4003/api/v1/chats/${chat.id}/roles/${data.role_id}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            const roleData = await roleResponse.json()

            currentNickname = data.nickname
            currentRoleId = data.role_id
            currentRoleName = roleData.name
            currentPermissions = roleData.permissions
            chat_id = chat.id
            console.log(data)
            console.log(roleData)
            if (!roleData.permissions.includes("CAN_SEND_FILE")) {
                attachButton.style.display = 'none';
            }
            if (!roleData.permissions.includes("CAN_WRITE_MESSAGE")) {
                chatInput.style.display = 'none';
            }
        }
        console.log(chat_id);
        
        // Загружаем историю сообщений
        await loadChatHistory(chat_id);
        
        // Открываем WebSocket соединение
        openWebSocketConnection(chat_id);
    }
    
    // Функция загрузки истории сообщений
    async function loadChatHistory(chatId) {
        try {
            let endpoint;
            if (isCurrentChatDirect) {
                endpoint = `http://localhost:4005/api/v1/messaging/history/direct/${chatId}`;
            } else {
                endpoint = `http://localhost:4005/api/v1/messaging/history/${chatId}`;
            }
            
            const response = await fetch(endpoint, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (response.ok) {
                const messages = await response.json();
                renderMessages(messages.messages);
            }
        } catch (error) {
            console.error('Ошибка загрузки истории:', error);
        }
    }
    
    // Функция отрисовки сообщений
    function renderMessages(messages) {
        chatMessages.innerHTML = '';
        
        messages.forEach(msg => {
            const messageElement = document.createElement('div');
            const messageBoxElement = document.createElement('div');
            const profilePicElement = document.createElement('div');
            messageBoxElement.className = 'message-box';
            profilePicElement.className = "message-profile-pic";

            messageElement.dataset.messageId = msg.message_id;

            const loadDirectSender = () => {
                const url = `http://localhost:4002/api/v1/users/profiles?userId=${msg.sender_id}`;
                const token = localStorage.getItem('token');
                const headers = {
                    Authorization: `Bearer ${token}`,
                };
                
                // Make the request and set the sender name
                fetch(url, { headers })
                    .then(response => response.json())
                    .then(async data => {
                        const senderName = data.name;
                        const roleName = '';
                        messageBoxElement.innerHTML = `
                            <div class="message-metadata">
                                <span class="message-sender-name">${senderName}</span>
                                <span class="message-role-name">${roleName}</span>
                            </div>
                            <div class="message-text">${msg.body}</div>
                            <div class="message-footer">
                                <span class="message-date">${new Date(msg.created_at).toLocaleString()}</span>
                            </div>
                        `;
                        
                        profilePicElement.innerHTML = `
                            <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${data.profile_pic}" class="user-avatar" width="128" height="128">
                        `;
                        
                        // Load images
                        if (msg.files.length > 0) {
                            console.log(msg.files.length)
                            const messageFilesElement = document.createElement('div');
                            messageFilesElement.className = "message-files";
                            messageBoxElement.appendChild(messageFilesElement);
                            
                            msg.files.forEach(file => {
                                const imageUrl = `http://localhost:4004/api/v1/media/uploads?mediaId=${file}`;
                                const image = document.createElement('img');
                                image.src = imageUrl;
                                image.className = 'message-file';
                                image.alt = 'Image';
                                
                                // Add Authorization header to image request
                                const imageProxy = document.createElement('img');
                                imageProxy.src = '';
                                imageProxy.onload = () => {
                                    image.src = imageProxy.src;
                                };
                                fetch(imageUrl, { headers })
                                    .then(response => response.blob())
                                    .then(blob => {
                                        const urlCreator = window.URL || window.webkitURL;
                                        const imageUrl = urlCreator.createObjectURL(blob);
                                        imageProxy.src = imageUrl;
                                    })
                                    .catch(error => console.error(error));
                                
                                messageFilesElement.appendChild(image);
                            });
                        }
                    })
                    .catch(error => console.error(error));
            };

            const loadSender = () => {
                const url = `http://localhost:4003/api/v1/chats/${msg.chat_id}/users/${msg.sender_id}`;
                const token = localStorage.getItem('token');
                const headers = {
                    Authorization: `Bearer ${token}`,
                };
                
                // Make the request and set the sender name
                fetch(url, { headers })
                    .then(response => response.json())
                    .then(async data => {
                        const roleResponse = await fetch(`http://localhost:4003/api/v1/chats/${msg.chat_id}/roles/${data.role_id}`, {
                            headers: {
                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                            }
                        })
                        const roleData = await roleResponse.json();
                        const senderName = data.nickname;
                        const roleName = roleData.name;
                        messageBoxElement.innerHTML = `
                            <div class="message-metadata">
                                <span class="message-sender-name">${senderName}</span>
                                <span class="message-role-name">${roleName}</span>
                            </div>
                            <div class="message-text">${msg.body}</div>
                            <div class="message-footer">
                                <span class="message-date">${new Date(msg.created_at).toLocaleString()}</span>
                            </div>
                        `;


                        const userResponse = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${msg.sender_id}`, {
                            headers: {
                                'Authorization': `Bearer ${localStorage.getItem('token')}`
                            }
                        })

                        const userData = await userResponse.json();
                        
                        profilePicElement.innerHTML = `
                            <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${userData.profile_pic}" class="user-avatar" width="128" height="128">
                        `;
                        
                        // Load images
                        if (msg.files.length > 0) {
                            const messageFilesElement = document.createElement('div');
                            messageFilesElement.className = "message-files";
                            messageBoxElement.appendChild(messageFilesElement);
                            
                            msg.files.forEach(file => {
                                const imageUrl = `http://localhost:4004/api/v1/media/uploads?mediaId=${file}`;
                                const image = document.createElement('img');
                                image.src = imageUrl;
                                image.className = 'message-file';
                                image.alt = 'Image';
                                
                                // Add Authorization header to image request
                                const imageProxy = document.createElement('img');
                                imageProxy.src = '';
                                imageProxy.onload = () => {
                                    image.src = imageProxy.src;
                                };
                                fetch(imageUrl, { headers })
                                    .then(response => response.blob())
                                    .then(blob => {
                                        const urlCreator = window.URL || window.webkitURL;
                                        const imageUrl = urlCreator.createObjectURL(blob);
                                        imageProxy.src = imageUrl;
                                    })
                                    .catch(error => console.error(error));
                                
                                messageFilesElement.appendChild(image);
                            });
                        }
                    })
                    .catch(error => console.error(error));
            };
            
            // Call the function to load the sender name
            if (msg.is_direct) {
                loadDirectSender();
            } else {
                loadSender();
            }

            const messageContainer = document.createElement('div');
            messageContainer.className = 'message-container';
            if (msg.sender_id !== currentUserId) {
                if (!msg.is_direct) {
                    const permissions = currentPermissions;
                    console.log(permissions)
                    console.log("can edit or delete others message")
                    const messageActions = document.createElement('div');
                    messageActions.className = 'message-actions';
                    if (permissions.includes("CAN_EDIT_OTHERS_MESSAGE") && !permissions.includes("CAN_DELETE_OTHERS_MESSAGE")) {
                        messageActions.innerHTML = `
                            <div class="message-action edit-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-pencil-alt"></i>
                            </div>
                        `
                    }
                    if (!permissions.includes("CAN_EDIT_OTHERS_MESSAGE") && permissions.includes("CAN_DELETE_OTHERS_MESSAGE")) {
                        messageActions.innerHTML = `
                            <div class="message-action delete-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-trash"></i>
                            </div>
                        `
                    }
                    if (permissions.includes("CAN_EDIT_OTHERS_MESSAGE") && permissions.includes("CAN_DELETE_OTHERS_MESSAGE")) {
                        messageActions.innerHTML = `
                            <div class="message-action edit-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-pencil-alt"></i>
                            </div>
                            <div class="message-action delete-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-trash"></i>
                            </div>
                        `
                    }
                    messageContainer.appendChild(messageActions);
                }
            } else {
                if (!msg.is_direct) {
                    const permissions = currentPermissions;
                    console.log(permissions)
                    console.log("can edit or delete message")
                    const messageActions = document.createElement('div');
                    messageActions.className = 'message-actions';
                    if (permissions.includes("CAN_EDIT_MESSAGE") && !permissions.includes("CAN_DELETE_MESSAGE")) {
                        messageActions.innerHTML = `
                            <div class="message-action edit-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-pencil-alt"></i>
                            </div>
                        `
                    }
                    if (!permissions.includes("CAN_EDIT_MESSAGE") && permissions.includes("CAN_DELETE_MESSAGE")) {
                        messageActions.innerHTML = `
                            <div class="message-action delete-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-trash"></i>
                            </div>
                        `
                    }
                    if (permissions.includes("CAN_EDIT_MESSAGE") && permissions.includes("CAN_DELETE_MESSAGE")) {
                        messageActions.innerHTML = `
                            <div class="message-action edit-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-pencil-alt"></i>
                            </div>
                            <div class="message-action delete-message" data-message-id="${msg.message_id}">
                                <i class="fas fa-trash"></i>
                            </div>
                        `
                    }
                    messageContainer.appendChild(messageActions);
                } else {
                    const messageActions = document.createElement('div');
                    messageActions.className = 'message-actions';
                    messageActions.innerHTML = `
                        <div class="message-action edit-message" data-message-id="${msg.message_id}">
                            <i class="fas fa-pencil-alt"></i>
                        </div>
                        <div class="message-action delete-message" data-message-id="${msg.message_id}">
                            <i class="fas fa-trash"></i>
                        </div>
                    `
                    messageContainer.appendChild(messageActions);
                }
            }

            if (msg.sender_id === currentUserId) {
                messageElement.className = 'own-message';
                messageElement.appendChild(messageBoxElement);
                messageElement.appendChild(messageContainer);
                messageElement.appendChild(profilePicElement);
            } else {
                messageElement.className = 'message';
                messageElement.appendChild(profilePicElement);
                messageElement.appendChild(messageBoxElement);
                messageElement.appendChild(messageContainer);
            }

            chatMessages.appendChild(messageElement);
        });
    
        // Добавляем обработчики для кнопок
        document.querySelectorAll('.edit-message').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const messageId = btn.dataset.messageId;
                editMessage(messageId);
            });
        });
        
        document.querySelectorAll('.delete-message').forEach(btn => {
            btn.addEventListener('click', (e) => {
                e.stopPropagation();
                const messageId = btn.dataset.messageId;
                deleteMessage(messageId);
            });
        });
        
        // Прокручиваем вниз
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }

    // Новые функции для работы с сообщениями
    async function deleteMessage(messageId) {
        try {
            const response = await fetch(`http://localhost:4005/api/v1/messaging/history/${messageId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            });
            
            if (response.ok) {
                console.log('Сообщение удалено');
                const messageElement = document.querySelector(`[data-message-id="${messageId}"]`);
                if (messageElement) {
                    messageElement.closest('.message, .own-message').remove();
                }
            } else {
                const errorText = await response.text();
                alert('Ошибка удаления сообщения: ' + errorText);
            }
        } catch (error) {
            console.error('Ошибка удаления сообщения:', error);
        }
    }

    async function editMessage(messageId) {
        const newText = prompt('Введите новый текст сообщения:', );
        if (newText) {
            try {
                const response = await fetch(`http://localhost:4005/api/v1/messaging/history`, {
                    method: 'PUT',
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`,
                        'Content-Type': 'application/json'
                    },
                    body: JSON.stringify({
                        message_id: messageId,
                        body: newText
                    })
                });
                
                if (response.ok) {
                    console.log('Сообщение обновлено');
                    const messageElement = document.querySelector(`[data-message-id="${messageId}"]`);
                    if (messageElement) {
                        const messageBoxElement = messageElement.querySelector('.message-text');
                        messageBoxElement.textContent = newText;
                    }
                } else {
                    const errorText = await response.text();
                    alert('Ошибка обновления сообщения: ' + errorText);
                }
            } catch (error) {
                console.error('Ошибка редактирования сообщения:', error);
            }
        }
    }
    
    // Функция открытия WebSocket соединения
    function openWebSocketConnection(chatId) {
        // Закрываем предыдущие соединения
        if (currentSocket) currentSocket.close();
        
        if (isCurrentChatDirect) {
            // Соединение для отправки сообщений
            currentSocket = new WebSocket(`ws://localhost:4005/api/v1/messaging/ws/direct/${currentChat.chat_id}?token=${localStorage.getItem('token')}`);
        } else {
            currentSocket = new WebSocket(`ws://localhost:4005/api/v1/messaging/ws/${chatId}?token=${localStorage.getItem('token')}`);
        }
        
        currentSocket.onopen = function() {
            console.log('WebSocket соединение установлено');
        };
        
        currentSocket.onmessage = function(event) {
            const message = JSON.parse(event.data);
            console.log(message);
            addNewMessage(message);
        };
        
        currentSocket.onclose = function() {
            console.log('WebSocket соединение закрыто');
        };
    }
    
    // Функция добавления нового сообщения
    async function addNewMessage(message) {
        const messageElement = document.createElement('div');
        const messageBoxElement = document.createElement('div');
        const profilePicElement = document.createElement('div');
        const messageFilesElement = document.createElement('div');
        messageBoxElement.className = 'message-box';
        profilePicElement.className = 'message-profile-pic';
        messageFilesElement.className = 'message-files';
        let user_id = null;
        let nickname = null;
        let role_name = null;
        let profile_pic = null;
        if (message.sender_id) {
            user_id = message.sender_id;
            if (message.is_direct) {
                const response = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${user_id}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                })
                const data = await response.json();

                nickname = data.name;
                role_name = '';
                profile_pic = data.profile_pic;
            } else {
                const response = await fetch(`http://localhost:4003/api/v1/chats/${message.chat_id}/users/${user_id}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                })

                const data = await response.json();
                nickname = data.nickname;

                const roleResponse = await fetch(`http://localhost:4003/api/v1/chats/${message.chat_id}/roles/${data.role_id}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                })
                const roleData = await roleResponse.json();
                role_name = roleData.name;

                const userResponse = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${user_id}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    }
                })

                const userData = await userResponse.json();
                profile_pic = userData.profile_pic;
            }
        } else {
            user_id = currentUserId;
            nickname = currentNickname;
            role_name = currentRoleName;
            profile_pic = currentProfilePic;
        }

        messageBoxElement.innerHTML = `
            <div class="message-metadata">
                <span class="message-sender-name">${nickname}</span>
                <span class="message-role-name">${role_name}</span>
            </div>
            <div class="message-text">${message.body}</div>
            <div class="message-footer">
                <span class="message-date">${new Date(message.created_at).toLocaleString()}</span>
            </div>
        `;
        profilePicElement.innerHTML = `
            <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${profile_pic}" class="user-avatar">
        `;

        const messageContainer = document.createElement('div');
        messageContainer.className = 'message-container';
        if (message.sender_id !== currentUserId) {
            if (!message.is_direct) {
                const permissions = currentPermissions;
                console.log(permissions)
                const messageActions = document.createElement('div');
                messageActions.className = 'message-actions';
                if (permissions.includes("CAN_EDIT_OTHERS_MESSAGE") && !permissions.includes("CAN_DELETE_OTHERS_MESSAGE")) {
                    messageActions.innerHTML = `
                        <div class="message-action edit-message" data-message-id="${message.message_id}">
                            <i class="fas fa-pencil-alt"></i>
                        </div>
                    `
                }
                if (!permissions.includes("CAN_EDIT_OTHERS_MESSAGE") && permissions.includes("CAN_DELETE_OTHERS_MESSAGE")) {
                    messageActions.innerHTML = `
                        <div class="message-action delete-message" data-message-id="${message.message_id}">
                            <i class="fas fa-trash"></i>
                        </div>
                    `
                }
                if (permissions.includes("CAN_EDIT_OTHERS_MESSAGE") && permissions.includes("CAN_DELETE_OTHERS_MESSAGE")) {
                    messageActions.innerHTML = `
                        <div class="message-action edit-message" data-message-id="${message.message_id}">
                            <i class="fas fa-pencil-alt"></i>
                        </div>
                        <div class="message-action delete-message" data-message-id="${message.message_id}">
                            <i class="fas fa-trash"></i>
                        </div>
                    `
                }
                messageContainer.appendChild(messageActions);
            }
        } else {
            if (!message.is_direct) {
                const permissions = currentPermissions;
                console.log(permissions)
                const messageActions = document.createElement('div');
                messageActions.className = 'message-actions';
                if (permissions.includes("CAN_EDIT_MESSAGE") && !permissions.includes("CAN_DELETE_MESSAGE")) {
                    messageActions.innerHTML = `
                        <div class="message-action edit-message" data-message-id="${message.message_id}">
                            <i class="fas fa-pencil-alt"></i>
                        </div>
                    `
                }
                if (!permissions.includes("CAN_EDIT_MESSAGE") && permissions.includes("CAN_DELETE_MESSAGE")) {
                    messageActions.innerHTML = `
                        <div class="message-action delete-message" data-message-id="${message.message_id}">
                            <i class="fas fa-trash"></i>
                        </div>
                    `
                }
                if (permissions.includes("CAN_EDIT_MESSAGE") && permissions.includes("CAN_DELETE_MESSAGE")) {
                    messageActions.innerHTML = `
                        <div class="message-action edit-message" data-message-id="${message.message_id}">
                            <i class="fas fa-pencil-alt"></i>
                        </div>
                        <div class="message-action delete-message" data-message-id="${message.message_id}">
                            <i class="fas fa-trash"></i>
                        </div>
                    `
                }
                messageContainer.appendChild(messageActions);
            } else {
                const messageActions = document.createElement('div');
                messageActions.className = 'message-actions';
                messageActions.innerHTML = `
                    <div class="message-action edit-message" data-message-id="${message.message_id}">
                        <i class="fas fa-pencil-alt"></i>
                    </div>
                    <div class="message-action delete-message" data-message-id="${message.message_id}">
                        <i class="fas fa-trash"></i>
                    </div>
                `
                messageContainer.appendChild(messageActions);
            }
        }

        if (message.sender_id) {
            messageElement.className = 'message';
            messageElement.appendChild(profilePicElement);
            messageElement.appendChild(messageBoxElement);
            messageElement.appendChild(messageContainer);
        } else {
            messageElement.className = 'own-message';
            messageElement.appendChild(messageBoxElement);
            messageElement.appendChild(messageContainer);
            messageElement.appendChild(profilePicElement);
        }
        if (message.files.length > 0) {
            message.files.forEach(file => {
                const imageUrl = `http://localhost:4004/api/v1/media/uploads?mediaId=${file}`;
                const image = document.createElement('img');
                image.className = 'message-file';
                image.src = imageUrl;
                image.alt = 'Image';
                messageFilesElement.appendChild(image);
            });
            messageBoxElement.appendChild(messageFilesElement);
        }
        chatMessages.appendChild(messageElement);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    
    // Функция отправки сообщения
    function sendMessage() {
        const text = messageInput.value.trim();
        if (!text && fileInput.files.length === 0) return;
        
        const message = {
            body: text,
            is_direct: isCurrentChatDirect,
            created_at: Date.now(),
            files: currentFilesToSend
        };
        
        if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
            console.log(message);
            currentSocket.send(JSON.stringify(message));
            messageInput.value = '';
            
            // Добавляем сообщение сразу в чат (оптимистичное обновление)
            addNewMessage(message);
            currentFilesToSend = [];
        }
    }
    
    // Функция обработки загрузки файлов
    async function handleFileUpload() {
        const files = fileInput.files;
        for (let i = 0; i < files.length; i++) {
            var form = new FormData();
            form.append('file', files[i]);
            const response = await fetch('http://localhost:4004/api/v1/media/uploads', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                },
                body: form
            });
            if (response.ok) {
                const data = await response.json();
                console.log(data);
                currentFilesToSend.push(data.media_id)
            };
        }
    }
    
    // Функция закрытия чата
    function closeChat() {
        if (currentSocket) currentSocket.close();
        
        currentChat = null;
        chatActive.style.display = 'none';
        noChatSelected.style.display = 'flex';
        chatMessages.innerHTML = '';
        messageInput.value = '';
    }
    
    // Обработчик клика на заголовок чата для открытия деталей
    chatTitle.addEventListener('click', openChatDetails);
    
    // Функция открытия деталей чата
    function openChatDetails() {
        if (!currentChat) return;
        
        if (isCurrentChatDirect) {
            const otherUserId = currentChat.first_user_id === currentUserId ? currentChat.second_user_id : currentChat.first_user_id;
            window.open(`/user_info?userId=${otherUserId}`, '_self');
        } else {
            window.open(`/chat_info?chatId=${currentChat.id}`, '_self');
        }
    }
});