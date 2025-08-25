# Data Vault Server

Серверная часть системы Data Vault - gRPC сервер для безопасного хранения и управления пользовательскими данными.

## Настройка

### Переменные окружения

Создайте файл `.env` или установите следующие переменные:

```bash
# Конфигурация сервера
SERVER_ADDRESS=localhost:8080
LOG_LEVEL=info

# База данных
DATABASE_DSN=postgres://user:password@localhost:5432/datavault?sslmode=disable

# Безопасность
JWT_SECRET=your-secret-key-here
ENCRYPTION_KEY=32-byte-encryption-key
```

### База данных

1. Установите PostgreSQL 14+
2. Создайте базу данных:
   ```sql
   CREATE DATABASE datavault;
   ```
3. Сервер автоматически создаст необходимые таблицы при первом запуске

## Запуск

### Разработка

```bash
go mod tidy
go run cmd/main.go
```

### Производство

```bash
go build -o server cmd/main.go
./server
```

## API

Сервер предоставляет следующие gRPC методы:

- `Register(RegisterRequest) RegisterResponse` - регистрация пользователя
- `Login(LoginRequest) LoginResponse` - вход в систему
- `PostData(PostDataRequest) PostDataResponse` - сохранение данных
- `GetData(GetDataRequest) GetDataResponse` - получение данных
- `DeleteData(DeleteDataRequest) DeleteDataResponse` - удаление данных
- `Ping(PingRequest) PingResponse` - проверка состояния сервера

## Тестирование

Запуск тестов:

```bash
go test ./...
```

Запуск тестов с покрытием:

```bash
go test -cover ./...
```

## Структура проекта

```
server/
├── cmd/main.go              # Точка входа
├── internal/
│   ├── config/             # Конфигурация
│   ├── handler/            # gRPC обработчики
│   ├── models/             # Модели данных
│   ├── service/            # Бизнес-логика
│   ├── storage/            # Работа с БД
│   └── transport/          # gRPC сервер
└── proto/                  # Protobuf определения
```