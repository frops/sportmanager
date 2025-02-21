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
DB_PASSWORD=your_database_password
HOST_IP=your_server_ip  # IP адрес сервера, где запущено приложение
```

### Локальный запуск

1. Запустите все сервисы:
```bash
docker compose up --build
```

2. Приложение будет доступно по адресам:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## Деплой на DigitalOcean Droplet

### Создание дроплета

1. Создайте новый дроплет в DigitalOcean:
   - Выберите Ubuntu 22.04 LTS
   - Выберите Basic plan
   - Выберите регион ближе к вашим пользователям
   - Выберите размер: минимум 1GB RAM / 1 CPU
   - Добавьте ваш SSH ключ для доступа

2. После создания дроплета, запишите его IP адрес

### Настройка дроплета

1. Подключитесь к дроплету:
```bash
ssh root@your_droplet_ip
```

2. Обновите систему и установите необходимые пакеты:
```bash
apt update && apt upgrade -y
apt install -y git curl
```

3. Установите Docker:
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh
```

4. Установите Docker Compose:
```bash
apt install -y docker-compose-plugin
```

### Деплой приложения

1. Склонируйте репозиторий:
```bash
mkdir -p /root/app && cd /root/app
git clone https://github.com/frops/sportmanager.git webservice
cd webservice
```

2. Создайте файл `.env` с необходимыми переменными окружения:
```bash
cat > .env << EOL
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
REACT_APP_GOOGLE_MAPS_API_KEY=your_google_maps_api_key
DB_PASSWORD=your_secure_password
HOST_IP=your_droplet_ip
EOL
```

3. Запустите приложение:
```bash
docker compose up -d
```

4. Приложение будет доступно по адресам:
   - Frontend: http://your_droplet_ip:3000
   - Backend API: http://your_droplet_ip:8080

### Настройка автоматического деплоя

1. Сгенерируйте SSH ключ для деплоя (если у вас его еще нет):
```bash
ssh-keygen -t rsa -b 4096 -C "deploy-key" -f ~/.ssh/deploy_key
```

2. Добавьте публичный ключ на ваш дроплет:
```bash
# Локально
cat ~/.ssh/deploy_key.pub | ssh root@your_droplet_ip "cat >> ~/.ssh/authorized_keys"
```

3. В репозитории GitHub перейдите в Settings -> Secrets and variables -> Actions:
   - Нажмите кнопку "New repository secret"
   - Добавьте следующие секреты:
     - Имя: `DROPLET_HOST`, Значение: IP адрес вашего дроплета
     - Имя: `DROPLET_USERNAME`, Значение: root
     - Имя: `DROPLET_SSH_KEY`, Значение: содержимое файла ~/.ssh/deploy_key (приватный ключ)

   Чтобы добавить приватный ключ:
   ```bash
   # Локально
   cat ~/.ssh/deploy_key
   ```
   Скопируйте ВСЁ содержимое файла (включая строки BEGIN и END) и вставьте его как значение для `DROPLET_SSH_KEY`

4. Проверьте настройку:
   - Перейдите в раздел Actions в вашем репозитории
   - Должен появиться workflow "Deploy to Server"
   - При следующем пуше в ветку `main` начнется автоматический деплой

### Полезные команды

- Просмотр логов:
```bash
docker compose logs -f
```
<!-- @import "[TOC]" {cmd="toc" depthFrom=1 depthTo=6 orderedList=false} -->


- Просмотр статуса контейнеров:
```bash
docker compose ps
```

- Просмотр использования ресурсов:
```bash
docker stats
```

- Перезапуск сервисов:
```bash
docker compose restart
```

- Обновление приложения вручную:
```bash
cd /root/app/webservice
git pull
docker compose down
docker compose up -d
```

### Рекомендации по безопасности

1. Настройте брандмауэр (UFW):
```bash
ufw allow OpenSSH
ufw allow 3000
ufw allow 8080
ufw enable
```

2. Создайте не-root пользователя (опционально):
```bash
adduser deploy
usermod -aG docker deploy
```

3. Регулярно обновляйте систему:
```bash
apt update && apt upgrade -y
```

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
