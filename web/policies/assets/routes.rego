package sargassum.pslive.web.policies.assets

import future.keywords

scope := {
	"/sw.js",
	"/favicon.ico",
	"/fonts/**",
	"/static/**",
	"/app/**",
}

allow if input.operation.method == "GET"

errors contains "unallowed method" if not allow
