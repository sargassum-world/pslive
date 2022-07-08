package sargassum.pslive.web.policies.privatechat

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

scope := {"/private-chats/**"}

default allow := false

# TODO: also add error message for each nonexistent user, and if subject isn't one of the two users
errors contains "unauthenticated user" if {
	not auth.is_authenticated(input.subject)
}

# TODO: SUB should ensure both users exist
allow if {
	input.operation.method in {"SUB", "MSG", "POST"}
	glob.match("/private-chats/*/*/chat/{users,messages}", ["/"], input.resource.path)
	auth.is_authenticated(input.subject)
	# TODO: check whether the subject is one of the two users
}
