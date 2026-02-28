package model

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// PermissionsToJSON marshals []string to datatypes.JSON for storage.
func PermissionsToJSON(perms []string) datatypes.JSON {
	if len(perms) == 0 {
		return nil
	}
	b, err := json.Marshal(perms)
	if err != nil {
		return nil
	}
	return datatypes.JSON(b)
}

// PermissionsFromJSON unmarshals datatypes.JSON to []string.
func PermissionsFromJSON(data datatypes.JSON) []string {
	if len(data) == 0 {
		return nil
	}
	var perms []string
	if err := json.Unmarshal(data, &perms); err != nil {
		return nil
	}
	return perms
}
