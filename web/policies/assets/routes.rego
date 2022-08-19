package sargassum.pslive.web.policies.assets

import future.keywords

import data.sargassum.godest.errors as e

# Policy Scope

in_scope if {
	"/sw.js" == input.resource.path
}

in_scope if {
	"/favicon.ico" == input.resource.path
}

in_scope if {
	glob.match("/fonts/*", ["#"], input.resource.path)
}

in_scope if {
	glob.match("/static/*", ["#"], input.resource.path)
}

in_scope if {
	glob.match("/app/*", ["#"], input.resource.path)
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
