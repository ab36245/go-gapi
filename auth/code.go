package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"strconv"

	"golang.org/x/oauth2"
)

func codeFromWeb(config *oauth2.Config) (string, error) {
	if err := codeClient(config); err != nil {
		return "", err
	}
	return codeServer(config)
}

func codeClient(config *oauth2.Config) error {
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// TODO: this will only work on a mac!
	cmd := exec.Command("open", url)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("could not start browser: %w", err)
	}
	return nil
}

func codeServer(config *oauth2.Config) (string, error) {
	code := ""
	var err error

	url, err := url.Parse(config.RedirectURL)
	if err != nil {
		return "", fmt.Errorf("unparsable URL: %w", err)
	}
	port := url.Port()
	if port == "" {
		port = "80"
	} else if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("invalid port %q: %w", port, err)
	}

	handler := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			err = codeServerFailure(w, "expected GET request, got %s instead", r.Method)
			server.Close()
			return
		}
		query := r.URL.Query()
		for k, v := range query {
			fmt.Printf("query %v = %v\n", k, v)
		}
		code = query.Get("code")
		if code == "" {
			err = codeServerFailure(w, "no code in query string")
			server.Close()
			return
		}
		scope := query.Get("scope")
		if scope == "" {
			err = codeServerFailure(w, "no scope in query string")
			server.Close()
			return
		}
		codeServerSuccess(w)
		server.Close()
	})
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return "", fmt.Errorf("server error: %w", err)
		}
	}
	return code, err
}

func codeServerFailure(response http.ResponseWriter, mesg string, args ...any) error {
	err := fmt.Errorf(mesg, args...)
	w := &writer{response}
	w.code(http.StatusForbidden)
	w.start()
	w.line("<head>")
	w.line("  <title>%s</title>", "Authorization failed")
	w.line("</head>")
	w.line("<body>")
	w.line("  <p>%s</p>", "Sorry, the authentication failed:")
	w.line("  <p>%s</p>", err.Error())
	w.line("</body>")
	w.end()
	return err
}

func codeServerSuccess(response http.ResponseWriter) {
	w := &writer{response}
	w.code(http.StatusOK)
	w.start()
	w.line("<head>")
	w.line("  <title>%s</title>", "Authorization succeeded")
	w.line("</head>")
	w.line("<body>")
	w.line("  <p>%s</p>", "The authentication succeeded")
	w.line("  <p>%s</p>", "You can close this page now")
	w.line("</body>")
	w.end()
}
