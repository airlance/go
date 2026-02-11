# Команда Update для автоматического обновления IP

## Описание

Команда `update` предназначена для автоматического обновления данных sender score для самого старого IP-адреса в базе данных.

## Что делает команда

1. Получает самый старый IP из базы данных (по полю `updated_at`)
2. Запрашивает актуальные данные с senderscore.org
3. Парсит полученный отчет
4. Обновляет данные в базе:
    - Обновляет score, spam_trap, blocklists, complaints
    - Добавляет или обновляет историю
    - Обновляет счетчики в группах
5. Логирует результат операции

## Использование

### Ручной запуск

```bash
# Запуск через go run
go run main.go update

# Или через Makefile
make update

# Или через скомпилированный бинарник
./bin/api update
```

### Автоматический запуск через cron

#### 1. Создайте лог-файл
```bash
sudo touch /var/log/senderscore-update.log
sudo chmod 666 /var/log/senderscore-update.log
```

#### 2. Настройте crontab
```bash
crontab -e
```

#### 3. Добавьте задание (примеры)

**Каждые 6 часов:**
```cron
0 */6 * * * cd /path/to/senderscore/api && /usr/local/go/bin/go run main.go update >> /var/log/senderscore-update.log 2>&1
```

**Каждый час:**
```cron
0 * * * * cd /path/to/senderscore/api && /usr/local/go/bin/go run main.go update >> /var/log/senderscore-update.log 2>&1
```

**Каждые 30 минут:**
```cron
*/30 * * * * cd /path/to/senderscore/api && /usr/local/go/bin/go run main.go update >> /var/log/senderscore-update.log 2>&1
```

**Раз в день в 3:00:**
```cron
0 3 * * * cd /path/to/senderscore/api && /usr/local/go/bin/go run main.go update >> /var/log/senderscore-update.log 2>&1
```

#### 4. С использованием скомпилированного бинарника (рекомендуется)

Сначала соберите проект:
```bash
make build
```

Затем в crontab:
```cron
0 */6 * * * cd /path/to/senderscore/api && ./bin/api update >> /var/log/senderscore-update.log 2>&1
```

## Переменные окружения

Команда использует те же переменные окружения, что и основное приложение:

```env
DB_DSN=postgresql://user:password@localhost:5432/dbname?sslmode=disable
LOG_LEVEL=info
```

Убедитесь, что файл `.env` находится в корне проекта или переменные установлены в системе.

## Логирование

Команда использует structured logging через logrus. Примеры логов:

```json
{
  "ip": "1.2.3.4",
  "level": "info",
  "msg": "Processing oldest IP",
  "time": "2025-02-12T10:00:00Z",
  "updated_at": "12.02.2025 10:00:00"
}

{
  "blocklists": "0",
  "complaints": "0.00%",
  "history_cnt": 30,
  "ip": "1.2.3.4",
  "level": "info",
  "msg": "Parsed sender score data",
  "score": 95,
  "spam_traps": 0,
  "time": "2025-02-12T10:00:05Z"
}

{
  "history_added": 5,
  "history_updated": 25,
  "ip": "1.2.3.4",
  "ip_created": false,
  "level": "info",
  "msg": "Successfully updated IP score",
  "time": "2025-02-12T10:00:10Z"
}
```

## Мониторинг

### Просмотр логов
```bash
tail -f /var/log/senderscore-update.log
```

### Проверка статуса cron
```bash
# Список задач cron
crontab -l

# Логи cron (зависит от системы)
grep CRON /var/log/syslog
# или
journalctl -u cron
```

### Ротация логов

Создайте файл `/etc/logrotate.d/senderscore`:

```
/var/log/senderscore-update.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0666 root root
}
```

## Обработка ошибок

Команда логирует все ошибки, но не останавливает выполнение при некритических ошибках:

- Если не удалось получить отчет с senderscore.org - команда завершится с ошибкой
- Если не удалось обновить счетчики группы - команда продолжит работу, но выдаст warning

## Рекомендации

1. **Частота обновления**: Рекомендуется обновлять каждые 6-12 часов, чтобы не создавать излишнюю нагрузку на senderscore.org
2. **Мониторинг**: Настройте алерты при появлении ошибок в логах
3. **Бэкап**: Регулярно делайте бэкапы базы данных
4. **Production**: Используйте скомпилированный бинарник вместо `go run` для лучшей производительности

## Troubleshooting

**Проблема**: Cron не выполняется
- Проверьте синтаксис crontab: `crontab -l`
- Проверьте права доступа к логу
- Проверьте пути в команде cron
- Проверьте логи cron: `grep CRON /var/log/syslog`

**Проблема**: Ошибка подключения к БД
- Проверьте переменную DB_DSN
- Убедитесь, что файл .env в правильной директории
- Проверьте доступность базы данных

**Проблема**: Ошибка при получении данных с senderscore.org
- Проверьте интернет-соединение
- Проверьте, не заблокирован ли IP
- Возможно, сайт временно недоступен