package sargassum.pslive.web.policies.users

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_user_get(id) := is_valid_user(id)

allow_user_chat_post(subject, id) if {
	is_valid_user(id)
	auth.is_authenticated(subject)
}

# Internal Attribute Checks

is_valid_user(user_id) if {
	user_id == "07d4a550-c503-4d90-a312-a41f2cd41344"
} else {
	user_id in input.context.users # TODO: implement
}
