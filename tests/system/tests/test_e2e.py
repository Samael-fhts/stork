from core.wrappers import WebUI
from playwright.sync_api import Page, BrowserContext, expect
import pytest
import re



def test_e2e_version(webui_service: WebUI, page: Page):

    server = webui_service.server()
    version = server.read_version().version

    page.goto(webui_service.url)
    expect(page.get_by_text("version:")).to_be_visible()
    expect(page.locator("app-login-screen")).to_contain_text("version: {}".format(version))

def test_e2e_inject_cookie(webui_service: WebUI, context: BrowserContext):
    server = webui_service.server()
    server.log_in_as_admin()
    cookie = server.session_cookie()
    print("Session cookie: {}".format(cookie))

    chunks = cookie.split(";")
    cookie_dict = {}
    for chunk in chunks:
        if "=" in chunk:
            key, value = chunk.strip().split("=", 1)
            cookie_dict[key] = value

    context.add_cookies([{"name": "session", "value": cookie_dict["session"], "url": webui_service.url}])
    page = context.new_page()
    page.goto(webui_service.url)

def test_e2e_login(webui_service: WebUI, page: Page):
    page.goto(webui_service.url)

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

@pytest.mark.skip
def test_e2e_codegen(webui_service: WebUI):

    print("playwright codegen {}".format(webui_service.url))
    while True:
        pass

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