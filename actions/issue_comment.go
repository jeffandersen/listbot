package actions

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/github"
)

func handleIssueComment(ctx context.Context, event github.IssueCommentEvent) error {
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

	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	comment, _, err := githubClient.Issues.GetComment(ctx, owner, repo, *event.Comment.ID)
	if err != nil {
		return err
	}

	switch *event.Action {
	case "edited":
		err = handleIssueCommentEdited(ctx, event, *comment)
	case "deleted":
		err = handleIssueCommentDeleted(ctx, event)
	}

	return err
}

func handleIssueCommentEdited(ctx context.Context, event github.IssueCommentEvent, comment github.IssueComment) error {
	var state string

	// naively determine if items are left un-checked, setting the status
	if HasCheckbox.MatchString(strings.TrimSpace(*comment.Body)) {
		state = "failure"
	} else {
		state = "success"
	}

	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	return setPullStatus(ctx, owner, repo, number, state, *event.Comment.HTMLURL)
}

func handleIssueCommentDeleted(ctx context.Context, event github.IssueCommentEvent) error {
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	return refreshPullStatus(ctx, owner, repo, number)
}

func refreshPullStatus(ctx context.Context, owner, repo string, number int) error {
	comments, err := getAllPullComments(ctx, owner, repo, number)
	if err != nil {
		log.Printf("failed to retrieve pull request comments: %s", err.Error())
		return err
	}

	count := 0
	var target *github.IssueComment
	for _, c := range comments {
		if isListbotListComment(c) {
			target = c
			count++
		}
	}

	// If no listbot comments are found, set status to successful
	if count < 1 {
		log.Print("status updated: no list")
		return setPullStatus(ctx, owner, repo, number, "success", "")
	}

	// naively determine if items are left un-checked, setting the status
	var state string
	if HasCheckbox.MatchString(*target.Body) {
		state = "failure"
	} else {
		state = "success"
	}

	return setPullStatus(ctx, owner, repo, number, state, *target.HTMLURL)
}

func setPullStatus(ctx context.Context, owner, repo string, number int, state, url string) error {
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
	pr, _, err := githubClient.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		log.Printf("failed to retrieve pull request: %s", err.Error())
		return err
	}
	sha := *pr.Head.SHA

	log.Printf("set state for %s as %+v", sha, status)

	// set the status for the pull request at the given sha
	_, _, err = githubClient.Repositories.CreateStatus(ctx, owner, repo, sha, &status)
	if err != nil {
		log.Printf("failed to update status for ref: %s", err.Error())
		return err
	}

	return nil
}

func isListbotListComment(c *github.IssueComment) bool {
	return strings.Contains(*c.Body, CommentTag)
}

func getAllPullComments(ctx context.Context, owner, repo string, number int) ([]*github.IssueComment, error) {
	opt := github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var comments []*github.IssueComment
	for {
		log.Printf("%s/%s #%d -- %+v\n", owner, repo, number, opt)
		results, resp, err := githubClient.Issues.ListComments(ctx, owner, repo, number, &opt)
		log.Printf("err: %+v\n", err)
		log.Printf("resp: %+v\n", resp)
		log.Printf("results: %+v\n", results)
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
