package dto

import "time"

// GrantRelationReq represents a request to grant a relation tuple
type GrantRelationReq struct {
	// Object components
	Namespace string `json:"namespace" binding:"required"`
	ObjectID  string `json:"objectId" binding:"required"`
	
	// Relation
	Relation string `json:"relation" binding:"required"`
	
	// Subject components
	SubjectNamespace string `json:"subjectNamespace" binding:"required"`
	SubjectObjectID  string `json:"subjectObjectId" binding:"required"`
	SubjectRelation  string `json:"subjectRelation,omitempty"` // Optional: for usersets
	
	// Optional metadata
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// RevokeRelationReq represents a request to revoke a relation tuple
type RevokeRelationReq struct {
	Namespace        string `json:"namespace" binding:"required"`
	ObjectID         string `json:"objectId" binding:"required"`
	Relation         string `json:"relation" binding:"required"`
	SubjectNamespace string `json:"subjectNamespace" binding:"required"`
	SubjectObjectID  string `json:"subjectObjectId" binding:"required"`
	SubjectRelation  string `json:"subjectRelation,omitempty"`
}

// CheckRelationReq represents a request to check if a relation exists
type CheckRelationReq struct {
	Namespace        string `json:"namespace" binding:"required"`
	ObjectID         string `json:"objectId" binding:"required"`
	Relation         string `json:"relation" binding:"required"`
	SubjectNamespace string `json:"subjectNamespace" binding:"required"`
	SubjectObjectID  string `json:"subjectObjectId" binding:"required"`
}

// CheckRelationResp represents the response of a relation check
type CheckRelationResp struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason,omitempty"`
}

// ListRelationsReq represents a request to list relation tuples
type ListRelationsReq struct {
	// Filter by object
	Namespace string `json:"namespace,omitempty"`
	ObjectID  string `json:"objectId,omitempty"`
	
	// Filter by relation
	Relation string `json:"relation,omitempty"`
	
	// Filter by subject
	SubjectNamespace string `json:"subjectNamespace,omitempty"`
	SubjectObjectID  string `json:"subjectObjectId,omitempty"`
	
	// Pagination
	PaginationReq
}

// RelationTupleResp represents a relation tuple response
type RelationTupleResp struct {
	ID               string     `json:"id"`
	Namespace        string     `json:"namespace"`
	ObjectID         string     `json:"objectId"`
	Relation         string     `json:"relation"`
	SubjectNamespace string     `json:"subjectNamespace"`
	SubjectObjectID  string     `json:"subjectObjectId"`
	SubjectRelation  string     `json:"subjectRelation,omitempty"`
	IsActive         bool       `json:"isActive"`
	ExpiresAt        *time.Time `json:"expiresAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// BulkGrantRelationReq represents a request to grant multiple relation tuples
type BulkGrantRelationReq struct {
	Relations []GrantRelationReq `json:"relations" binding:"required,min=1,dive"`
}

// BulkRevokeRelationReq represents a request to revoke multiple relation tuples
type BulkRevokeRelationReq struct {
	Relations []RevokeRelationReq `json:"relations" binding:"required,min=1,dive"`
}

// ExpandRelationReq represents a request to expand a relation (get all subjects)
type ExpandRelationReq struct {
	Namespace string `json:"namespace" binding:"required"`
	ObjectID  string `json:"objectId" binding:"required"`
	Relation  string `json:"relation" binding:"required"`
}

// RelationSubjectResp represents a subject in relation expansion
type RelationSubjectResp struct {
	Namespace string `json:"namespace"`
	ObjectID  string `json:"objectId"`
	Relation  string `json:"relation,omitempty"`
}

// ExpandRelationResp represents the response of relation expansion
type ExpandRelationResp struct {
	Subjects []RelationSubjectResp `json:"subjects"`
	Count    int                   `json:"count"`
}
