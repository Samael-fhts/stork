from core.wrappers.compose import ComposeServiceWrapper


class WebUI(ComposeServiceWrapper):
    def __init__(self, compose, service_name, server_service):
        super().__init__(compose, service_name)

        internal_port = 81
        mapped = self._compose.port(service_name, internal_port)
        self.url = f"http://{mapped[0]}:{mapped[1]}/"
        self._server_service = server_service

    def server(self):
        return self._server_service
    
    def log_in_as_admin(self):
        """Logs in an admin. Changes the password for the user. Returns the user info."""
        user = self._server_service.log_in_as_admin() 
        self._server_service.change_password(user.id, "admin", "r+YB4E3T['5n4RcShcw-")
        return user
    
    def inject_session_cookie(self, context):
        """Injects the session cookie into the browser context."""
        cookie = self._server_service.session_cookie()
        if cookie is None or "session" not in cookie:
            raise Exception("No session cookie found. Log in first.")
        chunks = cookie.split(";")
        cookie_dict = {}
        for chunk in chunks:
            if "=" in chunk:
                key, value = chunk.strip().split("=", 1)
                cookie_dict[key] = value

        context.add_cookies([{
            "name": "session", 
            "value": cookie_dict["session"], 
            "url": self.url, 
            }])
