package web

import (
	"testing"
)

func TestGetListenAddr(t *testing.T) {
	tests := []struct {
		name string
		host string
		port string
		want string
	}{
		{
			name: "defaults to loopback and 8080 when neither env var is set",
			host: "",
			port: "",
			want: "127.0.0.1:8080",
		},
		{
			name: "honors HOST env var",
			host: "0.0.0.0",
			port: "",
			want: "0.0.0.0:8080",
		},
		{
			name: "honors PORT env var",
			host: "",
			port: "3000",
			want: "127.0.0.1:3000",
		},
		{
			name: "honors both HOST and PORT",
			host: "192.168.1.1",
			port: "9000",
			want: "192.168.1.1:9000",
		},
		{
			name: "HOST=0.0.0.0 exposes on all interfaces",
			host: "0.0.0.0",
			port: "8080",
			want: "0.0.0.0:8080",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOST", tt.host)
			t.Setenv("PORT", tt.port)
			if got := getListenAddr(); got != tt.want {
				t.Errorf("getListenAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}
