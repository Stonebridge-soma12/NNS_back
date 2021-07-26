package service

import "testing"

func Test_validatePassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test 1",
			args: args{password: ""},
			wantErr: true,
		},
		{
			name: "Test 2",
			args: args{password: "abcde"},
			wantErr: true,
		},
		{
			name: "Test 3",
			args: args{password: "12345"},
			wantErr: true,
		},
		{
			name: "Test 4",
			args: args{password: "!@#$%"},
			wantErr: true,
		},
		{
			name: "Test 5",
			args: args{password: "abcde12345"},
			wantErr: true,
		},
		{
			name: "Test 6",
			args: args{password: "abcde!@#$%"},
			wantErr: true,
		},
		{
			name: "Test 7",
			args: args{password: "12345!@#$%"},
			wantErr: true,
		},
		{
			name: "Test 8",
			args: args{password: "abcde12345!@#$%"},
			wantErr: false,
		},
		{
			name: "Test 9",
			args: args{password: "12345abcde!@#$%"},
			wantErr: false,
		},
		{
			name: "Test 10",
			args: args{password: "!@#$512345abcde"},
			wantErr: false,
		},
		{
			name: "Test 11",
			args: args{password: "passWord1@"},
			wantErr: false,
		},
		{
			name: "Test 12",
			args: args{password: "passWord1_"},
			wantErr: false,
		},
		{
			name: "Test 13",
			args: args{password: "passWord1.| "},
			wantErr: false,
		},
		{
			name: "Test 14",
			args: args{password: "aA1!"},
			wantErr: true,
		},
		{
			name: "Test 15",
			args: args{password: "1234567890123456789012345678901234567890123456789012345678901234567890a!"},
			wantErr: false,
		},
		{
			name: "Test 15",
			args: args{password: "1234567890123456789012345678901234567890123456789012345678901234567890a!a"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validatePassword(tt.args.password); (err != nil) != tt.wantErr {
				t.Errorf("validatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
