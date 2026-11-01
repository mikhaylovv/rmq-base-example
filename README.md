# RabbitMQ + Go: базовый пример

## Что делает пример

Паблишер отправляет сообщение раз в секунду. Консьюмер получает его,
пишет в лог и подтверждает получение через `Ack`.

```text
publisher → exchange demo.topic (topic) → binding demo.* → очередь hello → consumer
```

## Быстрый запуск

Нужен работающий Docker с Docker Compose. Локальный Go не требуется.
Из корня репозитория, с обычным паблишером:

```sh
docker compose --profile basic up --build
```

Команда собирает и запускает RabbitMQ, паблишер и консьюмер.

Для паблишера с подтверждениями — отдельный файл `publisher-confirms/main.go`
и профиль `confirms`:

```sh
docker compose --profile confirms up --build
```

Только RabbitMQ и обычный паблишер, без консьюмера — профиль `publisher`:

```sh
docker compose --profile publisher up --build
```

Те же команды через Makefile (нужен `make`):

```sh
make basic
make confirms
make publisher
make clean
```

Профили `basic` и `confirms` запускают RabbitMQ, консьюмер и выбранный паблишер.
Профиль `publisher` запускает только RabbitMQ и обычный паблишер.
Без профиля запускается только RabbitMQ.
Перед переключением между уже запущенными вариантами остановите предыдущий запуск:
`make clean` (или `docker compose --profile basic --profile confirms --profile publisher down`).
Команда удаляет контейнеры и сеть всех трёх профилей, сохраняя данные RabbitMQ.

Данные RabbitMQ хранятся в папке `rabbitmq-data/` рядом с `compose.yaml`.
Она подключается в контейнер как `/var/lib/rabbitmq` и исключена из Git.

Вариант `confirms` раз в секунду отправляет батч из трёх сообщений без ожидания
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
