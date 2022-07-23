package sargassum.pslive.web.policies.privatechat

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_private_chat(subject, first, second) if {
	# is_valid_chat(first, second) # TODO: implement user validity check
	is_participant(subject, first, second)
	auth.is_authenticated(subject)
}

# Internal Attribute Checks

is_valid_user(user_id) if {
	user := input.context.db.users_user[_]
	to_number(user_id) == user.id
}

is_valid_chat(first, second) if {
	is_valid_user(first)
	is_valid_user(second)
}

is_participant(subject, first, second) if {
	subject.identity in {first, second}
}
