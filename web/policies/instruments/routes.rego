package sargassum.pslive.web.policies.instruments

import future.keywords

scope := {
	"/instruments",
	"/instruments/**",
}

default allow := false

# TODO: implement authz checks on these routes, and ensure the instrument exists
routes := {
	{"method": "POST", "path": "/instruments"},
	{"method": "POST", "path": "/instruments/*"},
	{"method": "POST", "path": "/instruments/*/name"},
	{"method": "POST", "path": "/instruments/*/description"},
	{"method": "POST", "path": "/instruments/*/cameras"},
	{"method": "POST", "path": "/instruments/*/cameras/*"},
	{"method": "POST", "path": "/instruments/*/controllers"},
	{"method": "POST", "path": "/instruments/*/controllers/*"},
	{"method": "POST", "path": "/instruments/*/controllers/*/pump"},
	{"method": "POST", "path": "/instruments/*/chat/messages"},
}

allow if input.operation.method == "GET"

# TODO: SUB to presence and and pump chat should ensure the instrument and controller exist
allow if input.operation.method == "SUB"

# TODO: UNSUB from presence and and pump chat should ensure the instrument and controller exist
allow if input.operation.method == "UNSUB"

# TODO: MSG on presence and chat and pump should ensure the instrument and controller exist
allow if input.operation.method == "MSG"

# TODO: PUB to pump should ensure the instrument exists
allow if input.operation.method == "PUB"

errors contains "not implemented" if {
	not allow
}
