package services

import (
	"testing"
)

func tokenHelper(id string) string {
	a, _ := NewToken(id, 3600, "manager", "manager")
	return "Bearer " + a
}

func TestExtractToken(t *testing.T) {
	type args struct {
		tokenString string
		secret      string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "successful case",
			args: args{
				tokenString: tokenHelper("1234-1234-1234-1234"),
				secret:      "manager",
			},
			want:    "1234-1234-1234-1234",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractToken(tt.args.tokenString, tt.args.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
