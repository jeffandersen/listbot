package actions

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	CommentEvent        = "comment"
	NewPullRequestEvent = "new_pull_request"
	CommentTag          = "<!-- listbot comment -->"
	StatusContext       = "listbot"
	StatusSuccessText   = "Checklist complete"
	StatusFailureText   = "Checklist incomplete"
	StatusStateFailure  = "failure"
	StatusStateSuccess  = "success"
	TemplateLocation    = ".github/listbot.md"
)

var HasCheckbox = regexp.MustCompile(`\-\s\[\s\]`)

var githubToken string
var githubLogin string
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
	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	bodyJSON := buf.Bytes()

	var event string

	var commentEvent github.IssueCommentEvent
	var pullRequestEvent github.PullRequestEvent

	cErr := json.Unmarshal(bodyJSON, &commentEvent)
	if cErr == nil && commentEvent.Issue != nil {
		event = CommentEvent
	}
	if event == "" {
		pErr := json.Unmarshal(bodyJSON, &pullRequestEvent)
		if pErr == nil && pullRequestEvent.PullRequest != nil {
			event = NewPullRequestEvent
		}
	}

	var err error
	switch event {
	case CommentEvent:
		err = handleIssueComment(commentEvent)
	case NewPullRequestEvent:
		err = handleNewPullRequest(pullRequestEvent)
	}

	if err != nil {
		log.Printf("error: %s", err.Error())
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.String(http.StatusOK, "")
}
