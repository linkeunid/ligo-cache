# CLAUDE.md

## Behavioral Foundation

1. Don't assume. Don't hide confusion. Surface tradeoffs.
2. Minimum code that solves the problem. Nothing speculative.
3. Touch only what you must. Clean up only your own mess.
4. Define success criteria. Loop until verified.

## Project

A caching extension package for [Ligo](https://github.com/linkeunid/ligo),
inspired by [@nestjs/cache-manager](https://docs.nestjs.com/techniques/caching).
Provides a pluggable `Store` interface, an in-memory default with per-entry TTL
and background eviction, a `Cache` service with get-or-compute semantics, and an
HTTP interceptor for automatic response caching.

- Framework: `github.com/linkeunid/ligo`
- License: MIT

## Architecture

```
Config → Module(cfg) → *Cache (singleton via DI)
                      ├── Store (pluggable interface)
                      │     └── InMemoryStore (default, TTL + janitor goroutine)
                      └── CacheInterceptor (opt-in per route group, GET-only)
```

**Key types:**
- `Store` interface — pluggable backend (`Get/Set/Del/Reset`)
- `InMemoryStore` — default with `sync.RWMutex`, per-entry TTL, janitor goroutine
- `Cache` — wraps one Store, adds `Wrap` (get-or-compute), default TTL
- `CacheInterceptor` — `ligo.Interceptor` for auto-caching GET responses
- `Config` — `{Store, DefaultTTL, JanitorInterval}`

**File layout:**
- `store.go` — `Store` interface + `InMemoryStore`
- `cache.go` — `Cache` struct
- `interceptor.go` — `CacheInterceptor`
- `module.go` — `Config`, `Module(cfg)`, `Provider(cfg)`

## Commands

```bash
go build ./...                          # Build
go test ./...                           # Run tests
go test -v ./...                        # Verbose tests
go test -race ./...                     # Race detector
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
go test -bench=. -benchmem ./...        # Benchmarks
golangci-lint run                       # Lint (config: .golangci.yml)
gofumpt -w .                            # Format (stricter than gofmt)
govulncheck ./...                       # CVE scan
go mod tidy                             # Tidy deps
```

## Go Best Practices (do)

- **Small, focused packages.** A package's name should describe a single
  responsibility. If you can't name it cleanly, it's doing too much.
- **Accept interfaces, return structs.** Consumers depend on the smallest
  surface possible; producers expose concrete types so behavior can grow.
- **Errors are values.** Wrap with `fmt.Errorf("doing X: %w", err)`. Check
  with `errors.Is` / `errors.As`, never `==` or type assertion.
- **Context first.** `func F(ctx context.Context, ...)` — always the first
  parameter, never stored in a struct.
- **Use `any`, not `interface{}`.** Modern Go.
- **Pre-allocate slices** when the size is known: `make([]T, 0, n)`.
- **`fmt.Errorf("%w", err)` to wrap, `errors.New` for sentinel errors.**

## Go Best Practices (don't)

- **Don't use `init()` for application logic.** It's untestable. Use a
  factory function and register it in a module's providers instead.
- **Don't `panic` in library code.** Return an error. Reserve panics for
  truly unrecoverable programmer mistakes (e.g. invariant violation).
- **Don't ignore errors.** `_ = doX()` only when you've actually decided
  the error doesn't matter, and leave a comment explaining why.
- **Don't share mutable state across goroutines without synchronization.**
  Channels for ownership transfer; `sync.Mutex` for shared mutation;
  `sync.RWMutex` only when reads vastly outnumber writes.
- **Don't put business logic in HTTP handlers / controllers.** They
  translate request ↔ response and delegate. Use cases own the rules.
- **Don't reach for global variables.** Inject via Ligo's DI container.

## Ligo Practices (do)

- **One module per bounded context.** `user`, `auth`, `billing` — each its
  own `ligo.NewModule(...)` with providers + controllers + middleware.
- **Constructor injection.** Factories take their dependencies as
  parameters; Ligo's DI resolves them.
  ```go
  func NewUserService(cache *ligo_cache.Cache) *UserService
  ligo.Factory[*UserService](NewUserService)
  ```
- **Validate at the edge.** Use `ligo.ValidationPipe(&Dto{})` on the route
  and `ligo.ValidatedBody[Dto](ctx)` in the handler. The use case can
  then trust its input.
- **Resolve dependencies with `ligo.MustResolve[T](app)` only after
  `app.Run()`.** Prefer `ligo.Resolve[T]` (returns `(T, error)`) when
  the failure is recoverable; reserve `MustResolve` for cases where a
  missing provider really should crash the process.
- **Pagination in the framework.** `ctx.Paginate(20, 100)` and
  `ctx.Paginated(items, page, perPage, total)` — don't roll your own.
- **Query binding in the framework.** `ctx.BindQuery(&filter)` with
  `query:"name"` tags. Don't parse `ctx.Request().URL.Query()` by hand.

## Ligo Practices (don't)

- **Don't store DI-resolved singletons in package-level vars.** Resolve at
  construction time and pass through the constructor chain.
- **Don't depend on resolution order between providers in the same
  module.** OnInit hooks run after construction; do cross-provider setup
  there, not in factories.
- **Don't bypass middleware by composing handlers manually.** Use
  `ligo.NewChainRouter(r.Group(prefix))` so middleware order is explicit.
- **Don't pin a Ligo minor version forever.** Track minor releases —
  they add API surface without breaking changes.

## Static Analysis

Every Ligo project ships `.golangci.yml` (schema v2) enabling
`errcheck`, `govet` (with `shadow` and `nilness`), `staticcheck`,
`unused`, `gofumpt`, `misspell`, `unconvert`, `unparam`, `revive`,
`bodyclose`, `errorlint`, `nolintlint`, `whitespace`, `tagalign`, `gci`.

`gci` enforces uniform import order: stdlib, third-party, local —
separated by blank lines, in that order. The `prefix(...)` matches this
project's module path so your own packages always group as "local".

Install the toolchain once:

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install mvdan.cc/gofumpt@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/daixiang0/gci@latest
go install github.com/4meepo/tagalign/cmd/tagalign@latest
go install golang.org/x/tools/gopls@latest
```

Pre-merge checklist:

```bash
gci write --skip-generated -s standard -s default --custom-order .
gofumpt -w .                                                     # formatting
tagalign -fix -sort $(find . -name '*.go' -not -path './vendor/*') # tag columns
go test -race ./...     # tests + race detector
golangci-lint run       # static checks (incl. gci, gofumpt, tagalign)
govulncheck ./...       # CVE scan
```

## CI

`.github/workflows/ci.yml` runs `golangci-lint` (v2), `go test -race`,
and `govulncheck` on every push to `main` and every pull request,
pinned to Node-24 action versions (checkout v6, setup-go v6,
golangci-lint-action v9). Lint catches gci / gofumpt / errcheck /
errorlint drift before merge.
