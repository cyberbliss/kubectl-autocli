# An autocomplete wrapper for kubectl

## Development

## Releasing
goreleaser (https://goreleaser.com/) is used to generate new releases and push them to Github. It also generates the homebrew tap formula.
1. Make/test changes in a feature branch
2. PR and merge into master
3. Switch to master and git pull locally
4. `git tag -a vn.n.n` where n.n.n obeys semver standards
5. `git push origin vn.n.n`
6. `goreleaser --rm-dist`