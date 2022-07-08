package sargassum.pslive.web.policies.cable

import future.keywords

scope := {"/cable"}

allow if input.operation.method == "GET"

errors contains "unallowed method" if not allow
