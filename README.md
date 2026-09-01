# Goa v3.31.0-preview findings

Minimal reproductions found while testing `goa.design/goa/v3@v3.31.0-preview.1` and
`v3.31.0-preview.3` against an existing multi-project HTTP codebase (no gRPC, no JSON-RPC, no
streaming, no unions, no interceptors).

Each directory is a self-contained Go module with a minimal design, a `check.sh` that runs a small
condition matrix against a given Goa version, and a README with the diagnosis and, where one was
found, a suggested fix.

| # | Issue | Status |
|:--|:--|:--|
| [01](01-required-array-panic/) | Panic when a result type has a required array of a user type | Cause found, one-line fix verified |
| [02](02-collection-name-collision/) | `CollectionOf` called twice in one service collides on the generated name | Cause found, one-line fix verified |
| [03](03-multiple-success-responses/) | Two untagged success responses: a planned conversion is never rendered | Cause found, not fixed |

All three generate successfully with every released version tested (v3.27.0, v3.28.0, v3.29.2,
v3.30.0) and fail with both previews, so all three are regressions introduced on the preview branch.

## Running

Each directory:

```sh
cd 01-required-array-panic
go run goa.design/goa/v3/cmd/goa gen repro/design -o .   # the primary reproduction
./check.sh v3.31.0-preview.3                             # condition matrix
./check.sh v3.30.0                                       # control (any released version works)
```

`check.sh` switches the module's Goa version with `go get` and leaves it at the version last used.

## Environment

- Goa: `v3.31.0-preview.1` and `v3.31.0-preview.3` (all three failures occur on both)
- Released versions checked: `v3.30.0`, `v3.29.2`, `v3.28.0`, `v3.27.0`
- Go: 1.25.0 and 1.27.0, darwin/arm64 (both reproduce)
- Transport: HTTP only, so `protoc` versions do not apply
- Plugins: none, every reproduction here uses plain Goa
- Each reproduction runs a complete `goa gen`, so client and server are regenerated together

None of the three findings is described in the preview's `UPGRADING.md`.
