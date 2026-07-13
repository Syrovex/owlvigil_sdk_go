package main

import "testing"

func TestOAuthEnabled(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		clientID     string
		clientSecret string
		want         bool
	}{
		{name: "access token", accessToken: "token", want: true},
		{name: "client credentials", clientID: "client", clientSecret: "secret", want: true},
		{name: "no OAuth credentials", want: false},
		{name: "client ID only", clientID: "client", want: false},
		{name: "client secret only", clientSecret: "secret", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oauthEnabled(tt.accessToken, tt.clientID, tt.clientSecret); got != tt.want {
				t.Errorf("oauthEnabled(%q, %q, %q) = %t, want %t", tt.accessToken, tt.clientID, tt.clientSecret, got, tt.want)
			}
		})
	}
}

func TestWriteSmokeEnabled(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{value: "1", want: true},
		{value: "true", want: false},
		{value: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := writeSmokeEnabled(tt.value); got != tt.want {
				t.Errorf("writeSmokeEnabled(%q) = %t, want %t", tt.value, got, tt.want)
			}
		})
	}
}
