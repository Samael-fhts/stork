
"""
Run Playwright UI tests against the same stable Docker Compose stack
used by the system tests.
"""

import os
import sys
from pathlib import Path
import pytest
from playwright.sync_api import Page


cur = Path(__file__).resolve()


tests_dir = cur.parents[3]              
system_dir = tests_dir / "system"        
if not (system_dir / "core").exists():
    # Fallback in case the repo layout differs slightly
    alt = cur.parents[4] / "tests" / "system"
    if (alt / "core").exists():
        system_dir = alt
if str(system_dir) not in sys.path:
    sys.path.insert(0, str(system_dir))


from core.compose_factory import create_docker_compose  # type: ignore


STORK_BASE_URL = os.getenv("STORK_BASE_URL", "http://localhost:42080")

@pytest.fixture(scope="session", autouse=True)
def stork_stack():
    """
    Bring up a fresh stack before the UI session starts, and tear it down after.
    Mirrors system tests' lifecycle: kill/down -> bootstrap -> wait healthy.
    """
    compose = create_docker_compose()

    compose.kill()
    compose.down()

    compose.bootstrap("server")
    compose.wait_for_operational("server")
    compose.wait_for_operational("postgres")

   
    compose.bootstrap("agent-kea")
    compose.wait_for_operational("agent-kea")
    compose.bootstrap("agent-bind9")
    compose.wait_for_operational("agent-bind9")

    yield

    compose.kill()
    compose.down()


@pytest.fixture(scope="session")
def stork_base_url() -> str:
    return STORK_BASE_URL


@pytest.fixture(scope="function")
def logged_in_page(page: Page, stork_base_url: str):
    page.goto(f"{stork_base_url}/login?returnUrl=%2Fdashboard")
    from tests.ui.playwright.pages.login_page import LoginPage  # type: ignore
    LoginPage(page).login("admin", "admin")
    return page
