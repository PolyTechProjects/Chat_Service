document.getElementById('loginForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const login = document.getElementById('login-login').value;
    const password = document.getElementById('login-password').value;
    const messageEl = document.getElementById('login-message');
    
    try {
        const response = await fetch('http://localhost:4001/api/v1/auth/login', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ login, password })
        });
        const token = await response.json();
        console.log(response);
        localStorage.setItem('token', token.access_token);
        messageEl.textContent = 'Вход выполнен успешно!';
        messageEl.className = 'message success';
        
        // Перенаправление на главную страницу через 1 секунду
        setTimeout(() => {
            window.location.pathname = '/';
        }, 1000);
    } catch (error) {
        messageEl.textContent = error.message || 'Ошибка входа';
        messageEl.className = 'message error';
    }
});
