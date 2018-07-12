# release-note-generator

## Introduction
This is a tool that generates a release note web page based from the history of a release branch

## Prerequisites
To use this tool you will need to:
- Have `Go` installed
- Have a GitHub Personal Access Token set in your bash profile

## Getting started
To build, you must first git clone it into your `$GOPATH` under `src/github.com/companieshouse` and navigate to the tools directory. Run `go build .` from the head directory of the tool which will build the binary in the current directory and then `go install .` which will build and install the binary into the`$GOPATH/bin` directory.

Running the tool is easy and only requires you to run it in the directory of the repository in which you want to generate the release note for. Additionally, you will need to be checked out under the release branch the release note is being generated for.

Example use: navigate to the directory of the repository you wish you generate the release note for, and run `$GOPATH/src/github.com/companieshouse/release-note-tool/release-note-tool` or if you have the `$GOPATH/bin` directory added to your path variable you should be able to run `release-note-tool`.
