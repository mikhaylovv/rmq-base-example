# RabbitMQ + Go: базовый пример

## Что делает пример

Паблишер отправляет сообщение раз в секунду. Консьюмер получает его,
пишет в лог и подтверждает получение через `Ack`.

```text
publisher → exchange demo.topic (topic) → binding demo.* → очередь hello → consumer
```

## Быстрый запуск

Нужен работающий Docker с Docker Compose. Локальный Go не требуется.
Из корня репозитория:

```sh
docker compose up --build
```

Команда собирает и запускает RabbitMQ, паблишер и консьюмер.

## Что должно получиться

В терминале появляются строки `sent` и `received` с одинаковым телом сообщения.
Пример логов (время сокращено):

```text
publisher-1 | ... sent exchange=demo.topic routing_key=demo.message body="Hello RabbitMQ! #1 at ..."
consumer-1  | ... received queue=hello exchange=demo.topic routing_key=demo.message body="Hello RabbitMQ! #1 at ..."
```

Админка RabbitMQ: [localhost:15672](http://localhost:15672).
Логин и пароль: `demo` / `demo`.

- **Exchanges** — exchange `demo.topic` типа `topic`.
- **Queues and Streams** — очередь `hello` и её сообщения.
- В карточке очереди, раздел **Bindings** — связь с `demo.topic` по шаблону `demo.*`.
