import re

from playwright.sync_api import Page, expect
from openapi_client.models import Leases


class LeaseSearchPage:
    def __init__(self, page: Page):
        self.page = page

    def input_search(self, text: str):
        search = self.page.get_by_label("Search leases:")
        expect(search).to_be_visible()
        search.fill(text)
        search.press("Enter")


    def expect_visible(self, leases: Leases):
        table = self.page.get_by_role("table")
        expect(table).to_be_visible()
        expect(table.get_by_role("row")).to_have_count(len(leases.items) + 1) # +1 for header row
            
        table.locator("td > a").first.click() # Click on the first lease link to expand details
        for lease in leases.items:
            row_test_id = "lease-row-{}-{}".format(lease.daemon_id, lease.id)
            expect(table.get_by_role("cell", name=lease.ip_address, exact=True)).to_be_visible()
        table.locator("td > a").first.click() # Collapse it to do not mess with other tests

    
    def lease_state_as_int (self,  state: str) -> int:
        if state == 'Valid':
            return 0
        elif state == 'Declined':
            return 1
        elif state == 'Expired/Reclaimed':
            return 2