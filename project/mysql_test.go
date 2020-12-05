package project

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// Run a mysql before running this test
// docker run --name some-mysql -e MYSQL_ROOT_PASSWORD=root -e MYSQL_DATABASE=test-database -e MYSQL_USER=user -e MYSQL_PASSWORD=password -d -p 3306:3306 mysql:5.7
func databaseConnProvider() *gorm.DB {
	connectionString := fmt.Sprintf("user:password@tcp(localhost:3306)/test-database?parseTime=True")
	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		panic(err)
	}
	return db
}

func Test_mysql_Create(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		e   Project
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successful case",
			fields: fields{
				db: databaseConnProvider(),
			},
			args: args{
				ctx: context.TODO(),
				e: Project{
					ID:           "1234",
					DateCreated:  time.Now(),
					DateModified: time.Now(),
					Status:       "created",
				},
			},
			wantErr: false,
		},
	}
	db := databaseConnProvider()
	db.AutoMigrate(&Project{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mysql{
				db: tt.fields.db,
			}
			if err := m.Create(tt.args.ctx, tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("mysql.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
