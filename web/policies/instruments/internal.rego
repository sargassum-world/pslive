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

allow_camera_get(instrument_id, camera_id) if {
	is_valid_instrument(instrument_id)
	is_valid_camera(instrument_id, camera_id)
}

allow_camera_post(subject, instrument_id, camera_id) if {
	is_valid_instrument(instrument_id)
	is_valid_camera(instrument_id, camera_id)
	is_instrument_admin(subject, instrument_id)
}

allow_controller_get(subject, instrument_id, controller_id) if {
	is_valid_instrument(instrument_id)
	is_valid_controller(instrument_id, controller_id)
}

allow_controller_post(subject, instrument_id, controller_id) if {
	is_valid_instrument(instrument_id)
	is_valid_controller(instrument_id, controller_id)
	is_instrument_admin(subject, instrument_id)
}

allow_controller_pump_post(subject, instrument_id, controller_id) if {
	is_valid_instrument(instrument_id)
	is_valid_controller(instrument_id, controller_id)
	is_instrument_operator(subject, instrument_id)
}

allow_instrument_chat_post(subject, instrument_id) if {
	is_valid_instrument(instrument_id)
	auth.is_authenticated(subject)
}

# Internal Attribute Checks

is_valid_instrument(instrument_id) if {
	instrument := input.context.db.instruments_instrument[_]

	to_number(instrument_id) == instrument.id
}

is_valid_camera(instrument_id, camera_id) if {
	camera := input.context.db.instruments_camera[_]
	to_number(camera_id) == camera.id
	to_number(instrument_id) == camera.instrument_id
}

is_valid_controller(instrument_id, controller_id) if {
	controller := input.context.db.instruments_controller[_]
	to_number(controller_id) == controller.id
	to_number(instrument_id) == controller.instrument_id
}

is_instrument_admin(subject, instrument_id) if {
	auth.is_authenticated(subject)
	instrument := input.context.db.instruments_instrument[_]
	to_number(instrument_id) == instrument.id
	subject.identity == instrument.admin_identity_id
}

is_instrument_operator(subject, instrument_id) if {
	auth.is_authenticated(subject) # TODO: implement operator permissions
}
