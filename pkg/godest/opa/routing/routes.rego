# METADATA
# description: Path-based routing utilities for godest apps
package sargassum.godest.routing

import future.keywords

import data.sargassum.godest.errors

# Routes

# METADATA
# description: |
#   Create a route with a method and a path-matching pattern, and a result associated with the
#   route (such as an authorizaton result). The result can either be a plain value (scalar or
#   composite), an error result object with a set of errors, or a key result object with a key for
#   indirectly (and dynamically) retrieving the actual value from some other structure. A indirect
#   retrieval is needed to prevent recursion if the value depends on any path parameters parsed by
#   the route.
route(method, pattern) := {
	"method": method,
	"pattern": pattern,
}

# METADATA
# description: |
#   Create a route for a GET method, e.g. for HTTP routing
get(pattern) := route("GET", pattern)

# METADATA
# description: |
#   Create a route for a HEAD method, e.g. for HTTP routing
head(pattern) := route("HEAD", pattern)

# METADATA
# description: |
#   Create a route for a POST method, e.g. for HTTP routing
post(pattern) := route("POST", pattern)

# METADATA
# description: |
#   Create a route for a PUT method, e.g. for HTTP routing
put(pattern) := route("PUT", pattern)

# METADATA
# description: |
#   Create a route for a PATCH method, e.g. for HTTP routing
patch(pattern) := route("PATCH", pattern)

# METADATA
# description: |
#   Create a route for a DELETE method, e.g. for HTTP routing
delete(pattern) := route("DELETE", pattern)

# METADATA
# description: |
#   Create a route for a CONNECT method, e.g. for HTTP routing
connect(pattern) := route("CONNECT", pattern)

# METADATA
# description: |
#   Create a route for a OPTIONS method, e.g. for HTTP routing
options(pattern) := route("OPTIONS", pattern)

# METADATA
# description: |
#   Create a route for a PUB method, e.g. for Turbo Streams event routing
pub(pattern) := route("PUB", pattern)

# METADATA
# description: |
#   Create a route for a SUB method, e.g. for Turbo Streams event routing
sub(pattern) := route("SUB", pattern)

# METADATA
# description: |
#   Create a route for a UNSUB method, e.g. for Turbo Streams event routing
unsub(pattern) := route("UNSUB", pattern)

# METADATA
# description: |
#   Create a route for a MSG method, e.g. for Turbo Streams event routing
msg(pattern) := route("MSG", pattern)

# METADATA
# description: |
#   Create a route for a SHOW method, e.g. for conditional rendering in templates
show(pattern) := route("SHOW", pattern)

# Error Handling

error_matching_routes(matching_routes) := error_no_matches if {
	count(matching_routes) == 0
	error_no_matches := errors.new("no matching route found")
} else := error_multiple_matches {
	count(matching_routes) > 1
	error_multiple_matches := errors.errorf(
		"multiple matching routes found: %s",
		[concat(
			", ",
			[sprintf("%s %s", route) | some route in matching_routes],
		)],
	)
}
