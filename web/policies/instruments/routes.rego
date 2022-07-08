package sargassum.pslive.web.policies.instruments

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

scope := {
	"/instruments",
	"/instruments/**",
}

default allow := false

allow if input.operation.method == "GET"

# TODO: SUB to presence and and pump chat should ensure the instrument and controller exist
allow if input.operation.method == "SUB"

# TODO: MSG on presence and chat and pump should ensure the instrument and controller exist
allow if input.operation.method == "MSG"

errors contains "unauthenticated user" if {
	input.operation.method == "POST"
	not auth.is_authenticated(input.subject)
}

# TODO: reduce repetition in these rules

# TODO: POST should ensure the instrument exists
# TODO: check admin role
allow if {
	input.operation.method == "POST"
	input.resource.path == "/instruments"
	auth.is_authenticated(input.subject)
}

# TODO: POST should ensure the instrument exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/name", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/description", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument and camera exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/cameras", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument and camera exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/cameras/*", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument and controller exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/controllers", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument and controller exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/controllers/*", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument and controller exists
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/controllers/*/pump", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check admin role
}

# TODO: POST should ensure the instrument exists
# TODO: refactor chat rules into reusable functions
allow if {
	input.operation.method == "POST"
	glob.match("/instruments/*/chat/messages", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
}
