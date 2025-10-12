import re
from playwright.sync_api import Page, expect, TimeoutError as PWTimeout


class LoginPage:
    """Encapsulates interactions with the login screen and forced password-change dialog."""

    def __init__(self, page: Page):
        self.page = page

    # ---------------- Navigation ----------------
    def open(self, base_url: str):
        self.page.goto(base_url, wait_until="domcontentloaded")

    # ---------------- Locators: Login form ----------------
    def username_locator(self):
        selector = (
            "input[type='email'], "
            "input[type='text'], "
            "input[formcontrolname*='login' i], "
            "input[placeholder*='login' i], "
            "input[placeholder*='email' i]"
        )
        return self.page.locator(selector).first

    def password_locator(self):
        return self.page.locator("input[type='password']").first

    def sign_in_button(self):
        btn = self.page.get_by_role(
            "button", name=re.compile(r"(sign in|log in|login)", re.I)
        )
        if not btn.count():
            btn = self.page.locator("button[type='submit']").first
        return btn

    # ---------------- Locators: Forced password-change dialog ----------------
    def old_password(self):
        return self.page.locator("#old-password")

    def new_password(self):
        return self.page.locator("#new-password")

    def confirm_password(self):
        return self.page.locator("#confirm-password")

    def save_new_password_button(self):
        return self.page.get_by_role("button", name=re.compile(r"save", re.I))

    # ---------------- Locators: Toasts / validation messages ----------------
    def toast_invalid_login(self):
        # Shown on login page when credentials are wrong
        return self.page.get_by_text("Invalid login or password")

    def toast_password_updated(self):
        # Success after saving new password
        return self.page.get_by_text("User password updated")

    def error_mismatch_confirm(self):
        # Inline validation on confirm-password field
        return self.page.get_by_text("Passwords must match.")

    def error_new_password_too_short(self):
        return self.page.get_by_text("This field value is too short.")

    def error_required_field(self):
        # All empty fields show the same text; first is sufficient to assert presence
        return self.page.get_by_text("This field is required.").first

    # ---------------- Actions ----------------
    def login(self, username: str, password: str):
        self.page.wait_for_load_state("networkidle")

        user = self.username_locator()
        pwd = self.password_locator()
        expect(user).to_be_visible(timeout=15000)
        expect(pwd).to_be_visible(timeout=15000)

        user.fill(username)
        pwd.fill(password)

        self.sign_in_button().click()

    def logout(self, username: str = None):
        """Logs out the current user."""
        if username:
            self.page.get_by_role("button", name=f"Logout ({username})").click()
        else:
            self.page.get_by_role("button", name="Logout").click()

    def is_password_change_required(self, timeout_ms: int = 2000) -> bool:
        """Detect if the forced password-change dialog is present."""
        try:

            self.old_password().wait_for(state="visible", timeout=timeout_ms)

            self.new_password().wait_for(state="visible", timeout=200)
            self.confirm_password().wait_for(state="visible", timeout=200)
            return True
        except PWTimeout as err:
            print(
                f"[login_page] Password-change dialog not detected within "
                f"{timeout_ms} ms at URL={self.page.url!r}: {err!r}"
            )
            return False

    def change_password(self, old_password: str, new_password: str):
        """Fill and save the password-change dialog."""
        self.old_password().fill(old_password)
        self.new_password().fill(new_password)
        self.confirm_password().fill(new_password)
        self.save_new_password_button().click()

    def await_dashboard(self, timeout_ms: int = 10_000):
        """ Wait until the dashboard is truly loaded. """
        try:
            self.page.wait_for_url("**/dashboard*", timeout=timeout_ms)


            menubar = self.page.locator("[data-pc-name='menubar']").first
            menubar.wait_for(state="visible", timeout=3000)


            welcome_panel = self.page.locator("[data-pc-name='panel']").filter(
                has_text=re.compile(r"^\s*Welcome to Stork!?$", re.I)
            )
            welcome_panel.first.wait_for(state="visible", timeout=3000)

        except PWTimeout:
            # Safety net: at minimum, we should be off the login form
            expect(self.password_locator()).not_to_be_visible(timeout=2000)
