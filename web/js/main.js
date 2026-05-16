document.addEventListener('DOMContentLoaded', () => {
    console.log('Habit Tracker App Loaded');

    const habitsList = document.getElementById('habitsList');

    async function loadHabits() {
        try {

            habitsList.innerHTML = '<p>Пока нет привычек. Добавьте первую!</p>';
        } catch (error) {
            console.error('Error loading habits:', error);
            habitsList.innerHTML = '<p>Ошибка загрузки данных.</p>';
        }
    }

    loadHabits();

    document.getElementById('addHabitBtn').addEventListener('click', () => {
        alert('Функция добавления привычки будет реализована на следующем этапе');
    });
});