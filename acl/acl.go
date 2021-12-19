// Package acl is meant to handle access control of projects
//
// Only basic version of acl - assumes 1 Project to 1 Owner mapping
package acl

import (
	"time"

	"github.com/gofrs/uuid"
)

type permission string

// Anonymous permission to resources - applied on project level
// Creates a temporary user on system. User does not need to login
// A temporary user with expiry for this user - only can access certain features
// Similar to reader
var Anonymous permission = "anonymous"

// Reader permission to resources - applied on project level
// Able to access projects on the platform and can read but is unable
// to alter anything on it
// Can view scripts and videos but all buttons will be grayed out
var Reader permission = "reader"

// Editor permission to resource - applied on project level
// Able to alter the following:
// - Edit scripts
// - Regenerate video segments
var Editor permission = "editor"

// Publisher permission to resource - applied on project level
// This permission is needed to provide some sort of granularity to
// publish video to a video publishing platform
var Publisher permission = "publisher"

// Owner permission to resource - applied on project level
var Owner permission = "owner"

// ACL - manage permission model of handling project resources
type ACL struct {
	ID           string     `json:"id"`
	ProjectID    string     `json:"project_id"`
	UserID       string     `json:"user_id"`
	Permission   permission `json:"permission"`
	DateCreated  time.Time  `json:"date_created"`
	DateModified time.Time  `json:"date_modified"`
}

func New(projectID, userID string) ACL {
	aclID, _ := uuid.NewV4()
	return ACL{
		ID:           aclID.String(),
		UserID:       userID,
		ProjectID:    projectID,
		Permission:   Owner,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}
}
