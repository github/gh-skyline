# Contributing

[fork]: https://github.com/github/gh-skyline/fork
[pr]: https://github.com/github/gh-skyline/compare
[style]: https://github.com/github/gh-skyline/blob/main/.github/linters/.golangci.yml

Hi there! We're thrilled that you'd like to contribute to this project. Your help is essential for keeping it great.

Contributions to this project are [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license) to the public under the [project's open source license](LICENSE).

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.

## Prerequisites for running and testing code

### GitHub Codespace

The repository includes a [pre-configured devcontainer](.devcontainer/devcontainer.json) that handles most prerequisites. To use it:

1. Create a fork of the repository
1. Click the green "Code" button on the repository
1. Select the "Codespaces" tab
1. Click "Create Codespace on main" (or on the branch you want to work on)

This will create a cloud-based development environment with:

- Go installation
- Required development tools
- Project dependencies
  - GitHub CLI (gh)
- Several Visual Studio Code extensions for Go development and GitHub integration
- Pre-configured linting and testing tools

The environment will be ready to use in a few minutes.

### Local development environment

These are one-time installations required to be able to test your changes locally as part of the pull request (PR) submission process.

1. install Go [through download](https://go.dev/doc/install) | [through Homebrew](https://formulae.brew.sh/formula/go)
1. [install golangci-lint](https://golangci-lint.run/welcome/install/#local-installation)

### Building the extension

You can build from source using the following command:

```bash
go build -o gh-skyline
```

However, you'll want to test your changes in the GitHub CLI before you raise a Pull Request. To make that easier in local development, you could consider using the provided [rebuild script](rebuild.sh):

```bash
./rebuild.sh
```

This script will:

1. Remove any existing installation of the `gh-skyline` extension
2. Build the extension from source
3. Install the local version for testing

### Testing

Run the full test suite with:

```bash
go test ./...
```

## Submitting a pull request

1. [Fork][fork] and clone the repository
1. Configure and install the dependencies: `script/bootstrap`
1. Make sure the tests pass on your machine: `go test -v ./...`
1. Make sure linter passes on your machine: `golangci-lint run`
1. Create a new branch: `git checkout -b my-branch-name`
1. Make your change, add tests, and make sure the tests and linter still pass
1. Push to your fork and [submit a pull request][pr]
1. Pat yourself on the back and wait for your pull request to be reviewed and merged.

Here are a few things you can do that will increase the likelihood of your pull request being accepted:

- Follow the [style guide][style].
- Write tests.
- Keep your change as focused as possible. If there are multiple changes you would like to make that are not dependent upon each other, consider submitting them as separate pull requests.
- Write a [good commit message](http://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html).

## Resources

- [How to Contribute to Open Source](https://opensource.guide/how-to-contribute/)
- [Using Pull Requests](https://help.github.com/articles/about-pull-requests/)
- [GitHub Help](https://help.github.com)
