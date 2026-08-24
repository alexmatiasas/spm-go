# {{PROJECT_NAME}}

Command line tool scaffolded by spm.

## Setup

```bash
uv sync
just setup
```

## Usage

```bash
just run
just test
just lint
```

Run `just --list` to see every available task.

## Git hooks

Hooks run on every commit via [prek](https://prek.j178.dev) (drop-in compatible with pre-commit):

```bash
prek install
```

## Type checking

[Pyright](https://microsoft.github.io/pyright/) checks types (`uvx pyright`). Configure it in `pyrightconfig.json`.
