let currentHabitId = null;
let currentCalMonth = new Date().getMonth() + 1;
let currentCalYear = new Date().getFullYear();
const monthNames = ["Январь","Февраль","Март","Апрель","Май","Июнь","Июль","Август","Сентябрь","Октябрь","Ноябрь","Декабрь"];

let calendarModal, closeCalendarBtn, calendarTitle, monthYearEl;
let prevMonthBtn, nextMonthBtn, calendarGrid;

document.addEventListener('DOMContentLoaded', () => {
    calendarModal = document.getElementById('calendarModal');
    closeCalendarBtn = document.getElementById('closeCalendar');
    calendarTitle = document.getElementById('calendarTitle');
    monthYearEl = document.getElementById('monthYear');
    prevMonthBtn = document.getElementById('prevMonth');
    nextMonthBtn = document.getElementById('nextMonth');
    calendarGrid = document.getElementById('calendarGrid');

    if(closeCalendarBtn) closeCalendarBtn.addEventListener('click', closeCalendar);
    if(calendarModal) calendarModal.addEventListener('click', (e) => { if(e.target === calendarModal) closeCalendar(); });
    if(prevMonthBtn) prevMonthBtn.addEventListener('click', () => changeMonth(-1));
    if(nextMonthBtn) nextMonthBtn.addEventListener('click', () => changeMonth(1));
});

function openCalendar(habitId, title) {
    if (!calendarModal) return;
    currentHabitId = habitId;
    calendarTitle.textContent = `📅 ${title}`;

    // Сбрасываем на текущий месяц при открытии
    const now = new Date();
    currentCalMonth = now.getMonth() + 1;
    currentCalYear = now.getFullYear();

    calendarModal.classList.remove('hidden');
    loadCalendarData();
}

function closeCalendar() {
    const modal = document.getElementById('calendarModal');
    if (modal) {
        modal.classList.add('hidden');
        document.querySelectorAll('.btn-open-cal').forEach(btn => {
            btn.classList.remove('active');
        });
    }
}

function changeMonth(delta) {
    currentCalMonth += delta;
    if (currentCalMonth > 12) { currentCalMonth = 1; currentCalYear++; }
    if (currentCalMonth < 1) { currentCalMonth = 12; currentCalYear--; }
    loadCalendarData();
}

async function loadCalendarData() {
    if (!currentHabitId) return;

    monthYearEl.textContent = `${monthNames[currentCalMonth-1]} ${currentCalYear}`;
    const token = localStorage.getItem('token');

    try {
        const res = await fetch(`/api/habits/logs?habit_id=${currentHabitId}&month=${currentCalMonth}&year=${currentCalYear}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const data = await res.json();
        renderCalendarGrid(data.dates || []);
    } catch (e) {
        console.error('Calendar load error:', e);
    }
}

// Отрисовка сетки
function renderCalendarGrid(completedDates) {
    if (!calendarGrid) return;
    calendarGrid.innerHTML = '';

    const firstDay = new Date(currentCalYear, currentCalMonth - 1, 1).getDay();
    const daysInMonth = new Date(currentCalYear, currentCalMonth, 0).getDate();

    const startOffset = (firstDay + 6) % 7;
    for (let i = 0; i < startOffset; i++) {
        calendarGrid.innerHTML += `<div class="calendar-day empty"></div>`;
    }

    for (let d = 1; d <= daysInMonth; d++) {
        const dateStr = `${currentCalYear}-${String(currentCalMonth).padStart(2,'0')}-${String(d).padStart(2,'0')}`;
        const isCompleted = completedDates.includes(dateStr);

        const div = document.createElement('div');
        div.className = `calendar-day ${isCompleted ? 'completed' : ''}`;
        div.textContent = d;
        div.addEventListener('click', () => toggleDay(div, dateStr));
        calendarGrid.appendChild(div);
    }
}


async function toggleDay(div, dateStr) {
    const token = localStorage.getItem('token');
    const wasCompleted = div.classList.contains('completed');

    div.classList.toggle('completed');

    try {
        const res = await fetch('/api/habits/toggle', {
            method: 'PATCH',
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ habit_id: parseInt(currentHabitId), date: dateStr })
        });

        if (!res.ok) {
            const errorText = await res.text();
            console.error('Failed to toggle calendar day:', res.status, errorText);
            div.classList.toggle('completed');
            alert(errorText || 'Ошибка при сохранении');
        } else {
            if (window.refreshHabitsList) {
                window.refreshHabitsList();
            }
        }
    } catch (err) {
        console.error(err);
        div.classList.toggle('completed');
    }
}

function toggleCalendar(habitId, title) {
    const modal = document.getElementById('calendarModal');

    if (!modal.classList.contains('hidden') && currentHabitId === habitId) {
        closeCalendar();
    } else {
        openCalendar(habitId, title);
    }
}

window.toggleCalendar = toggleCalendar;
window.openCalendar = openCalendar;
