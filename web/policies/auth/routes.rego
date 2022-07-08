package sargassum.pslive.web.policies.auth

import future.keywords

scope := {
	"/csrf",
	"/login",
	"/login?return=**",
	"/sessions",
}

allow if input.operation.method == "GET"

allow if {
	input.operation.method in {"GET", "POST"}
	input.resource.path == "/sessions"
}

errors contains "unallowed method" if not allow
