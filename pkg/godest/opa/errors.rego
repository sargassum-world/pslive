# METADATA
# description: Golang-like error handling utilities
package sargassum.godest.errors

import future.keywords

# Errors

# METADATA
# description: |
#   Create a result consisting of an error
new(message) := {
	"type": "error",
	"message": message,
}

# METADATA
# description: |
#   Check if the result consists of an error
is_error(result) if {
	result.type == "error"
	some "message", _ in result
} else = false {
	true
}

# METADATA
# description: |
#   Create an error result with a message from the provided sprintf arguments
errorf(format, values) := new(sprintf(format, values))

# METADATA
# description: |
#   Return the provided result if it's not an error; otherwise, wrap the result's error message with
#   the provided message as an annotation
wrap(result, annotation) := result if {
	not is_error(result)
} else := error {
	error := errorf("%s: %s", [annotation, result.message])
}

# METADATA
# description: |
#   Return the provided result if it's not an error; otherwise, wrap the result's error message with
#   the provided message sprintf arguments as an annotation
wrapf(result, annotation_format, annotation_values) := wrap(
	result,
	sprintf(annotation_format, annotation_values),
)

# Result processing

# METADATA
# description: |
#   Merge the provided errors into a single error whose message is a list of the messages in all
#   provided error messages.
merge(errors) := errors if {
	is_error(errors)
} else := selected_error {
	deduplicated := {error | some error in errors}
	count(deduplicated) == 1
	some selected_error in deduplicated
} else := merged_error {
	count(errors) > 0
	merged_error := errorf("multiple errors: %s", [concat(
		", ",
		{sprintf("[%s]", [error.message]) | some error in errors},
	)])
}
