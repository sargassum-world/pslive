package sargassum.pslive.web.policies.home

import future.keywords

scope := {"/"}

allow if input.operation.method == "GET"

errors contains "unallowed method" if not allow
