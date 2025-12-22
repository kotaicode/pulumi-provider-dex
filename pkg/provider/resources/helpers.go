package resources

// ============================================================================
// Helper functions for connector implementations
// ============================================================================

// GetString extracts a string value from a map, returning empty string if not found.
func GetString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// GetStringPtr extracts a string value from a map, returning nil if not found or empty.
func GetStringPtr(m map[string]any, key string) *string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok && str != "" {
			return &str
		}
	}
	return nil
}

// GetBoolPtr extracts a boolean value from a map, returning nil if not found.
func GetBoolPtr(m map[string]any, key string) *bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return &b
		}
	}
	return nil
}
