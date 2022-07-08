package sargassum.pslive.web.policies.assets

import future.keywords

scope := {
	"/sw.js",
	"/favicon.ico",
	"/fonts/**",
	"/static/**",
	"/app/**",
}

allow := true

errors := set()
