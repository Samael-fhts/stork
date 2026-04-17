from core.wrappers import WebUI
from playwright.sync_api import Page, expect
import pytest
import re



def test_e2e_version(webui_service: WebUI):

    server = webui_service.server()
    version = server.read_version().version

    page = webui_service.new_page()

    expect(page.get_by_text("version:")).to_be_visible()
    expect(page.locator("app-login-screen")).to_contain_text("version: {}".format(version))


def test_e2e_version_popup(webui_service: WebUI):
    """Checks if the version tooltip is shown and contains correct version information."""
    server = webui_service.server()
    version = server.read_version().version
    webui_service.log_in_as_admin()
    page = webui_service.new_page()

    page.get_by_role("link").filter(has_text=re.compile(r"^$")).hover()
    tooltip = page.get_by_role("tooltip")

    expect(tooltip).to_be_visible()
    expect(tooltip).to_contain_text("Version: {}".format(version))

def test_e2e_login_api(webui_service: WebUI):
    webui_service.log_in_as_admin()

    page = webui_service.new_page()

    expect(page.get_by_role("button", name="Logout")).to_be_visible()


def test_e2e_custom_not_found_page(webui_service: WebUI):
    """Checks that the custom 404 page is shown when navigating to a non-existent path. And user is able to navigate
       back to the dashboard using the link on 404 page."""

    webui_service.log_in_as_admin() 
    page = webui_service.new_page("/not/existent/path") 

    expect(page.get_by_role("alert")).to_contain_text("Page Not Found")
    goto_dashboard = page.get_by_role("link", name="Go to Dashboard page")
    expect(goto_dashboard).to_be_visible()
    goto_dashboard.click()
    expect(page.get_by_text("Welcome to Stork!")).to_be_visible()


def test_e2e_login_ui(webui_service: WebUI, page: Page):
    page.goto(webui_service._url)

    dropdown = page.get_by_role("button", name="dropdown trigger")

    dropdown.click()
    page.get_by_text("Internal").click()

    username = username_locator(page)
    expect(username).to_be_visible()
    username.fill("admin")
    expect(username).to_have_value("admin")
    password = password_locator(page)
    expect(password).to_be_visible()
    password.fill("admin")
    expect(password).to_have_value("admin")
    sign_in_button = page.get_by_role("button", name=re.compile(r"(sign in|log in|login)", re.I))
    sign_in_button.click()

    page.locator("#old-password").fill("admin")
    page.locator("#new-password").fill("r+YB4E3T['5n4RcShcw-")
    page.locator("#confirm-password").fill("r+YB4E3T['5n4RcShcw-")
    page.get_by_role("button", name=" Save").click()

    expect(page.get_by_role("button", name="Logout")).to_be_visible()

def username_locator(page: Page):
    selector = (
        "input[type='email'], "
        "input[type='text'], "
        "input[formcontrolname*='login' i], "
        "input[placeholder*='login' i], "
        "input[placeholder*='email' i]"
    )
    return page.locator(selector).first

def password_locator(page: Page):
    return page.locator("input[type='password']").first