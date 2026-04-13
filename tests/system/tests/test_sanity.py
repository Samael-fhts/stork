from core.wrappers import WebUI


def test_login_screen_version(webui_service: WebUI):

    assert webui_service.url == "http://localhost:40081/"
