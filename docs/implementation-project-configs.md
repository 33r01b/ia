# Implementation Plan: Global And Project Config Files

## Goal

Реализовать файловую конфигурацию для `ia` по схеме:

```text
~/.config/ia/
  config.toml
  projects/
    <project>.toml
```

С сохранением текущего поведения:
- env `IA_*` продолжает работать
- CLI-флаги `--mask-file` и `--mask-dir` остаются самым сильным override layer
- отсутствие файлового конфига не является ошибкой

## Scope For V1

Поддержать в file config только эти поля:

```toml
[docker]
all_proxy = "..."
http_proxy = "..."
https_proxy = "..."
no_proxy = "..."
add_host = "..."
mask_files = ["..."]
mask_dirs = ["..."]
```

Не включать в v1:
- `agents.*.image`
- `agents.*.state_mount`
- `agents.*.config_source`

Причина: это снижает риск и ограничивает первый релиз runtime policy-слоем, не меняя container/runtime wiring и не вводя лишнюю agent-specific схему.

## Merge Model

Порядок merge:

1. built-in defaults
2. `~/.config/ia/config.toml`
3. `~/.config/ia/projects/<project>.toml`
4. env `IA_*`
5. CLI flags

Семантика merge:
- scalar fields: last writer wins
- `mask_files`, `mask_dirs`: union with dedup, preserving first-seen order
- `http_proxy` и `https_proxy` fallback на `all_proxy` только после merge всех слоев

## Code Changes

## 1. Update config loading API

Текущее состояние:
- [config.go](/app/go/ia/internal/app/config.go) экспортирует `loadConfig() (Config, error)`
- [app.go](/app/go/ia/internal/app/app.go) вызывает `loadConfig()` без project context

Нужно:

```go
func loadConfig(project string) (Config, error)
```

Новый flow:
1. собрать built-in defaults
2. применить global file config
3. применить project file config
4. наложить env overrides
5. применить пост-merge normalization

В [app.go](/app/go/ia/internal/app/app.go) нужно передавать `project` в `loadConfig(project)`.

## 2. Split config assembly into layers

В [config.go](/app/go/ia/internal/app/config.go) стоит разделить логику на небольшие функции:

```go
func defaultConfig() Config
func envConfigOverrides() (envConfig, error)
func applyEnvConfig(cfg *Config, envCfg envConfig)
func finalizeConfig(cfg *Config)
```

Почему это нужно:
- сейчас `loadConfig()` сразу собирает финальный runtime config из env
- для file config нужен промежуточный mergeable state
- так проще тестировать порядок наложения слоев

## 3. Add file config decode layer

Новый файл: [file_config.go](/app/go/ia/internal/app/file_config.go)

Нужные типы:

```go
type FileConfig struct {
    Project *string          `toml:"project"`
    Docker  FileDockerConfig `toml:"docker"`
}

type FileDockerConfig struct {
    AllProxy   *string  `toml:"all_proxy"`
    HTTPProxy  *string  `toml:"http_proxy"`
    HTTPSProxy *string  `toml:"https_proxy"`
    NoProxy    *string  `toml:"no_proxy"`
    AddHost    *string  `toml:"add_host"`
    MaskFiles  []string `toml:"mask_files"`
    MaskDirs   []string `toml:"mask_dirs"`
}
```

Почему pointer fields:
- нужно различать отсутствие поля и пустое значение
- это влияет на корректный merge scalar полей

## 4. Add file discovery helpers

В [file_config.go](/app/go/ia/internal/app/file_config.go):

```go
func configHome() (string, error)
func globalConfigPath() (string, error)
func projectConfigPath(project string) (string, error)
func loadGlobalConfigFile() (FileConfig, bool, error)
func loadProjectConfigFile(project string) (FileConfig, bool, error)
```

Рекомендации:
- использовать `os.UserHomeDir()` для построения `~/.config/ia/...`
- `projectConfigPath(project)` не должен дополнительно доверять path traversal, даже если `project` уже валидируется выше
- если файла нет, возвращать `(FileConfig{}, false, nil)`
- parse error возвращать как ошибку запуска

## 5. Add TOML parsing

Для TOML использовать `BurntSushi/toml`, зависимость уже есть в `go.sum`.

Примерно так:

```go
func decodeFileConfig(path string) (FileConfig, error)
```

Важно:
- включить strict decoding, чтобы неизвестные поля не игнорировались silently
- оборачивать ошибки путем файла, чтобы diagnostics были понятными

Ожидаемый формат ошибок:
- `decode global config /home/user/.config/ia/config.toml: ...`
- `decode project config /home/user/.config/ia/projects/billing.toml: ...`

## 6. Apply file config onto runtime config

В [config.go](/app/go/ia/internal/app/config.go) добавить:

```go
func (c *Config) applyFileConfig(f FileConfig)
func (c *Config) applyFileDockerConfig(d FileDockerConfig)
```

Merge rules:
- если pointer field non-nil, overwrite scalar field
- `mask_files` merge в `Docker.NullFiles`
- `mask_dirs` merge в `Docker.TmpfsDirs`

## 7. Add file config validation

В [file_config.go](/app/go/ia/internal/app/file_config.go) или [config.go](/app/go/ia/internal/app/config.go) добавить:

```go
func validateProjectFileConfig(project string, cfg FileConfig) error
```

Правила:
- если `cfg.Project != nil`, значение должно совпадать с CLI `project`
- `mask_files` и `mask_dirs` должны пройти через существующую `MountTargets.Validate()` после merge
- итоговая runtime config должна пройти текущую `validateForAgent()`

На v1 достаточно валидировать `project` и положиться на итоговую runtime validation для mount targets и agent config.

## 8. Preserve existing defaults behavior

Текущее special-case поведение:
- `.idea` автоматически добавляется в `TmpfsDirs`
- `HTTPProxy` и `HTTPSProxy` наследуются из `AllProxy`, если не заданы

Это должно остаться, но выполняться после merge всех слоев.

Рекомендуемая схема:

```go
func finalizeConfig(c *Config) {
    if c.Docker.HTTPProxy == "" {
        c.Docker.HTTPProxy = c.Docker.AllProxy
    }

    if c.Docker.HTTPSProxy == "" {
        c.Docker.HTTPSProxy = c.Docker.AllProxy
    }

    c.Docker.TmpfsDirs.Add(".idea")
}
```

Это важнее, чем текущее поведение внутри `loadConfig()`, потому что file config может изменить `AllProxy` позже defaults.

## 9. Update docs

Обновить [README.md](/app/go/ia/README.md):
- секцию `Configuration`
- описать file config layout
- описать merge order
- дать по одному примеру global и project TOML
- зафиксировать, что file config в v1 поддерживает только `docker.*`

## Testing Plan

Новый файл: `internal/app/file_config_test.go`

Минимальный набор тестов:

1. `loadConfig(project)` без файлов сохраняет текущее поведение.
2. global config применяет `docker.mask_files` и `docker.mask_dirs`.
3. project config переопределяет scalar values из global config.
4. env переопределяет значения из file config.
5. CLI masks дополняют file config и имеют наивысший приоритет.
6. `http_proxy` и `https_proxy` корректно fallback на итоговый `all_proxy`.
7. `project = "billing"` в `billing.toml` проходит validation.
8. `project = "other"` в `billing.toml` дает ошибку.
9. неизвестный TOML key дает ошибку decode.
10. отсутствие `config.toml` и `projects/<project>.toml` не дает ошибку.

## Error Handling

Ошибки должны быть user-facing и понятными.

Желательные форматы:
- `error: load config: decode global config /home/.../.config/ia/config.toml: ...`
- `error: load config: decode project config /home/.../.config/ia/projects/billing.toml: ...`
- `error: load config: project config mismatch: expected "billing", got "other"`

## Implementation Order

1. Ввести `FileConfig` и decode helpers.
2. Добавить discovery для global/project TOML.
3. Перестроить `loadConfig(project)` вокруг layered merge.
4. Добавить validation project file.
5. Обновить `app.go` на новый API.
6. Покрыть tests.
7. Обновить README.

## Acceptance Criteria

Реализация считается готовой, если:
- `ia` корректно запускается без файловых конфигов
- `~/.config/ia/config.toml` меняет глобальные defaults
- `~/.config/ia/projects/<project>.toml` меняет только выбранный проект
- env overrides побеждают file config
- CLI mask flags побеждают file config
- validation и error messages понятны
- README описывает новый контракт
- agent-specific настройки остаются в env и не дублируются в file config

## Deferred

Откладываем на потом:
- override `image`
- override `state_mount`
- override `config_source` через file config
- `ia config show`
- lookup по `language + project`
- поддержка нескольких форматов кроме TOML
