from playwright.sync_api import Page, expect


def test_example(page: Page) -> None:
    page.goto("/")
    assert page.title() == "Stork"
    with page.expect_popup() as page1_info:
        page.get_by_role("link", name="ISC Kea").click()

    page1 = page1_info.value
    assert page1.title() == "Internet Systems Consortium"
    page1.get_by_role("img", name="Kea logo").click()

    page.locator("#identifier").click()
    page.locator("#identifier").click()
    page.locator("#identifier").fill("admin")
    page.locator("#identifier").press("Tab")
    page.locator("#secret").fill("admin")
    page.get_by_role("button", name="Sign In").click()
    page.get_by_role("heading", name="DHCPv4").click()
    page.get_by_role("heading", name="DHCPv6").click()
    page.get_by_role("button", name="Logout (admin)").click()

# @pytest.mark.skip_browser("firefox")
# @pytest.mark.only_browser("firefox")
def test_example_that_will_fail(page: Page) -> None:
    ## let's do some stuff so we could see something in collected logs
    page.goto("/")
    page.locator("#identifier").click()
    page.locator("#identifier").fill("admin")
    page.locator("#identifier").press("Tab")
    page.locator("#secret").fill("admin")
    page.get_by_role("button", name="Sign In").click()
    page.get_by_role("menuitem", name="Services").click()
    page.get_by_role("menuitem", name=" Machines").click()
    # with second execution of this test, this will fail
    page.get_by_role("button", name="Unauthorized (8)").click()
    page.get_by_role("row", name="agent-kea agent-kea:").locator("#show-machines-menu").click()
    page.get_by_role("menuitem", name=" Authorize").click()
    page.get_by_role("menuitem", name="DHCP").click()
    page.get_by_role("menuitem", name=" Dashboard").click()
    page.get_by_role("heading", name="DHCPv4").click()
    assert page.title() == "Stork"