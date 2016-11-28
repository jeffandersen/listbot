package actions

import (
	"log"
	"strings"

	"github.com/google/go-github/github"
)

func handleNewPush(event github.PushEvent) error {
	if event.After == nil {
		return nil
	}

	fullname := *event.Repo.FullName
	owner := strings.Split(fullname, "/")[0]
	repo := strings.Split(fullname, "/")[1]

	pulls, err := getAllPulls(owner, repo)
	if err != nil {
		log.Print("failed to retrieve pull requests")
		return err
	}

	var target *github.PullRequest
	for _, p := range pulls {
		if strings.Contains(*p.StatusesURL, *event.After) {
			target = p
		}
	}

	if target == nil {
		log.Printf("pull request not found for sha %s", *event.After)
		return nil
	}

	err = refreshPullStatus(owner, repo, *target.Number)
	if err != nil {
		log.Print("failed to set pull request status")
		return err
	}

	return nil
}

func getAllPulls(owner, repo string) ([]*github.PullRequest, error) {
	opt := github.PullRequestListOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var pulls []*github.PullRequest
	for {
		results, resp, err := githubClient.PullRequests.List(owner, repo, &opt)
		if err != nil {
			return nil, err
		}
		pulls = append(pulls, results...)
		if resp.NextPage == 0 {
			break
		}
		opt.ListOptions.Page = resp.NextPage
	}

	return pulls, nil
}
