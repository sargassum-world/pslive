# METADATA
# description: Internal auth-related utilities for the pslive app
package sargassum.pslive.internal.app.pslive.auth

# METADATA
# description: |
#   Check the subject to determine whether it is authenticated
is_authenticated(subject) {
	subject.authenticated
	subject.identity != ""
} else = false
