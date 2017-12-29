package utils

import ()

func ValidateUserID(id string) bool {
	result := true
	if id == "" {
		result = false
	}
	return result
}

func ValidatePassword(id string) bool {
	// to be implemented ...
	if id == "" {
		return false
	}
	return true
}
