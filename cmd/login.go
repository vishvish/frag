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
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/99designs/keyring"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("login called")
		checkAuth()
	},
}

var (
	conf         *oauth2.Config
	ctx          context.Context
	serviceName  = "com.tfto.myFreeAgent."
	clientID     = "***REMOVED***"
	clientSecret = "***REMOVED***"
	srv          http.Server
)

func init() {
	rootCmd.AddCommand(loginCmd)

	//s := retrieveItem()
	//
	//var t oauth2.Token
	//err := json.Unmarshal(s.Data, &t)
	//if err != nil {
	//	initAuth()
	//} else {
	//	log.Printf("Access: %s", t.AccessToken)
	//	log.Printf("Refresh: %s", t.RefreshToken)
	//}

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkAuth() {
	s := retrieveItem()
	var t oauth2.Token
	err := json.Unmarshal(s.Data, &t)
	if err != nil {
		initAuth()
	} else {
		log.Println("Token found in keychain.")
		// log.Printf("Access: %s", t.AccessToken)
		// log.Printf("Refresh: %s", t.RefreshToken)
	}
}

func initAuth() {
	ctx = context.Background()
	conf = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"openid", "profile"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.freeagent.com/v2/approve_app",
			TokenURL: "https://api.freeagent.com/v2/token_endpoint",
		},
		// my own callback URL
		RedirectURL: "http://127.0.0.1:9999/oauth/callback",
	}

	// add transport for self-signed certificate to context
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslcli := &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslcli)

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)

	log.Println(color.CyanString("You will now be taken to your browser for authentication"))
	time.Sleep(1 * time.Second)
	open.Run(url)
	time.Sleep(1 * time.Second)
	log.Printf("Authentication URL: %s\n", url)

	http.HandleFunc("/oauth/callback", callbackHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))

	//srv = *startHTTPServer()
}

func startHTTPServer() *http.Server {
	s := &http.Server{Addr: ":9999"}

	http.HandleFunc("/oauth/callback", callbackHandler)
	log.Fatal(s.ListenAndServe())

	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	io.WriteString(w, "hello world\n")
	//})

	go func() {
		// returns ErrServerClosed on graceful close
		if err := s.ListenAndServe(); err != http.ErrServerClosed {
			// NOTE: there is a chance that next line won't have time to run,
			// as main() doesn't wait for this goroutine to stop. don't use
			// code with race conditions like these for production. see post
			// comments below on more discussion on how to handle this.
			log.Fatalf("ListenAndServe(): %s", err)
		}
		fmt.Println("Server gracefully stopped")
	}()

	// returning reference so caller can call Shutdown()
	return &srv
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

func itemToToken(item keyring.Item) oauth2.Token {
	var t oauth2.Token
	err := json.Unmarshal(item.Data, &t)
	if err != nil {
		panic(err)
	}
	return t
}

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

func retrieveItem() keyring.Item {
	log.Println("Getting " + serviceName + "token")

	ring, _ := keyring.Open(keyring.Config{
		ServiceName:  serviceName + "token",
		KeychainName: "login",
	})

	i, err := ring.Get(serviceName + "token")
	if err != nil {
		initAuth()
	}

	return i
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {

	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// Use the authorization code that is pushed to the redirect
	// URL.
	code := queryParts["code"][0]
	//log.Printf("code: %s\n", code)

	// Exchange will do the handshake to retrieve the initial access token.
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	storeItem(tokenToItem(tok))

	// The HTTP Client returned by conf.Client will refresh the token as necessary.
	client := conf.Client(ctx, tok)

	resp, err := client.Get("https://api.freeagent.com/v2/users/me")
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(color.CyanString("Authentication successful\n"))
	}

	defer resp.Body.Close()
	// show success page
	msg := "<p><strong>Success!</strong></p>"
	msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
	fmt.Fprintf(w, msg)
}
