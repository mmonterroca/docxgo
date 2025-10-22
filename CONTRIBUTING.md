# Contributing to go-docx (SlideLang Fork)

Thank you for your interest in contributing to the SlideLang fork of go-docx! This document provides guidelines and workflow information for contributors.

## Git Flow Workflow

We use a simplified Git Flow branching strategy to maintain code quality and stability:

### Branch Structure

- **`master`**: Production-ready code only. This branch contains stable releases and is tagged with semantic versions (e.g., `v0.1.0-slidelang`, `v0.2.0-slidelang`).
- **`dev`**: Integration branch where features are tested before release. All development work merges here first.
- **Feature branches**: Short-lived branches for specific features, bug fixes, or improvements. Named with prefixes like `feature/`, `fix/`, `docs/`, etc.

### Contributing Process

Follow these steps to contribute:

#### 1. Fork and Clone

Fork the repository to your GitHub account, then clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/go-docx.git
cd go-docx
```

#### 2. Add Upstream Remote

Add the original repository as upstream (if not already added):

```bash
git remote add upstream https://github.com/SlideLang/go-docx.git
git remote -v  # Verify remotes
```

#### 3. Create Feature Branch

Always branch from `dev`:

```bash
git checkout dev
git pull upstream dev  # Get latest changes
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/feature-name` - New features
- `fix/bug-description` - Bug fixes
- `docs/what-changed` - Documentation changes
- `test/what-tested` - Test additions
- `refactor/what-refactored` - Code refactoring
- `perf/what-improved` - Performance improvements

#### 4. Make Changes

Write your code following the project's style:
- Run `go fmt ./...` before committing
- Add tests for new features
- Update documentation as needed
- Ensure existing tests pass: `go test ./...`

#### 5. Commit Changes

Use descriptive commit messages following [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git add .
git commit -m "feat: add support for STYLEREF field"
```

Commit message format:
```
<type>: <short description>

[optional body with more details]
[optional footer with breaking changes or issue references]
```

Types:
- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `test:` Test additions or modifications
- `refactor:` Code refactoring without feature changes
- `perf:` Performance improvements
- `chore:` Maintenance tasks (dependencies, build, etc.)

Examples:
```bash
git commit -m "feat: add HYPERLINK field support"
git commit -m "fix: prevent empty RunProperties XML elements"
git commit -m "docs: update TOC examples in README"
git commit -m "test: add coverage for bookmark generation"
```

#### 6. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

#### 7. Open Pull Request

1. Go to the [original repository](https://github.com/SlideLang/go-docx)
2. Click "New Pull Request"
3. **Important**: Set base branch to `dev` (NOT `master`)
4. Set compare branch to your feature branch
5. Fill in the PR template:
   - Clear description of changes
   - Reference related issues
   - List any breaking changes
   - Add screenshots/examples if applicable

#### 8. Code Review

- Wait for maintainer review
- Address feedback by pushing additional commits
- Engage in discussion if needed
- Once approved, maintainers will merge to `dev`

#### 9. Release Process

Periodically, maintainers will:
1. Merge `dev` → `master`
2. Tag the release with semantic version
3. Create GitHub release with changelog

## What We're Looking For

We welcome contributions in these areas:

### High Priority
- ✅ **Bug fixes**: Resolve issues or Word compatibility problems
- ✅ **Additional field codes**: STYLEREF, HYPERLINK, IF, DATE, etc.
- ✅ **Extended style support**: More heading levels, custom styles
- ✅ **Test coverage**: Improve code reliability

### Medium Priority
- ✅ **Performance improvements**: Optimize parsing or generation
- ✅ **Documentation**: Better examples, API docs, tutorials
- ✅ **Headers/Footers API**: Enhanced section properties

### Future Enhancements
- ✅ **Advanced formatting**: SmartArt, equations, charts
- ✅ **Section management**: Multiple sections, page layout
- ✅ **Cross-references**: Figure numbers, table references

## Development Guidelines

### Code Quality

- **Go Idioms**: Follow Go best practices and idioms
- **Error Handling**: Always handle errors properly
- **Naming**: Use clear, descriptive names
- **Comments**: Document exported functions and complex logic
- **Tests**: Aim for >60% coverage

### Testing

Run tests before submitting:

```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -cover ./...             # With coverage
go test -race ./...              # Race detection
```

### Documentation

Update documentation when adding features:
- API Reference in README.md
- Inline code comments
- Example usage in demo files
- Update CHANGELOG if significant

## Community

- **Issues**: Report bugs or request features via [GitHub Issues](https://github.com/SlideLang/go-docx/issues)
- **Discussions**: Ask questions or share ideas in [GitHub Discussions](https://github.com/SlideLang/go-docx/discussions)
- **Code of Conduct**: Be respectful and constructive

## Questions?

If you have questions about contributing:
1. Check existing issues and PRs
2. Read the documentation in README.md
3. Open a discussion or issue

Thank you for contributing to go-docx! 🎉
