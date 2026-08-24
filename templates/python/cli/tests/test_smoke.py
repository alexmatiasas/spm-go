"""Smoke test: the entry point must stay importable."""


def test_entrypoint_imports():
    from src.__main__ import main  # noqa: F401


def test_main_runs(capsys):
    from src.__main__ import main

    main()
    assert "hello" in capsys.readouterr().out
