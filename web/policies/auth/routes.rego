package sargassum.pslive.web.policies.auth

import future.keywords

scope := {
	"/csrf",
	"/login",
	"/login?return=**",
	"/sessions",
}

allow := true

errors := set()
