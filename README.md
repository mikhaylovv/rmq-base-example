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

Для паблишера с подтверждениями — отдельный файл `publisher-confirms/main.go`
и профиль `confirms`:

```sh
docker compose --profile confirms up --build consumer publisher-confirms
```

Если обычный паблишер уже работает, сначала остановите его:
`docker compose stop publisher`. Указывайте сервисы в команде выше: запуск
только с `--profile confirms` поднимет оба паблишера.

Этот вариант раз в секунду отправляет батч из трёх сообщений без ожидания
между отправками, затем ждёт Ack всех трёх. Следующий батч отправляется после
полного подтверждения предыдущего и тика таймера. При Nack, закрытии канала
или таймауте 10 секунд на весь батч — `panic`, без повторной отправки
и автоматического перезапуска контейнера. При штатной остановке текущий батч
дожидается подтверждений в пределах того же таймаута.

## Что должно получиться

В терминале появляются строки `sent` и `received` с одинаковым телом сообщения.
Пример логов (время сокращено):

```text
publisher-1 | ... sent exchange=demo.topic routing_key=demo.message body="Hello RabbitMQ! #1 at ..."
consumer-1  | ... received queue=hello exchange=demo.topic routing_key=demo.message body="Hello RabbitMQ! #1 at ..."
```

В варианте с подтверждениями после трёх `sent` появляются три `confirmed`
с соответствующими `tag`, затем `batch confirmed messages=3`.
`confirmed` — подтверждение брокера, а не обработки сообщения консьюмером.

Админка RabbitMQ: [localhost:15672](http://localhost:15672).
Логин и пароль: `demo` / `demo`.

- **Exchanges** — exchange `demo.topic` типа `topic`.
- **Queues and Streams** — очередь `hello` и её сообщения.
- В карточке очереди, раздел **Bindings** — связь с `demo.topic` по шаблону `demo.*`.
