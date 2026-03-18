# ia

<p align="center">
  <img src="docs/images/donkey.png" width="220" alt="IA mascot"/>
</p>

<p align="center">
  <b>isolated agent runner</b>
</p>

<p align="center">
  <i>does things its own way</i>
</p>

---

Инструмент для изолированного запуска agent CLI в контейнерах поверх локальных проектов.

`ia` поднимает отдельное контейнерное окружение для выбранного агента, монтирует в него локальный проект и дает запускать `claude`, `codex` и другие agent CLI без смешивания их состояния с хостовой системой.

Сейчас поддерживаются два режима запуска:
- `claude`
- `codex`

Формат команды:

```bash
ia <agent> <language> <project> [--dry-run]
```

Примеры:

```bash
ia codex go calc
ia claude php billing --dry-run
```

При запуске проект монтируется так:
- host: `./<project>`
- container: `/app/<language>/<project>`

`language` и `project` должны быть безопасными сегментами пути: без `/`, `.` и `..`.

Важно:
- `project` — это директория относительно текущей директории на хосте
- `language` используется только для выбора пути внутри контейнера (`./<project>` -> `/app/<language>/<project>`)
Внутрь `claude`, `codex` и других agent CLI он не передается как отдельный параметр и не включает какой-то специальный режим языка.

После запуска `ia ...` ты попадаешь внутрь контейнера с подготовленным окружением агента.
Сама утилита не выполняет `codex`, `claude` или другой agent CLI автоматически: нужную команду внутри контейнера нужно запускать вручную.

## Requirements

- Go
- Docker
- экспортированные env-переменные для конфигурации раннера

## Configuration

Конфигурация читается через `cleanenv` из переменных окружения.

Обязательные переменные:

```bash
export IA_ALL_PROXY=http://host.docker.internal:8888
```

Почему в примере `host.docker.internal`:
- агент запускается внутри Docker-контейнера
- `host.docker.internal` позволяет контейнеру обратиться к сервису на хост-машине
- это полезно, если локальный proxy слушает на хосте

Если у тебя другая схема сети, используй свое значение `IA_ALL_PROXY`.

Поддерживаемые переменные:

```bash
IA_ALL_PROXY
IA_HTTP_PROXY
IA_HTTPS_PROXY
IA_NO_PROXY
IA_DOCKER_ADD_HOST

IA_CLAUDE_IMAGE
IA_CLAUDE_STATE_MOUNT
IA_CLAUDE_CONFIG_SOURCE

IA_CODEX_IMAGE
IA_CODEX_STATE_MOUNT
IA_CODEX_CONFIG_SOURCE
```

Логика proxy:
- если `IA_HTTP_PROXY` не задан, используется `IA_ALL_PROXY`
- если `IA_HTTPS_PROXY` не задан, используется `IA_ALL_PROXY`

## Make

Корневой `Makefile`:

```bash
make build
make install
make lint
make run ARGS='codex go calc'
make dry-run ARGS='claude php billing'
```

`make build` собирает бинарник в `./bin/ia`.

`make install` ставит бинарник в `${INSTALL_DIR}/ia`, по умолчанию в `~/.local/bin`.

`make lint` запускает `golangci-lint run`.

## Docker Images

В каталоге `docker/` лежат Dockerfile и make-цели для образов агентов.

Примеры:

```bash
make -C docker build-claude
make -C docker run-claude
make -C docker build-codex
make -C docker run-codex
```

Для явного создания state volumes:

```bash
make -C docker init
make -C docker init-claude
make -C docker init-codex
make -C docker/gemini init
```

`docker/Makefile` использует те же `IA_*` переменные окружения, что и Go CLI.
Для `init`-целей имя docker volume извлекается из `IA_*_STATE_MOUNT`.

Для `claude` дополнительно нужен bind mount файла:

```bash
export IA_CLAUDE_CONFIG_SOURCE=$HOME/.config/ia/claude/.claude.json
```

В `docker/gemini/Makefile` переменные тоже переведены на `IA_*`, хотя сам `gemini` пока не участвует в Go CLI.

## Project Layout

```text
cmd/ia               # entrypoint
internal/app         # CLI, config, validation, docker args
docker/              # Dockerfile и make-цели для агентских образов
Makefile             # build/run/install для CLI
```

## Notes

- `claude` и `codex` валидируются как фиксированный набор агентов.
- `project` берется как директория относительно текущей директории.
- `language` не влияет на host path и используется только для пути проекта внутри контейнера.
- Claude state хранится в docker volume, а `~/.claude.json` монтируется как отдельный файл через `IA_CLAUDE_CONFIG_SOURCE`.

This project was developed with assistance from AI coding tools.

## TODO

1. Добавить поддержку списка файлов, которые нужно монтировать как `/dev/null`, и директорий, которые нужно монтировать через `tmpfs`.
2. Сделать proxy опциональным.
