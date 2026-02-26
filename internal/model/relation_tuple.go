package model

import "time"

// RelationTuple represents a Zanzibar-style relation tuple.
// Format: <object>#<relation>@<subject>
// Example: document:readme#viewer@user:alice
//
// In Zanzibar terminology:
// - Object: The resource being accessed (namespace:object_id)
// - Relation: The relationship type (e.g., viewer, editor, owner)
// - Subject: The entity with the relation (user:id or object#relation for usersets)
type RelationTuple struct {
	BaseModel
	
	// Object components
	Namespace string `gorm:"type:varchar(255);not null;index:idx_object"`
	ObjectID  string `gorm:"type:varchar(255);not null;index:idx_object"`
	
	// Relation
	Relation string `gorm:"type:varchar(255);not null;index:idx_relation"`
	
	// Subject components (can be a user or a userset)
	SubjectNamespace string `gorm:"type:varchar(255);not null;index:idx_subject"`
	SubjectObjectID  string `gorm:"type:varchar(255);not null;index:idx_subject"`
	SubjectRelation  string `gorm:"type:varchar(255);index:idx_subject"` // Optional: for usersets
	
	// Metadata
	IsActive  bool       `gorm:"type:boolean;default:true;index"`
	ExpiresAt *time.Time `gorm:"index"` // Optional: for temporary permissions
}

func (RelationTuple) TableName() string {
	return "relation_tuples"
}

// Object returns the full object identifier (namespace:object_id)
func (rt *RelationTuple) Object() string {
	return rt.Namespace + ":" + rt.ObjectID
}

// Subject returns the full subject identifier
// Format: namespace:object_id or namespace:object_id#relation (for usersets)
func (rt *RelationTuple) Subject() string {
	subject := rt.SubjectNamespace + ":" + rt.SubjectObjectID
	if rt.SubjectRelation != "" {
		subject += "#" + rt.SubjectRelation
	}
	return subject
}

// String returns the canonical tuple representation
// Format: <object>#<relation>@<subject>
func (rt *RelationTuple) String() string {
	return rt.Object() + "#" + rt.Relation + "@" + rt.Subject()
}

// IsExpired checks if the tuple has expired
func (rt *RelationTuple) IsExpired() bool {
	if rt.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*rt.ExpiresAt)
}

// IsValid checks if the tuple is currently valid (active and not expired)
func (rt *RelationTuple) IsValid() bool {
	return rt.IsActive && !rt.IsExpired()
}
