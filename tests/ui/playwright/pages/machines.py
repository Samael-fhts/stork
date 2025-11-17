from playwright.sync_api import Page, expect
import re


class MachinesPage:
    def __init__(self, page: Page):
        self.page = page

    # Navigation: Navigation → Services → Machines
    def open(self):
        self.page.get_by_role("button", name="Navigation").click()
        self.page.locator("a").filter(has_text="Services").click()
        self.page.get_by_role("link", name=" Machines").click()
        # Verify we landed on Machines tab
        self.page.get_by_role("tab", name="Machines").click()

    # Sections: Unauthorized / Authorized
    def switch_to_unauthorized(self):
        self.page.locator("#unauthorized-select-button").get_by_text(
            "Unauthorized"
        ).click()

    def switch_to_authorized(self):
        self.page.locator("#unauthorized-select-button").get_by_text(
            "Authorized", exact=True
        ).click()

    def search(self, text: str):
        self.page.get_by_role("textbox", name="Search machines").click()
        self.page.get_by_role("textbox", name="Search machines").fill(text)

    def clear_filters(self):
        self.page.get_by_role("button", name=re.compile(r"\bClear\b", re.I)).click()

    def refresh_list(self):
        self.page.get_by_role(
            "button", name=re.compile(r"\bRefresh\s+List\b", re.I)
        ).click()

    def expect_no_results_row(self):
        expect(
            self.page.get_by_role(
                "cell", name=re.compile(r"No machines found\.\s*Clear", re.I)
            )
        ).to_be_visible(timeout=3000)

    def click_clear_in_no_results_row(self):
        row = self.page.get_by_role(
            "row", name=re.compile(r"No machines found\.\s*Clear", re.I)
        )
        row.get_by_role("button", name=re.compile(r"\bClear\b", re.I)).click()

    def select_machine_row(self, row_text: str):
        self.page.get_by_role("row", name=row_text).get_by_role("checkbox").check()

    def authorize_selected(self):
        self.page.get_by_role(
            "button", name=re.compile(r"\bAuthorize\s+selected\b", re.I)
        ).click()

    def open_machine(self, link_text: str):
        self.page.get_by_role("cell", name=link_text).click()

    # Detail page verifications and actions
    def expect_detail_headings(self):
        self.page.get_by_role("heading", name="System Information").click()
        self.page.get_by_role("heading", name="Applications").click()
        self.page.get_by_role("heading", name="Events").click()

    def expect_detail_ip_fragment(self, fragment: str):
        self.page.get_by_text(fragment).first.click()

    def get_latest_state(self):
        self.page.get_by_role(
            "button", name=re.compile(r"\bGet\s+Latest\s+State\b", re.I)
        ).click()

    def dump_troubleshooting(self):
        with self.page.expect_download() as download_info:
            self.page.get_by_role(
                "button", name=re.compile(r"\bDump\s+Troubleshooting\s+Data\b", re.I)
            ).click()
        return download_info.value

    def back_to_machines_tab(self):
        self.page.get_by_role("tab", name="Machines").click()

    def wait_for_row(self, row_text: str, timeout_ms: int = 3000):
        """Ensure the target row is present before acting on it."""
        self.page.get_by_role("row", name=row_text, exact=True).wait_for(
            timeout=timeout_ms
        )

    def open_actions_menu(self):
        self.page.locator("#show-machines-menu-1").click()

    def actions_authorize_from_menu(self):
        self.page.get_by_role("menuitem", name="Authorize").locator("a").click()

    def actions_refresh_state_from_menu(self):
        self.page.get_by_role("menuitem", name="Refresh machine state").locator(
            "a"
        ).click()

    def actions_download_archive_from_menu(self):
        with self.page.expect_download() as dl:
            self.page.get_by_title("Download data archive for").click()
        return dl.value

    def actions_remove_machine_from_menu(self):
        self.page.get_by_title("Remove machine from Stork").click()
        self.page.get_by_text("Confirm", exact=True).click()
        self.page.get_by_role("button", name=re.compile(r"^\s*Yes\s*$", re.I)).click()

    # Installing Stork Agent dialog
    def open_install_dialog(self):
        self.page.get_by_role(
            "button", name=re.compile(r"Installing\s+Stork\s+Agent", re.I)
        ).click()

    def expect_install_dialog_title(self):
        expect(
            self.page.get_by_text("Agent Installation Instructions", exact=True)
        ).to_be_visible(timeout=3000)

    def assert_docs_link_opens_new_tab(self):
        """Clicks the “the Stork agent installation” link in the Install Agent dialog
        and asserts that it opens in a new tab/window.
        Scope:
        - Verifies that a popup is created (new Page).
        - Waits for the popup to reach 'domcontentloaded'.
        - Closes the popup.
        Out of scope:
        - No content/URL validation of the target page (we only verify the redirect occurred).
        """

        with self.page.expect_popup() as popup_info:
            self.page.get_by_role("link", name="the Stork agent installation").click()
        popup = popup_info.value
        try:
            popup.wait_for_load_state("domcontentloaded", timeout=3000)
        finally:
            popup.close()

    def expect_wget_snippet_visible(self):
        expect(self.page.get_by_text("wget http://localhost:42080/")).to_be_visible(
            timeout=3000
        )

        """Asserts that the Install Agent dialog shows the shell snippet starting with:
        'wget http://localhost:42080/'
         This confirms the command block is rendered for the local controller."""

        expect(self.page.get_by_text("wget http://localhost:42080/")).to_be_visible(
            timeout=3000
        )

    def click_copy_first(self):
        self.page.get_by_role("button", name="").first.click()

    def regenerate_token_and_wait(self):
        """Regenerates the server token and verifies the result without exposing the token:
        1) Snapshot current token value from the Agent Installation dialog input.
        2) Click 'Regenerate' and assert PUT /api/machines-server-token succeeds.
        3) Read the new token and assert it is non-empty and different.
        4) Click 'Copy server token to clipboard' and assert clipboard == new token.

          Note: token values are never printed or logged."""
        # 1) read current token
        token_input = (
            self.page.get_by_role("dialog", name="Agent Installation")
            .locator("input")
            .first
        )
        old_token = token_input.input_value()

        # 2) regenerate and assert backend call
        with self.page.expect_response(
            lambda r: r.request.method == "PUT"
            and r.url.endswith("/api/machines-server-token")
        ) as resp_info:
            self.page.get_by_role("button", name=" Regenerate").click()
        resp = resp_info.value
        assert (
            resp.ok
        ), f"Regenerate token failed: {resp.status} {getattr(resp, 'status_text', lambda: '')()}"

        # 3) verify token changed
        new_token = token_input.input_value()
        assert new_token, "New token is empty"
        assert new_token != old_token, "Token was not regenerated (value unchanged)"

        # 4) verify clipboard copy matches the new token
        self.page.context.grant_permissions(["clipboard-read", "clipboard-write"])
        self.page.locator("[ptooltip='Copy server token to clipboard']").click()
        clipboard_value = self.page.evaluate("navigator.clipboard.readText()")
        assert (
            clipboard_value == new_token
        ), "Copied token does not match the current token"

    def close_install_dialog(self):
        self.page.get_by_role("button", name=re.compile(r"\bClose\b", re.I)).click()

    def expect_dhcp_badges_on_row(self, row_key: str):
        row = self.page.get_by_role("row", name=row_key, exact=True)
        expect(row.get_by_role("cell", name=re.compile(r"DHCPv4.*DHCPv6.*CA", re.I))).to_be_visible(timeout=3000)

    def open_app_from_badges_cell(self, row_key: str):
        self.page.get_by_role("cell", name=re.compile(r"DHCPv4.*DHCPv6.*CA", re.I)).click()

    def expect_app_tabs_present(self):
        # Left-side daemon tabs
        for tab in ("DHCPv4", "DHCPv6", "CA"):
            expect(self.page.get_by_role("tab", name=re.compile(rf"\b{tab}\b", re.I))).to_be_visible(timeout=3000)
        # Main action buttons row
        for btn in ("Host Reservations", "Subnets", "Shared Networks", "Global Configuration", "Raw configuration"):
            expect(self.page.get_by_role("button", name=re.compile(rf"\b{btn}\b", re.I))).to_be_visible(timeout=3000)

    def app_click_refresh(self):
        btn = self.page.get_by_role("button", name=re.compile(r"\bRefresh\s+App\b", re.I))
        expect(btn).to_be_visible(timeout=3000)
        btn.click()

    def app_toggle_monitoring_off_on(self):
        sw = self.page.get_by_role("switch", name="Monitoring")
        sw.uncheck()
        sw.check()

    def app_open_dhcp4(self):
        self.page.get_by_role("tab", name=re.compile(r"\bDHCPv4\b", re.I)).click()

    def app_open_host_reservations(self):
        self.page.get_by_role("button", name=re.compile(r"\bHost Reservations\b", re.I)).click()

    def host_reservations_expect_loaded(self):
        # Verify the left nav item is present
        expect(
            self.page.get_by_role("list").locator("a").filter(has_text="Host Reservations")
        ).to_be_visible(timeout=3000)

    # -------- Host Reservations: dialogs & actions --------
    def host_click_migrate_to_db_and_cancel(self):
        self.page.get_by_role("button", name=re.compile(r"\bMigrate to Database\b", re.I)).click()
        expect(self.page.get_by_text("Migrate host reservations to database")).to_be_visible(timeout=3000)
        self.page.get_by_role("button", name=re.compile(r"^\s*No\s*$", re.I)).click()

    def host_click_new_host_expect_tx_error_then_back(self):
        self.page.get_by_role("button", name=re.compile(r"\bNew Host\b", re.I)).click()
        expect(self.page.get_by_text("Cannot create new transaction")).to_be_visible(timeout=3000)
        self.page.get_by_role("button", name=re.compile(r"^\s*Back\s*$", re.I)).click()

    # -------- Host Reservations: filtering & totals --------
    def host_filter_check_global_conflict(self):
        self.page.get_by_role("checkbox", name="Global Conflict").check()

    def host_expect_total_hosts_text(self, n: int):
        expect(self.page.get_by_text(f"Total: {n} hosts")).to_be_visible(timeout=3000)

    # -------- Host Reservations: table row & details --------
    def host_click_first_row_link(self, link_text: str):
        # e.g., "hw-address=(00:01:02:03:04:02)"
        self.page.get_by_role("link", name=link_text).click()

    def host_detail_expect_sections(self):
        # Verify the host detail sections are present
        expect(self.page.get_by_text("[1] Host in subnet 192.0.2.0/")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="DHCP Servers Using the Host")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="DHCP Identifiers")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="Hostname /  All Servers")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="IP Reservations /  All Servers")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="Boot Fields /  All Servers")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="Client Classes /  All Servers")).to_be_visible(timeout=3000)

    def host_click_leases_then_expect_dhcp_options_present(self):
        self.page.get_by_role("button", name=re.compile(r"\bLeases\b", re.I)).click()
        expect(self.page.get_by_role("button", name=re.compile(r"\bDHCP Options\s*/\s* All Servers\b", re.I))).to_be_visible(timeout=3000)

    def host_click_refresh_list(self):
        self.page.get_by_role("button", name=re.compile(r"\bRefresh\s+List\b", re.I)).click()

    def app_open_subnets(self):
        self.page.get_by_role("button", name=re.compile(r"\bSubnets\b", re.I)).click()

    def subnets_expect_loaded(self):
        # Verify the Subnets tab is visible
        expect(self.page.get_by_role("tab", name=re.compile(r"\bSubnets\b", re.I))).to_be_visible(timeout=3000)

    # -------- Subnets: totals, search, open result --------
    def subnets_expect_total(self, n: int):
        # Matches the footer text e.g. "Total: 9 subnets"
        expect(self.page.get_by_text(f"Total: {n} subnets")).to_be_visible(timeout=3000)

    def subnets_search(self, query: str):
        self.page.get_by_role("textbox", name="Search IP or identifier").click()
        self.page.get_by_role("textbox", name="Search IP or identifier").fill(query)

    def subnets_open_search_result(self):
        self.page.get_by_text("-192.0.5.50").click()
        self.page.get_by_role("link", name="/24").click()

    # -------- Subnet detail: header, edit/back, sections --------
    def subnets_detail_expect_header(self, text: str):
        expect(self.page.get_by_text(text)).to_be_visible(timeout=3000)

    def subnets_click_edit_expect_tx_error_then_back(self):
        self.page.get_by_role("button", name=re.compile(r"\bEdit\b", re.I)).click()
        expect(self.page.get_by_text("Cannot create new transaction")).to_be_visible(timeout=3000)
        self.page.get_by_role("button", name=re.compile(r"^\s*Back\s*$", re.I)).click()

    def subnets_detail_expect_sections(self):
        # Verify the subnet detail sections are present
        expect(self.page.get_by_role("group", name="DHCP Servers Using the Subnet")).to_be_visible(timeout=3000)
        expect(self.page.get_by_text("Pools / All Servers")).to_be_visible(timeout=3000)
        expect(self.page.get_by_text("-192.0.5.50")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="Statistics")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="User Context /  All Servers")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="DHCP Parameters")).to_be_visible(timeout=3000)
        expect(self.page.get_by_role("group", name="DHCP Options /  All Servers")).to_be_visible(timeout=3000)

    # -------- Subnets: back/clear/new/refresh --------
    def subnets_back_to_tab(self):
        self.page.get_by_role("tab", name="Subnets").click()

    def subnets_click_new_subnet_expect_error_then_back(self):
        self.page.get_by_role("button", name=re.compile(r"\bNew\s+Subnet\b", re.I)).click()
        expect(self.page.get_by_text("Cannot create new transaction")).to_be_visible(timeout=3000)
        self.page.get_by_role("button", name=re.compile(r"^\s*Back\s*$", re.I)).click()

    def subnets_click_refresh_list(self):
        self.page.get_by_role("button", name=re.compile(r"\bRefresh\s+List\b", re.I)).click()

