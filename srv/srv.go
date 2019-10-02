// MIT License
//
// Copyright 2018 Canonical Ledgers, LLC
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package srv

import (
	"net/http"

	jrpc "github.com/AdamSLevy/jsonrpc2/v11"
	"github.com/Factom-Asset-Tokens/fatd/flag"
	"github.com/goji/httpauth"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var srv http.Server

// Start the server in its own goroutine. If stop is closed, the server is
// closed and any goroutines will exit. The done channel is closed when the
// server exits for any reason. If the done channel is closed before the stop
// channel is closed, an error occurred. Errors are logged.
func Start(stop <-chan struct{}) (done <-chan struct{}) {
	// Set up JSON RPC 2.0 handler with correct headers.
	jrpc.DebugMethodFunc = true
	jrpcHandler := jrpc.HTTPRequestHandler(jrpcMethods)

	var handler http.Handler = http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			jrpcHandler(w, r)
		})
	if flag.HasAuth {
		authOpts := httpauth.AuthOptions{
			User:     flag.Username,
			Password: flag.Password,
			UnauthorizedHandler: http.HandlerFunc(
				func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte(`{}`))
				}),
		}
		handler = httpauth.BasicAuth(authOpts)(handler)
	}

	// Set up server.
	srvMux := http.NewServeMux()

	srvMux.Handle("/", handler)
	srvMux.Handle("/v1", handler)

	cors := cors.New(cors.Options{AllowedOrigins: []string{"*"}})
	srv = http.Server{Handler: cors.Handler(srvMux)}

	srv.Addr = flag.APIAddress

	// Start server.
	_done := make(chan struct{})
	log.Infof("Listening on %v...", flag.APIAddress)
	go func() {
		var err error
		if flag.HasTLS {
			err = srv.ListenAndServeTLS(flag.TLSCertFile, flag.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != http.ErrServerClosed {
			log.Errorf("srv.ListenAndServe(): %v", err)
		}
		close(_done)
	}()
	// Listen for stop signal.
	go func() {
		select {
		case <-stop:
			if err := srv.Shutdown(nil); err != nil {
				log.Errorf("srv.Shutdown(): %v", err)
			}
		case <-_done:
		}
	}()
	return _done
}
