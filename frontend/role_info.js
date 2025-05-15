document.addEventListener('DOMContentLoaded', function() {
    const urlParams = new URLSearchParams(window.location.search);
    const roleId = urlParams.get('roleId');
    const chatId = urlParams.get('chatId');
    
    // Элементы формы
    const roleNameInput = document.getElementById('roleName');
    const roleColorInput = document.getElementById('roleColor');
    const roleDescriptionInput = document.getElementById('roleDescription');
    const permissionsList = document.getElementById('permissionsList');
    const saveRoleBtn = document.getElementById('saveRoleBtn');
    const deleteRoleBtn = document.getElementById('deleteRoleBtn');
    
    let allPermissions = [];
    let currentPermissions = [];
    
    // Загрузка информации о роли
    loadRoleInfo(roleId, chatId);
    
    // Обработчики событий
    saveRoleBtn.addEventListener('click', function() {
        saveRoleChanges(roleId, chatId);
    });
    
    deleteRoleBtn.addEventListener('click', function() {
        deleteRole(roleId, chatId);
    });
    
    function loadRoleInfo(roleId, chatId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles/${roleId}`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => response.json())
        .then(role => {
            console.log(role);
            roleNameInput.value = role.name;
            roleColorInput.value = role.color;
            roleDescriptionInput.value = role.description;
            currentPermissions = role.permissions;
            
            // Загрузка всех возможных прав
            loadAllPermissions(chatId);
        })
        .catch(error => {
            console.error('Ошибка загрузки информации о роли:', error);
        });
    }
    
    function loadAllPermissions(chatId) {
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/permissions`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        })
        .then(response => response.json())
        .then(permissions => {
            allPermissions = permissions;
            renderPermissions();
        });
    }
    
    function renderPermissions() {
        permissionsList.innerHTML = '';
        
        allPermissions.forEach(permission => {
            const permissionItem = document.createElement('div');
            permissionItem.className = 'checkbox-item';
            permissionItem.innerHTML = `
                <input type="checkbox" 
                       id="perm-${permission}" chat_info?chatId=c83328d8-dcdc-4f5d-85ff-646bc1b8ef19
                       ${currentPermissions.includes(permission) ? 'checked' : ''}>
                <label for="perm-${permission}">${formatPermissionName(permission)}</label>
            `;
            permissionsList.appendChild(permissionItem);
        });
    }
    
    function formatPermissionName(permission) {
        return permission
            .split('_')
            .map(word => word.charAt(0) + word.slice(1).toLowerCase())
            .join(' ');
    }
    
    function saveRoleChanges(roleId, chatId) {
        const updatedRole = {
            name: roleNameInput.value,
            color: roleColorInput.value,
            description: roleDescriptionInput.value,
            chat_id: chatId,
            role_id: roleId,
            permissions: getSelectedPermissions()
        };
        
        fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles/${roleId}`, {
            method: 'PUT',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(updatedRole)
        })
        .then(response => {
            if (response.ok) {
                window.close();
            }
        });
    }
    
    function getSelectedPermissions() {
        const selected = [];
        const checkboxes = permissionsList.querySelectorAll('input[type="checkbox"]');
        
        checkboxes.forEach(checkbox => {
            if (checkbox.checked) {
                selected.push(checkbox.id.replace('perm-', ''));
            }
        });
        
        return selected;
    }
    
    function deleteRole(roleId, chatId) {
        if (confirm('Вы уверены, что хотите удалить эту роль?')) {
            fetch(`http://localhost:4003/api/v1/chats/${chatId}/roles/${roleId}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            })
            .then(response => {
                if (response.ok) {
                    window.close();
                }
            });
        }
    }
});