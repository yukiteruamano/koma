# Changelog

Todos los cambios notables de este proyecto se documentan aquí. Formato basado en Keep a Changelog.

## [Unreleased] - 2026-08-20

### Cambiado
- Módulo renombrado `github.com/metafates/gache` → `github.com/yukiteruamano/gache`.
- Bump mínimo Go `1.19` → `1.26` (`go.mod:3`). CI actualizado a `actions/checkout@v4` / `setup-go@v5` y matriz `1.26`.

### Añadido
- `Cache.Clear() error` y `Cache.IsExpired() bool` (`gache.go`).
- `GobEncoderDecoder` y helper `WithGob(*Options)` (`encoder.go`, `options.go`).
- `Options.Validate()` y limpieza de `Path`.
- Soporte `FileSystem.Stat/Rename/Remove` para escritura atómica (`filesystem.go`).
- `.golangci.yml`, `Makefile` (test/vet/lint/cover), CI job `lint` separado (`workflows/test.yml`).

### Corregido
- **Data race** en `init()`/`Get()`/`Set()` por uso de `RLock` + escritura (`gache.go:22`, `api.go:37`, `init.go:3`). Ahora doble-checked locking en `init()` y `Get()` con upgrade a `Lock` para expiración.
- **FD leak** en `load.go:18` (faltaba `Close`).
- **Doble asignación de `Time`** (`api.go:24` + `save.go:28`). Ahora solo `saveLocked()` actualiza `Time` (y `saveClearedLocked()` preserva `nil` en expiración/clear).
- **Expiración inconsistente**: `Get()` solo retornaba `expired=true` sin limpiar estado ni llamar `ExpirationHook` ni persistir. Ahora limpia, persiste atómicamente con `saveClearedLocked()` y llama hook fuera del lock.
- **Permisos** `0777`/`0666` → `0755`/`0644` (`load.go:13`, `save.go:15,20`).
- **Escritura no atómica** → `*.tmp` + `Rename` (`save.go:9`).

### Seguridad
- Escritura atómica evita fichero vacío en crash.
- `.gitignore` ya no ignora directorio `test/` (corregido a `*.test` + `/testdata/tmp/`).

## [Anterior] - upstream metafates/gache
- Ver historial `git log` en `metafates/gache`.
