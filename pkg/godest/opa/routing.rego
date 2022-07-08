# METADATA
# description: Policy routing utilities for godest apps
package sargassum.godest.routing

import future.keywords

# METADATA
# description: |
#   Return the list of scopes in the policy matching the path
match_scopes(path, policy) := matching_scopes if {
	matching_scopes := {s | some s in policy.scope; glob.match(s, ["/"], path)}
}

# METADATA
# description: |
#   Return the list of policies in the policy set matching the path
match_policies(path, policies) := matching_policies if {
	matching_policies := {name: {
		"scopes": matching_scopes,
		"results": policies[name],
	} |
		some name, policy in policies
		matching_scopes := match_scopes(path, policy)
		count(matching_scopes) > 0
	}
}

# METADATA
# description: |
#   Generate an error message for troubleshooting overlapping scopes between policies
overlapping_matches_error(matching_policies) := error if {
	overlap := {sprintf("%s (%s)", [name, concat(", ", scopes)]) |
		some name, policy in matching_policies
		scopes := policy.scopes
	}

	error := sprintf("%d matching policies: %s", [count(matching_policies), concat(", ", overlap)])
}

# METADATA
# description: |
#   Wrap error messages with the name of the policy which produced those errors
wrap_error(policy_name, error) := wrapped if {
	wrapped := sprintf("%s: %s", [policy_name, error])
}
