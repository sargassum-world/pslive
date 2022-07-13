package sargassum.pslive.web.policies.auth

import future.keywords

import data.sargassum.godest.routing as r

# Policy Scope

in_scope if {
	"/csrf" == input.resource.path
}

in_scope if {
	"/login" == input.resource.path
}

in_scope if {
	"/sessions" == input.resource.path
}

# Policy Result & Error

matching_routes contains route if {
	"GET" == input.operation.method
	["csrf"] == r.to_parts(input.resource.path)
	route := r.get("/csrf")
}

allow if {
	"GET" == input.operation.method
	["csrf"] == r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"GET" == input.operation.method
	["login"] == r.to_parts(input.resource.path)
	route := r.get("/login")
}

allow if {
	"GET" == input.operation.method
	["login"] == r.to_parts(input.resource.path)
}

matching_routes contains route if {
	"POST" == input.operation.method
	["sessions"] == r.to_parts(input.resource.path)
	route := r.post("/sessions")
}

allow if {
	"POST" == input.operation.method
	["sessions"] == r.to_parts(input.resource.path)
}

errors contains error_matching if {
	in_scope
	error_matching := r.error_matching_routes(matching_routes)
}
