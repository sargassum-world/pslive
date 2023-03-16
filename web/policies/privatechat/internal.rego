package sargassum.pslive.web.policies.privatechat

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_private_chat_get(subject, first, second) if {
	is_valid_chat(first, second)
	is_participant(subject, first, second)
	auth.is_authenticated(subject)
}

allow_private_chat_post(subject, first, second) if {
	allow_private_chat_get(subject, first, second)
}

# Internal Attribute Checks

# TODO: implement (right now we don't have a users db table)
is_valid_user(_) = true

is_valid_chat(first, second) if {
	first != second
	is_valid_user(first)
	is_valid_user(second)
}

is_participant(subject, first, second) if {
	subject.identity in {first, second}
}
