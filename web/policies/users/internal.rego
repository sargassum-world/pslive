package sargassum.pslive.web.policies.users

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_user_get(id) = true # TODO: implement user validity check

# is_valid_user(id) # TODO: implement user validity check

allow_user_chat_post(subject, id) if {
	# is_valid_user(id) # TODO: implement user validity check
	auth.is_authenticated(subject)
}

# Internal Attribute Checks

is_valid_user(user_id) if {
	user := input.context.db.users_user[_]
	to_number(user_id) == user.id
}
