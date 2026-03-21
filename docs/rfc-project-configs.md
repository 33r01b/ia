# RFC: Global And Project Config Files

## Status

Draft

## Summary

Добавить файловую конфигурацию для `ia` в структуре:

```text
~/.config/ia/
  config.toml
  projects/
    project-a.toml
    project-b.toml
```

Где:
- `config.toml` задает глобальные defaults
- `projects/<project>.toml` задает overrides для конкретного проекта
- CLI-флаги текущего запуска остаются самым приоритетным слоем

Это добавляет устойчивый файловый конфиг поверх текущего env-based подхода и позволяет хранить настройки по проектам без повторения `IA_*` переменных и `--mask-*` флагов.

## Motivation

Сейчас `ia` читает конфигурацию только из env и частично из CLI-опций запуска.
Этого недостаточно, если:

1. Нужны постоянные глобальные defaults без обязательного экспорта env.
2. У разных проектов разные mount masks и proxy-настройки.
3. Хочется иметь явный и редактируемый config layer вне shell profile.

По текущему коду:
- `loadConfig()` читает только env через `cleanenv`
- `project` используется как ключ запуска и безопасный path segment
- постоянного project-level config layer нет

## Goals

- Добавить global config file: `~/.config/ia/config.toml`
- Добавить project config files: `~/.config/ia/projects/<project>.toml`
- Сохранить совместимость с текущими `IA_*` переменными окружения
- Сделать merge order простым и предсказуемым
- Не менять основной CLI contract `ia <agent> <language> <project>`

## Non-Goals

- Не хранить конфиг внутри самих project repositories
- Не вводить сложное наследование между project config files
- Не добавлять шаблоны, includes и генерацию конфига
- Не менять semantics аргумента `project`
- Не добавлять agent-specific file config в v1

## Proposed Layout

Целевая структура:

```text
~/.config/ia/
  config.toml
  projects/
    billing.toml
    calc.toml
```

### Rationale

Почему так:
- один корневой каталог для всех конфигов `ia`
- `config.toml` очевидно выглядит как global baseline
- `projects/` изолирует per-project overrides
- TOML хорошо подходит для ручного редактирования и уже есть в зависимостях проекта
- `docker.*` покрывает основной практический кейс без лишней схемы

## Discovery

### Global config

Если существует файл:

```text
~/.config/ia/config.toml
```

он загружается как global file config.

Если файла нет, это не ошибка.

### Project config

Для запуска:

```bash
ia codex go billing
```

`ia` ищет файл:

```text
~/.config/ia/projects/billing.toml
```

Ключ lookup на первом этапе: только `project`.

Если файла нет, это не ошибка.

## Merge Order

Предлагаемый приоритет, от меньшего к большему:

1. built-in defaults
2. global file config: `~/.config/ia/config.toml`
3. project file config: `~/.config/ia/projects/<project>.toml`
4. environment variables: `IA_*`
5. CLI run options: `--mask-file`, `--mask-dir`

### Why env after files

Если оставить env выше файловых конфигов:
- shell environment остается аварийным override-механизмом
- проще временно переопределять file config без редактирования TOML
- сохраняется привычная логика текущего инструмента, где env уже считается внешним источником правды

## Proposed Schema

## Global config example

```toml
[docker]
all_proxy = "http://host.docker.internal:8888"
no_proxy = "host.docker.internal,localhost"
add_host = "host.docker.internal:host-gateway"
mask_files = [".env"]
mask_dirs = [".idea"]
```

## Project config example

```toml
project = "billing"

[docker]
mask_files = [".env", ".secrets/local.yaml"]
mask_dirs = [".cache", "tmp/runtime"]
no_proxy = "host.docker.internal,localhost,internal.service"
```

## Schema

```toml
project = "optional string"

[docker]
all_proxy = "string"
http_proxy = "string"
https_proxy = "string"
no_proxy = "string"
add_host = "string"
mask_files = ["string"]
mask_dirs = ["string"]
```

## Merge Semantics

Scalar fields:
- last writer wins

List fields:
- `mask_files`, `mask_dirs` merge через union с deduplication
- порядок сохраняется по первому появлению

Derived fields:
- `http_proxy` fallback на `all_proxy`, если после merge оно пустое
- `https_proxy` fallback на `all_proxy`, если после merge оно пустое

## Validation

Нужны следующие правила:
- если в project file задан `project`, он должен совпадать с именем `<project>.toml` и CLI-аргументом
- `mask_files` и `mask_dirs` проходят текущую валидацию mount targets
- неизвестные ключи в TOML лучше считать ошибкой

## Runtime Flow

Предлагаемая последовательность:

1. Собрать built-in defaults.
2. Прочитать `~/.config/ia/config.toml`, если файл существует.
3. Прочитать `~/.config/ia/projects/<project>.toml`, если файл существует.
4. Наложить env overrides из `IA_*`.
5. Наложить CLI run options.
6. Выполнить итоговую валидацию.

## Implementation Sketch

### Suggested API

```go
func loadConfig(project string) (Config, error)
func loadGlobalConfigFile() (FileConfig, bool, error)
func loadProjectConfigFile(project string) (FileConfig, bool, error)
func (c *Config) applyFileConfig(f FileConfig)
```

### Suggested types

Отдельный decode-layer:
- `FileConfig`
- `FileDockerConfig`

Это лучше, чем декодировать TOML напрямую в runtime `Config`, потому что нужно различать:
- поле отсутствует
- поле задано пустым значением

### Suggested files

- `internal/app/config.go`
  runtime config, merge logic
- `internal/app/file_config.go`
  TOML decode, discovery, helpers
- `internal/app/file_config_test.go`
  tests на parsing, merge order, validation

## Backward Compatibility

Обратная совместимость должна сохраниться:
- если ни `config.toml`, ни project file не существуют, поведение не меняется
- текущие `IA_*` переменные продолжают работать
- текущие agent-specific настройки остаются в env
- текущие CLI-флаги продолжают работать и остаются самым верхним override layer для mount masks

## Open Questions

1. Нужен ли отдельный debug-режим для merged config, например `ia config show ...`?
2. Оставлять ли `project = "..."` внутри project file обязательным или только optional validation field?
3. Нужны ли когда-либо agent-specific file configs, или лучше держать их только в env?

## Recommendation

Для первой итерации:
- принять TOML-layout с `~/.config/ia/config.toml` и `~/.config/ia/projects/<project>.toml`
- global file использовать как baseline
- project file использовать как per-project override
- env оставить выше файлов
- CLI оставить самым приоритетным слоем
- ограничить file config только полями `docker.*`

Эта схема компактная, легко объясняется и хорошо ложится на текущую архитектуру `ia`.
