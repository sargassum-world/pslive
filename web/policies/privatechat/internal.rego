package sargassum.pslive.web.policies.privatechat

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_private_chat(subject, first, second) if {
	is_valid_chat(first, second)
	is_participant(subject, first, second)
	auth.is_authenticated(subject)
}

# Internal Attribute Checks

is_valid_user(user_id) if {
	user_id == "07d4a550-c503-4d90-a312-a41f2cd41344"
} else {
	user_id in input.context.users # TODO: implement
}

is_valid_chat(first, second) if {
	is_valid_user(first)
	is_valid_user(second)
}

is_participant(subject, first, second) if {
	subject.identity in {first, second}
}
