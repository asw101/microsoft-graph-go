package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/microsoft"
)

var config = struct {
	Config *oauth2.Config
	Token  *oauth2.Token
}{}

func main() {
	clientID := os.Getenv("AZ_CLIENT_ID")
	clientSecret := os.Getenv("AZ_CLIENT_SECRET")
	tenantID := os.Getenv("AZ_TENANT_ID")
	scopes := os.Getenv("AZ_GRAPH_SCOPES")

	config.Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     microsoft.AzureADEndpoint(tenantID),
		Scopes:       strings.Split(scopes, " "),
		RedirectURL:  "http://localhost:8080/auth",
	}

	switch os.Getenv("MODE") {
	case "web":
		web()
	default:
		cli(config.Config)
	}

}

func cli(conf *oauth2.Config) {
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url1 := conf.AuthCodeURL("state", oauth2.AccessTypeOnline)
	log.Printf("Visit URL: %v\nPaste URL: ", url1)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}

	if url1, err := url.Parse(code); err == nil {
		code = url1.Query().Get("code")
	}

	ctx := context.Background()
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	b, _ := json.Marshal(tok)
	fmt.Printf("%s\n", b)

}

func web() {
	http.HandleFunc("/me", httpMe)
	http.HandleFunc("/login", httpLogin)
	http.HandleFunc("/auth", httpAuth)
	http.HandleFunc("/token", httpToken)

	listenPort := ":8080"
	fmt.Printf("Listening on %s\n", listenPort)
	err := http.ListenAndServe(listenPort, httpLog(http.DefaultServeMux))
	if err != nil {
		log.Fatal(err)
	}

}

func httpLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
			handler.ServeHTTP(w, r)
		})

}

func httpMe(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	client := config.Config.Client(ctx, config.Token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	fmt.Fprintf(w, "%s\n", b)

}

func httpLogin(w http.ResponseWriter, r *http.Request) {
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url1 := config.Config.AuthCodeURL("state", oauth2.AccessTypeOnline)
	http.Redirect(w, r, url1, 302)

}

func httpAuth(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Parameter required: code", 500)
		return
	}
	tok, err := config.Config.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	config.Token = tok
	b, _ := json.Marshal(tok)
	fmt.Fprintf(w, "%s\n", b)

}

func httpToken(w http.ResponseWriter, r *http.Request) {
	b, _ := json.Marshal(config.Token)
	fmt.Fprintf(w, "%s\n", b)

}
