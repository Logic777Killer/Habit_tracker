document.addEventListener('DOMContentLoaded', () => {
    const loginBtn = document.getElementById('loginBtn');
    const registerBtn = document.getElementById('registerBtn');
    
    registerBtn.addEventListener('click', async () => {
        const username = prompt("Введите имя пользователя:");
        const email = prompt("Введите email:");
        const password = prompt("Введите пароль:");

        if (!username || !email || !password) return;

        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, email, password })
            });

            if (response.ok) {
                alert('Регистрация успешна! Теперь войдите.');
            } else {
                const errText = await response.text();
                alert('Ошибка: ' + errText);
            }
        } catch (e) {
            console.error(e);
            alert('Ошибка сети');
        }
    });

    loginBtn.addEventListener('click', async () => {
        const email = prompt("Введите email:");
        const password = prompt("Введите пароль:");

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (response.ok) {
                const data = await response.json();
                localStorage.setItem('token', data.token);
                localStorage.setItem('role', data.role);
                alert('Вход выполнен! Токен сохранен.');
                console.log("Token:", data.token);
            } else {
                alert('Неверный email или пароль');
            }
        } catch (e) {
            console.error(e);
            alert('Ошибка сети');
        }
    });
});