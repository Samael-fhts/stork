from core.wrappers.compose import ComposeServiceWrapper



class WebUI(ComposeServiceWrapper):
    def __init__(self, compose, service_name, server_service):
        super().__init__(compose, service_name)

        internal_port = 81
        mapped = self._compose.port(service_name, internal_port)
        self.url = f"http://{mapped[0]}:{mapped[1]}/"
        self._server_service = server_service
