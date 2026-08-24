# {{PROJECT_NAME}}

HTTP service scaffolded by spm.

## Setup

Requires Go 1.22+ and [lefthook](https://lefthook.dev) for git hooks.

## Usage

```bash
just run       # build and start cmd/server
just test      # race-enabled test suite
just lint      # golangci-lint
```

Run `just --list` to see every available task.

## Git hooks

```bash
lefthook install
```
