package main

import (
	"context"
	"fmt"
	"os"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

func getPullRequestTotal() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("INPUT_GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)

	var query struct {
		User struct {
			PullRequests struct {
				TotalCount int
			}
		} `graphql:"user(login: $login)"`
	}

	vars := map[string]interface{}{
		"login": githubv4.String("JackPlowman"), // TODO: get username from GITHUB_TOKEN
	}

	err := client.Query(context.Background(), &query, vars)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Total pull requests by %s: %d\n", vars["login"], query.User.PullRequests.TotalCount)
}
