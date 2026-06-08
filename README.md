# eDoc Blockchain

Курсовой проект: безопасное хранение и передача файлов с проверкой SHA-256.

## Backend

Требуется PostgreSQL. Перед первым запуском можно выполнить:

```sql
\i database/schema.sql
```

Если уже была старая БД с таблицами `users`, `documents`, `transfers`, используйте:

```sql
\i database/migration_from_old_schema.sql
```

Backend также выполняет совместимую инициализацию схемы при старте.

Переменные окружения:

```powershell
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5433/blockchain?sslmode=disable"
$env:API_PORT="8080"
$env:P2P_ADDR="localhost:8001"
$env:UPLOAD_DIR="./uploads"
```

Запуск:

```powershell
cd backend
go run ./cmd/server
```

## Frontend

Клиент: .NET MAUI. Для сборки нужны MAUI workloads Visual Studio/.NET.

```powershell
cd frontend/BlockchainClient/BlockchainClient
dotnet build -f net9.0-windows10.0.19041.0
```

## Основная логика

- Регистрация: имя, возраст, телефон, пароль.
- ID пользователя генерируется UUID в PostgreSQL.
- Чужие профили в каталоге показывают только ID.
- Файл сохраняется с SHA-256, датой, временем, именем загрузившего и размером.
- Отправитель выбирает получателя и файл, затем видит SHA-256 файла.
- Получатель принимает файл только после ввода SHA-256.
- При отказе статус записывается как `declined`, в клиентской истории отображается `Файл не принят`.
- Полученные файлы отображаются только после успешной проверки хеша.

## Аватары

Поддерживаются JPG и PNG до 5 MB. Backend проверяет формат, разрешение от 128x128 до 2048x2048 и разумные пропорции.
