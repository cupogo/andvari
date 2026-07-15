---
title: Migrate stores/pgx Logging to log/slog
type: refactor
status: active
date: 2026-07-15
origin: docs/brainstorms/2026-07-15-pgx-slog-logging-migration-requirements.md
---

# Migrate stores/pgx Logging to log/slog

## Summary

Replace `zlog.Logger`-based `Infow/Debugw/Errorw/Warnw` calls throughout `stores/pgx/` with `log/slog`'s `LogAttrs(ctx, level, msg, slog.Attr...)`, drop the dormant `zlog.Set(zap.Sugar())` injection from `TestMain`, and leave `utils/zlog` and the public pgx API untouched.

## Problem Frame

See origin's Problem Frame for full context. Restating briefly: pgx is the only `utils/zlog` user in this repository; the abstraction adds a layer over a stdlib type without earning it.

## Requirements

R-IDs trace to the origin document (`docs/brainstorms/2026-07-15-pgx-slog-logging-migration-requirements.md`).

- R1. `stores/pgx/logger.go` exposes `func logger() *slog.Logger` (origin R1).
- R2. Every `Infow/Debugw/Errorw/Warnw` call in the seven production files under `stores/pgx/` is converted to `LogAttrs` with typed `slog.Attr` arguments (origin R2).
- R3. Each modified file drops the `utils/zlog` import and adds `context` / `log/slog` imports as needed (origin R3).
- R4. `stores/pgx/z0_main_test.go` no longer constructs a `*zap.Logger`, no longer calls `zlog.Set(...)`, and no longer imports either zap or zlog (origin R4).
- R5. When the host-configured slog handler has `HandlerOptions.AddSource=true`, emitted records carry a `source` attribute pointing at the pgx call site (origin R5).
- R6. No file under `utils/zlog` is modified (origin R6).
- R7. No `pgx.SetLogger(...)` (or similar) entry point is exposed in this iteration (origin R7).

**Origin actors:** none (no Actor ID triggers).
**Origin flows:** none (refactor is not flow-shaped).
**Origin acceptance examples:** none (no behavioral-conditional requirements).

## Scope Boundaries

- `utils/zlog` is preserved as-is. Its deprecated `Debug`/`Info`/`Printf`-style methods and `LoggerX` context-aware variants remain even after this work makes them dormant in this repository.
- No `Set` / `SetLogger` entry point is added on pgx; see Q1 in origin.
- No wrapper, interface, or `runtime.Callers`-based caller-attribution layer is introduced between pgx and `*slog.Logger`.
- No other package outside `stores/pgx/` is touched.
- The default logger that pgx sources from is `slog.Default()` at package-init time, captured in a package-level `_logger` variable. A subsequent `slog.SetDefault(...)` from the host does NOT redirect pgx logger output, because pgx already snapshot the value at init. This is the user-confirmed origin-Q1 direction; changing it is out of scope for this iteration. pgx does not install its own internal default.

### Deferred to Follow-Up Work

- A `pgx.SetLogger(*slog.Logger)` accessor, if future iterations surface demand from host-side code that cannot reach pgx via `slog.SetDefault`. Today's decision: not required because `slog.SetDefault` covers the same surface at the host level.

## Context & Research

### Relevant Code and Patterns

- The seven production files under `stores/pgx/` use zap-sugar-style key/value pairs consistently. Conversion is mechanical: error→`slog.Any("err", err)`, string→`slog.String`, integer→`slog.Int`, anything else→`slog.Any`. No bespoke typed wrappers exist.
- `utils/zlog/log.go` already imports `log/slog` and constructs `&logger{slg: slog.Default()}` at init; that is the local precedent for "pgx-style" slog integration and confirms `slog.Default()` is a reasonable choice from this codebase's perspective.
- `stores/pgx/context.go` already establishes the `context.Context` parameter convention used throughout pgx methods; nearly every logger call site is inside a function that already has a `ctx context.Context` in scope.
- `stores/pgx/db.go:OpenDB(dsn string)` (line 85) is the only ctx-less function that emits logger calls. Two call sites at lines 95 and 98 fall in this category and must use `context.Background()`.

### Institutional Learnings

- None. `docs/solutions/` does not exist in this repository.

### External References

- Go standard library source at `/opt/local/lib/go/src/log/slog/logger.go` — the `runtime.Callers(3, …)` at `Logger.logAttrs` is the load-bearing detail that proves R5 is satisfiable without a pgx wrapper. Verified empirically at `/tmp/slogverify/main.go` before authoring this plan.

## Key Technical Decisions

- **Slog-only caller attribution.** R5 is satisfied by `slog.HandlerOptions.AddSource=true` with no pgx-side code. Empirical verification at `/tmp/slogverify/main.go` showed source.file/line resolving to the pgx call site under both the `Info(...any)` and `LogAttrs(...Attr)` paths.
- **No `pgx` wrapper type or interface.** Returning `*slog.Logger` directly keeps pgx's surface a single standard-library type away from any caller. If a future `SetLogger` is added, that surface decision becomes a one-line wrap later — but doing it now would pre-commit to abstractions we have no caller-side evidence we need.
- **`var _logger = slog.Default()` at package init.** Per origin Q1, user picked option (a): pull-through from `slog.Default()`. The accessor itself is `func logger() *slog.Logger { return _logger }` after the binding; the package snapshots `slog.Default()` at init, and a later `slog.SetDefault(...)` from the host does NOT redirect pgx's logger output. This is a known v1 limitation; future iteration may switch to per-call re-fetch or add a `SetLogger` accessor if it becomes a real problem.
- **Note on F-5 (open consideration):** if the host calls `slog.SetDefault(...)` after pgx has been loaded, pgx keeps using the original default. The implementer-facing effect is that pgx's `_logger` snapshot is stable across the process lifetime. If this becomes a real concern (e.g., test frameworks that re-`SetDefault` mid-run), either replace the accessor with `func logger() *slog.Logger { return slog.Default() }` or add a `pgx.SetLogger(...)` entry point — both are explicitly out of scope for this iteration but cheap to add later.
- **Existing ctx in scope at all but two sites.** Almost every call site is inside a function or method whose signature already includes `context.Context`. The two exceptions (`db.go:OpenDB` lines 95 and 98) use `context.Background()`. Adding a `ctx` parameter to `OpenDB` would change a public exported function's signature, which is out of scope.
- **Slog attr type heuristics.** Errors→`slog.Any("err", err)`. Strings→`slog.String`. Integers→`slog.Int`. Booleans→`slog.Bool`. `time.Time`→`slog.Time`. Anything implementing `fmt.Stringer` or `slog.LogValuer`→`slog.Any(v)` (which respects their `LogValue`/`String` representation). Everything else→`slog.Any`. These match origin R2's explicit mapping and what downstream log aggregators most often index on.

## Open Questions

### Resolved During Planning

- **Q1 (origin, plan-time confirmation):** `var _logger = slog.Default()`. Resolved: pull-through from `slog.Default()` — no internal handler installed. Caller-accuracy is a host-side responsibility.
- **Q2 (origin):** ctx at each call site. Resolved: prefer the in-scope `ctx context.Context` parameter when one exists; otherwise use `context.Background()`. The two affected sites are `stores/pgx/db.go:95` and `stores/pgx/db.go:98` inside `OpenDB`.

### Deferred to Implementation

- Exact attribute key naming for ad-hoc fields beyond `err`. The migration must mirror the existing keys (`addr`, `db`, `user`, `id`, `name`, `model`, `field`, `during`, `cfg`, `query`, `table`, `schema`, `txt`, `ts`, `x`, `argc`, `result`, `pager`, `op`, `key`, `val`, `ot`, `obj`, `name`, `id`, `text`, `cats`, `slug`, `k`, etc.). Same-key equivalence matters for downstream log-query compatibility; the implementer should preserve every original key verbatim.
- Test runtime decision: whether `TestMain` should call `slog.SetDefault(...)` to give tests deterministic JSON output, or rely on the bare `slog.Default()`. The simplest path that satisfies R4 (no zap, no zlog.Set) is to leave slog alone; if test logs are unreadable in CI, an `init()` in `z0_main_test.go` that sets a JSON handler is acceptable scope.

## Implementation Units

### U1. Rewrite the package-local logger accessor

**Goal:** Replace `stores/pgx/logger.go` with the accessor that returns a `*slog.Logger` direct from `slog.Default()`.

**Requirements:** R1, R6

**Dependencies:** None

**Files:**
- Modify: `stores/pgx/logger.go`

**Approach:**
- Strip out the `zlog.Logger` return type and the `utils/zlog` import.
- Introduce a package-level `var _logger = slog.Default()` per origin Q1.
- Keep `func logger() *slog.Logger { return _logger }` as the only entry point. `_logger` (unexported) and `logger()` (unexported) match the existing visibility and avoid touching call sites that already use `logger()`.
- Note: at this point U1 alone *will* still compile, because `*slog.Logger` has matching `Infow/Debugw/Errorw/Warnw` methods. The migration to `LogAttrs` is a quality/perf/caller-attribution choice, not a compile-necessity.
- Do not introduce a `Set`/`SetLogger` here (R7).

**Patterns to follow:**
- The pattern used by `utils/zlog/log.go` of a package-level logger value bound to `slog.Default()` is the closest local precedent. The pgx version is simpler: no interface, no method wrappers.

**Test scenarios:**
- Test expectation: none — no behavioral change in isolation. U1 alone already compiles because `*slog.Logger` exposes the same variadic `Infow`/`Debugw`/`Errorw`/`Warnw` methods that the call sites currently use. Verification is via U2's mechanical rewrite and the regression test added in U2.

**Verification:**
- `go build ./stores/pgx/...` succeeds after U1 alone (callers still find `Infow`/`Debugw`/etc. on `*slog.Logger`).
- After U2 lands, the same build continues to compile; U3 clears the final `zlog.Set`/zap imports.

---

### U2. Convert every logger call site in production files

**Goal:** Replace every `Infow/Debugw/Errorw/Warnw` call under `stores/pgx/` with `LogAttrs`, using typed `slog.Attr` arguments and the existing function-scope `ctx` (or `context.Background()` for the two `OpenDB` sites).

**Requirements:** R2, R3, R5

**Dependencies:** U1

**Files:**
- Modify: `stores/pgx/alter.go` (4 logger calls)
- Modify: `stores/pgx/db.go` (6 logger calls; lines 95 in `OpenDB` and 325 in `patchPool` use `context.Background()`, lines 98 / 250 / 284 / 288 use enclosing `ctx`)
- Modify: `stores/pgx/event.go` (1 logger call)
- Modify: `stores/pgx/ops.go` (36 logger calls; all in ctx-scoped functions)
- Modify: `stores/pgx/registry.go` (1 logger call)
- Modify: `stores/pgx/sift.go` (3 logger calls)
- Modify: `stores/pgx/textsearch.go` (2 logger calls)
- Modify: `stores/pgx/trash.go` (9 logger calls; all in ctx-scoped functions)
- Modify: `stores/pgx/utilfs.go` (5 logger calls including one `Warnf`-style call shape; ctx availability to be checked per site)
- Create: `stores/pgx/logger_test.go` (regression test for `AddSource=true` attribution through `slog.Default()`)

**Approach:**
- Per call site:
  - Convert the method-name suffix to the matching `slog.Level*` constant (`Infow`→`LevelInfo`, `Debugw`→`LevelDebug`, `Errorw`→`LevelError`, `Warnw`→`LevelWarn`; `Warnf`-style calls map to `LevelWarn` as well).
  - Pass the in-scope `ctx`; use `context.Background()` at `db.go:95` (`OpenDB`), `db.go:325` (`patchPool`), and any other ctx-less call site the implementer surfaces.
  - Convert each `key, value` pair into the most specific typed `slog.Attr` possible: `slog.Any("err", err)` for errors, `slog.String` for strings, `slog.Int` for ints, `slog.Bool` for booleans, `slog.Time` for `time.Time`, `slog.Any` for anything else (preserves `fmt.Stringer`/`slog.LogValuer` representations).
  - Preserve every original key verbatim.
- Per file:
  - Add `import "context"` and `import "log/slog"` where missing; remove `"github.com/cupogo/andvari/utils/zlog"`.
- New permanent test in `stores/pgx/logger_test.go` (`TestPgxSlogSource`):
  - Wires `slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{AddSource: true})))` in a test-local setup.
  - Restores the previous default in test teardown.
  - Triggers a single log line from a known call site (e.g., a no-op invocation pattern that compiles through pgx).
  - Asserts the captured record's `source.file` ends with the expected pgx filename and `source.line` matches the expected line.

**Patterns to follow:**
- The existing per-file pattern of `ctx context.Context` already in scope at the call site. Reread each call site's enclosing function before substitution.
- Theattr-key conventions from the original `Infow(...)` arguments — preserve them bit-for-bit; downstream log queries may key on them.

**Test scenarios:**
- Happy path: each call site compiles; pre-existing tests still pass with an `AddSource=true` handler at runtime.
- Edge case: `OpenDB` with an unreachable DSN — the failure log path goes through `context.Background()` and the emitted record still carries source attribution when `AddSource=true` is configured.
- Error path: `EnsureSchema` / `EnsureExtension` / `Insert model fail` / `migrate fail` — all flow through their enclosing `ctx`. Verify the error attribute is `"err"` and typed `slog.Any`.
- Integration: one production call site (e.g., `db.go:98` `connected OK`) is run in a test that wires `slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{AddSource: true})))` and asserts the JSON record's `source.file` ends with `db.go` and `source.line == 98`.

**Verification:**
- `go build ./stores/pgx/...` succeeds.
- `go vet ./stores/pgx/...` reports no shadowing or unused-import warnings.
- A short characterization script (could be a one-off test, not a permanent file) emits one log line at `db.go:98` and confirms `source.file == "...stores/pgx/db.go"` and `source.line == 98`.

---

### U3. Clean up z0_main_test.go

**Goal:** Remove the `zlog.Set(zap.Sugar())` injection and the zap dependency from the test entrypoint, satisfying R4.

**Requirements:** R4, R6

**Dependencies:** U1

**Files:**
- Modify: `stores/pgx/z0_main_test.go`

**Approach:**
- Drop the `lgr, _ := zap.NewDevelopment()` construction.
- Drop `defer func() { _ = lgr.Sync() }()`.
- Drop `sugar := lgr.Sugar()`.
- Drop `zlog.Set(sugar)`.
- Remove imports: `"github.com/cupogo/andvari/utils/zlog"` and `"go.uber.org/zap"`.
- Keep `os.Setenv("DB_ALLOW_LEFT_WILDCARD", "1")` and the rest of `TestMain` intact.

**Patterns to follow:**
- Other test main functions in this repo if they exist; otherwise the minimum-trim pattern: `TestMain` should be `m.Run()` with the env-var setup only.

**Test scenarios:**
- Integration: `go test ./stores/pgx/...` runs end-to-end. Test output should appear via `slog.Default()` (whatever that is in the test environment).
- Edge case: even without `AddSource=true` configured externally, the test runs without panicking — missing-source is acceptable; missing-log-on-error is not.

**Verification:**
- `go test ./stores/pgx/...` passes against a PostgreSQL instance reachable via `TEST_PG_STORE_DSN`.
- `grep "zlog.Set\|zap.New" stores/pgx/` returns no results in production code or test code outside any preserved `// TODO` lines (there should be none).

## System-Wide Impact

- **Interaction graph:** `logger()` is the single fan-in point inside pgx. After this work, every existing call site routes through `*slog.Logger.LogAttrs` to `slog.Default().Handler().Handle(ctx, record)`. The downstream handler is whatever the host has wired up — there is no pgx-side handler in the chain.
- **Error propagation:** Logger calls do not return errors; they cannot change error propagation paths. They do remain non-blocking — that is a property of slog, not pgx.
- **State lifecycle risks:** `_logger` snapshots `slog.Default()` at package-init time and does not re-fetch. A host-side `slog.SetDefault(...)` issued after pgx is loaded does NOT redirect pgx output (the F-5 limitation described in Key Technical Decisions). This is the user-confirmed origin-Q1 direction for this iteration.
- **API surface parity:** Unchanged. `logger()` is unexported. Every exported pgx function keeps its current signature. Downstream code that does not import `utils/zlog` is unaffected. Downstream code that *does* import `utils/zlog` (none, per repo scan) is unaffected by this PR but loses the lever of redirecting pgx logs via `zlog.Set(...)` — they would have to call `slog.SetDefault(...)` instead.
- **Integration coverage:** Two layers worth crossing — slog handler config and the pgx call site. The new `TestPgxSlogSource` in `stores/pgx/logger_test.go` covers both: configure a JSON handler with `AddSource=true`, trigger a log line, verify the recorded `source` resolves to the pgx call line.
- **Unchanged invariants:** `utils/zlog` source is untouched (R6). Every public pgx function signature is unchanged. `OpenDB` does not gain a `ctx` parameter; `patchPool` continues to be ctx-less internally.

## Risks & Dependencies

| Risk | Mitigation |
|------|------------|
| Caller-accurate `source` requires host-side `slog.SetDefault(...)` configuration. If a host forgets `AddSource=true`, pgx logs lose line attribution. | Documented in origin as an explicit host responsibility (R5 wording). Future iteration may add a `pgx.SetLogger(...)` so hosts can configure pgx-specifically, but not in this PR. |
| Mechanical translation across ~58 call sites (across 9 files) can drop attribute keys or mis-type values, breaking downstream log queries. | Per-call-site review preserves original key strings verbatim; type heuristics in U2 are conservative (use `slog.Any` when in doubt); a final `go vet` + `go build` guards compile-level breakage; the new permanent regression test in `logger_test.go` catches regressions in source-attribution wiring. |
| `context.Background()` at three ctx-less sites (`OpenDB`, `patchPool`, and any utilfs or trash site the implementer surfaces) is a step backward from the rest of pgx's "always ctx-in-scope" pattern. | Accepted trade-off: adding `ctx` to `OpenDB`'s signature would change a public exported function — out of scope for this iteration. A follow-up could refactor `OpenDB` to `OpenDBContext(ctx, dsn)` if needed. |
| Test logs may be unreadable in CI without an explicit handler configured. | Deferred to implementation per Open Questions; a minimal `init()` in `z0_main_test.go` setting a JSON handler to `os.Stderr` is acceptable if needed. |

## Documentation / Operational Notes

- No README or external doc changes — pgx's logger behavior was implicit before and stays implicit after. A brief mention in `docs/changelog.md` (if such a file exists in this repo) is appropriate but is not a blocker.
- No operational rollout concerns. Host applications that previously relied on `zlog.Set(zap.Sugar())` to redirect pgx logs must migrate to `slog.SetDefault(...)`. That migration is the host application's responsibility, not this PR's.

## Sources & References

- **Origin document:** [docs/brainstorms/2026-07-15-pgx-slog-logging-migration-requirements.md](../../brainstorms/2026-07-15-pgx-slog-logging-migration-requirements.md)
- **Related code:** `stores/pgx/logger.go`, the seven production files under `stores/pgx/`, and `stores/pgx/z0_main_test.go`
- **External docs / source:** Go standard library `log/slog` at `/opt/local/lib/go/src/log/slog/logger.go` and `handler.go`; empirical test bench `/tmp/slogverify/main.go`
- **Repository CLAUDE.md / project guidance:** [../../CLAUDE.md](../../CLAUDE.md) (no constraints conflict)
