package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// withPort returns hostport with port when it names none, for the guard of
// a run to check the same host and port the action dials. An IPv6 address
// may be given bare or in brackets.
func withPort(hostport, port string) string {
	if _, _, err := net.SplitHostPort(hostport); err == nil {
		return hostport
	}
	return net.JoinHostPort(strings.Trim(hostport, "[]"), port)
}

// hostOfURL returns the host and port of u, taken at the port of its scheme
// when it names none: 443 for https, and 80 for any other.
func hostOfURL(u *url.URL) string {
	if u.Port() != "" {
		return u.Host
	}
	port := "80"
	if u.Scheme == "https" {
		port = "443"
	}
	return net.JoinHostPort(u.Hostname(), port)
}

// decodeJSON returns the JSON object or array s holds, or, for anything
// else, a map whose error_message says why it is not one.
func decodeJSON(s string) any {
	s = strings.TrimSpace(s)
	var err error
	switch {
	case s == "":
		err = fmt.Errorf("empty string")
	case s[0] == '{':
		var obj map[string]any
		if err = json.Unmarshal([]byte(s), &obj); err == nil {
			return obj
		}
	case s[0] == '[':
		var arr []any
		if err = json.Unmarshal([]byte(s), &arr); err == nil {
			return arr
		}
	default:
		err = fmt.Errorf("input is not a JSON object or array")
	}
	return map[string]any{"error_message": fmt.Sprintf("invalid JSON: %s", err)}
}
