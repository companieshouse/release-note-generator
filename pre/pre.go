package pre

import (
	"github.com/google/go-github/github"
	"html/template"
	"net/http"
	"fmt"
	"context"
	"encoding/json"
	"os/exec"
	"os"
	"strconv"
	"golang.org/x/oauth2"
	"strings"
)

var organisation string = "companieshouse"
var repo string = getRepoName()
var toolHome string = os.Getenv("GOPATH") + "/src/github.com/release-note-generator"

type PR struct {
	Number		string
	Name        string
	URL			string
}

type NewFeature struct {
	PRs 		[]PR
}

type BugFix struct {
	PRs 		[]PR
}

type Improvement struct {
	PRs 		[]PR
}

type Information struct {
	Populated			bool
	Improvement 		Improvement
	BugFix				BugFix
	NewFeature 			NewFeature
}

var client *github.Client
var ctx context.Context

var information Information
var improvements 	Improvement
var bugFixes 		BugFix
var newFeatures 	NewFeature
var prs 	[]PR

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
	
	createClient()

	stringSlice, err := retrieveMergeCommits()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pulls, err := retrievePRs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, pull := range pulls {
		if pull.MergeCommitSHA == nil {
			continue
		}
		commitSHA, err := getMergeCommitSHA(pull)
		if err != nil {
			fmt.Printf("Error getting merge commit SHA from PR: %s", err)
			os.Exit(1)
		}

		if contains(stringSlice, commitSHA) {
			prNum, err := getPRNum(pull)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		    
			pr, err := getPR(prNum)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			

			prType, err := getPRType(pr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

    		name, err := getPRName(pr)
    		if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			

			url, err := getPRURL(pr)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			displayPR := PR {
				Number: convertPRNumToString(prNum),
				Name: name,
				URL: url}

			sortPRType(prType, displayPR)		
		}
 	}

	information = Information {
		Populated: true,
		Improvement: improvements,
		BugFix:		bugFixes,
		NewFeature:	newFeatures}

	os.Chdir(toolHome)

	t, err := template.ParseFiles("./assets/template.htm")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	t.ExecuteTemplate(w, "template.htm", information)
}

func getRepoName() string {
	getRepoNameCmd := "basename $(git remote get-url origin) .git "

	repoName, err := exec.Command("sh","-c", getRepoNameCmd).Output()
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

func sortPRType(prType string, pr PR) {
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

func getPRType(pr *github.PullRequest) (string, error) {
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

func createClient() {
	ctx = context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client = github.NewClient(tc)
}

func retrieveMergeCommits() ([]string, error) {
	// retrieves a list of all commits SHAs between the range of the first and second - merges only
	prCommand := "git rev-list HEAD ^origin/master --merges "

	prCommandOut, err := exec.Command("sh","-c", prCommand).Output()
	if err != nil {
		return nil, err
    }

	stringSlice := strings.Split(string(prCommandOut), "\n")

	return stringSlice, nil
}

func retrievePRs() ([]*github.PullRequest, error) {
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

func getMergeCommitSHA(pr *github.PullRequest) (string, error) {
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

func getPR(number int) (*github.PullRequest, error) {
	pr, _, err := client.PullRequests.Get(ctx, organisation, repo, number)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

func getPRURL(pr *github.PullRequest) (string, error) {
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

func getPRNum(pr *github.PullRequest) (int, error) {
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

func getPRName(pr *github.PullRequest) (string, error) {
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

func convertPRNumToString(prNum int) string {
    return strconv.Itoa(prNum)
}

func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}