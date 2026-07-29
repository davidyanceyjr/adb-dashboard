# Cycle History

| Date | Cycle ID | Mode | Slice | Result | Focused evidence | Broad evidence | Commit |
|---|---|---|---|---|---|---|---|
| 2026-07-29 | CYCLE-20260729-M1-S1 | feature | M1-S1 | committed | `GOCACHE=$PWD/.codex/cache/go-build GOMODCACHE=$PWD/.codex/cache/go-mod go test -count=1 ./...` exited `0`; real CLI path exercised for help, version, `--version`, unknown command, unknown option, unknown option after global option, and missing argument | `go test -race ./...` exited `0`; `go vet ./...` exited `0` with repo-local Go caches | `bae1abe108150bea75aecce76699de2f7aee8ef7` |
| 2026-07-29 | CYCLE-20260729-M1-S2 | feature | M1-S2 | CYCLE READY | `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -count=1 ./...` exited `0`; real CLI path exercised for `doctor` with config file, environment, and CLI overrides; data/temp directories created; fake `adb` and `xdg-open` were not invoked | `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go test -race ./...` exited `0`; `GOPATH=$PWD/.codex/cache/go-path GOCACHE=$PWD/.codex/cache/go-build go vet ./...` exited `0` | `a45785cabf6dd57c102c1b9bf7bce0fac334b578` |
