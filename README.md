# RabbitMQ + Go: базовый пример

Два самостоятельных executable-приложения: `publisher/main.go` отправляет
сообщение раз в секунду, `consumer/main.go` сразу получает его и пишет в лог.
Каждое сообщение содержит порядковый номер и время отправки в UTC.

```text
publisher (demo.message) → exchange demo.topic (topic) → binding demo.* → очередь hello → consumer → лог + ack
```

## Запуск одной командой

Нужен запущенный Docker с Docker Compose v2; локальный Go для этого не нужен.
Из этой папки:

```sh
docker compose up --build
```

Compose поднимает RabbitMQ и запускает приложения после успешного healthcheck.
Go-приложения используют `network_mode: "service:rabbitmq"`: у них общая
сетевая среда с брокером, поэтому фиксированный `localhost:5672` работает
и внутри контейнеров. Устаревшие аргументы командной строки из Compose удалены.
Описание режима — [Docker Compose network_mode](https://docs.docker.com/reference/compose-file/services/#network_mode).
Первый запуск скачивает образы и Go-зависимость. Логи всех трёх сервисов идут
в текущий терминал. Тело сообщения в строках `sent` и `received` совпадает:

```text
publisher | publisher 2026/10/10 10:00:01 sent exchange=demo.topic routing_key=demo.message body="Hello RabbitMQ! #1 at 2026-10-10T10:00:01Z"
consumer  | consumer 2026/10/10 10:00:01 received queue=hello exchange=demo.topic routing_key=demo.message body="Hello RabbitMQ! #1 at 2026-10-10T10:00:01Z"
```

Это иллюстрация формата; время и префиксы контейнеров зависят от запуска.

Веб-интерфейс: [localhost:15672](http://localhost:15672), логин / пароль —
`demo` / `demo`. Очередь `hello` видна на вкладке **Queues and Streams**.
AMQP доступен на `localhost:5672`.

Только логи приложений (в другом терминале):

```sh
docker compose logs -f publisher consumer
```

Интервал отправки — одна секунда, задан в `time.NewTicker(time.Second)` в паблишере.

`Ctrl+C` останавливает сервисы. Удалить контейнеры и сеть:

```sh
docker compose down
```

Данные брокера остаются в Docker volume. Полностью сбросить демо вместе
с очередью и сообщениями:

```sh
docker compose down -v
```

## Если Docker не знает команду compose на macOS

На машине Виктора Docker установлен через Homebrew, runtime — Colima.
Для установки недостающих компонентов:

```sh
brew install docker-compose docker-buildx docker-credential-helper
```

В существующем `~/.docker/config.json` должны быть поля ниже; остальные
поля конфигурации сохраняются:

```json
{
  "cliPluginsExtraDirs": ["/opt/homebrew/lib/docker/cli-plugins"],
  "credsStore": "osxkeychain"
}
```

`cliPluginsExtraDirs` подключает плагины Homebrew —
[инструкция Homebrew](https://formulae.brew.sh/formula/docker-compose).
`osxkeychain` использует установленный `docker-credential-osxkeychain` для
хранения учётных данных в macOS Keychain —
[документация Docker credential helpers](https://github.com/docker/docker-credential-helpers#with-the-docker-engine).

Запуск runtime и проверка:

```sh
colima start
docker context use colima
docker compose version
docker info
```

Colima предоставляет локальный Docker daemon —
[инструкция Colima](https://github.com/abiosoft/colima#docker).
В этом рабочем окружении контекст `colima` уже создан и подключён к
`unix:///Users/no/.colima/default/docker.sock`.

## Сборка и запуск Go отдельно

Нужен Go 1.21 или новее с включённым автоматическим выбором toolchain
(стандартное поведение): версия компилятора Go 1.27.2 задана в `go.mod` и при
необходимости скачивается автоматически. Сначала поднять только брокер:

```sh
docker compose up -d rabbitmq
go build -o bin/publisher ./publisher
go build -o bin/consumer ./consumer
```

Когда брокер готов, сначала запустить консьюмер и дождаться лога `connected`:

```sh
./bin/consumer
```

Затем в другом терминале запустить паблишер:

```sh
./bin/publisher
```

Собранные executable сохраняются в локальную папку `bin/`, исключённую из Git.
После клонирования репозитория собрать их командами выше для своей ОС и архитектуры.

Настройки заданы прямо в коде: `amqp://demo:demo@localhost:5672/`,
exchange `demo.topic`, routing key `demo.message`, очередь `hello`, binding `demo.*`.
`Ctrl+C` завершает программу и закрывает AMQP-соединение.

## Что показывает код

- Соединение → канал → объявление exchange и очереди → binding → публикация / потребление.
- Publisher объявляет durable exchange `demo.topic` типа `topic`
  и публикует в него с routing key `demo.message`.
- Consumer объявляет тот же exchange, durable-очередь `hello`
  и связывает их через `QueueBind` с шаблоном `demo.*`.
- В topic binding `*` заменяет ровно одно слово между точками:
  `demo.message` подходит под `demo.*`, а `demo.message.extra` — нет.
- При первом запуске консьюмер создаёт очередь и binding. Сообщения, отправленные
  до их создания, в этом примере будут отброшены, если нет других подходящих bindings.
- Консьюмер получает сообщения от брокера по мере поступления, без опроса по таймеру.
  После лога отправляет ручной `Ack`.
- Паблишер помечает сообщения как persistent. Publisher confirms в этом базовом
  примере нет: `sent` означает успешный вызов публикации на клиенте,
  а `received` — фактическое получение консьюмером.

## Версии и первоисточники

Проверено 2026-10-10:

- RabbitMQ **4.3.6**, последний стабильный релиз на дату проверки:
  [таблица релизов](https://www.rabbitmq.com/release-information),
  [официальный Docker-тег `4.3.6-management`](https://github.com/docker-library/official-images/blob/master/library/rabbitmq).
  Версия закреплена в `compose.yaml` для повторяемого демо.
- Клиент [`rabbitmq/amqp091-go` v1.15.0](https://github.com/rabbitmq/amqp091-go/releases/tag/v1.15.0),
  протокол AMQP 0-9-1.
- [Hello World на Go](https://www.rabbitmq.com/tutorials/tutorial-one-go) —
  подключение, каналы и объявление очереди.
- [Topics на Go](https://www.rabbitmq.com/tutorials/tutorial-five-go) —
  именованный topic exchange, `QueueBind`, routing key и шаблон `demo.*`;
  топология примера обновлена 2026-10-14.
- [Work Queues на Go](https://www.rabbitmq.com/tutorials/tutorial-two-go) —
  ручные подтверждения, durable / persistent и prefetch.

## Проверка примера

2026-10-14: полный запуск Compose в Colima прошёл успешно. RabbitMQ 4.3.6
healthy, оба Go-контейнера работают. Проверены 15 сообщений с совпадающими
телами в логах publisher и consumer, binding `demo.*`, один консьюмер очереди
`hello`; `messages_ready=0`, `messages_unacknowledged=0`. Management API
доступен на `localhost:15672`. [Логи проверки](../tmp/rmq-base-example-smoke-2026-10-14.log).

2026-10-14: после перехода на topic exchange оба executable пересобраны
для macOS ARM64; сборка и `go vet ./...` прошли успешно. Живой обмен
не проверен — брокер на `localhost:5672` недоступен.

2026-10-10: оба executable собраны Go 1.27.2 для macOS ARM64, проверены запуск
справки и отказ при нулевом интервале. Статическая сборка для Linux ARM64
(как в Dockerfile), `go vet ./...` и `docker compose config` с интервалами
`1s` и `500ms` прошли успешно.

На проверках до настройки окружения 2026-10-14 полный запуск Compose и обмен
с реальным брокером не проверялись: Docker daemon был недоступен.
