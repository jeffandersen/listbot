package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Available constants
const (
	CommentTag       = "<!-- listbot comment -->"
	TemplateLocation = ".github/listbot.md"

	CommentEvent        = "comment"
	NewPullRequestEvent = "new_pull_request"
	StatusContext       = "listbot"
	StatusStateFailure  = "failure"
	StatusStateSuccess  = "success"
	StatusSuccessText   = "Checklist complete"
	StatusFailureText   = "Checklist incomplete"
)

// HasCheckbox is a naive regex to determine if a blob contains a md checkbox
var HasCheckbox = regexp.MustCompile(`\-\s\[\s\]`)

var githubToken string
var githubClient *github.Client

func init() {
	// githubToken is a personal access token issued for listbot
	githubToken = os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("missing GITHUB_TOKEN")
		return
	}

	// initialize authenticated github client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	githubClient = github.NewClient(tc)
}

// HandleWebhook identities the event type and responds to the request
func HandleWebhook(c *gin.Context) {
	ctx := context.Background()

	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	bodyJSON := buf.Bytes()

	var err error

	// Unmarshal comment event
	var commentEvent github.IssueCommentEvent
	cErr := json.Unmarshal(bodyJSON, &commentEvent)
	if cErr == nil && commentEvent.Issue != nil {
		log.Print("comment event")
		err = handleIssueComment(ctx, commentEvent)
	} else {

		// Unmarshal pull request event
		var pullRequestEvent github.PullRequestEvent
		prErr := json.Unmarshal(bodyJSON, &pullRequestEvent)
		if prErr == nil && pullRequestEvent.PullRequest != nil {
			log.Print("pull request event")
			err = handleNewPullRequest(ctx, pullRequestEvent)
		} else {

			// Unmarshal push event
			var pushEvent github.PushEvent
			pErr := json.Unmarshal(bodyJSON, &pushEvent)
			if pErr == nil {
				log.Print("push event")
				err = handleNewPush(ctx, pushEvent)
			}
		}
	}
	if err != nil {
		log.Printf("error: %s", err.Error())
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusOK, "")
}
