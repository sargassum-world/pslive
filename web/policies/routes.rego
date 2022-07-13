package sargassum.pslive.web.policies

import future.keywords

import data.sargassum.godest.errors as e
import data.sargassum.godest.routing as r

import data.sargassum.pslive.web.policies.assets
import data.sargassum.pslive.web.policies.auth
import data.sargassum.pslive.web.policies.cable
import data.sargassum.pslive.web.policies.home
import data.sargassum.pslive.web.policies.instruments

matching_policies contains "assets" if {
	assets.in_scope
}

allow if {
	assets.in_scope
	assets.allow
}

policy_errors["assets"] := error if {
	some error in assets.errors
}

matching_policies contains "auth" if {
	auth.in_scope
}

allow if {
	auth.in_scope
	auth.allow
}

policy_errors["auth"] := error if {
	some error in auth.errors
}

matching_policies contains "cable" if {
	cable.in_scope
}

allow if {
	cable.in_scope
	cable.allow
}

policy_errors["cable"] := error if {
	some error in cable.errors
}

matching_policies contains "home" if {
	home.in_scope
}

allow if {
	home.in_scope
	home.allow
}

policy_errors["home"] := error if {
	some error in home.errors
}

matching_policies contains "instruments" if {
	instruments.in_scope
}

allow if {
	instruments.in_scope
	instruments.allow
}

policy_errors["instruments"] := error if {
	some error in instruments.errors
}

errors contains error_matching if {
	error_matching := r.error_matching_policies(matching_policies)
}

error := merged if {
	merged := r.merge_policy_errors(errors, policy_errors)
} else := error_unknown {
	not allow
	error_unknown := e.new("unknown error")
}
