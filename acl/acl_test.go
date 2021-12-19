// Package acl is meant to handle access control of projects
//
// Only basic version of acl - assumes 1 Project to 1 Owner mapping
package acl

import (
	"testing"
	"time"
)

func TestACL_IsAuthorized(t *testing.T) {
	type fields struct {
		ID           string
		ProjectID    string
		UserID       string
		Permission   permission
		DateCreated  time.Time
		DateModified time.Time
	}
	type args struct {
		p permission
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "editor but need reader",
			fields: fields{
				Permission: Editor,
			},
			args: args{
				Reader,
			},
			want: true,
		},
		{
			name: "owner but need reader",
			fields: fields{
				Permission: Owner,
			},
			args: args{
				Reader,
			},
			want: true,
		},
		{
			name: "reader but need editor",
			fields: fields{
				Permission: Reader,
			},
			args: args{
				Editor,
			},
			want: false,
		},
		{
			name: "anonymous but need reader",
			fields: fields{
				Permission: Anonymous,
			},
			args: args{
				Reader,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := ACL{
				ID:           tt.fields.ID,
				ProjectID:    tt.fields.ProjectID,
				UserID:       tt.fields.UserID,
				Permission:   tt.fields.Permission,
				DateCreated:  tt.fields.DateCreated,
				DateModified: tt.fields.DateModified,
			}
			if got := a.IsAuthorized(tt.args.p); got != tt.want {
				t.Errorf("ACL.IsAuthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}
