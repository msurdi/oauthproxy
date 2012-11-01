package server

import (
	"testing"
)

func TestConfig(t *testing.T) {
	LoadConfig("oauthproxy.conf.test")
}
