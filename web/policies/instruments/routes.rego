package sargassum.pslive.web.policies.instruments

import future.keywords

import data.sargassum.godest.routing as r

import data.sargassum.pslive.internal.app.pslive.auth

# Policy Scope

in_scope if {
	"/instruments" == input.resource.path
}

in_scope if {
	glob.match("/instruments/*", [], input.resource.path)
}

# Policy Result & Error

matching_routes contains route if {
	"GET" == input.operation.method

	# TODO: use golang templating to pre-process routes and reduce copy-pasting
	["instruments"] == r.to_parts(input.resource.path)
	route := r.get("/instruments")
}

allow if {
	"GET" == input.operation.method
	["instruments"] == r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments"] == r.to_parts(input.resource.path)
	route := r.post("/instruments")
}

allow if {
	"POST" == input.operation.method
	["instruments"] == r.to_parts(input.resource.path)
	allow_instruments_post(input.subject)
}

matching_routes contains route if {
	"GET" == input.operation.method
	["instruments", id] = r.to_parts(input.resource.path)
	route := r.get("/instruments/*")
}

allow if {
	"GET" == input.operation.method
	["instruments", id] = r.to_parts(input.resource.path)
	allow_instrument_get(id)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*")
}

allow if {
	"POST" == input.operation.method
	["instruments", id] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "name"] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/name")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "name"] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "description"] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/description")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "description"] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"SUB" == input.operation.method
	["instruments", id, "users"] = r.to_parts(input.resource.path)
	route := r.sub("/instruments/*/users")
}

allow if {
	"SUB" == input.operation.method
	["instruments", id, "users"] = r.to_parts(input.resource.path)
	allow_instrument_get(id)
}

matching_routes contains route if {
	"UNSUB" == input.operation.method
	["instruments", id, "users"] = r.to_parts(input.resource.path)
	route := r.unsub("/instruments/*/users")
}

allow if {
	"UNSUB" == input.operation.method
	["instruments", id, "users"] = r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"MSG" == input.operation.method
	["instruments", id, "users"] = r.to_parts(input.resource.path)
	route := r.msg("/instruments/*/users")
}

allow if {
	"MSG" == input.operation.method
	["instruments", id, "users"] = r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "cameras"] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/cameras")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "cameras"] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "cameras", camera_id] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/cameras/*")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "cameras", camera_id] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "controllers"] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/controllers")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "controllers"] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "controllers", controller_id] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/controllers/*")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "controllers", controller_id] = r.to_parts(input.resource.path)
	allow_instrument_admin_post(input.subject, id)
}

matching_routes contains route if {
	"SUB" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
	route := r.sub("/instruments/*/controllers/*/pump")
}

allow if {
	"SUB" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
	allow_instrument_get(id)
}

matching_routes contains route if {
	"PUB" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
	route := r.pub("/instruments/*/controllers/*/pump")
}

allow if {
	"PUB" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"MSG" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
	route := r.msg("/instruments/*/controllers/*/pump")
}

allow if {
	"MSG" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/controllers/*/pump")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "controllers", controller_id, "pump"] = r.to_parts(input.resource.path)
	allow_instrument_operator_post(input.subject, id)
}

matching_routes contains route if {
	"SUB" == input.operation.method
	["instruments", id, "chat", "messages"] = r.to_parts(input.resource.path)
	route := r.sub("/instruments/*/chat/messages")
}

allow if {
	"SUB" == input.operation.method
	["instruments", id, "chat", "messages"] = r.to_parts(input.resource.path)
	allow_instrument_get(id)
}

matching_routes contains route if {
	"MSG" == input.operation.method
	["instruments", id, "chat", "messages"] = r.to_parts(input.resource.path)
	route := r.msg("/instruments/*/chat/messages")
}

allow if {
	"MSG" == input.operation.method
	["instruments", id, "chat", "messages"] = r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["instruments", id, "chat", "messages"] = r.to_parts(input.resource.path)
	route := r.post("/instruments/*/chat/messages")
}

allow if {
	"POST" == input.operation.method
	["instruments", id, "chat", "messages"] = r.to_parts(input.resource.path)
	allow_instrument_chatter_post(input.subject, id)
}

errors contains error_matching if {
	in_scope
	error_matching := r.error_matching_routes(matching_routes)
}

# Internal Route Checks

allow_instruments_post(subject) := is_authenticated(subject)

allow_instrument_get(instrument_id) := is_valid_instrument(instrument_id)

allow_instrument_admin_post(subject, instrument_id) if {
	is_valid_instrument(instrument_id)
	is_instrument_admin(subject, instrument_id)
}

allow_instrument_operator_post(subject, instrument_id) if {
	is_valid_instrument(instrument_id)
	is_instrument_operator(subject, instrument_id)
}

allow_instrument_chatter_post(subject, instrument_id) if {
	is_valid_instrument(instrument_id)
	is_authenticated(subject, instrument_id)
}

# Internal Attribute Checks

is_authenticated(subject) if {
	auth.is_authenticated(subject)
}

is_valid_instrument(instrument_id) if {
	to_number(instrument_id) == 1
} else {
	to_number(instrument_id) in input.context.instruments # TODO: implement
}

is_instrument_admin(subject, instrument_id) if {
	is_authenticated(subject)
	# subject.identity == input.context.instrument.admin_identity_id # TODO: implement
}

is_instrument_operator(subject, instrument_id) if {
	is_authenticated(subject) # TODO: implement operator permissions
}
