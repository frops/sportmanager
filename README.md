# Футбольный менеджер

Сервис для организации футбольных матчей с возможностью бронирования мест и автоматическим распределением игроков по командам.

## Функциональность

- Просмотр расписания матчей
- Бронирование места на матч
- Автоматическое распределение игроков по командам
- Интеграция с Telegram ботом
- Отображение места проведения на карте

## Технологии

- Frontend: React, TypeScript, Material-UI
- Backend: Go, Gin
- База данных: PostgreSQL
- Дополнительно: Telegram Bot API, Google Maps API

## Установка и запуск

### Предварительные требования

- Docker и Docker Compose
- Telegram Bot Token
- Google Maps API Key

### Настройка переменных окружения

1. Создайте файл `.env` в корневой директории:

```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
REACT_APP_GOOGLE_MAPS_API_KEY=your_google_maps_api_key
```

### Запуск

1. Запустите все сервисы:

```bash
docker-compose up --build
```

2. Приложение будет доступно по адресам:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## API Endpoints

### Матчи

- `GET /api/matches` - получить список матчей
- `POST /api/matches` - создать новый матч
- `POST /api/matches/:id/join` - присоединиться к матчу
- `DELETE /api/matches/:id/leave` - покинуть матч

### Игроки

- `GET /api/players` - получить список игроков
- `POST /api/players` - создать нового игрока

## Telegram Bot

Бот поддерживает следующие команды:

- `/start` - начало работы с ботом
- `/help` - список доступных команд
- `/matches` - показать ближайшие матчи
- `/join [ID]` - записаться на матч
- `/leave [ID]` - отписаться от матча 