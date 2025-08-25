"""
Playwright UI tests using the exact same Docker stack as system tests.
"""

import os
import sys
import subprocess
from pathlib import Path
import pytest
from playwright.sync_api import Page
from core.compose_factory import create_docker_compose  # type: ignore

cur = Path(__file__).resolve()
tests_dir = next(p for p in cur.parents if p.name == "tests")
system_dir = tests_dir / "system"
if str(system_dir) not in sys.path:
    sys.path.insert(0, str(system_dir))


STORK_BASE_URL = os.getenv("STORK_BASE_URL", "http://localhost:42080")
STORK_UI_SERVICE = os.getenv("STORK_UI_SERVICE", "server")
STORK_PROJECT = os.getenv("STORK_PROJECT", "stork_tests")
COMPOSE_FILE = str(system_dir / "docker-compose.yaml")


def _run(cmd: list[str], check: bool = True) -> subprocess.CompletedProcess:
    return subprocess.run(
        cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, check=check
    )


def _hard_cleanup():
    """Remove orphans and conflicting networks from previous runs."""
    _run(
        [
            "docker",
            "compose",
            "--project-directory",
            str(tests_dir.parent),
            "-p",
            STORK_PROJECT,
            "-f",
            COMPOSE_FILE,
            "down",
            "--remove-orphans",
            "--volumes",
        ],
        check=False,
    )

    ps = _run(["docker", "ps", "-a", "--format", "{{.ID}} {{.Names}}"], check=False)
    stray_ids = [
        line.split()[0] for line in ps.stdout.splitlines() if STORK_PROJECT in line
    ]
    if stray_ids:
        _run(["docker", "rm", "-f", *stray_ids], check=False)

    # Remove unused networks (safe: only unused)
    _run(["docker", "network", "prune", "-f"], check=False)


@pytest.fixture(scope="session")
def stork_base_url() -> str:
    """
    Bring up the SAME environment as system tests, or reuse an already running one.
    """

    os.environ.setdefault("IPWD", os.getcwd())
    os.environ.setdefault("DOCKER_DEFAULT_PLATFORM", "linux/amd64")

    if os.getenv("STORK_REUSE") == "1":
        return STORK_BASE_URL

    _hard_cleanup()

    compose = create_docker_compose()

    compose.kill()
    compose.down()

    compose.bootstrap("postgres")
    compose.bootstrap(STORK_UI_SERVICE)
    compose.bootstrap("agent-kea")

    def _try_register():
        compose.run("register", "register --non-interactive")

    try:
        _try_register()
    except Exception as e1:

        _hard_cleanup()
        compose.bootstrap("postgres")
        compose.bootstrap(STORK_UI_SERVICE)
        compose.bootstrap("agent-kea")
        try:
            _try_register()
        except Exception as e2:

            print(
                "WARN: Agent registration failed after retry. "
                "Continuing so UI tests/debugging can proceed.\n"
                f"First error: {e1}\nSecond error: {e2}"
            )

    return STORK_BASE_URL


@pytest.fixture(scope="function")
def logged_in_page(page: Page, stork_base_url: str):
    """Open login and authenticate with seeded admin credentials."""
    from tests.ui.playwright.pages.login_page import LoginPage  # type: ignore

    lp = LoginPage(page)
    lp.open(stork_base_url)
    lp.login("admin", "admin")
    return page
