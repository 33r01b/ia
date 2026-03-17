# agent

CLI для запуска агентских контейнеров поверх локальных проектов.

Сейчас поддерживаются два режима запуска:
- `claude`
- `codex`

Формат команды:

```bash
agent <agent> <language> <project> [--dry-run]
```

Примеры:

```bash
agent codex go calc
agent claude php billing --dry-run
```

При запуске проект монтируется так:
- host: `./<project>`
- container: `/app/<language>/<project>`

`language` и `project` должны быть безопасными сегментами пути: без `/`, `.` и `..`.

Важно:
- `project` — это директория относительно текущей директории на хосте
- `language` используется только для выбора пути внутри контейнера (`./<project>` -> `/app/<language>/<project>`)
Внутрь `claude`, `codex` и других agent CLI он не передается как отдельный параметр и не включает какой-то специальный режим языка.

После запуска `agent ...` ты попадаешь внутрь контейнера с подготовленным окружением агента.
Сама утилита не выполняет `codex`, `claude` или другой agent CLI автоматически: нужную команду внутри контейнера нужно запускать вручную.

## Requirements

- Go
- Docker
- экспортированные env-переменные для конфигурации раннера

## Configuration

Конфигурация читается через `cleanenv` из переменных окружения.

Обязательные переменные:

```bash
export AGENT_ALL_PROXY=http://host.docker.internal:8888
```

Дополнительно для `claude`:

```bash
export AGENT_CLAUDE_CONFIG_SOURCE=$HOME/.agent/claude/.claude.json
```

Почему в примере `host.docker.internal`:
- агент запускается внутри Docker-контейнера
- `host.docker.internal` позволяет контейнеру обратиться к сервису на хост-машине
- это полезно, если локальный proxy слушает на хосте

Если у тебя другая схема сети, используй свое значение `AGENT_ALL_PROXY`.

Поддерживаемые переменные:

```bash
AGENT_ALL_PROXY
AGENT_HTTP_PROXY
AGENT_HTTPS_PROXY
AGENT_NO_PROXY
AGENT_DOCKER_ADD_HOST

AGENT_CLAUDE_IMAGE
AGENT_CLAUDE_STATE_MOUNT
AGENT_CLAUDE_CONFIG_SOURCE
AGENT_CLAUDE_CONFIG_TARGET

AGENT_CODEX_IMAGE
AGENT_CODEX_STATE_MOUNT
AGENT_CODEX_CONFIG_SOURCE
AGENT_CODEX_CONFIG_TARGET
```

Логика proxy:
- если `AGENT_HTTP_PROXY` не задан, используется `AGENT_ALL_PROXY`
- если `AGENT_HTTPS_PROXY` не задан, используется `AGENT_ALL_PROXY`

## Make

Корневой `Makefile`:

```bash
make build
make install
make lint
make run ARGS='codex go cmd'
make dry-run ARGS='claude php billing'
```

`make build` собирает бинарник в `./bin/agent`.

`make install` ставит бинарник в `${INSTALL_DIR}/agent`, по умолчанию в `~/.local/bin`.

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

`docker/Makefile` использует те же `AGENT_*` переменные окружения, что и Go CLI.
Для `init`-целей имя docker volume извлекается из `AGENT_*_STATE_MOUNT`.

Для `claude` там также обязательны:

```bash
export AGENT_CLAUDE_CONFIG_SOURCE=$HOME/.agent/claude/.claude.json
```

В `docker/gemini/Makefile` переменные тоже переведены на `AGENT_*`, хотя сам `gemini` пока не участвует в Go CLI.

## Project Layout

```text
cmd/agent            # entrypoint
internal/app         # CLI, config, validation, docker args
docker/              # Dockerfile и make-цели для агентских образов
Makefile             # build/run/install для CLI
```

## Notes

- `claude` и `codex` валидируются как фиксированный набор агентов.
- `project` берется как директория относительно текущей директории.
- `language` не влияет на host path и используется только для пути проекта внутри контейнера.
- `AGENT_CLAUDE_CONFIG_SOURCE` обязателен только для запуска `claude`, но не для `codex`.
- Этот README описывает текущее состояние кода; команды в этом окружении не запускались.

This project was developed with assistance from AI coding tools.
