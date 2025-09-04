import openapi_client
from openapi_client.api.users_api import UsersApi
from openapi_client.api.services_api import ServicesApi
from openapi_client.api.dhcp_api import DHCPApi

def test_primary():
    url = "http://localhost:8080/api"
    configuration = openapi_client.Configuration(host=url)
    api_client = openapi_client.ApiClient(configuration)

    api_instance = UsersApi(api_client)
    http_info = api_instance.create_session_with_http_info(
        credentials={
            "identifier": "admin",
            "secret": "admin",
            "authentication_method_id": "internal",
        },
    )
    headers = http_info.headers
    user = http_info.data

    session_cookie = headers["Set-Cookie"]
    api_client.cookie = session_cookie

    params = {"start": 0, "limit": 10}
    api_instance = ServicesApi(api_client)
    machines = api_instance.get_machines(**params)

    for machine in machines.items:
        machine.authorized = True

        api_instance.update_machine(
            id=machine.id,
            machine=machine,
        )


    state = api_instance.get_machine_state(id=machines.items[0].id)
    assert state is not None

    params = {"text": "192.0.2.1"}
    api_instance = DHCPApi(api_client)
    leases = api_instance.get_leases(**params)
    assert leases.total == 1
