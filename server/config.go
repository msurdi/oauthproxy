/*
 Copyright 2012 Matias Surdi

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package server

import (
	"code.google.com/p/goauth2/oauth"
	"errors"
	"github.com/kless/goconfig/config"
	"log"
	"net/url"
	"regexp"
	"strings"
)

// Implements configuration for the server process
type ServerConfig struct {
	ListenAddress    string
	ListenAddressTLS string
	ProxyURL         url.URL
	CookieName       string
	CallbackPath     string
	ProtectPath      string
	SSLKey           string
	SSLCert          string
}

// Implements configuration for an OAuth provider
type OAuthProviderConfig struct {
	EmailRegexp regexp.Regexp
	UserInfoAPI string
	oauthConfig oauth.Config
}

// Actual storage for oauthproxy parameters
var serverConfig ServerConfig

// Actual storage for the oauth provider
var oauthProviderConfig OAuthProviderConfig

// readConfigString reads a given option from a given section from the
// provided config.Config instance, returning the value as a string.
// If the option is not present, then a string with the default value
// provided in 'def' is returned.
func readConfigString(configFile *config.Config, section string, option string, def string) string {
	value, err := configFile.String(section, option)
	if err != nil {
		return def
	}
	return strings.TrimSpace(value)
}

// readConfigRegexp reads a given option from a given section from the
// provided config.Config instance, returning the value as a regexp.Regexp.
// If the option is not present, then a regexp.Regexp with the default value
// provided in the string 'def' is returned.
func readConfigRegexp(configFile *config.Config, section string, option string, def string) regexp.Regexp {
	value, err := configFile.String(section, option)
	if err != nil {
		regexp, _ := regexp.Compile(def)
		return *regexp
	}
	regexp, err := regexp.Compile(value)
	if err != nil {
		log.Fatal(err)
	}
	return *regexp
}

// readConfigURL reads a given option from a given section from the
// provided config.Config instance, returning the value as a url.URL.
// If the option is not present, then an url.URL with the default value
// provided in the string 'def' is returned.
func readConfigURL(configFile *config.Config, section string, option string, def string) url.URL {
	value, err := configFile.String(section, option)
	if err != nil {
		url_, _ := url.Parse(def)
		return *url_
	}
	url_, err := url.Parse(value)
	if err != nil {
		log.Fatal(err)
	}
	return *url_
}

// LoadConfig reads and validates the server and oauth provider configuration from
// the provided config file path. If for some reason it cannot load the settings
// defined in the config file, or the config file cannot be read at all, it will
// return an error.
func LoadConfig(path string) error {
	// load config file
	configFile, err := config.ReadDefault(path)
	if err != nil {
		log.Fatal("Can't read config file:", err)
	}
	// validate server configuration
	serverConfig.ListenAddress = readConfigString(configFile, "server", "ListenAddress", "localhost:8080")
	serverConfig.ListenAddressTLS = readConfigString(configFile, "server", "ListenAddressTLS", "localhost:4443")
	if serverConfig.ListenAddress == "" && serverConfig.ListenAddressTLS == "" {
		return errors.New("At least one of ListenAddress or ListenAddressTLS must be present")
	}

	// Validate certs if SSL is enabled
	if serverConfig.ListenAddressTLS != "" {
		serverConfig.SSLCert = readConfigString(configFile, "server", "SSLCert", "")
		serverConfig.SSLKey = readConfigString(configFile, "server", "SSLKey", "")

		if serverConfig.SSLCert == "" {
			return errors.New("A SSL Certificate is required")
		}

		if serverConfig.SSLKey == "" {
			return errors.New("A SSL Key is required")
		}
	}

	serverConfig.CallbackPath = readConfigString(configFile, "server", "CallbackPath", "/oauth2callback")
	serverConfig.ProtectPath = readConfigString(configFile, "server", "ProtectPath", "/")
	serverConfig.CookieName = readConfigString(configFile, "server", "CookieName", "oauthproxy")
	serverConfig.ProxyURL = readConfigURL(configFile, "server", "ProxyURL", "http://example.com/")
	// Validate OAuth settings
	oauthProviderConfig.EmailRegexp = readConfigRegexp(configFile, "oauth", "EmailRegexp", ".*")
	oauthProviderConfig.oauthConfig.ClientId = readConfigString(configFile, "oauth", "ClientId", "")
	oauthProviderConfig.oauthConfig.ClientSecret = readConfigString(configFile, "oauth", "ClientSecret", "")
	scope := readConfigURL(configFile, "oauth", "Scope", "https://www.googleapis.com/auth/userinfo.email")
	oauthProviderConfig.oauthConfig.Scope = scope.String()
	authURL := readConfigURL(configFile, "oauth", "AuthURL", "https://accounts.google.com/o/oauth2/auth")
	oauthProviderConfig.oauthConfig.AuthURL = authURL.String()
	tokenURL := readConfigURL(configFile, "oauth", "TokenURL", "https://accounts.google.com/o/oauth2/token")
	oauthProviderConfig.oauthConfig.TokenURL = tokenURL.String()
	redirectURL := readConfigURL(configFile, "oauth", "RedirectURL", "http://testsite.com/oauth2callback")
	oauthProviderConfig.oauthConfig.RedirectURL = redirectURL.String()
	userInfoAPI := readConfigURL(configFile, "oauth", "UserInfoAPI", "https://www.googleapis.com/oauth2/v1/userinfo")
	oauthProviderConfig.UserInfoAPI = userInfoAPI.String()
	return nil
}
