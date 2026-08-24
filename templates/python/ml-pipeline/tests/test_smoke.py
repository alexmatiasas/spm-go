"""Smoke test: the pipeline module must stay importable."""


def test_pipeline_imports():
    from src.pipeline import run  # noqa: F401
