---
date: 2026-07-15
topic: pgx-slog-logging-migration
---

# Migrate stores/pgx Logging to log/slog

## Summary

Migrate the data-access layer's logging from a project-internal zlog abstraction to Go's standard `log/slog` package via `LogAttrs(ctx, level, msg, slog.Attr...)`, without introducing wrappers, interface layers, or caller-attribution machinery on top of the standard library.

## Problem Frame

`stores/pgx/` currently routes every log call through `zlog.Logger`, an interface with zap-sugar-style `Infow/Debugw/Errorw/Warnw` methods backed by a default `slog.Default()` instance inside `utils/zlog`. In tests an external `zlog.Set(zap.Sugar())` substitutes a real `*zap.SugaredLogger`, which is what actually pins caller-attribution to the pgx call sites today. A pre-migration scan showed `utils/zlog` is the only other file in the entire repository that imports zlog — pgx is the package effectively keeping zlog alive.

The bridge through zlog now costs more than it earns: it carries an opaque abstraction over a stdlib type, drags in zap as the only practical injector, and offers no path for downstream projects to extend it. The standard library already provides what is needed.

## Requirements

**Migration mechanics**

- R1. The `stores/pgx/logger.go` accessor is rewritten so that `logger()` returns `*slog.Logger` (the standard library type), no other abstraction in between.
- R2. Every `Infow` / `Debugw` / `Errorw` / `Warnw` call site in the seven production files under `stores/pgx/` (alter.go, db.go, event.go, ops.go, registry.go, sift.go, textsearch.go) is converted to `LogAttrs(ctx, slog.LevelXxx, msg, slog.Attr...)`. Each `key, value` pair becomes a typed `slog.Attr` (string → `slog.String`, error → `slog.Any("err", err)`, integer → `slog.Int`, time → `slog.Time`, anything else → `slog.Any`).
- R3. Each modified file removes the import of `github.com/cupogo/andvari/utils/zlog`, and adds `import "context"` and `import "log/slog"` where they are not already present.
- R4. `stores/pgx/z0_main_test.go` `TestMain` no longer constructs a `*zap.Logger`, no longer calls `zlog.Set(...)`, and no longer imports either zap or zlog.

**Caller attribution**

- R5. When the configured slog handler has `HandlerOptions.AddSource=true`, each emitted record carries a `source` attribute whose `file` and `line` point to the pgx call site — the line that invokes `logger().LogAttrs(...)` — and not to a line inside the `log/slog` package or the pgx wrapper.

**Out of migration scope (preserved as-is)**

- R6. No source file in the `utils/zlog` package is modified by this work.
- R7. The replacement accessor does not expose a `pgx.SetLogger(...)` (or similarly named) entry point in this iteration.

## Success Criteria

- The package builds with `go build ./stores/pgx/...` after the migration, and the existing tests under `stores/pgx/` still pass in environments with a PostgreSQL instance reachable via `TEST_PG_STORE_DSN`.
- Downstream runs that previously redirected pgx logs by calling `zlog.Set(zapSugaredLogger)` no longer see pgx log lines unless the host application has configured `slog.SetDefault(...)` itself.
- A trace from one chosen production call site (for example the connect-OK line in `db.go`) shows, with `AddSource=true`, a `source` attribute whose `file` resolves to that line — verifiable by eye on a single stderr line.

## Scope Boundaries

- The `utils/zlog` package is not modified. Its deprecated `Debug`/`Info`/`Printf`-style methods and the dual `LoggerX` context-aware variants remain as-is for any other current or future user.
- No `Set` / `SetLogger` entry point is added to `pgx` in v1. Future iterations may add one if a need surfaces; the migration itself does not require it because `slog.SetDefault` covers the same surface at the host level.
- No `pgx.Logger` interface, custom struct wrapper, or `runtime.Callers`-based caller-injection layer is introduced. The standard library `HandlerOptions.AddSource` is the only caller-attribution mechanism.
- No other package outside `stores/pgx/` is migrated. The `utils/zlog` import only appears in two files today (pgx production and pgx test); once those references are gone, `utils/zlog` becomes effectively dormant inside this repository but is still preserved.
- The default `*slog.Logger` that pgx sources its logger from is whatever `slog.Default()` returns at runtime. pgx does not install its own internal default handler in this iteration, unless the user confirms otherwise before planning.

## Key Decisions

- **Caller attribution via stdlib only.** Empirical verification at a scratch module (`/tmp/slogverify/main.go`) showed that with `HandlerOptions.AddSource=true`, slog's internal `runtime.Callers(3, …)` already lands source.file/line on the calling line of `LogAttrs` (or any other slog public API). An earlier hypothesis that pgx needed its own `runtime.Callers` injection wrapper was wrong; the experiment disproved it. Keeping pgx minimal also makes the migration robust against future Go versions where slog's internal stack-frame count might change.
- **No interface abstraction.** Returning `*slog.Logger` directly keeps the API one type away from the standard library and avoids forcing every caller to learn a custom abstraction they cannot reuse.
- **No `Set` entry in v1.** A `pgx.SetLogger(*slog.Logger)` is plausible future work but not required for the migration to land. `slog.SetDefault` already covers the same need at the host level.

## Dependencies / Assumptions

- Go toolchain ≥ 1.21 (the release that introduced `log/slog`); verified during the brainstorm against `go version` reporting `go1.26.5 darwin/arm64`.
- Host applications that want pgx logs to carry a `source` attribute are responsible for configuring their slog handler with `AddSource=true` (e.g., via `slog.SetDefault(slog.New(handler))` with `HandlerOptions.AddSource: true`). pgx does not enforce this.
- The only `zlog.Set(...)` injection in this repository is in `z0_main_test.go`. Once removed, no test in the pgx package relies on a non-default handler, and the test output goes through whatever `slog.Default()` happens to be.
- The `zlog.Set(zap.Sugar())` test injection is genuinely the only thing keeping today's pgx logs working in tests — its removal is part of R4, not an oversight.

## Outstanding Questions

### Resolve Before Planning

- Q1. **[Affects R1]** Should the package-local default handler live inside pgx (e.g., a hardcoded `slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug, AddSource: true}))`) or remain pulled-through from `slog.Default()` as the brainstorm originally landed? Trade-off: an internal handler makes pgx self-sufficient and guarantees `AddSource=true` at the test boundary; pull-through keeps pgx zero-side-effect but means caller-accuracy depends on whoever sets `slog.SetDefault` in the host process.

### Deferred to Planning

- Q2. **[Affects R2]** At each call site, the `context.Context` argument to `LogAttrs` should be the existing function parameter when one is already in scope; otherwise `context.Background()` is the fallback. The per-site choice is mechanical and fits inside the implementation step; planning does not need a separate decision.
