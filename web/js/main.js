document.addEventListener('DOMContentLoaded', () => {
    // Элементы DOM
    const authSection = document.getElementById('authSection');
    const appSection = document.getElementById('appSection');
    const authNav = document.getElementById('authNav');
    const userNameDisplay = document.getElementById('userNameDisplay');
    const logoutBtn = document.getElementById('logoutBtn');
    const tabBtns = document.querySelectorAll('.tab-btn');
    const loginForm = document.getElementById('loginForm');
    const registerForm = document.getElementById('registerForm');
    const loginError = document.getElementById('loginError');
    const regError = document.getElementById('regError');

    // Проверка авторизации при загрузке
    checkAuth();

    // Переключение вкладок Вход/Регистрация
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            tabBtns.forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            const tab = btn.dataset.tab;

            if (tab === 'login') {
                loginForm.classList.remove('hidden');
                registerForm.classList.add('hidden');
            } else {
                loginForm.classList.add('hidden');
                registerForm.classList.remove('hidden');
            }
            clearErrors();
        });
    });

    // Обработка входа
    loginForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        clearErrors();

        const email = document.getElementById('loginEmail').value.trim();
        const password = document.getElementById('loginPassword').value;

        try {
            const res = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (!res.ok) {
                const errText = await res.text();
                loginError.textContent = errText || 'Неверный email или пароль';
                return;
            }

            const data = await res.json();
            localStorage.setItem('token', data.token);
            localStorage.setItem('role', data.role);
            localStorage.setItem('username', data.username);
            updateUI(true);
        } catch (err) {
            loginError.textContent = 'Ошибка сети. Проверь подключение к интернету.';
        }
    });

    // Обработка регистрации
    registerForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        clearErrors();

        const username = document.getElementById('regUsername').value.trim();
        const email = document.getElementById('regEmail').value.trim();
        const password = document.getElementById('regPassword').value;

        try {
            // 1. Регистрация
            const regRes = await fetch('/api/register', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, email, password })
            });

            if (!regRes.ok) {
                const errText = await regRes.text();
                regError.textContent = errText || 'Ошибка регистрации. Возможно, email уже занят.';
                return;
            }

            // 2. Автоматический вход
            const loginRes = await fetch('/api/login', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email, password })
            });

            if (loginRes.ok) {
                const data = await loginRes.json();
                localStorage.setItem('token', data.token);
                localStorage.setItem('role', data.role);
                localStorage.setItem('username', data.username);
                updateUI(true);
            } else {
                regError.textContent = 'Аккаунт создан, но войти не удалось. Попробуй войти вручную.';
            }
        } catch (err) {
            regError.textContent = 'Ошибка сети. Попробуй позже.';
        }
    });

    // Выход
    logoutBtn.addEventListener('click', () => {
        localStorage.removeItem('token');
        localStorage.removeItem('role');
        localStorage.removeItem('username');
        updateUI(false);
    });

    // Вспомогательные функции
    function checkAuth() {
        const token = localStorage.getItem('token');
        updateUI(!!token);
    }

    function updateUI(isLoggedIn) {
        if (isLoggedIn) {
            authSection.classList.add('hidden');
            appSection.classList.remove('hidden');
            authNav.classList.remove('hidden');
            userNameDisplay.textContent = localStorage.getItem('username') || 'Пользователь';
        } else {
            authSection.classList.remove('hidden');
            appSection.classList.add('hidden');
            authNav.classList.add('hidden');
            loginForm.reset();
            registerForm.reset();
            clearErrors();
        }
    }

    function clearErrors() {
        loginError.textContent = '';
        regError.textContent = '';
    }
});