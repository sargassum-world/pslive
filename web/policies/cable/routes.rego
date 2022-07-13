package sargassum.pslive.web.policies.cable

import future.keywords

import data.sargassum.godest.errors as e

# Policy Scope

in_scope if {
	"/cable" == input.resource.path
}

# Policy Result & Error

allow if {
	input.operation.method == "GET"
}

errors contains error_method if {
	in_scope
	not allow
	error_method := e.new("route not implemented")
}
