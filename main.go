package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

type Config struct {
	owner    string
	repo     string
	tag      string
	watch    bool
	interval int
	token    string
}

func main() {
	config := parseFlags()

	client := createGitHubClient(config.token)

	tag := config.tag
	if tag == "" {
		var err error
		tag, err = getLatestTag(client, config.owner, config.repo)
		if err != nil {
			color.Red("Error fetching latest tag: %v", err)
			os.Exit(1)
		}
		color.Yellow("Using latest tag: %s", tag)
	}

	allCompleted := checkAndDisplayStatus(client, config.owner, config.repo, tag)

	if config.watch && !allCompleted {
		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		for !allCompleted {
			time.Sleep(time.Duration(config.interval) * time.Second)
			fmt.Printf("\n")
			s.Start()
			s.Suffix = " Checking status..."
			allCompleted = checkAndDisplayStatus(client, config.owner, config.repo, tag)
			s.Stop()

			if allCompleted {
				color.Green("\nAll workflows completed!")
			}
		}
	}
}

func parseFlags() Config {
	owner := flag.String("o", "Orange-Health", "GitHub owner or organization name (default: Orange-Health)")
	repo := flag.String("r", "", "GitHub repository name (required)")
	tag := flag.String("t", "", "Git tag to check (default: latest tag)")
	watch := flag.Bool("w", false, "Enable watch mode to wait for completion")
	interval := flag.Int("i", 10, "Polling interval in watch mode (seconds)")
	flag.Parse()

	if *repo == "" {
		color.Red("Error: repository name is required. Use -r to specify it.")
		os.Exit(1)
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		color.Red("Error: GITHUB_TOKEN environment variable is not set")
		os.Exit(1)
	}

	return Config{
		owner:    *owner,
		repo:     *repo,
		tag:      *tag,
		watch:    *watch,
		interval: *interval,
		token:    token,
	}
}

func createGitHubClient(token string) *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func getLatestTag(client *github.Client, owner, repo string) (string, error) {
	ctx := context.Background()
	opts := &github.ListOptions{PerPage: 1}
	tags, _, err := client.Repositories.ListTags(ctx, owner, repo, opts)
	if err != nil {
		return "", err
	}
	if len(tags) == 0 {
		return "", fmt.Errorf("no tags found in this repository")
	}
	return *tags[0].Name, nil
}

func checkAndDisplayStatus(client *github.Client, owner, repo, tag string) bool {
	ctx := context.Background()

	opts := &github.ListWorkflowRunsOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	runs, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, owner, repo, opts)
	if err != nil {
		color.Red("Error fetching workflow runs: %v", err)
		return false
	}

	filteredRuns := []*github.WorkflowRun{}
	for _, run := range runs.WorkflowRuns {
		if (run.HeadBranch != nil && *run.HeadBranch == tag) ||
			(run.HeadSHA != nil && *run.HeadSHA == tag) {
			filteredRuns = append(filteredRuns, run)
		}
	}

	if len(filteredRuns) == 0 {
		color.Yellow("No workflow runs found for tag '%s'. They might not have started yet.", tag)
		return false
	}

	bold := color.New(color.Bold).SprintFunc()
	dim := color.New(color.Faint).SprintFunc()

	fmt.Printf("\n%s\n", bold(fmt.Sprintf("Workflow status for tag: %s", tag)))
	fmt.Println(dim(strings.Repeat("-", 60)))

	allCompleted := true

	for _, run := range filteredRuns {
		var statusColor *color.Color
		var statusSymbol string

		completed := *run.Status == "completed"
		success := completed && *run.Conclusion == "success"

		if completed {
			if success {
				statusColor = color.New(color.FgGreen)
				statusSymbol = "✓"
			} else {
				statusColor = color.New(color.FgRed)
				statusSymbol = "✗"
			}
		} else {
			statusColor = color.New(color.FgYellow)
			statusSymbol = "●"
			allCompleted = false
		}

		workflowName := *run.Name
		status := *run.Status
		conclusion := ""
		if completed {
			conclusion = fmt.Sprintf("(%s)", *run.Conclusion)
		}

		createdAt := run.CreatedAt.Time.Format("2006-01-02 15:04:05")

		fmt.Printf("%s %s - %s %s %s\n",
			statusColor.Sprint(statusSymbol),
			bold(workflowName),
			statusColor.Sprint(status),
			statusColor.Sprint(conclusion),
			dim(createdAt))
	}

	fmt.Println(dim(strings.Repeat("-", 60)))
	fmt.Printf("URL: %s\n", color.BlueString(fmt.Sprintf("https://github.com/%s/%s/actions", owner, repo)))

	return allCompleted
}