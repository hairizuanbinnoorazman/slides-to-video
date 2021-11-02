// Package projectroles is to provide a table that maps user to projects
package projectroles

import "time"

type role string

var (
	Owner     role = "owner"
	Editor    role = "editor"
	Viewer    role = "viewer"
	Publisher role = "publisher"
	Guest     role = "guest"
)

type ProjectRole struct {
	ProjectID string `gorm:"primary_key"`
	// Can be user or group
	EntityID     string `gorm:"primary_key"`
	Role         role   `gorm:"type:varchar(80)"`
	DateCreated  time.Time
	DateModified time.Time
}

func New(projectID, entityID string, currentRole role) ProjectRole {
	return ProjectRole{
		ProjectID:    projectID,
		EntityID:     entityID,
		Role:         currentRole,
		DateCreated:  time.Now(),
		DateModified: time.Now(),
	}
}

func setRole(r role) func(*ProjectRole) error {
	return func(a *ProjectRole) error {
		a.Role = r
		return nil
	}
}
