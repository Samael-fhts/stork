from core.wrappers.compose import ComposeServiceWrapper



class WebUI(ComposeServiceWrapper):
    def __init__(self, compose, service_name, server_service):
        super().__init__(compose, service_name)
        self.server_service = server_service