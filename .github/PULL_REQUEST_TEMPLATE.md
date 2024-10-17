# Pull Request Template

## Description

Please provide a summary of the changes you made. Include context and why the change was necessary. You can reference
issues related to the pull request (e.g., "Closes #123").

**Example:**
> Added Docker integration to support SSH-based Docker commands on a remote server. This allows developers to manage
> containerized environments directly via the CLI, making deployments and testing more convenient.

## Type of Change

Please delete options that are not relevant:

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?

Describe the tests that you ran to verify your changes. Provide instructions so that reviewers can reproduce the test
results.

**Example:**
> - Unit Tests: Ran `go test ./...` and confirmed all tests passed.
> - Manual Test: Connected to a remote Docker host and verified that KasmLink could successfully execute SSH commands
    for container management.

## Checklist

Please ensure you have completed the following before submitting the pull request:

- [ ] I have performed a self-review of my code.
- [ ] I have added or updated relevant documentation (if applicable).
- [ ] I have added tests that prove my fix is effective or that my feature works.
- [ ] I have ensured that my changes do not introduce any new vulnerabilities.
- [ ] Any dependent changes have been merged and published in downstream modules.

## Screenshots (if applicable)

Include screenshots here if your changes involve any UI elements.

## Related Issue(s)

Link any related issues here:

- Resolves #123
- Related to #456

## Dependencies

Mention any PRs or other dependencies required for this change to work:

- Requires PR #789 to be merged first.

## Additional Notes

Provide any additional context regarding your changes that may be important for the reviewer.

**Example:**
> This PR includes several refactoring efforts that touch multiple files. I've verified all existing tests, but I'd
> appreciate it if reviewers could focus on changes involving Docker integration.
