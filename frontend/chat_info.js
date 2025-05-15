document.addEventListener('DOMContentLoaded', function() {
    const urlParams = new URLSearchParams(window.location.search);
    const isCreateMode = urlParams.get('mode') === 'create';
    const chatId = urlParams.get('chatId');
    
    // Элементы формы
    const chatInfoContainer = document.getElementById('chatInfoContainer');
    const pageTitle = document.getElementById('pageTitle');
    const successMessage = document.getElementById('successMessage');
    const formContent = document.getElementById('formContent');
    const chatNameInput = document.getElementById('chatName');
    const isChannelCheckbox = document.getElementById('isChannel');
    const isPrivateCheckbox = document.getElementById('isPrivate');
    const inviteLinkContainer = document.getElementById('inviteLinkContainer');
    const inviteLinkInput = document.getElementById('inviteLink');
    const chatDescriptionInput = document.getElementById('chatDescription');
    const avatarPreview = document.getElementById('avatarPreview');
    const chatAvatarInput = document.getElementById('chatAvatar');
    const participantsList = document.getElementById('participantsList');
    const permissionsContainer = document.getElementById('permissionsContainer');
    const saveBtn = document.getElementById('saveBtn');
    const cancelBtn = document.getElementById('cancelBtn');
    const updateBtn = document.getElementById('updateBtn');
    
    let avatarFile = null;
    let profile_pic = null;
    let participants = [];
    let selectedParticipants = [];
    let defaultPermissions = {
        CAN_WRITE_MESSAGE: false,
        CAN_PIN_MESSAGE: false,
        CAN_EDIT_OWN_MESSAGE: false,
        CAN_EDIT_OTHERS_MESSAGE: false,
        CAN_DELETE_OWN_MESSAGE: false,
        CAN_DELETE_OTHERS_MESSAGE: false,
        CAN_SEND_FILE: false,
        CAN_CHANGE_OWN_NICKNAME: false,
        CAN_CHANGE_OTHERS_NICKNAME: false,
        CAN_EDIT_CHAT: false,
        CAN_DELETE_CHAT: false,
        CAN_CREATE_ROLE: false,
        CAN_EDIT_ROLE: false,
        CAN_DELETE_ROLE: false,
        CAN_SET_ROLES: false,
        CAN_ADD_USERS: false,
        CAN_DELETE_USERS: false
    };
    
    // Инициализация страницы
    if (isCreateMode) {
        initCreateMode();
    } else if (chatId) {
        initViewMode(chatId);
    } else {
        window.location.href = '/';
    }
    
    // Обработчики событий
    isChannelCheckbox.addEventListener('change', function() {
        updatePermissionsDefault();
        toggleInviteLinkField();
    });
    
    chatAvatarInput.addEventListener('change', function(e) {
        if (e.target.files && e.target.files[0]) {
            avatarFile = e.target.files[0];
            const reader = new FileReader();
            reader.onload = function(event) {
                avatarPreview.src = event.target.result;
            };
            reader.readAsDataURL(avatarFile);
        }
    });
    
    saveBtn.addEventListener('click', saveChat);
    cancelBtn.addEventListener('click', function() {
        window.location.pathname = '/'; // Вернуться на предыдущую страницу
    });
    
    function initCreateMode() {
        pageTitle.textContent = 'Создание нового чата';
        saveBtn.textContent = 'Создать чат';
        loadParticipants();
        renderPermissions();
        updatePermissionsDefault();
    }

    function initViewMode(chatId) {
        pageTitle.textContent = 'Информация о чате';
        saveBtn.style.display = 'none';
        updateBtn.style.display = 'block';
        updateBtn.textContent = 'Сохранить изменения';
        updateBtn.addEventListener('click', function() {
            updateChat(chatId);
        });
        loadChatInfo(chatId);
        
        // Обработчик кнопки удаления чата
        document.getElementById('deleteChatBtn').addEventListener('click', function() {
            if (confirm('Вы уверены, что хотите удалить этот чат?')) {
                deleteChat(chatId);
            }
        });
    }

    function loadRoles(chatId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => response.json())
        .then(roles => {
            console.log(roles);
            renderRoles(roles);
        })
        .catch(error => {
            console.error('Ошибка загрузки ролей:', error);
        });
    }

    function renderRoles(roles) {
        const rolesList = document.getElementById('rolesList');
        rolesList.innerHTML = '';
        
        roles.roles.forEach(role => {
            const roleItem = document.createElement('div');
            roleItem.className = 'role-item';
            console.log(role);
            roleItem.innerHTML = `
                <div>
                    <span class="role-color" style="background-color: ${role.color};"></span>
                    <strong>${role.name}</strong> - ${role.description}
                </div>
                <div class="role-actions">
                    <button class="action-btn view-btn" data-role-id="${role.role_id}">Подробнее</button>
                </div>
            `;
            rolesList.appendChild(roleItem);
            
            // Обработчик кнопки просмотра роли
            roleItem.querySelector('.view-btn').addEventListener('click', function() {
                window.open(`/role_info?roleId=${role.role_id}&chatId=${chatId}`, '_blank');
            });
        });
    }

    function loadParticipantsWithRoles(chatId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => response.json())
        .then(data => {
            participants = data.participants;
            renderParticipantsWithRoles(participants, data.chat.id);
        })
        .catch(error => {
            console.error('Ошибка загрузки участников:', error);
        });
    }

    function renderParticipantsWithRoles(participants, chat_id) {
        participantsList.innerHTML = '';
        
        participants.forEach(async participant => {
            const roleResponse = await fetch(`http://localhost:4003/api/v1/chats/${chat_id}/roles/${participant.role_id}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            const roleData = await roleResponse.json();
            const userResponse = await fetch(`http://localhost:4002/api/v1/users/profiles?userId=${participant.user_id}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            const userData = await userResponse.json();
            const participantItem = document.createElement('div');
            participantItem.className = 'participant-item';
            
            participantItem.innerHTML = `
                <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${userData.profile_pic}" class="participant-avatar">
                <span>${participant.nickname}</span>
                <span class="participant-role">
                    ${roleData.name}
                </span>
                <div class="participant-actions">
                    <button class="action-btn edit-nickname-btn" data-user-id="${participant.user_id}" title="Изменить никнейм">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path>
                        </svg>
                    </button>
                    <button class="action-btn change-btn" data-user-id="${participant.user_id}">Изменить роль</button>
                    <button class="action-btn remove-btn" data-user-id="${participant.user_id}">Исключить</button>
                </div>
            `;
            
            participantsList.appendChild(participantItem);
            console.log(chat_id)
            
            // Обработчики кнопок
            participantItem.querySelector('.change-btn').addEventListener('click', function() {
                changeUserRole(participant.user_id, chat_id);
            });
            
            participantItem.querySelector('.remove-btn').addEventListener('click', function() {
                removeUserFromChat(participant.user_id, chat_id);
            });

            // Обработчик кнопки редактирования никнейма
            participantItem.querySelector('.edit-nickname-btn').addEventListener('click', function() {
                const userId = this.getAttribute('data-user-id');
                editUserNickname(userId, chat_id, participant.nickname);
            });
        });
    }

    function changeUserRole(userId, chatId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        .then(response => response.json())
        .then(data => {
            // Создаем модальное окно
            const modal = document.createElement('div');
            modal.style.position = 'fixed';
            modal.style.top = '0';
            modal.style.left = '0';
            modal.style.width = '100%';
            modal.style.height = '100%';
            modal.style.backgroundColor = 'rgba(0,0,0,0.5)';
            modal.style.display = 'flex';
            modal.style.justifyContent = 'center';
            modal.style.alignItems = 'center';
            modal.style.zIndex = '1000';
            
            modal.innerHTML = `
                <div style="background: white; padding: 20px; border-radius: 8px; width: 300px;">
                    <h3 style="margin-top: 0;">Изменить роль пользователя</h3>
                    <select id="roleSelect" style="width: 100%; padding: 8px; margin-bottom: 15px;">
                        ${data.roles.map(role => 
                            `<option value="${role.role_id}">${role.name}</option>`
                        ).join('')}
                    </select>
                    <div style="display: flex; justify-content: flex-end; gap: 10px;">
                        <button id="cancelBtn">Отмена</button>
                        <button id="confirmBtn" style="background: #4CAF50; color: white;">Подтвердить</button>
                    </div>
                </div>
            `;
            
            document.body.appendChild(modal);
            
            // Обработчики событий
            modal.querySelector('#cancelBtn').addEventListener('click', () => {
                document.body.removeChild(modal);
            });
            
            modal.querySelector('#confirmBtn').addEventListener('click', () => {
                const selectedRoleId = modal.querySelector('#roleSelect').value;
                updateUserRole(userId, chatId, selectedRoleId);
                document.body.removeChild(modal);
            });
        });
    }

    function updateUserRole(userId, chatId, roleId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/users/${userId}/role`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                role_id: roleId,
                chat_id: chatId,
                target_user_ids: [userId]
            })
        })
        .then(async response => {
            if (response.ok) {
                loadParticipantsWithRoles(chatId);
            } else {
                const errorText = await response.text();
                alert('Ошибка назначения роли: ' + errorText);
            }
        });
    }

    function removeUserFromChat(userId, chatId) {
        if (confirm('Вы уверены, что хотите исключить этого пользователя из чата?')) {
            fetch(`http://localhost:4003/api/v1/chats/${chatId}/users`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({
                    users_ids: [userId],
                    chat_id: chatId
                })
            })
            .then(response => {
                if (response.ok) {
                    loadParticipantsWithRoles(chatId);
                } else {
                    alert('Ошибка исключения пользователя');
                }
            });
        }
    }

    function deleteChat(chatId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => {
            if (response.ok) {
                window.location.href = '/';
            } else {
                alert('Ошибка удаления чата');
            }
        });
    }
    
    function loadParticipants() {
        const currentUserId = localStorage.getItem('userId');
        // Загрузка списка пользователей для добавления в чат
        fetch('http://localhost:4003/api/v1/chats', {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => response.json())
        .then(data => {
            // Фильтруем только direct чаты
            data.direct_chats.forEach(chat => {
                participants.push(chat.first_user_id === currentUserId ? chat.second_user_id : chat.first_user_id);
            })
            renderParticipants();
        })
        .catch(error => {
            console.error('Ошибка загрузки участников:', error);
        });
    }
    
    function renderParticipants() {
        participantsList.innerHTML = '';
        participants.forEach(participant => {
            console.log(participant)
            fetch(`http://localhost:4002/api/v1/users/profiles?userId=${participant}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            }).then(response => response.json())
            .then(data => {
                const participantItem = document.createElement('div');
                participantItem.className = 'participant-item';
                participantItem.innerHTML = `
                    <input type="checkbox" id="participant-${data.user_id}" class="participant-checkbox" data-user-id="${data.user_id}">
                    <img src="http://localhost:4004/api/v1/media/uploads?mediaId=${data.profile_pic}" class="participant-avatar">
                    <label for="participant-${data.user_id}">${data.name}</label>
                `;
                participantsList.appendChild(participantItem);
                const checkbox = participantItem.querySelector(`[data-user-id="${data.user_id}"]`);
                checkbox.addEventListener('change', function() {
                    updateSelectedParticipants(data.user_id, this.checked);
                });
            })
        });
    }

    function updateSelectedParticipants(participantId, isSelected) {
        if (isSelected) {
            if (!selectedParticipants.includes(participantId)) {
                selectedParticipants.push(participantId);
            }
        } else {
            selectedParticipants = selectedParticipants.filter(id => id !== participantId);
        }
        console.log('Выбранные участники:', selectedParticipants);
    }
    
    function renderPermissions() {
        permissionsContainer.innerHTML = '';
        for (const [permission, value] of Object.entries(defaultPermissions)) {
            const permissionItem = document.createElement('div');
            permissionItem.className = 'checkbox-item';
            permissionItem.innerHTML = `
                <input type="checkbox" id="${permission}" ${value ? 'checked' : ''}>
                <label for="${permission}">${formatPermissionName(permission)}</label>
            `;
            permissionsContainer.appendChild(permissionItem);
        }
    }
    
    function formatPermissionName(permission) {
        return permission
            .split('_')
            .map(word => word.charAt(0) + word.slice(1).toLowerCase())
            .join(' ');
    }
    
    function updatePermissionsDefault() {
        if (isChannelCheckbox.checked) {
            // Для канала все права по умолчанию false
            for (const permission in defaultPermissions) {
                defaultPermissions[permission] = false;
            }
        } else {
            // Для обычного чата устанавливаем стандартные права
            defaultPermissions = {
                CAN_WRITE_MESSAGE: true,
                CAN_PIN_MESSAGE: false,
                CAN_EDIT_OWN_MESSAGE: true,
                CAN_EDIT_OTHERS_MESSAGE: false,
                CAN_DELETE_OWN_MESSAGE: true,
                CAN_DELETE_OTHERS_MESSAGE: false,
                CAN_SEND_FILE: true,
                CAN_CHANGE_OWN_NICKNAME: true,
                CAN_CHANGE_OTHERS_NICKNAME: false,
                CAN_EDIT_CHAT: false,
                CAN_DELETE_CHAT: false,
                CAN_CREATE_ROLE: false,
                CAN_EDIT_ROLE: false,
                CAN_DELETE_ROLE: false,
                CAN_SET_ROLES: false,
                CAN_ADD_USERS: false,
                CAN_DELETE_USERS: false
            };
        }
        renderPermissions();
    }
    
    function toggleInviteLinkField() {
        if (isChannelCheckbox.checked) {
            inviteLinkContainer.style.display = 'block';
            inviteLinkInput.value = generateInviteLink();
        } else {
            inviteLinkContainer.style.display = 'none';
            inviteLinkInput.value = '';
        }
    }
    
    function generateInviteLink() {
        return `http://localhost:3000/join/${Math.random().toString(36).substring(2, 10)}`;
    }
    
    async function saveChat() {
        if (avatarFile) {
            var formData = new FormData();
            formData.append('file', avatarFile);
            const fileResponse = await fetch('http://localhost:4004/api/v1/media/uploads', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: formData
            });
            const fileData = await fileResponse.json();
            profile_pic = fileData.media_id;
            console.log(profile_pic);
            console.log(fileData);
        }
        
        // Собираем выбранные права
        const permissions = [];
        for (const permission in defaultPermissions) {
            if (document.getElementById(permission).checked) {
                permissions.push(permission);
            }
        }
        
        var body = JSON.stringify({
            name: chatNameInput.value,
            is_channel: isChannelCheckbox.checked,
            is_closed: isPrivateCheckbox.checked,
            description: chatDescriptionInput.value,
            profile_pic: profile_pic,
            permissions: permissions,
            participants_ids: selectedParticipants,
            creator_id: localStorage.getItem('userId')
        });
        console.log(body)
        
        // Отправка на сервер
        fetch('http://localhost:4003/api/v1/chats', {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: body
        })
        .then(response => {
            if (response.ok) {
                showSuccessMessage();
            } else {
                throw new Error('Ошибка создания чата');
            }
        })
        .catch(error => {
            console.error('Ошибка:', error);
            alert('Не удалось создать чат');
        });
    }
    
    async function updateChat(chatId) {
        if (avatarFile) {
            var formData = new FormData();
            formData.append('file', avatarFile);
            const fileResponse = await fetch('http://localhost:4004/api/v1/media/uploads', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: formData
            });
            const fileData = await fileResponse.json();
            profile_pic = fileData.media_id;
            console.log(profile_pic);
            console.log(fileData);
        }
        
        var body = JSON.stringify({
            chat_id: chatId,
            name: chatNameInput.value,
            description: chatDescriptionInput.value,
            join_link: inviteLinkInput.value,
            profile_pic: profile_pic,
        });
        console.log(body)
        
        // Отправка на сервер
        fetch(`http://localhost:4003/api/v1/chats/${chatId}`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: body
        })
        .then(response => {
            if (response.ok) {
                showSuccessMessage();
            } else {
                throw new Error('Ошибка обновления чата');
            }
        })
        .catch(error => {
            console.error('Ошибка:', error);
            alert('Не удалось обновить чат');
        });
    }
    
    function showSuccessMessage() {
        formContent.style.display = 'none';
        successMessage.style.display = 'block';
        
        setTimeout(() => {
            window.location.pathname = '/'; // Перенаправление на главную страницу
        }, 1000);
    }
    
    function loadChatInfo(chatId) {
        // Загрузка информации о чате для просмотра
        fetch(`http://localhost:4003/api/v1/chats/${chatId}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => response.json())
        .then(chat => {
            console.log(chat);
            // Заполняем форму данными чата
            chatNameInput.value = chat.chat.name
            isChannelCheckbox.checked = chat.chat.is_channel;
            isPrivateCheckbox.checked = chat.chat.is_closed;
            chatDescriptionInput.value = chat.chat.description;
            
            if (chat.chat.profile_pic) {
                avatarPreview.src = `http://localhost:4004/api/v1/media/uploads?mediaId=${chat.chat.profile_pic}`;
            }
            
            // Показываем участников
            participants = chat.participants.map(p => p.user_id);
            console.log(participants);
            
            // Показываем роли
            var roles = new Set(chat.participants.map(p => p.role_id));
            console.log(roles);
            
            // Загрузка ролей
            loadRoles(chatId);
            
            // Загрузка участников с ролями
            loadParticipantsWithRoles(chatId);
        })
        .catch(error => {
            console.error('Ошибка загрузки информации о чате:', error);
        });
    }

    // Показываем модальное окно создания роли
    function showCreateRoleModal(chatId) {
        const modal = document.getElementById('createRoleModal');
        const baseRoleSelect = document.getElementById('baseRoleSelect');
        const permissionsContainer = document.getElementById('newRolePermissions');
        
        // Загружаем список ролей для выбора базовой
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        .then(response => response.json())
        .then(data => {
            // Заполняем select ролями
            baseRoleSelect.innerHTML = data.roles.map(role => 
                `<option value="${role.role_id}">${role.name}</option>`
            ).join('');
            
            // Загружаем права первой роли (по умолчанию)
            if (data.roles.length > 0) {
                loadBaseRolePermissions(data.roles[0].role_id, chatId);
            }
            
            // При изменении базовой роли
            baseRoleSelect.addEventListener('change', () => {
                loadBaseRolePermissions(baseRoleSelect.value, chatId);
            });
        });
        
        // Показываем модальное окно
        modal.style.position = 'fixed';
        modal.style.top = '0';
        modal.style.left = '0';
        modal.style.width = '100%';
        modal.style.height = '100%';
        modal.style.backgroundColor = 'rgba(0,0,0,0.5)';
        modal.style.display = 'flex';
        modal.style.justifyContent = 'center';
        modal.style.alignItems = 'center';
        modal.style.zIndex = '1000';
        
        // Обработчики кнопок
        document.getElementById('cancelCreateRoleBtn').onclick = () => {
            modal.style.display = 'none';
        };
        
        document.getElementById('confirmCreateRoleBtn').onclick = () => {
            createNewRole(chatId);
        };
    }

    // Загружаем права базовой роли
    function loadBaseRolePermissions(roleId, chatId) {
        const permissionsContainer = document.getElementById('newRolePermissions');
        
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles/${roleId}`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        .then(response => response.json())
        .then(async role => {
            // Отображаем права с чекбоксами
            permissionsContainer.innerHTML = '';
            
            // Предполагаем, что у нас есть список всех возможных прав
            const allPermissionsResp = await fetch(`http://localhost:4003/api/v1/chats/${chatId}/permissions`);
            const allPermissions = await allPermissionsResp.json();
            
            allPermissions.forEach(permission => {
                const isChecked = role.permissions.includes(permission);
                const permissionItem = document.createElement('div');
                permissionItem.className = 'checkbox-item';
                permissionItem.innerHTML = `
                    <input type="checkbox" id="perm-${permission}" 
                        ${isChecked ? 'checked' : ''}>
                    <label for="perm-${permission}">${formatPermissionName(permission)}</label>
                `;
                permissionsContainer.appendChild(permissionItem);
            });
        });
    }

    // Создаем новую роль
    function createNewRole(chatId) {
        const baseRoleId = document.getElementById('baseRoleSelect').value;
        const name = document.getElementById('newRoleName').value;
        const color = document.getElementById('newRoleColor').value;
        const description = document.getElementById('newRoleDescription').value;
        
        // Собираем выбранные права
        const permissions = [];
        const checkboxes = document.querySelectorAll('#newRolePermissions input[type="checkbox"]');
        
        checkboxes.forEach(checkbox => {
            if (checkbox.checked) {
                permissions.push(checkbox.id.replace('perm-', ''));
            }
        });
        
        // Формируем JSON для отправки
        const roleData = {
            base_role_id: baseRoleId,
            name: name,
            color: color,
            description: description,
            permissions: permissions,
            chat_id: chatId
        };
        
        // Отправляем запрос на сервер
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(roleData)
        })
        .then(response => {
            if (response.ok) {
                alert('Роль успешно создана!');
                document.getElementById('createRoleModal').style.display = 'none';
                loadRoles(chatId); // Обновляем список ролей
            } else {
                throw new Error('Ошибка создания роли');
            }
        })
        .catch(error => {
            console.error('Ошибка:', error);
            alert('Не удалось создать роль');
        });
    }

    // Добавляем обработчик кнопки создания роли
    document.getElementById('createRoleBtn').addEventListener('click', function() {
        showCreateRoleModal(chatId);
    });

    // Редактирование никнейма пользователя
    function editUserNickname(userId, chatId, currentNickname) {
        const modal = document.getElementById('editNicknameModal');
        const nicknameInput = document.getElementById('nicknameInput');
        
        // Заполняем текущим никнеймом
        nicknameInput.value = currentNickname;
        
        // Показываем модальное окно
        modal.style.display = 'flex';
        
        // Обработчики кнопок
        document.getElementById('cancelNicknameEdit').onclick = () => {
            modal.style.display = 'none';
        };
        
        document.getElementById('saveNicknameEdit').onclick = () => {
            const newNickname = nicknameInput.value.trim();
            if (newNickname && newNickname !== currentNickname) {
                updateUserNickname(userId, chatId, newNickname);
            }
            modal.style.display = 'none';
        };
    }

    // Обновление никнейма на сервере
    function updateUserNickname(userId, chatId, newNickname) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/users/${userId}/nickname`, {
            method: 'PATCH',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                nickname: newNickname,
                chat_id: chatId,
                user_id: userId
            })
        })
        .then(response => {
            if (response.ok) {
                // Обновляем список участников
                loadParticipantsWithRoles(chatId);
            } else {
                alert('Ошибка изменения никнейма');
            }
        })
        .catch(error => {
            console.error('Ошибка:', error);
            alert('Не удалось изменить никнейм');
        });
    }
});