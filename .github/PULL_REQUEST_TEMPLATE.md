<!--
  Thanks for contributing to gh-skyline! Please fill out the sections below
  to help reviewers understand and validate your change.
-->

## Description

<!-- A clear and concise description of what this PR does and why. -->

## Related issues

<!-- Link any related issues, e.g. "Closes #123" or "Refs #456". -->

Closes #

## Type of change

<!-- Check all that apply by replacing [ ] with [x]. -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would change existing behavior)
- [ ] Refactor / code cleanup (no functional change)
- [ ] Documentation update
- [ ] CI / build / tooling change
- [ ] Dependency update

## Areas affected

<!-- Check all that apply. -->

- [ ] CLI commands / flags (`cmd/`)
- [ ] GitHub API client (`internal/github/`)
- [ ] STL generation / geometry (`internal/stl/`)
- [ ] ASCII art rendering (`internal/ascii/`)
- [ ] Logger / errors / utils (`internal/logger`, `internal/errors`, `internal/utils`)
- [ ] Types (`internal/types/`)
- [ ] Tests / fixtures / mocks (`internal/testutil/`)
- [ ] Documentation (README, CONTRIBUTING, etc.)
- [ ] CI workflows / linters (`.github/`)

## How has this been tested?

<!--
  Describe the testing you performed. At minimum, run:

    go test ./...
    golangci-lint run

  If your change affects STL output, please verify the file is valid
  (e.g. opens correctly in a 3D viewer / slicer) and remains float32
  to comply with the binary STL format.
-->

- [ ] `go test ./...` passes locally
- [ ] `golangci-lint run` passes locally
- [ ] Manually ran the extension locally (e.g. `go run main.go --user <user> --year <year>` or `./rebuild.sh && gh skyline ...`)
- [ ] Verified the generated STL opens in a 3D viewer / slicer (if applicable)
- [ ] Added or updated unit tests for new / changed behavior

### Commands run

```bash
# Paste the commands you ran and a brief summary of the output
go test ./...
```

## Screenshots / output

<!--
  If your change affects user-facing output, attach screenshots of:
    - the ASCII art preview, and/or
    - the rendered STL (3D viewer screenshot), and/or
    - relevant terminal output.
-->

## Checklist

- [ ] My code follows the style of this project (see [`.github/linters/.golangci.yml`](.github/linters/.golangci.yml))
- [ ] I have added GoDoc comments for any new exported packages, functions, types, or constants
- [ ] Functions are focused and reasonably sized (generally under 50 lines)
- [ ] Final STL output uses `float32` to adhere to the STL format (if applicable)
- [ ] I have updated the README / docs where relevant (new flags, behavior changes, etc.)
- [ ] I have read the [Contributing guide](../CONTRIBUTING.md) and [Code of Conduct](../CODE_OF_CONDUCT.md)

## Additional notes for reviewers

<!-- Anything else reviewers should know: trade-offs, follow-ups, open questions, etc. -->
