document.addEventListener('DOMContentLoaded', () => {
    // === Элементы авторизации ===
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

    // === Элементы привычек ===
    const habitsList = document.getElementById('habitsList');
    const addHabitBtn = document.getElementById('addHabitBtn');

    // === Элементы модального окна добавления привычки ===
    const addHabitModal = document.getElementById('addHabitModal');
    const closeAddHabit = document.getElementById('closeAddHabit');
    const addHabitForm = document.getElementById('addHabitForm');
    const habitError = document.getElementById('addHabitError');
    const colorPicker = document.getElementById('habitColor');
    const colorPreview = document.getElementById('colorPreview');

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

    // === Загрузка привычек ===
    async function loadHabits() {
        const token = localStorage.getItem('token');
        if (!token) return;
        try {
            const res = await fetch('/api/habits', {
                headers: { 'Authorization': `Bearer ${token}` }
            });
            if (!res.ok) {
                if (res.status === 401) updateUI(false);
                return;
            }
            const habits = await res.json();
            renderHabits(habits);
        } catch (e) {
            console.error('Failed to load habits:', e);
            habitsList.innerHTML = '<p class="placeholder">Ошибка загрузки данных</p>';
        }
    }

    // === Отрисовка привычек ===
    function renderHabits(habits) {
        if (!habits || habits.length === 0) {
            habitsList.innerHTML = '<p class="placeholder">Нет привычек. Добавь первую!</p>';
            return;
        }
        habitsList.innerHTML = habits.map(habit => {
            const isDone = habit.is_completed_today;
            const btnClass = isDone ? 'done' : '';
            const btnText = isDone ? '✓' : '○';
            const safeTitle = habit.title.replace(/'/g, "\\'");
            return `
            <div class="habit-item" style="border-left: 4px solid ${habit.color || '#4caf50'}">
                <div>
                    <strong>${escapeHtml(habit.title)}</strong>
                    ${habit.description ? `<br><small style="color:var(--text-muted)">${escapeHtml(habit.description)}</small>` : ''}
                </div>
                <div style="display:flex; gap:8px; align-items:center;">
                    <button class="btn-open-cal" 
                            onclick="openCalendar(${habit.id}, '${safeTitle}')"
                            title="Календарь">📅</button>
                    <button class="btn-toggle ${btnClass}" 
                            data-id="${habit.id}" 
                            data-done="${isDone}">${btnText}</button>
                    <button class="btn-delete" onclick="deleteHabit(${habit.id})" title="Удалить">🗑️</button>
                </div>
            </div>`;
        }).join('');

        document.querySelectorAll('.btn-toggle').forEach(btn => {
            btn.addEventListener('click', toggleHabit);
        });
        document.querySelectorAll('.btn-open-cal').forEach(btn => {
            btn.addEventListener('click', (e) => e.stopPropagation());
        });
    }

    // === Переключение статуса привычки ===
    async function toggleHabit(e) {
        const btn = e.currentTarget;
        const habitId = btn.dataset.id;
        const isDone = btn.dataset.done === 'true';
        const token = localStorage.getItem('token');
        const newDone = !isDone;
        btn.dataset.done = newDone.toString();
        btn.classList.toggle('done');
        btn.textContent = newDone ? '✓' : '○';
        try {
            const res = await fetch('/api/habits/toggle', {
                method: 'PATCH',
                headers: {
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ habit_id: parseInt(habitId) })
            });
            if (!res.ok) {
                const errorText = await res.text();
                console.error('Failed to toggle habit:', res.status, errorText);
                btn.dataset.done = isDone.toString();
                btn.classList.toggle('done');
                btn.textContent = isDone ? '✓' : '○';
            }
        } catch (err) {
            console.error('Failed to toggle habit:', err);
            btn.dataset.done = isDone.toString();
            btn.classList.toggle('done');
            btn.textContent = isDone ? '✓' : '○';
        }
    }

    // === Модальное окно: открытие/закрытие ===
    function openAddHabitModal() {
        if (!addHabitModal) return;
        addHabitForm.reset();
        habitError.textContent = '';
        if (colorPicker && colorPreview) colorPreview.textContent = colorPicker.value;
        addHabitModal.classList.remove('hidden');
    }
    function closeAddHabitModal() {
        if (addHabitModal) addHabitModal.classList.add('hidden');
    }

    // === Обработчики модального окна ===
    if (closeAddHabit) closeAddHabit.addEventListener('click', closeAddHabitModal);
    if (addHabitModal) {
        addHabitModal.addEventListener('click', (e) => {
            if (e.target === addHabitModal) closeAddHabitModal();
        });
    }
    if (colorPicker && colorPreview) {
        colorPicker.addEventListener('input', (e) => {
            colorPreview.textContent = e.target.value;
        });
    }

    // === Отправка формы добавления привычки ===
    if (addHabitForm) {
        addHabitForm.addEventListener('submit', async (e) => {
            e.preventDefault();
            habitError.textContent = '';
            const title = document.getElementById('habitTitle').value.trim();
            const description = document.getElementById('habitDesc').value.trim();
            const color = colorPicker ? colorPicker.value : '#4caf50';
            const token = localStorage.getItem('token');
            if (!title) {
                habitError.textContent = 'Название обязательно';
                return;
            }
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
                    closeAddHabitModal();
                    if (window.refreshHabitsList) window.refreshHabitsList();
                } else {
                    const errText = await res.text();
                    habitError.textContent = errText || 'Не удалось создать';
                }
            } catch (err) {
                console.error(err);
                habitError.textContent = 'Ошибка сети';
            }
        });
    }

    // === Удаление привычки ===
    async function deleteHabit(habitId) {
        if (!confirm('Удалить привычку? История выполнения будет потеряна.')) return;
        const token = localStorage.getItem('token');
        try {
            const res = await fetch('/api/habits/delete', {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' },
                body: JSON.stringify({ habit_id: habitId })
            });
            if (res.ok) {
                if (window.refreshHabitsList) window.refreshHabitsList();
                else location.reload();
            } else {
                const errorText = await res.text();
                alert(errorText || 'Ошибка при удалении');
            }
        } catch(e) { alert('Ошибка сети'); }
    }

    // === Вспомогательные функции ===
    function escapeHtml(text) {
        if (!text) return '';
        const map = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#039;' };
        return text.replace(/[&<>"']/g, m => map[m]);
    }
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

    // === Обработчик кнопки "+ Добавить" ===
    if (addHabitBtn) {
        addHabitBtn.addEventListener('click', openAddHabitModal);
    }

    window.openAddHabitModal = openAddHabitModal;
    window.refreshHabitsList = loadHabits;
    window.deleteHabit = deleteHabit;
});
