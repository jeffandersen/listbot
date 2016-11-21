package actions

import (
	"log"
	"strings"

	"github.com/google/go-github/github"
)

func handleIssueComment(event github.IssueCommentEvent) error {
	// if the issue or comment are nil, ignore it
	if event.Issue == nil || event.Comment == nil {
		return nil
	}
	// only pay attention to pull request comments
	if event.Issue.PullRequestLinks == nil {
		return nil
	}
	// double check that the comment is tagged with a list
	if !strings.Contains(*event.Comment.Body, CommentTag) {
		return nil
	}

	var err error
	switch *event.Action {
	case "edited":
		err = handleIssueCommentEdited(event)
	case "deleted":
		err = handleIssueCommentDeleted(event)
	}

	return err
}

func handleIssueCommentEdited(event github.IssueCommentEvent) error {
	var state string

	// naively determine if items are left un-checked, setting the status
	if HasCheckbox.MatchString(*event.Comment.Body) {
		state = "failure"
	} else {
		state = "success"
	}

	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	return setPullStatus(owner, repo, number, state, *event.Comment.HTMLURL)
}

func handleIssueCommentDeleted(event github.IssueCommentEvent) error {
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number

	comments, err := getAllPullComments(owner, repo, number)
	if err != nil {
		log.Print("failed to retrieve pull request comments: %s", err.Error())
		return err
	}

	count := 0
	for _, c := range comments {
		if isListbotListComment(c) {
			count++
		}
	}

	// If no listbot comments are found, set status to successful
	if count < 1 {
		return setPullStatus(owner, repo, number, "success", *event.Issue.HTMLURL)
	}

	return nil
}

func setPullStatus(owner, repo string, number int, state, url string) error {
	status := github.RepoStatus{
		TargetURL: &url, //event.Comment.HTMLURL, // link the checklist as details
		Context:   addrStr(StatusContext),
	}

	if state == "success" {
		status.Description = addrStr(StatusSuccessText)
		status.State = addrStr(StatusStateSuccess)
	} else {
		status.Description = addrStr(StatusFailureText)
		status.State = addrStr(StatusStateFailure)
	}

	// identify the pull request sha
	pr, _, err := githubClient.PullRequests.Get(owner, repo, number)
	if err != nil {
		log.Printf("failed to retrieve pull request: %s", err.Error())
		return err
	}
	sha := *pr.Head.SHA

	log.Print("set state for %s as %+v", sha, status)

	// set the status for the pull request at the given sha
	_, _, err = githubClient.Repositories.CreateStatus(owner, repo, sha, &status)
	if err != nil {
		log.Printf("failed to update status for ref: %s", err.Error())
		return err
	}

	return nil
}

func isListbotListComment(c *github.PullRequestComment) bool {
	return strings.Contains(*c.Body, CommentTag)
}

func getAllPullComments(owner, repo string, number int) ([]*github.PullRequestComment, error) {
	opt := github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var comments []*github.PullRequestComment
	for {
		results, resp, err := githubClient.PullRequests.ListComments(owner, repo, number, &opt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, results...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	return comments, nil
}

func addrStr(str string) *string {
	return &str
}
