# mitm-proxy

Man-In-The-Middle сервер для проксирования http и https запросов.

## Перед началом работы

Перед первым запуском необходимо сгенерировать корневой сертификат (CA):

```bash
chmod +x scripts/gen_ca.sh
./scripts/gen_ca.sh
```

Будут созданы файлы:
- `certs/ca.key` - Приватный ключ центра сертификации
- `certs/ca.crt` - Корневой сертификат
- `certs/cert.key` - Ключ для сертификатов доменов


Корневой самоподписный сертификат необходимо добавить в систему, для **Linux/macOS:**:
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

## Использование

### Запуск напрямую:
```bash
go run cmd/proxy/main.go
```

### Команды Docker:
```bash
make docker-build  # Собрать образ
make docker-start  # Запустить контейнер
make docker-stop   # Остановить контейнер 
make docker-clean  # Удалить контейнер
```

Прокси работает на `localhost:8080`
