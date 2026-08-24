# {{PROJECT_NAME}}

ML pipeline scaffolded by spm.

## Setup

```bash
uv sync
just setup
```

## Usage

```bash
just test
just lint
```

Run `just --list` to see every available task. Pipeline logic lives in `src/pipeline.py`.

## Git hooks

Hooks run on every commit via [prek](https://prek.j178.dev) (drop-in compatible with pre-commit):

```bash
prek install
```

## Type checking

[Pyright](https://microsoft.github.io/pyright/) checks types (`uvx pyright`). Configure it in `pyrightconfig.json`.
