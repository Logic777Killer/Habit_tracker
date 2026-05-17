document.addEventListener('DOMContentLoaded', () => {
    // === Элементы DOM (Auth) ===
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

    // === Элементы DOM (Habits) ===
    const habitsList = document.getElementById('habitsList');
    const addHabitBtn = document.getElementById('addHabitBtn');

    // Проверка авторизации при загрузке
    checkAuth();

    // === Переключение вкладок Вход/Регистрация ===
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

    // === Обработка входа ===
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

    // === Обработка регистрации ===
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

    // === Выход ===
    logoutBtn.addEventListener('click', () => {
        localStorage.removeItem('token');
        localStorage.removeItem('role');
        localStorage.removeItem('username');
        updateUI(false);
    });

    // === Функции для работы с привычками ===

    // Загрузка списка привычек с сервера
    async function loadHabits() {
        const token = localStorage.getItem('token');
        if (!token) return;

        try {
            const res = await fetch('/api/habits', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (!res.ok) {
                if (res.status === 401) updateUI(false); // Токен протух — разлогиниваем
                return;
            }

            const habits = await res.json();
            renderHabits(habits);
        } catch (e) {
            console.error('Failed to load habits:', e);
            habitsList.innerHTML = '<p class="placeholder">Ошибка загрузки данных</p>';
        }
    }

    // Отрисовка списка привычек в HTML
    function renderHabits(habits) {
        if (!habits || habits.length === 0) {
            habitsList.innerHTML = '<p class="placeholder">Нет привычек. Добавь первую!</p>';
            return;
        }

        habitsList.innerHTML = habits.map(habit => {
            // Определяем статус кнопки
            const isDone = habit.is_completed_today;
            const btnClass = isDone ? 'done' : '';
            const btnText = isDone ? '✓' : '○';

            return `
            <div class="habit-item" style="border-left: 4px solid ${habit.color || '#4caf50'}">
                <div>
                    <strong>${escapeHtml(habit.title)}</strong>
                    ${habit.description ? `<br><small style="color:var(--text-muted)">${escapeHtml(habit.description)}</small>` : ''}
                </div>
                <button class="btn-toggle ${btnClass}" 
                        data-id="${habit.id}" 
                        data-done="${isDone}">
                    ${btnText}
                </button>
            </div>
            `;
        }).join('');

        // Навешиваем обработчики клика на кнопки переключения
        document.querySelectorAll('.btn-toggle').forEach(btn => {
            btn.addEventListener('click', toggleHabit);
        });
    }

    // Переключение статуса "выполнено сегодня"
    async function toggleHabit(e) {
        const btn = e.currentTarget;
        const habitId = btn.dataset.id;
        const isDone = btn.dataset.done === 'true';
        const token = localStorage.getItem('token');

        // Оптимистичное обновление интерфейса (сразу меняем вид)
        const newDone = !isDone;
        btn.dataset.done = newDone.toString();
        btn.classList.toggle('done');
        btn.textContent = newDone ? '✓' : '○';

        try {
            console.log('Sending toggle:', { habit_id: parseInt(habitId), type: typeof parseInt(habitId) });
            const res = await fetch('/api/habits/toggle', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ habit_id: parseInt(habitId) })
            });

            if (!res.ok) {
                // Если сервер вернул ошибку — откатываем изменение
                btn.dataset.done = isDone.toString();
                btn.classList.toggle('done');
                btn.textContent = isDone ? '✓' : '○';
            }
        } catch (err) {
            console.error('Failed to toggle habit:', err);
            // Откат при ошибке сети
            btn.dataset.done = isDone.toString();
            btn.classList.toggle('done');
            btn.textContent = isDone ? '✓' : '○';
        }
    }

    // Добавление новой привычки
    async function addHabit() {
        const title = prompt('Название привычки:');
        if (!title) return;

        const description = prompt('Описание (необязательно):') || '';
        const color = prompt('Цвет (HEX, например #4caf50):', '#4caf50') || '#4caf50';
        const token = localStorage.getItem('token');

        try {
            const res = await fetch('/api/habits/create', {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ title, description, color })
            });

            if (res.ok) {
                loadHabits(); // Перезагружаем список после успешного создания
            } else {
                alert('Не удалось создать привычку');
            }
        } catch (e) {
            console.error(e);
            alert('Ошибка сети');
        }
    }

    // Вспомогательная функция для защиты от XSS
    function escapeHtml(text) {
        if (!text) return '';
        const map = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' };
        return text.replace(/[&<>"']/g, m => map[m]);
    }

    // === Вспомогательные функции (Auth) ===
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

            // Загружаем привычки после успешного входа
            loadHabits();
        } else {
            authSection.classList.remove('hidden');
            appSection.classList.add('hidden');
            authNav.classList.add('hidden');
            loginForm.reset();
            registerForm.reset();
            clearErrors();
            habitsList.innerHTML = '<p class="placeholder">Здесь будут твои привычки</p>';
        }
    }

    function clearErrors() {
        loginError.textContent = '';
        regError.textContent = '';
    }

    // === Обработчик кнопки "Добавить привычку" ===
    if (addHabitBtn) {
        addHabitBtn.addEventListener('click', addHabit);
    }
});