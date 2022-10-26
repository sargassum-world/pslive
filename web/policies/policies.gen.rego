# Code generated by github.com/hairyhenderson/gomplate DO NOT EDIT.

package sargassum.pslive.web.policies

import future.keywords

import data.sargassum.godest.routing

import data.sargassum.pslive.web.policies.assets
import data.sargassum.pslive.web.policies.auth
import data.sargassum.pslive.web.policies.cable
import data.sargassum.pslive.web.policies.home
import data.sargassum.pslive.web.policies.instruments
import data.sargassum.pslive.web.policies.privatechat
import data.sargassum.pslive.web.policies.users
import data.sargassum.pslive.web.policies.videostreams

# Policy assets

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

# Policy auth

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

# Policy cable

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

# Policy home

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

# Policy instruments

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

# Policy privatechat

matching_policies contains "privatechat" if {
	privatechat.in_scope
}

allow if {
	privatechat.in_scope
	privatechat.allow
}

policy_errors["privatechat"] := error if {
	some error in privatechat.errors
}

# Policy users

matching_policies contains "users" if {
	users.in_scope
}

allow if {
	users.in_scope
	users.allow
}

policy_errors["users"] := error if {
	some error in users.errors
}

# Policy videostreams

matching_policies contains "videostreams" if {
	videostreams.in_scope
}

allow if {
	videostreams.in_scope
	videostreams.allow
}

policy_errors["videostreams"] := error if {
	some error in videostreams.errors
}

# Error handling

errors contains error_matching if {
	error_matching := routing.error_matching_policies(matching_policies)
}

error := merged if {
	merged := routing.merge_policy_errors(errors, policy_errors)
}
