// Copyright © 2019 Vish Vishvanath <vishvish@users.noreply.github.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/99designs/keyring"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/url"
	"os"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate to the FreeAgent API",
	Long:  `Login and fetch a token from the FreeAgent API, and store it in your secure system credential store.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("login called")
		checkAuth()
	},
}

var (
	conf         *oauth2.Config
	ctx          context.Context
	serviceName  = "io.koothooloo.frag."
	clientID     = os.Getenv("FREEAGENT_CLIENT_ID")
	clientSecret = os.Getenv("FREEAGENT_CLIENT_SECRET")
	fragUrl      string
	srv          http.Server
)

func init() {
	rootCmd.AddCommand(loginCmd)
}

func checkAuth() {
	s, retrieveError := retrieveItem()
	if retrieveError != nil {
		initAuth()
	} else {
		log.Println("Token found in keychain.")
		var t oauth2.Token
		err := json.Unmarshal(s.Data, &t)
		if err != nil {
			initAuth()
		}
	}
}

func initAuth() {
	conf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.freeagent.com/v2/approve_app",
			TokenURL: "https://api.freeagent.com/v2/token_endpoint",
		},
		// my own callback URL
	}

	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	fragUrl = conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Println(color.CyanString("You will now be taken to your browser for authentication"))
	//time.Sleep(1 * time.Second)
	log.Printf("Authentication URL: %s\n", fragUrl)
	open.Run(fragUrl)
	//time.Sleep(1 * time.Second)

	startHTTPServer()
	log.Printf("Finished login")
}

func startHTTPServer() {
	m := http.NewServeMux()
	srv = http.Server{Addr: ":9999", Handler: m}
	ctx2, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		queryParts, _ := url.ParseQuery(r.URL.RawQuery)

		// Use the authorization code that is pushed to the redirect
		// URL.
		code := queryParts["code"][0]
		//log.Printf("code: %s\n", code)

		// Exchange will do the handshake to retrieve the initial access token.
		tok, err := conf.Exchange(ctx2, code)
		if err != nil {
			log.Fatal(err)
		}

		storeItem(tokenToItem(tok))

		// The HTTP Client returned by conf.Client will refresh the token as necessary.
		client := conf.Client(ctx2, tok)

		resp, err := client.Get("https://api.freeagent.com/v2/users/me")
		if err != nil {
			log.Fatal(err)
		} else {
			msg := "<p><strong>Success!</strong></p>"
			msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
			fmt.Fprintf(w, msg)

			cancel()
			log.Println(color.CyanString("Authentication successful"))
		}
		defer resp.Body.Close()
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	select {
	case <-ctx2.Done():
		// Shutdown the server when the context is canceled
		srv.Shutdown(ctx2)
	}
}

func tokenToItem(tok *oauth2.Token) keyring.Item {
	bytes, err := json.Marshal(tok)
	if err != nil {
		panic(err)
	}
	//fmt.Println("Encoded JSON ", bytes)
	item := keyring.Item{
		Key:                         serviceName + "token",
		Data:                        bytes,
		Label:                       serviceName + "token",
		Description:                 "oauth2 token",
		KeychainNotTrustApplication: false,
		KeychainNotSynchronizable:   false,
	}
	return item
}

//func itemToToken(item keyring.Item) oauth2.Token {
//	var t oauth2.Token
//	err := json.Unmarshal(item.Data, &t)
//	if err != nil {
//		panic(err)
//	}
//	return t
//}

func storeItem(item keyring.Item) {
	log.Println("Setting " + item.Key)

	ring, _ := keyring.Open(keyring.Config{
		ServiceName:  item.Key,
		KeychainName: "login",
	})

	err := ring.Set(item)
	if err != nil {
		log.Fatal(err)
	}
}

func retrieveItem() (keyring.Item, error) {
	log.Println("Getting " + serviceName + "token")

	ring, _ := keyring.Open(keyring.Config{
		ServiceName:  serviceName + "token",
		KeychainName: "login",
	})

	i, err := ring.Get(serviceName + "token")

	return i, err
}
