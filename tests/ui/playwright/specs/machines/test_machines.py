import os
import pytest
from tests.ui.playwright.pages.login_page import LoginPage
from tests.ui.playwright.pages.machines import MachinesPage

BASE_URL = os.getenv("STORK_BASE_URL", "http://localhost:42080")
ADMIN_USER = os.getenv("STORK_ADMIN_USER", "admin")
ADMIN_PASS = os.getenv("STORK_ADMIN_PASS", "admin")
NEW_ADMIN_PASS = os.getenv("STORK_NEW_PASS", "A123456a!")


@pytest.mark.ui
@pytest.mark.usefixtures("clean_env")
def test_machines_unauthorized_to_authorized_flow(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()

    # Navigate to Machines
    mp.open()

    # Go to Unauthorized
    mp.switch_to_unauthorized()

    # Negative search
    mp.search("182")
    mp.expect_no_results_row()
    mp.click_clear_in_no_results_row()

    # Positive search (present)
    mp.search("172")
    mp.select_machine_row("172.42.42.100:8080")

    # Authorize
    mp.authorize_selected()

    # Switch to Authorized, clear and refresh to match your steps
    mp.switch_to_authorized()
    mp.clear_filters()
    mp.refresh_list()

    # Negative search
    mp.search("negativetest")
    mp.expect_no_results_row()
    mp.click_clear_in_no_results_row()

    # Search actual authorized machine and open it
    mp.search("agent")
    mp.open_machine("agent-kea")

    # Verify elements on the Machines details page
    mp.expect_detail_headings()
    mp.expect_detail_ip_fragment("172.42.42.100:")

    # Get latest state
    mp.get_latest_state()

    # Back and clear filters
    mp.back_to_machines_tab()
    mp.clear_filters()

    # Logout
    lp.logout("admin")


@pytest.mark.ui
@pytest.mark.usefixtures("clean_env")
def test_machines_authorize_via_actions_and_cleanup(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()

    # Navigate: Navigation → Services → Machines → Unauthorized
    mp.open()
    mp.switch_to_unauthorized()

    row_key = "172.42.42.100:8080"

    # Select row and authorize via row-scoped Actions menu
    mp.select_machine_row(row_key)
    mp.wait_for_row(row_key)
    mp.open_actions_menu()
    mp.actions_authorize_from_menu()

    # Switch to Authorized and perform actions: refresh, download, then remove
    mp.switch_to_authorized()
    mp.wait_for_row(row_key)

    mp.open_actions_menu()
    mp.actions_refresh_state_from_menu()

    mp.open_actions_menu()
    mp.actions_download_archive_from_menu()

    mp.open_actions_menu()
    mp.actions_remove_machine_from_menu()

    # Logout
    lp.logout("admin")


@pytest.mark.ui
def test_machines_installing_agent_dialog(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()
    mp.open()

    # Open install dialog
    mp.open_install_dialog()
    mp.expect_install_dialog_title()

    mp.assert_docs_link_opens_new_tab()

    # Verify the command snippet, try the copy button, regenerate token, then close
    mp.expect_wget_snippet_visible()
    mp.click_copy_first()
    mp.regenerate_token_and_wait()
    mp.close_install_dialog()

    # Logout
    lp.logout("admin")

@pytest.mark.ui
def test_machines_dhcp_badges_and_app_tabs_present(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login and reach Machines page
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()
    mp.open()

    row_key = "172.42.42.100:8080"

    # Verify badges on Machines row
    mp.expect_dhcp_badges_on_row(row_key)

    # Enter the App details by clicking the badges cell
    mp.open_app_from_badges_cell(row_key)

    # Verify presence only (no tab clicks)
    mp.expect_app_tabs_present()

    # Click refresh and toggle monitoring
    mp.app_click_refresh()
    mp.app_toggle_monitoring_off_on()

    lp.logout("admin")

@pytest.mark.ui
def test_machines_host_reservations_filters_and_sections(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login and reach Machines
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()
    mp.open()

    row_key = "172.42.42.100:8080"
    mp.open_app_from_badges_cell(row_key)
    mp.app_open_dhcp4()
    mp.app_open_host_reservations()
    mp.host_reservations_expect_loaded()

    # Migrate to DB dialog -> verify pop-up window -> cancel
    mp.host_click_migrate_to_db_and_cancel()

    # New Host -> expect transaction error -> Back to Host Reservations tab
    mp.host_click_new_host_expect_tx_error_then_back()

    # Filters: check "Global Conflict" → expect total 3 hosts
    mp.host_filter_check_global_conflict()
    mp.host_expect_total_hosts_text(3)

    # Clear filters
    mp.clear_filters()

    # Click first row link, verify sections, click Leases and ensure DHCP Options exists
    mp.host_click_first_row_link("hw-address=(00:01:02:03:04:02)")
    mp.host_detail_expect_sections()
    mp.host_click_leases_then_expect_dhcp_options_present()

    # Back to Host Reservations tab and refresh list
    page.get_by_role("tab", name="Host Reservations").click()
    mp.host_click_refresh_list()

    # Logout
    lp.logout("admin")

@pytest.mark.ui
def test_machines_subnets_filters_and_detail_flow(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login and reach Machines
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()
    mp.open()

    # Enter app from the badges cell, then open Subnets
    row_key = "172.42.42.100:8080"
    mp.open_app_from_badges_cell(row_key)
    mp.app_open_subnets()
    mp.subnets_expect_loaded()

    # Expect total 9 subnets
    mp.subnets_expect_total(9)

    # Search "192.0.5.0/24", open the resulting subnet
    mp.subnets_search("192.0.5.0/24")
    mp.subnets_open_search_result()

    # Verify detail header text and attempt Edit -> error -> Back
    mp.subnets_detail_expect_header("Subnet 192.0.5.0/24 in shared")
    mp.subnets_click_edit_expect_tx_error_then_back()

    # Verify all the specified sections
    mp.subnets_detail_expect_sections()

    # Back to Subnets, clear filters, New Subnet -> error -> Back, Refresh List
    mp.subnets_back_to_tab()
    mp.clear_filters()
    mp.subnets_click_new_subnet_expect_error_then_back()
    mp.subnets_click_refresh_list()

    lp.logout("admin")

@pytest.mark.ui
def test_machines_shared_networks_filters_and_detail_flow(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login and reach Machines
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()
    mp.open()

    # Enter app via badges cell, then open Shared Networks
    row_key = "172.42.42.100:8080"
    mp.open_app_from_badges_cell(row_key)
    mp.app_open_shared_networks()
    mp.shared_networks_expect_loaded()

    # New Shared Network -> expect error -> Back
    mp.shared_click_new_shared_network_expect_error_then_back()

    # Expect total 2 shared networks
    mp.shared_expect_total(2)

    # Search for "frog" and open it
    mp.shared_search("frog")
    mp.shared_open_result_by_name("frog")

    # Verify detail header, Edit -> error -> Back
    mp.shared_detail_expect_header("frog")
    mp.shared_click_edit_expect_error_then_back()

    # Verify presence of all visbile sections
    mp.shared_detail_expect_sections()

    # Back to Shared Networks, Clear filters, Refresh list
    mp.shared_back_to_tab()
    mp.clear_filters()
    mp.shared_click_refresh_list()

    lp.logout("admin")

@pytest.mark.ui
def test_machines_global_configuration_edit_flow(page):
    lp = LoginPage(page)
    mp = MachinesPage(page)

    # Login and reach Machines
    lp.open(BASE_URL)
    lp.login(ADMIN_USER, NEW_ADMIN_PASS if NEW_ADMIN_PASS else ADMIN_PASS)
    lp.await_dashboard()
    mp.open()

    # Enter the DHCP app from badges cell
    row_key = "172.42.42.100:8080"
    mp.open_app_from_badges_cell(row_key)

    # Open Global Configuration and verify sections
    mp.app_open_global_configuration()
    mp.global_config_expect_sections()

    # Edit: verify edit sections, add options, delete one, submit, expect success
    mp.global_config_click_edit()
    mp.global_config_expect_edit_sections()
    mp.global_config_add_more_options()
    mp.global_config_delete_option_at(5)
    mp.global_config_submit()
    mp.global_config_expect_submit_success()

    # Back to kea page, then logout
    mp.global_config_back_to_kea()
    lp.logout("admin")
