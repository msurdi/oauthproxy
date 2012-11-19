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

// OAuth Proxy server

package server

import (
	"code.google.com/p/goauth2/oauth"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
)

//store stores all server related configurations
var store sessions.CookieStore

// getUser checks if a given request belongs to an authenticated
// session
//
// A single *http.Request argument is required
func getUser(r *http.Request) string {
	session, _ := store.Get(r, serverConfig.CookieName)
	authenticated, ok := session.Values["auth"]
	if ok == true && authenticated == true {
		email := session.Values["email"]
		if email != nil {
			return email.(string)
		}
	}
	return ""
}

// proxy proxies a request to a backend server
//
//
func proxy(w http.ResponseWriter, r *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(&serverConfig.ProxyURL)
	proxy.ServeHTTP(w, r)
}

// Redirect the user to the OAuth provider for entering login information
func askForLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, oauthProviderConfig.oauthConfig.AuthCodeURL(""), http.StatusFound)
}

// Handle every non-callback request
func genericRequestHandler(w http.ResponseWriter, r *http.Request) {
	email := getUser(r)
	if email != "" {
		log.Println(email, r.URL.String())
		proxy(w, r)
	} else {
		session, _ := store.Get(r, serverConfig.CookieName)
		session.Values["next"] = r.URL.String()
		session.Save(r, w)
		log.Println("Asking for login, request from unknown user to:", r.URL.String())
		askForLogin(w, r)
	}
}

// isAuthorized checks if the email regular expression provided by
// configuration matches anything in the oauth provider response.
// If it does, then that match is considered the user email and
// a valid account to access any backend server
func isAuthorized(body []byte) (bool, string) {
	email := oauthProviderConfig.EmailRegexp.FindString(string(body))
	if email != "" {
		return true, email
	}
	return false, ""
}

// authorizeEmail configures the current session as an authenticated user
// being the user email provided as first argument the email of the 
// logged in user.
func authorizeEmail(email string, w http.ResponseWriter, r *http.Request) {
	log.Println("User", email, "logged in")
	session, _ := store.Get(r, serverConfig.CookieName)
	session.Values["auth"] = true
	session.Values["email"] = email
	session.Save(r, w)
}

// oauthCallbackHandler handles the request that happens after the
// oauth provider completes user identification/approval request and
// redirects back to us the user.
func oauthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	transport := &oauth.Transport{Config: &oauthProviderConfig.oauthConfig}
	transport.Exchange(r.FormValue("code"))
	client := transport.Client()
	response, err := client.Get(oauthProviderConfig.UserInfoAPI)
	if err != nil {
		log.Printf("Error while contacting '%s': %s\n", oauthProviderConfig.UserInfoAPI, err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("Error while parsing response from '%s': %s\n", oauthProviderConfig.UserInfoAPI, err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}
	response.Body.Close()
	authorized, email := isAuthorized(body)
	if authorized {
		authorizeEmail(email, w, r)
		log.Println("User", email, "logged in")
		session, _ := store.Get(r, serverConfig.CookieName)
		if next, ok := session.Values["next"]; ok {
			http.Redirect(w, r, next.(string), http.StatusFound)
		}
	} else {
		log.Println("Access Denied: Couldn't match an email address in the server response.")
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	}
}

// Run is the server entrypoint. It will listen on the addresses provided
// in the configuration file, for http and https.
func Run() {
	store = *sessions.NewCookieStore(securecookie.GenerateRandomKey(10))
	http.HandleFunc(serverConfig.CallbackPath, oauthCallbackHandler)
	http.HandleFunc(serverConfig.ProtectPath, genericRequestHandler)

	done := make(chan bool)

	go func() {
		if serverConfig.ListenAddress != "" {
			err := http.ListenAndServe(serverConfig.ListenAddress, nil)
			if err != nil {
				log.Fatal(err)
			}
			done <- true
		}
	}()

	go func() {
		if serverConfig.ListenAddressTLS != "" {
			err := http.ListenAndServeTLS(serverConfig.ListenAddressTLS, serverConfig.SSLCert, serverConfig.SSLKey, nil)
			if err != nil {
				log.Fatal(err)
			}
			done <- true
		}
	}()
	<-done
}
