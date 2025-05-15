document.getElementById('registerForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const username = document.getElementById('reg-username').value;
    const firstname = document.getElementById('reg-firstname').value;
    const lastname = document.getElementById('reg-lastname').value;
    const login = document.getElementById('reg-login').value;
    const password = document.getElementById('reg-password').value;
    const messageEl = document.getElementById('register-message');
    
    try {
        await fetch('http://localhost:4001/api/v1/auth/register', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                username,
                firstname,
                lastname,
                login,
                password
            })
        });
        
        messageEl.textContent = 'Регистрация успешна! Перенаправляем на страницу входа...';
        messageEl.className = 'message success';
        
        // Перенаправление на страницу входа через 2 секунды
        setTimeout(() => {
            window.location.href = '/login.html';
        }, 2000);
    } catch (error) {
        messageEl.textContent = error.message || 'Ошибка регистрации';
        messageEl.className = 'message error';
    }
});
