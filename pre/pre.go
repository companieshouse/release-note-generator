package pre

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var organisation string = "companieshouse"
var repo string = GetRepoName()
var toolHome string = os.Getenv("GOPATH") + "/src/github.com/release-note-generator"

type PR struct {
	Number string
	Name   string
	URL    string
}

type NewFeature struct {
	PRs []PR
}

type BugFix struct {
	PRs []PR
}

type Improvement struct {
	PRs []PR
}

type Information struct {
	Populated   bool
	Improvement Improvement
	BugFix      BugFix
	NewFeature  NewFeature
}

var client *github.Client
var ctx context.Context

var information Information
var improvements Improvement
var bugFixes BugFix
var newFeatures NewFeature
var prs []PR

// ServeTemplatePre populates the pre release data to the template for the main
// server to use and server
func ServeTemplatePre(w http.ResponseWriter, r *http.Request) {
	if information.Populated == true {
		t, err := template.ParseFiles("./assets/template.htm")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		t.ExecuteTemplate(w, "template.htm", information)
		return
	}

	CreateClient()

	stringSlice, err := RetrieveMergeCommits()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pulls, err := RetrievePRs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, pull := range pulls {
		if pull.MergeCommitSHA == nil {
			continue
		}
		commitSHA, err := GetMergeCommitSHA(pull)
		if err != nil {
			fmt.Printf("Error getting merge commit SHA from PR: %s", err)
			os.Exit(1)
		}

		if Contains(stringSlice, commitSHA) {
			prNum, err := GetPRNum(pull)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			pr, err := GetPR(prNum)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			prType, err := GetPRType(pr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			name, err := GetPRName(pr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			url, err := GetPRURL(pr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			displayPR := PR{
				Number: ConvertPRNumToString(prNum),
				Name:   name,
				URL:    url}

			SortPRType(prType, displayPR)
		}
	}

	information = Information{
		Populated:   true,
		Improvement: improvements,
		BugFix:      bugFixes,
		NewFeature:  newFeatures}

	os.Chdir(toolHome)

	t, err := template.ParseFiles("./assets/template.htm")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t.ExecuteTemplate(w, "template.htm", information)
}

// GetRepoName retrieves the repo of the current folder by using git
// this saves passing the repo as a param to tool or hardcoding repo names
func GetRepoName() string {
	getRepoNameCmd := "basename $(git remote get-url origin) .git "

	repoName, err := exec.Command("sh", "-c", getRepoNameCmd).Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	name := strings.TrimSuffix(string(repoName), "\n")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return name
}

// SortPRType sorts the pull requets into the correct categories
// depending on whether they are bugs, new features, or improvements
func SortPRType(prType string, pr PR) {
	if prType == "bug" {
		bugFixes.PRs = append(bugFixes.PRs, pr)
	}
	if prType == "new-feature" {
		newFeatures.PRs = append(newFeatures.PRs, pr)
	}
	if prType == "improvement" {
		improvements.PRs = append(improvements.PRs, pr)
	}
}

// GetPRType retrieves the PR type(bug fix, improvement, new feature)
// from the body using a simple contains function. The PRs are
// then sorted later on.
func GetPRType(pr *github.PullRequest) (string, error) {
	prBodyMarshalled, err := json.Marshal(pr.Body)
	if err != nil {
		return "", err
	}

	body, err := strconv.Unquote(string(prBodyMarshalled))
	if err != nil {
		return "", err
	}

	if strings.Contains(body, "* [x] Bug fix") || strings.Contains(body, "* [X] Bug fix") {
		return "bug", nil
	}
	if strings.Contains(body, "* [x] New feature") || strings.Contains(body, "* [X] New feature") {
		return "new-feature", nil
	}
	if strings.Contains(body, "* [x] Improvement") || strings.Contains(body, "* [X] Improvement") {
		return "improvement", nil
	}

	return "", nil
}

// CreateClient creates the GitHub client for the accessing of the
// GitHub APIs
func CreateClient() {
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)
}

// RetrieveMergeCommits gets all of the merge SHAs and stores them in a slice
// which is then returned for further processing
func RetrieveMergeCommits() ([]string, error) {
	// retrieves a list of all commits SHAs between the range of the first and second - merges only
	prCommand := "git rev-list HEAD ^origin/master --merges "

	prCommandOut, err := exec.Command("sh", "-c", prCommand).Output()
	if err != nil {
		return nil, err
	}

	stringSlice := strings.Split(string(prCommandOut), "\n")

	return stringSlice, nil
}

// RetrievePRs does a full retrieve of all PRs from within the repostory
// and returns them as an array
func RetrievePRs() ([]*github.PullRequest, error) {
	options := &github.PullRequestListOptions{
		State: "all",
	}

	// retrieves all PRs within the specified repo with the specified options
	pulls, _, err := client.PullRequests.List(ctx, organisation, repo, options)
	if err != nil {
		return nil, err
	}

	return pulls, nil
}

// GetMergeCommitSHA retrieves the merge commit SHA from the passed
// in pull request
func GetMergeCommitSHA(pr *github.PullRequest) (string, error) {
	commitSHAMarshalled, err := json.Marshal(pr.MergeCommitSHA)
	if err != nil {
		return "", err
	}

	commitSHA, err := strconv.Unquote(string(commitSHAMarshalled))
	if err != nil {
		fmt.Print(pr)
		return "", err
	}

	return commitSHA, nil
}

// GetPR gets the specific PR object from the repository using GitHub API
// and returns the PR for further processing
func GetPR(number int) (*github.PullRequest, error) {
	pr, _, err := client.PullRequests.Get(ctx, organisation, repo, number)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// GetPRURL returns the URL of the PR for the linking of the PR
// within the web-page
func GetPRURL(pr *github.PullRequest) (string, error) {
	prUrlMarshalled, err := json.Marshal(pr.HTMLURL)
	if err != nil {
		return "", err
	}

	url, err := strconv.Unquote(string(prUrlMarshalled))
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetPRNum retrieves the pull request number from the PR object returned
// from GitHub API and returns it as an int
func GetPRNum(pr *github.PullRequest) (int, error) {
	prNum, err := json.Marshal(pr.Number)
	if err != nil {
		return 1, err
	}

	num, err := strconv.Atoi(fmt.Sprintf("%s", prNum))
	if err != nil {
		return 1, err
	}
	return num, nil
}

// GetPRName retrieves the PR title from the GitHub API returned PR object
// and returns it as a string
func GetPRName(pr *github.PullRequest) (string, error) {
	prNameMarshalled, err := json.Marshal(pr.Title)
	if err != nil {
		return "", err
	}

	name, err := strconv.Unquote(string(prNameMarshalled))
	if err != nil {
		return "", err
	}

	return name, nil
}

// ConvertPRNumToString converts the PR number to a string
// for easier handling
func ConvertPRNumToString(prNum int) string {
	return strconv.Itoa(prNum)
}

// Contains accepts an array of strings and checks if the passed
// string is contained with that array (this is for the merge commits)
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
