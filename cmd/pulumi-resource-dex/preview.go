package main

import (
	"fmt"
)

// wrapError wraps a Dex API error with context to make it more user-friendly.
// This provides better error messages that include the operation, resource type,
// and resource ID for easier debugging.
func wrapError(operation, resourceType, resourceID string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("dex %s %s %q: %w", operation, resourceType, resourceID, err)
}

