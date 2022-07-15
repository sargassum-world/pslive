package sargassum.pslive.web.policies.instruments

import future.keywords

import data.sargassum.pslive.internal.app.pslive.auth

# Internal Route Checks

allow_instruments_post(subject) := auth.is_authenticated(subject)

allow_instrument_get(instrument_id) := is_valid_instrument(instrument_id)

allow_instrument_post(subject, instrument_id) if {
	is_valid_instrument(instrument_id)
	is_instrument_admin(subject, instrument_id)
}

allow_camera_post(subject, instrument_id, camera_id) if {
	is_valid_instrument(instrument_id)
	is_valid_camera(camera_id)
	is_instrument_admin(subject, instrument_id)
}

allow_controller_get(subject, instrument_id, controller_id) if {
	is_valid_instrument(instrument_id)
	is_valid_controller(controller_id)
}

allow_controller_post(subject, instrument_id, controller_id) if {
	is_valid_instrument(instrument_id)
	is_valid_controller(controller_id)
	is_instrument_admin(subject, instrument_id)
}

allow_controller_pump_post(subject, instrument_id, controller_id) if {
	is_valid_instrument(instrument_id)
	is_instrument_operator(subject, instrument_id)
}

allow_instrument_chat_post(subject, instrument_id) if {
	is_valid_instrument(instrument_id)
	auth.is_authenticated(subject, instrument_id)
}

# Internal Attribute Checks

is_valid_instrument(instrument_id) if {
	to_number(instrument_id) == 1
} else {
	to_number(instrument_id) in input.context.instruments # TODO: implement
}

is_valid_camera(camera_id) if {
	to_number(camera_id) == 1
} else {
	to_number(camera_id) in input.context.cameras # TODO: implement
}

is_valid_controller(controller_id) if {
	to_number(controller_id) == 1
} else {
	to_number(controller_id) in input.context.controllers # TODO: implement
}

is_instrument_admin(subject, instrument_id) if {
	auth.is_authenticated(subject)
	# subject.identity == input.context.instrument.admin_identity_id # TODO: implement
}

is_instrument_operator(subject, instrument_id) if {
	auth.is_authenticated(subject) # TODO: implement operator permissions
}
