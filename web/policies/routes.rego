package sargassum.pslive.web

import future.keywords

import data.sargassum.godest.routing

matches := routing.match_policies(input.resource.path, data.sargassum.pslive.web.policies)

active_policy := active if {
	count(matches) == 1
	some name, match in matches
	active := {
		"name": name,
		"matching_scopes": match.scopes,
		"results": match.results,
	}
}

default allow := false

errors contains "no matching policies" if {
	count(matches) == 0
}

errors contains message if {
	count(matches) > 1
	message := routing.overlapping_matches_error(matches)
}

allow if {
	active_policy.results.allow
}

errors contains message if {
	some error in active_policy.results.errors
	message := routing.wrap_error(active_policy.name, error)
}
