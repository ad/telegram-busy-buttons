package main

import "testing"

func Test_shortenButtonText(t *testing.T) {
	type args struct {
		command  string
		name     string
		lastname string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{
				command:  "free-testing-1",
				name:     "",
				lastname: "",
			},
			want: "",
		},
		{
			name: "normal",
			args: args{
				command:  "free-testing-1",
				name:     "veryveryveryveryvery",
				lastname: "veryveryveryveryveryveryveryvery2",
			},
			want: "veryveryveryveryvery v.",
		},
		{
			name: "long",
			args: args{
				command:  "free-testing-1",
				name:     "veryveryveryveryvery1",
				lastname: "veryveryveryveryveryveryveryvery",
			},
			want: "veryveryveryveryvery1 v.",
		},
		{
			name: "long name",
			args: args{
				command:  "free-testing-1",
				name:     "veryveryveryveryveryveryveryveryveryveryveryvery",
				lastname: "veryveryveryveryveryveryveryvery",
			},
			want: "veryveryveryveryveryveryveryvery",
		},
		{
			name: "long lastname",
			args: args{
				command:  "free-testing-1",
				name:     "1veryveryveryveryveryveryveryver",
				lastname: "2veryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryveryvery",
			},
			want: "2veryveryveryveryveryveryveryver",
		},
		{
			name: "normal",
			args: args{
				command:  "free-testing-1",
				name:     "Firstname",
				lastname: "Firstname",
			},
			want: "Firstname Firstname",
		},
		{
			name: "normal cyrrilic",
			args: args{
				command:  "free-testing-1",
				name:     "ИмяОтчество",
				lastname: "Фамилия",
			},
			want: "ИмяОтчество Ф.",
		},
		{
			name: "long cyrrilic name",
			args: args{
				command:  "free-testing-1",
				name:     "ИмяОтчество12345678901234567890123123123123123123123123123131231223131231231231312312123123123123123123123",
				lastname: "Фамилия",
			},
			want: "ИмяОтчество123456789012345678901",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortenUsername(tt.args.command, tt.args.name, tt.args.lastname)

			if got != tt.want {
				t.Errorf("shortenUsername(%s, %s, %s) = %v, want %v", tt.args.command, tt.args.name, tt.args.lastname, got, tt.want)
			}
		})
	}
}
