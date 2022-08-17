package sargassum.pslive.web.policies.users

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_user_get(id) if {
	is_valid_user(id)
}

allow_user_get_info(id) if {
	is_valid_user(id)
}

allow_user_get_info_email(subject, id) if {
	is_valid_user(id)
	auth.is_authenticated(subject)
}

allow_user_chat_post(subject, id) if {
	is_valid_user(id)
	auth.is_authenticated(subject)
}

# Internal Attribute Checks

# TODO: implement (right now we don't have a users db table)
is_valid_user(user_id) = true
