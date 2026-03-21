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
ia <agent> <language> <project> [--dry-run] [--shell] [--mask-file <path>] [--mask-dir <path>]
```

Аргументы:
- `agent` выбранный agent CLI: `claude` или `codex`
- `language` сегмент пути внутри контейнера: `/app/<language>/<project>`
- `project` идентификатор проекта: используется как директория проекта относительно текущей директории на хосте, как имя файла `~/.config/ia/projects/<project>.toml` и как сегмент пути внутри контейнера

Опции:
- `--dry-run` вывести итоговую команду `docker run` без запуска
- `--shell` открыть `bash` вместо agent CLI
- `--mask-file <path>` замаскировать файл через mount в `/dev/null`
- `--mask-dir <path>` замаскировать директорию через `tmpfs`

Примеры:

```bash
ia codex go calc
ia claude php billing --dry-run
ia codex go calc --mask-file .env --mask-dir .cache
ia codex go calc --shell
```

При запуске проект монтируется так:
- host: `./<project>`
- container: `/app/<language>/<project>`

`language` и `project` должны быть безопасными сегментами пути: без `/`, `.` и `..`.

Важно:
- `project` одновременно задает host directory, project config key и final path segment внутри контейнера
- `language` используется только для выбора пути внутри контейнера (`./<project>` -> `/app/<language>/<project>`)
- внутрь agent CLI ни `language`, ни `project` не передаются как отдельные параметры

По умолчанию `ia ...` сразу запускает agent CLI внутри контейнера.
Если нужен интерактивный shell вместо агента, добавь `--shell`.

## Requirements

- Go
- Docker

## Configuration

`ia` собирает итоговую конфигурацию из нескольких слоев.

Структура файловых конфигов:

```text
~/.config/ia/
  config.toml
  projects/
    billing.toml
    calc.toml
```

Где:
- `~/.config/ia/config.toml` — глобальный baseline
- `~/.config/ia/projects/<project>.toml` — override для конкретного проекта, который выбирается по positional аргументу `project`
- `IA_*` переменные окружения — временный override поверх файловых конфигов
- CLI-флаги `--mask-file` и `--mask-dir` — самый приоритетный слой

Порядок merge:

1. built-in defaults
2. `~/.config/ia/config.toml`
3. `~/.config/ia/projects/<project>.toml`
4. `IA_*`
5. CLI flags

Если файлового конфига нет, это не ошибка.

## Supported File Config Fields

File config сейчас поддерживает только `docker.*`:

```toml
[docker]
all_proxy = "http://host.docker.internal:8888"
http_proxy = "http://host.docker.internal:8888"
https_proxy = "http://host.docker.internal:8888"
no_proxy = "host.docker.internal,localhost"
add_host = "host.docker.internal:host-gateway"
mask_files = [".env"]
mask_dirs = [".idea"]
```

Через file config пока не настраиваются:
- `image`
- `state_mount`
- `config_source`

Эти параметры по-прежнему задаются через `IA_*` env-переменные.

## Global Config Example

```toml
[docker]
no_proxy = "host.docker.internal,localhost"
add_host = "host.docker.internal:host-gateway"
mask_dirs = [".idea"]
```

## Project Config Example

Для запуска:

```bash
ia codex go billing
```

можно создать файл:

```text
~/.config/ia/projects/billing.toml
```

с таким содержимым:

```toml
project = "billing"

[docker]
mask_files = [".env", ".secrets/local.yaml"]
mask_dirs = [".cache", "tmp/runtime"]
no_proxy = "host.docker.internal,localhost,internal.service"
```

Поле `project` внутри project file опционально, но если задано, оно должно совпадать с именем проекта из CLI и именем файла `billing.toml`.

## Environment Variables

Поддерживаемые переменные окружения:

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

Когда использовать env:
- для временного override поверх file config
- для настройки `image`, `state_mount` и agent-specific `config_source`
- для shell-specific сценариев

Пример:

```bash
export IA_HTTP_PROXY=http://host.docker.internal:8888
ia codex go calc
```

Логика proxy:
- если итоговый `HTTP_PROXY` не задан, используется итоговый `ALL_PROXY`
- если итоговый `HTTPS_PROXY` не задан, используется итоговый `ALL_PROXY`
- если ни одна proxy-переменная не задана ни в одном слое, контейнер запускается без proxy

## Mount Overrides

Для разового запуска можно передавать пути через CLI-опции:
- `--mask-file <path>` монтирует файл как `/dev/null`
- `--mask-dir <path>` монтирует директорию через `tmpfs`
- обе опции можно повторять или передавать список через запятую

Правила путей:
- относительные пути считаются от корня проекта внутри контейнера: `/app/<language>/<project>`
- абсолютные пути используются как есть
- `--mask-dir` дополняет стандартный tmpfs mount для `.idea`
- file config и CLI masks merge'ятся с deduplication

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
docs/                # RFC и implementation notes
```

## Notes

- `claude` и `codex` валидируются как фиксированный набор агентов.
- по умолчанию `ia` запускает агент сразу; `--shell` оставляет контейнер в `bash`.
- `project` определяет сразу три вещи: host directory `./<project>`, lookup файла `~/.config/ia/projects/<project>.toml` и final path segment внутри контейнера.
- `language` не влияет на host path и используется только для пути проекта внутри контейнера.
- Claude state хранится в docker volume, а config file монтируется через `IA_CLAUDE_CONFIG_SOURCE`.
- неизвестные ключи в TOML считаются ошибкой конфигурации.

This project was developed with assistance from AI coding tools.

## Docs

- [RFC: Global And Project Config Files](docs/rfc-project-configs.md)
- [Implementation Plan: Global And Project Config Files](docs/implementation-project-configs.md)
