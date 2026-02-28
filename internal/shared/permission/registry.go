package permission

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/hiamthach108/dreon-auth/config"
)

// Permission represents a single permission from config
type Permission struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// IRegistry is the interface for permission registry (load and validate)
type IRegistry interface {
	List() []Permission
	ValidateCodes(codes []string) error
}

// Registry holds loaded permissions and validates permission codes
type Registry struct {
	list   []Permission
	byCode map[string]Permission
}

// NewRegistry loads permissions from a JSON file and returns a Registry
func NewRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read permissions config: %w", err)
	}

	var list []Permission
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("parse permissions config: %w", err)
	}

	byCode := make(map[string]Permission, len(list))
	for _, p := range list {
		if p.Code == "" {
			continue
		}
		byCode[p.Code] = p
	}

	return &Registry{
		list:   list,
		byCode: byCode,
	}, nil
}

// List returns all permissions
func (r *Registry) List() []Permission {
	if r == nil {
		return nil
	}
	return r.list
}

// ValidateCodes returns an error if any code is not in the registry
func (r *Registry) ValidateCodes(codes []string) error {
	if r == nil {
		return nil
	}
	for _, code := range codes {
		if code == "" {
			continue
		}
		if _, ok := r.byCode[code]; !ok {
			return fmt.Errorf("invalid permission code: %s", code)
		}
	}
	return nil
}

// GetByCode returns the permission for the given code and true if found
func (r *Registry) GetByCode(code string) (Permission, bool) {
	if r == nil {
		return Permission{}, false
	}
	p, ok := r.byCode[code]
	return p, ok
}

const defaultPermissionsPath = "config/permissions.json"

// NewRegistryFromConfig loads registry from path in AppConfig.Permissions.FilePath (env: PERMISSIONS_FILE), or default config/permissions.json
func NewRegistryFromConfig(cfg *config.AppConfig) (*Registry, error) {
	path := cfg.Permissions.FilePath
	if path == "" {
		path = defaultPermissionsPath
	}
	return NewRegistry(path)
}
