"""Example Playwright test for Stork UI."""

from playwright.sync_api import Page
from ..pages.login_page import LoginPage
from ..pages.navigation import Navigation
from ..pages.shared_network_page import SharedNetworkPage


def test_shared_network_edit_bug(page: Page):
    """Test the shared network edit bug."""
    login_page = LoginPage(page)
    navigation_page = Navigation(page)
    shared_page = SharedNetworkPage(page)

    page.goto("http://localhost:8080/login?returnUrl=%2Fdashboard")
    login_page.login("admin", "A123456a!")

    navigation.go_to_shared_network("esperanto")

    shared.edit_network(valid_lifetime="50", min_valid_lifetime="100")
    shared.expect_failure_toast()

    shared.edit_network(min_valid_lifetime="40")
    shared.expect_network_without_refresh()
    shared.open_shared_network("esperanto")
