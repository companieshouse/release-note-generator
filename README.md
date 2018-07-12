# release-note-generator

## Introduction
This is a tool that generates a release note web page based from the history of a release branch

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
