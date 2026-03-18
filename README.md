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
ia <agent> <language> <project> [--dry-run] [--mask-file <path>] [--mask-dir <path>]
```

Примеры:

```bash
ia codex go calc
ia claude php billing --dry-run
ia codex go calc --mask-file .env --mask-dir .cache
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

## Proxy

Для включения proxy можно задать общий адрес:

```bash
export IA_ALL_PROXY=http://host.docker.internal:8888
```

Или отдельно для HTTP и HTTPS:

```bash
export IA_HTTP_PROXY=http://host.docker.internal:8888
export IA_HTTPS_PROXY=http://host.docker.internal:8888
```

Примеры запуска:

```bash
export IA_ALL_PROXY=http://host.docker.internal:8888
ia codex go calc

export IA_HTTP_PROXY=http://host.docker.internal:8888
export IA_HTTPS_PROXY=http://host.docker.internal:8888
ia claude php billing --dry-run
```

Почему в примере `host.docker.internal`:
- агент запускается внутри Docker-контейнера
- `host.docker.internal` позволяет контейнеру обратиться к сервису на хост-машине
- это полезно, если локальный proxy слушает на хосте

Если у тебя другая схема сети, используй свои значения proxy-переменных.

Логика proxy:
- если `IA_HTTP_PROXY` не задан, используется `IA_ALL_PROXY`
- если `IA_HTTPS_PROXY` не задан, используется `IA_ALL_PROXY`
- если ни одна proxy-переменная не задана, контейнер запускается без proxy

## Mount Overrides

Для разового запуска можно передавать пути через CLI-опции:
- `--mask-file <path>` монтирует файл как `/dev/null`
- `--mask-dir <path>` монтирует директорию через `tmpfs`
- обе опции можно повторять или передавать список через запятую

Правила путей:
- относительные пути считаются от корня проекта внутри контейнера: `/app/<language>/<project>`
- абсолютные пути используются как есть
- `--mask-dir` дополняет стандартный tmpfs mount для `.idea`

Примеры:

```bash
ia codex go calc --mask-file .env --mask-file .secrets/local.yaml --mask-dir .cache
ia codex go calc --mask-file=.env,.secrets/local.yaml --mask-dir=.cache,tmp/runtime
```

Что это даст внутри контейнера:
- `.env` и `.secrets/local.yaml` будут замонтированы как `/dev/null`
- `.cache`, `tmp/runtime` и `.idea` будут замонтированы через `tmpfs`

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

`docker/Makefile` использует часть тех же `IA_*` переменных окружения, что и Go CLI.
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

1. Добавить поддержку конфигов по проектам в `~/.confing/ia/project/`
2. Добавить опцию входа в shell, по умолчанию сразу запускать агента в контейнере
