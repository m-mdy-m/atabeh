# Contributing to Atabeh

Thanks for your interest in making Atabeh better. Whether you're fixing a typo in documentation or implementing a new protocol parser, your contributions help make internet connectivity more accessible for people who need it.

## Getting Started

Before diving into code, take some time to understand what Atabeh is trying to accomplish. Read through the README, browse existing issues, and maybe run the core locally to see how things work. This context will help you contribute more effectively.

If you're new to the project, look for issues tagged with "good first issue" or "help wanted". These are usually smaller, well-defined tasks that don't require deep knowledge of the entire codebase.

## Types of Contributions

**Code**: Implementing parsers, improving testing logic, optimizing performance, fixing bugs, adding features

**Documentation**: Writing guides, improving API docs, creating examples, translating content

**Testing**: Writing tests, reporting bugs, verifying fixes, testing edge cases

**Design**: UX/UI work for future clients, API design, architecture proposals

**Community**: Helping others in discussions, triaging issues, reviewing pull requests

All of these matter. Pick what matches your skills and interests.

## Development Workflow

### Setting Up Your Environment

1. Fork the repository to your GitHub account
2. Clone your fork locally: `git clone https://github.com/your-username/atabeh.git`
3. Add the upstream repository: `git remote add upstream https://github.com/m-mdy-m/atabeh.git`
4. Install Go 1.21 or later
5. Install project dependencies: `go mod download`
6. Run tests to verify everything works: `go test ./...`

### Making Changes

1. Create a new branch from `main`: `git checkout -b feature/your-feature-name`
2. Make your changes, keeping commits focused and logical
3. Write or update tests as needed
4. Ensure all tests pass: `go test ./...`
5. Update documentation if you're changing behavior or adding features
6. Commit your changes with clear, descriptive messages

### Commit Messages

Write commit messages that explain what changed and why, not just what you did. Good commit messages help reviewers understand your thinking and help future maintainers understand the codebase's history.

Format:
```
Short summary of change (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
Explain the problem this commit solves and why you chose
this particular solution.

Fixes #123
```

### Submitting a Pull Request

1. Push your branch to your fork: `git push origin feature/your-feature-name`
2. Open a pull request against the `main` branch
3. Fill out the pull request template completely
4. Link any related issues
5. Be responsive to feedback and questions

Pull requests go through review by maintainers and sometimes other contributors. This isn't about judging your code—it's about ensuring quality, discussing trade-offs, and maintaining consistency across the codebase. Reviews often involve back-and-forth discussion, and that's healthy.

## Code Standards

### Go Style

Follow standard Go conventions:
- Use `gofmt` to format code
- Run `go vet` to catch common issues
- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Keep functions focused and reasonably sized
- Write comments for exported functions and complex logic
- Handle errors explicitly

### Testing

Tests aren't just a checkbox—they're documentation of how code should behave and insurance against future breakage. When you write or modify code:

- Write tests that cover the happy path
- Write tests that cover edge cases and error conditions
- Make tests readable and maintainable
- Use table-driven tests where appropriate
- Don't test implementation details, test behavior

### Architecture Principles

Atabeh's architecture is modular by design:

- Keep parsers independent of each other
- Don't let the core depend on any specific client
- Use interfaces to define contracts between components
- Avoid global state where possible
- Make it easy to add new protocols without changing existing code

## Adding a New Parser

If you're implementing support for a new VPN/proxy protocol:

1. Create a new package under `pkg/parsers/`
2. Implement the `Parser` interface
3. Handle malformed inputs gracefully
4. Return normalized config objects
5. Write comprehensive tests including real-world examples
6. Document the protocol format and any quirks
7. Add integration tests for end-to-end validation

Look at existing parsers for examples. Don't reinvent patterns that already work.

## Documentation

Code is read far more often than it's written. Documentation multiplies the value of your work by making it accessible to others.

- Keep the README up to date with new features
- Document public APIs with clear explanations and examples
- Write comments that explain "why", not just "what"
- Update the CHANGELOG with user-facing changes
- If you're changing behavior, update relevant guides

## Issue Guidelines

### Reporting Bugs

Good bug reports save everyone time. When filing a bug:

- Use a clear, specific title
- Describe what you expected to happen
- Describe what actually happened
- Provide steps to reproduce
- Include your environment (OS, Go version, Atabeh version)
- Add relevant logs or error messages
- Note if you've found any workarounds

### Suggesting Features

Feature requests are welcome, but remember that every feature has a maintenance cost. When proposing features:

- Explain the problem you're trying to solve
- Describe your proposed solution
- Consider alternative approaches
- Think about edge cases and potential issues
- Be open to discussion about whether/how to implement

Not every feature request will be accepted, and that's okay. Sometimes it's because the feature doesn't align with Atabeh's goals, sometimes it's because the maintenance burden is too high, and sometimes it's just not the right time.

## Communication

### Be Patient

Contributors and maintainers are often volunteers with other commitments. Response times vary, especially for complex questions or during busy periods. If you don't get an immediate response, that's normal.

### Be Constructive

When giving feedback, focus on the work, not the person. When receiving feedback, remember that it's about making the project better, not about you being wrong.

### Ask Questions

If something isn't clear, ask. If you're stuck, ask. If you disagree with a decision, ask about the reasoning. The only bad question is the one that goes unasked and leads to confusion later.

## Pull Request Review Process

Reviews typically consider:

- Does the code solve the stated problem?
- Is the solution maintainable and consistent with the project's architecture?
- Are there tests covering the changes?
- Is documentation updated as needed?
- Are there any security implications?
- Does this change break existing functionality?

Reviews might request changes. This is normal and not a rejection. It's a conversation about how to make the contribution as good as it can be.

## Getting Help

Stuck on something? Here's how to get unstuck:

- Check existing documentation and issues
- Look at similar code in the project for patterns
- Open an issue describing what you're trying to do and where you're stuck
- Join discussions to learn from others

## Recognition

Contributors are recognized in several ways:

- Listed in the project's CONTRIBUTORS file
- Mentioned in release notes for significant contributions
- Referenced in commit messages and pull requests
- Built into the project's collective knowledge and capability

More importantly, your work helps people access the internet in challenging situations. That impact is worth more than any attribution.

## Legal

By contributing to Atabeh, you agree that your contributions will be licensed under the MIT License, the same license covering the project. You also confirm that you have the right to make these contributions and that they don't violate anyone else's rights.

## Final Thoughts

Contributing to open source can feel intimidating at first. Everyone starts somewhere, and every expert was once a beginner. Don't be afraid to try, and don't be discouraged if your first contribution isn't perfect. The community is here to help you grow, and the project benefits from diverse perspectives and skills.

The best way to contribute is to just start. Pick something small, give it a shot, and learn as you go. Welcome to Atabeh.