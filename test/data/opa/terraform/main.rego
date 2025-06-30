package terraform

import rego.v1

violations contains violation if {
	# Filter out "no-op" changes - only count actual modifications
	actual_changes := [rc |
		some rc in input.resource_changes
		not "no-op" in rc.change.actions
	]
	num_changes := count(actual_changes)
	num_changes >= 2

	violation := {
		"kind": "violation",
		"reason": "cannot modify multiple resources in a single plan",
	}
}

# Default decision (allow)
default decision := [{
	"kind": "allow",
	"reason": "changes are all valid",
}]

# Override decision if there are violations
decision := violations if {
	count(violations) > 0
}
