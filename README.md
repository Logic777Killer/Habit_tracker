# Habit Tracker

Трекер формирования привычек, разработанный в рамках курсового проекта.

## Стек технологий
- **Backend**: Go (Golang)
- **Frontend**: HTML, CSS, JavaScript
- **Database**: PostgreSQL
- **Deployment**: Docker, Render

## Структура проекта
- `cmd/server`: Точка входа
- `internal`: Логика приложения
- `web`: Фронтенд часть

## Запуск локально
1. Установите Go и PostgreSQL.
2. Создайте базу данных.
3. Задайте переменные окружения для подключения к БД и `JWT_SECRET`.
4. Запустите `go run cmd/server/main.go`.