package common

import "dpv/dpv/src/repository/t"

// AssignStringIfPresent assigns a string value from updates to patch if it's present.
// It sets the key in patch to nil if the value is empty, otherwise sets it to the value.
// It returns an error if the field is present but condition is false.
func AssignStringIfPresent(updates, patch map[string]interface{}, key string, condition bool) error {
	if val, ok := updates[key].(string); ok {
		if !condition {
			return t.Errorf("cannot edit %s field", key)
		}
		if val == "" {
			patch[key] = nil
		} else {
			patch[key] = val
		}
	}
	return nil
}

// AssignMandatoryString assigns a string value from updates to patch if present,
// but returns an error if the value is empty.
// It returns an error if the field is present but condition is false.
func AssignMandatoryString(updates, patch map[string]interface{}, key string, condition bool) error {
	if val, ok := updates[key].(string); ok {
		if !condition {
			return t.Errorf("cannot edit %s field", key)
		}
		if val == "" {
			return t.Errorf("%s is mandatory", key)
		}
		patch[key] = val
	}
	return nil
}
