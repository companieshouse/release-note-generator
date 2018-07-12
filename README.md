# release-note-generator

## Introduction
This is a tool that generates a release note web page based from the history of a release branch using Golang, Git and GitHub API

## Prerequisites
To use this tool you will need to:
- Have `Go` installed
- Have a GitHub Personal Access Token set in your bash profile

## Getting started
To build and install you must:
- Clone it into your `$GOPATH` under `src/github.com/companieshouse`
- Navigate to the tools directory. Run `go get ./...` to resolve and download any dependencies
- Run `go build .` from the head directory of the tool which will build the binary in the current directory
- Run `go install .` which will build and install the binary into the`$GOPATH/bin` directory.

NOTE: There is no reason to do a `go build .` as the install itself shall build and install the binary into the `$GOPATH/bin` directory which is easier to access. 

## Using the tool
Running the tool is easy and only requires you to run it in the directory of the repository in which you want to generate the release note for. Additionally, you will need to be checked out under the release branch the release note is being generated for.

Example use: navigate to the directory of the repository you wish to generate the release note for, and run `release-note-generator` as long as  you have the `$GOPATH/bin` directory added to your path variable you should be able to access the runnable binary. If not either add it, or do the `go build .` step above and just call the runnable binary that is built in the tools directory. It is recommended you choose the path option due to easier access.

## In-depth
As described in the introduction, this tool generates a release note in the form of a webpage for the chosen repository. The way it does this is by scanning the merge commits of the checked out release branch and compares them with what is in `master`. It determines that what ever merge commits exist inside of the release branch and not the `master` branch - these are the changes(PRs) that are being released. It couples these together in a list of pull requests and then sorts them into their correct categories (new feature, improvement, bug fix) whilst outputting the data onto a webpage which is easy to read.

### Walkthrough
The standard process of the tools is as follows:
1) Establishes the web-server to listen and serve to `localhost:8080`
2) Runs the `ServeTemplate` method that kicks off the `pre` functionality that populates the template that the server serves to the user (the `pre` functionality is the creation of a release note during its pre-release stage - this details changes that are about to go into the next release. There can be a post if required but that will generate backdated release notes, but that come as an improvement in future - if needed)
4) GitHub client is created to access the GitHub APIs
5) Retrieve all merge commits SHA that are on the current checked out branch (will be the release branch) that aren't in `master`.
6) Retrieves all PRs in the repository using GitHub API
7) Iterates over all retrieved PRs from repository and gets their merge commit SHA (as some PRs may not have been merged yet)
8) Compares the merge commit SHAs with the list of merge commit SHAs retrieved from release branch and only does the following functionality on those that exist in both of those lists.
(the reason we are comparing two lists of SHAs is because in the one list, we have the merges that are in the release branch, the second list, we are retrieving all PRs from the repository (of all states: open, closed etc) and then cross referencing those that exist in the first list. It has been done this way because there is no way of searching for a PR by its commit SHA through the GitHub APIs, so instead we have to interate through all PRs and check each individual commit SHA to determine if its the one we want)
9) Retrieves the PR number from the iterated PR object
10) Uses the retrieve PR number to get the full PR object from GitHub using their APIs
11) Using the retrieve PR object, we get the PR type by checking that its body contains a checkbox with its type in (if it doesn't have a type, then it doesn't get selected and gets ignored)
12) Get the PR name from the passed PR object
13) Get the PR URL from the passed PR object
14) Creates a display PR for the template to use
15) Sorts the PR type into the correct category so it ends up in the correct place on the web page
16) All information including the categories and their PRs are added to an information object
17) The information object is then served to the template as data which is then rendered by the template system to the webpage
