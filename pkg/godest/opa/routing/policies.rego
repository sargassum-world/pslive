package sargassum.godest.routing

import future.keywords

import data.sargassum.godest.errors

# Error handling

merge_policy_errors(top_level_errors, policy_errors) := merged if {
	wrapped_policy_errors := {errors.wrapf(error, "policy %s", [name]) |
		some name, error in policy_errors
	}

	all_errors := top_level_errors | wrapped_policy_errors
	count(all_errors) > 0
	merged := errors.merge(all_errors)
}

error_matching_policies(matching_policies) := error_no_matches if {
	count(matching_policies) == 0
	error_no_matches := errors.new("no matching policy found")
} else := error_multiple_matches {
	count(matching_policies) > 1
	error_multiple_matches := errors.errorf(
		"multiple matching policies found: %s",
		[concat(", ", matching_policies)],
	)
}
