# Contributing

Thank you for your interest in this project.
If you wish to submit a bug report, propose a feature or contribute a change,
please follow the guidelines described in this document.

## Bug reports and proposals

We use the [Github issue tracker](https://github.com/synthesio/zconfig/issues) for bugs
and feature proposals. Please check both the open and closed lists before posting
a new issue.
Please be as specific as you can: issues should have descriptive titles and relevant labels.

Bug reports must contain:

- the expected behaviour
- why this behaviour is expected
- the actual behaviour
- detailed steps allowing to reproduce the problem
- the used package version

## Making changes

Before you start working on an open issue, add a comment telling that you're going to tackle it.

If it's your first contribution, start by forking the repository.

Do not forget to:

- add new test cases whenever they are needed
- run unit tests before asking for review
- add an entry in the [CHANGELOG](https://github.com/synthesio/zconfig/blob/master/CHANGELOG.md) file

Commit messages should follow these [guidelines](https://chris.beams.io/posts/git-commit/):

1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how

When your changes are ready for review, open a [pull request](https://github.com/synthesio/zconfig/pulls) and reference the corresponding issue.

## Coding conventions

Go source files must be formatted using the [`goimports`](https://godoc.org/golang.org/x/tools/cmd/goimports) tool.
