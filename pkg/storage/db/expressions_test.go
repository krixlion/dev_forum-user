package db

import "testing"

func Test_findField(t *testing.T) {
	type args struct {
		input string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Test if created_at field is found",
			args: args{input: "created_at"},
			want: "created_at",
		},
		{
			name: "Test if name field is found",
			args: args{input: "name"},
			want: "name",
		},
		{
			name: "Test if id field is found",
			args: args{input: "id"},
			want: "id",
		},
		{
			name: "Test if email field is found",
			args: args{input: "email"},
			want: "email",
		},
		{
			name: "Test if password field is found",
			args: args{input: "password"},
			want: "password",
		},
	}
	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findField(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("findField() set: %d\n error = %v, wantErr %v", i, err, tt.wantErr)
				return
			}
			if got != tt.want && !tt.wantErr {
				t.Errorf("findField() set: %d\n got = %+v\n want %+v\n", i, got, tt.want)
			}
		})
	}
}
