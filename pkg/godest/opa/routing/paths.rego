package sargassum.godest.routing

import future.keywords

to_path(path_or_parts) := path if {
	path := sprintf("/%s", [concat("/", path_or_parts)])
}

to_parts(path) := split(trim_prefix(path, "/"), "/")
