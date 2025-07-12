# go-grpc-todo-list

## Описание проекта

`go-grpc-todo-list` — простой пример сервиса Go с gRPC API и HTTP-шлюзом для работы со списком задач.

## Особенности

* gRPC API для CRUD-операций над todo-задачами
* Чистая структура проекта (cmd, internal, proto, etc)
* Поддерживается генерация OpenAPI-документации (Swagger)
* HTMX-интерфейс как облегчённый способ взаимодействия с API

## Структура проекта

* `cmd/` — точка входа в приложение (main.go)
* `internal/todo/service` — реализация gRPC сервисов
* `internal/todo/http` — рендер htmx страниц
* `proto/` — .proto файлы для сервисов
* `third_party` — внешние пакеты и библиотеки

## Требования

* Go 1.24.2 или выше
* protoc (Protocol Buffers compiler)
* protoc-gen-go, protoc-gen-go-grpc

## Установка

### Установка через консоль

1. Клонируйте репозиторий:

```sh
git clone https://github.com/yourname/go-grpc-todo-list.git
cd go-grpc-todo-list
```

2. Установите Go зависимости:

```sh
go mod download
```

3. Сгенерируйте gRPC код:

```sh
make gen
```

4. Запустите сервер:

```sh
go build -o .build/main.exe ./cmd/app/main && .build/main
```

### Запуск с использованием Docker

1. Постройте Docker образ:

    ```sh
    docker build -t go-grpc-todo-list .
    ```

2. Запустите контейнер:

    ```sh
    docker run -p 80:80 go-grpc-todo-list
    ```

3. Откройте браузер и перейдите по адресу `http://localhost:80`.

### Запуск с использованием Docker Compose

1. Запустите контейнер:

```sh
docker-compose up -d
```

## Лицензия

Этот проект лицензирован по лицензии MIT. Подробности см. в файле `LICENSE`.
