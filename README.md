OAuthProxy
=========
Developed by Matias Surdi <matias.surdi@gmail.com>

Licensed under the Apache 2.0 License.

Disclaimer: Use at your own risk, I can't warranty this software has no security flaws.


Introduction
------------
OAuthproxy is a network daemon that will listen for incoming HTTP(S) requests on a given port, 
setup a session, request authentication via OAuth, and if the email address obtained from the 
oauth authentication process is allowed, then it will proxy every ongoing request from that 
session to a backend server. The process is totally transparent for the end user, except for the
authentication step.

Feel free to contact me if you need any help with getting it up and running, any suggestion to improve it, etc.

Known alternatives
------------------
The only similar alternative I've found to this is Apache's [mod_auth_openid](http://findingscience.com/mod_auth_openid/)
module. But it is based on openid and not on OAuth as you have guessed.

If You find any other alternative, feel free to suggest it for updating this document.

Deployment
----------
  1. Download the latest stable binary release from the [Downloads](https://github.com/msurdi/oauthproxy/downloads) section.
  2. Uncompress and upload all the files to your server.
  3. Copy the provided _oauthproxy.conf.example_ to _oauthproxy.conf_ and adapt to your needs.
  4. If you enable HTTPS, replace the provided cert.pem and key.pem for the right ones for your domain
  5. Run the daemon with ./oauthproxy -config /path/to/oauthproxy.conf


Development/Contributing
------------------------
Any contribution is welcome. Testing, documentation, or feedback. If you want to contribute code,
fork the project on Github and submit a pull request.

OAuthproxy has been developed with the [Go](http://golang.org) programming language.

ToDo/Ideas
-------------
    * Support multiple backend servers
    * Support backend selection reading the Host: header from the client request
    * Test with more oauth providers
    * Support multiple OAuth providers
    * Write more tests
    * Logging to syslog
    * Provide session storage alternatives
    * Send the logged in user email as a header or url parameter?

Current limitations
-------------------
    * Oauth 2.0
    * Tested only with google OAuth service

F.A.Q.
------
  1. Why [Go](http://golang.org)?

  Because I like it and I felt like this project was a good fit for it.
  
  2. Why OAuthProxy?
  
  Because sometimes setting up a VPN client on every company laptop is too much work. OAuthproxy
  won't be always the alternative of course, but for many cases it is a cheap and easy way to protect
  intranet applications from outsiders without all the hassle of a VPN. I also needed a good excuse
  to get my hands dirty with Go.

  
