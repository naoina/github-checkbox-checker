package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
	"github.com/naoina/denco"
)

const (
	defaultAddr = "0.0.0.0:4567"
)

var (
	githubAccessToken   = os.Getenv("GITHUB_ACCESS_TOKEN")
	githubWebHookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
)

type githubWebHookHandler struct {
	client *github.Client
	ctx    context.Context
}

func newGitHubWebHookHandler() *githubWebHookHandler {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: githubAccessToken,
		},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &githubWebHookHandler{
		client: client,
		ctx:    ctx,
	}
}

func (h *githubWebHookHandler) issuesEventHandler(event *github.IssuesEvent) error {
	switch event.GetAction() {
	case "closed":
		body := event.Issue.GetBody()
		if strings.Contains(body, "- [ ]") {
			owner := event.Repo.Owner.GetLogin()
			repo := event.Repo.GetName()
			number := event.Issue.GetNumber()
			if _, _, err := h.client.Issues.CreateComment(h.ctx, owner, repo, number, &github.IssueComment{
				Body: github.String(fmt.Sprintf("@%s This issue has non-checked checkbox in Description yet.\nPlease try to close this issue after checking all the checkboxes.", event.Sender.GetLogin())),
			}); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
			if _, _, err := h.client.Issues.Edit(h.ctx, owner, repo, number, &github.IssueRequest{
				State: github.String("open"),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *githubWebHookHandler) payloadHandler(w http.ResponseWriter, r *http.Request, params denco.Params) {
	payload, err := github.ValidatePayload(r, []byte(githubWebHookSecret))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	switch event := event.(type) {
	case *github.IssuesEvent:
		err = h.issuesEventHandler(event)
	}
	if err != nil {
		fmt.Fprintln(w, err)
	}
}

func main() {
	if githubAccessToken == "" {
		fmt.Fprintln(os.Stderr, "GITHUB_ACCESS_TOKEN is not set")
		os.Exit(1)
	}
	if githubWebHookSecret == "" {
		fmt.Fprintln(os.Stderr, "GITHUB_WEBHOOK_SECRET is not set")
		os.Exit(1)
	}
	mux := denco.NewMux()
	githubWebHookHandler := newGitHubWebHookHandler()
	handler, err := mux.Build([]denco.Handler{
		mux.POST("/payload", githubWebHookHandler.payloadHandler),
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listening on %s\n", defaultAddr)
	log.Fatal(http.ListenAndServe(defaultAddr, handler))
}
