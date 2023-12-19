package main

import "testing"

func Test_toLatin(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				"АаБбВвГгДдЕеЁёЖжЗзИиЙйКкЛлМмНнОоПпРрСсТтУуФфХхЦцЧчШшЩщЪъЫыЬьЭэЮюЯя",
			},
			want: "AaБ6BвГrДgEeEeЖж3зUuUuKkЛлMmHнOoПnPpCcTтYyФфXxЦцЧчШшЩщЪъЫыbьЭэЮюЯя",
		},
		{
			name: "latin",
			args: args{
				"abcdefghijklmnopqrstuvwxyz01234567890",
			},
			want: "abcdefghijklmnopqrstuvwxyz01234567890",
		},
		{
			name: "diacritic",
			args: args{
				"Pâté",
			},
			want: "Pate",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toLatin(tt.args.str); got != tt.want {
				t.Errorf("toLatin() = %v, want %v", got, tt.want)
			}
		})
	}
}
