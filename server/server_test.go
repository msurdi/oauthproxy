package server

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

var server *httptest.Server

func init() {
	LoadConfig("oauthproxy.conf.test")
}

func genericRequestHandlerAuthWrapper(w http.ResponseWriter, r *http.Request) {
	authorizeEmail("test@example.com", w, r)
	genericRequestHandler(w, r)
}

func TestUnauthRequest(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(serverConfig.ProtectPath, genericRequestHandler)
	mux.HandleFunc(serverConfig.CallbackPath, oauthCallbackHandler)
	server = httptest.NewServer(mux)

	req, _ := http.NewRequest("GET", server.URL, nil)
	client := new(http.Transport)
	response, err := client.RoundTrip(req)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	location := response.Header.Get("Location")
	if !strings.HasPrefix(location, "https://accounts.google.com/o/oauth2") {
		t.Errorf("Expected Location header incorrect: %q", location)
	}
}

func TestAuthRequest(t *testing.T) {
	const backendResponse = "I am the backend"
	const backendStatus = 200
	mux := http.NewServeMux()
	mux.HandleFunc(serverConfig.ProtectPath, genericRequestHandlerAuthWrapper)
	mux.HandleFunc(serverConfig.CallbackPath, oauthCallbackHandler)
	server = httptest.NewServer(mux)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TransferEncoding) > 0 {
			t.Errorf("backend got unexpected TransferEncoding: %v", r.TransferEncoding)
		}
		if r.Header.Get("X-Forwarded-For") == "" {
			t.Errorf("didn't get X-Forwarded-For header")
		}

		w.Header().Set("X-Foo", "bar")
		http.SetCookie(w, &http.Cookie{Name: "flavor", Value: "chocolateChip"})
		w.WriteHeader(backendStatus)
		w.Write([]byte(backendResponse))
	}))
	defer backend.Close()
	backendURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatal(err)
	}

	serverConfig.ProxyURL = *backendURL
	// Test request & response
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Host = "some-name"
	req.Header.Set("Connection", "close")
	req.Close = true

	// Handle the request
	transport := new(http.Transport)
	response, err := transport.RoundTrip(req)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if g, e := response.StatusCode, backendStatus; g != e {
		t.Errorf("got response.StatusCode %d; expected %d", g, e)
	}
	bodyBytes, _ := ioutil.ReadAll(response.Body)
	if g, e := string(bodyBytes), backendResponse; g != e {
		t.Errorf("got body %q; expected %q", g, e)
	}
	if g, e := response.Header.Get("X-Foo"), "bar"; g != e {
		t.Errorf("got X-Foo %q; expected %q", g, e)
	}
	if cookie := response.Cookies()[0]; cookie.Name != "flavor" {
		t.Errorf("unexpected cookie %q", cookie.Name)
	}
}

func TestIsAuthorized(t *testing.T) {
	re, _ := regexp.Compile(".*@example.com")
	oauthProviderConfig.EmailRegexp = *re
	if authorized, email := isAuthorized([]byte("test@other.com")); authorized {
		t.Error("Invalid authorization, email:", email)
	}

	re, _ = regexp.Compile(".*@example.com")
	oauthProviderConfig.EmailRegexp = *re
	if authorized, email := isAuthorized([]byte("test@example.com")); !authorized {
		t.Error("Authorization failed, email:", email)
	}

}
