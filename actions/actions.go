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

	var err error

	// Unmarshal comment event
	var commentEvent github.IssueCommentEvent
	cErr := json.Unmarshal(bodyJSON, &commentEvent)
	if cErr == nil && commentEvent.Issue != nil {
		log.Print("comment event")
		err = handleIssueComment(commentEvent)
	} else {

		// Unmarshal pull request event
		var pullRequestEvent github.PullRequestEvent
		prErr := json.Unmarshal(bodyJSON, &pullRequestEvent)
		if prErr == nil && pullRequestEvent.PullRequest != nil {
			log.Print("pull request event")
			err = handleNewPullRequest(pullRequestEvent)
		} else {

			// Unmarshal push event
			var pushEvent github.PushEvent
			pErr := json.Unmarshal(bodyJSON, &pushEvent)
			if pErr == nil {
				log.Print("push event")
				err = handleNewPush(pushEvent)
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
