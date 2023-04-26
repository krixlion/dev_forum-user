package db

import "testing"

func Test_verifyField(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test if created_at field is found",
			args: args{input: "created_at"},
		},
		{
			name: "Test if name field is found",
			args: args{input: "name"},
		},
		{
			name: "Test if id field is found",
			args: args{input: "id"},
		},
		{
			name: "Test if email field is found",
			args: args{input: "email"},
		},
		{
			name: "Test if password field is found",
			args: args{input: "password"},
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := verifyField(tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("verifyField() set: %d\n error = %v, wantErr %v", i, err, tt.wantErr)
			}
		})
	}
}
