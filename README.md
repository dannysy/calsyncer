# Синхронизатор calDAV-Todoist

Этот сервис синхронизирует события календаря (calDAV) с таск-менеджером Todoist.

## Конфигурация

Сервис можно настроить с помощью переменных окружения. Вот доступные переменные:

- **CALDAV_URL**: URL calDAV сервера
- **TODOIST_TOKEN**: Токен API Todoist

## Сборка под linux x86_64

    env GOOS=linux GOARCH=amd64 go build -o dist/calsyncer main.go

