from core.wrappers import Server, Kea
from core.fixtures import kea_parametrize

@kea_parametrize("agent-kea-premium-cb-database")
def test_config_backend_fetch_data(kea_service: Kea, server_service: Server):
    server_service.log_in_as_admin()
    server_service.authorize_all_machines()
    server_service.wait_for_next_machine_states()

    subnets = server_service.list_subnets()
    # 12 subnets in the JSON config and 2 subnets in the CB database.
    assert subnets.total == 12 + 2

