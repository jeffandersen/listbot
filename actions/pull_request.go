package actions

import (
	"errors"
	"log"

	"github.com/google/go-github/github"
)

func handleNewPullRequest(event github.PullRequestEvent) error {
	if *event.Action != "opened" {
		return nil
	}

	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	filename := TemplateLocation
	file, _, _, err := githubClient.Repositories.GetContents(owner, repo, filename, nil)
	if err != nil {
		log.Print("could not retrieve checklist template")
		return err
	}
	if file == nil {
		err = errors.New("template file not found")
		log.Print(err.Error())
		return err
	}

	contents, err := file.GetContent()
	if err != nil {
		log.Print("failed to read checklist template")
		return err
	}

	contents = CommentTag + "\n" + contents
	number := *event.PullRequest.Number
	body := github.IssueComment{
		Body: &contents,
	}

	comment, _, err := githubClient.Issues.CreateComment(owner, repo, number, &body)
	if err != nil {
		log.Print("failed to post checklist template")
		return err
	}
	if comment.HTMLURL == nil {
		log.Print("could not set status: missing comment url")
		return errors.New("unknown comment URL")
	}

	err = setPullStatus(owner, repo, number, "failure", *comment.HTMLURL)
	if err != nil {
		log.Print("failed to set pull request status")
		return err
	}

	return nil
}
