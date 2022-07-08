package sargassum.pslive.web.policies.users

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

scope := {
	"/users",
	"/users/**",
}

default allow := false

allow if input.operation.method == "GET"

# TODO: SUB to chat should ensure the user exists
allow if input.operation.method == "SUB"

# TODO: UNSUB from chat should ensure the user exists
allow if input.operation.method == "UNSUB"

# TODO: MSG on chat should ensure the user exists
allow if input.operation.method == "MSG"

# TODO: POST to chat should ensure the user exists
allow if {
	input.operation.method == "POST"
	auth.is_authenticated(input.subject)
}

errors contains message if {
	input.operation.method == "POST"
	not auth.is_authenticated(input.subject)
	message := "unauthenticated user"
}

errors contains "not implemented" if {
	not allow
}