---
name: Release Checklist
about: Create a ticket to track a release
title: ''
labels: release
assignees: ''

---

## Release

The release version and a description of the planned changes to be included in the release.

## Issues

Link the major issues planned to be included in the release.

## Documentation

Link the relevant documentation PRs for this release.

<<<<<<< HEAD
## Checklist

- [ ] Update version in plugin/evm/version.go
- [ ] Bump Lux Node dependency in go.mod for RPCChainVM Compatibility
- [ ] Update Lux Node dependency in scripts/versions.sh for e2e tests.
- [ ] Add new entry in compatibility.json for RPCChainVM Compatibility
- [ ] Update Lux Node compatibility in README
- [ ] Deploy to TestNet
=======
**Checklist**
- [ ] Update version in scripts/versions.sh and plugin/evm/version.go
- [ ] Bump Luxd dependency for RPCChainVM Compatibility
- [ ] Add new entry in compatibility.json for RPCChainVM Compatibility
- [ ] Update Luxd compatibility in README
- [ ] Bump cmd/simulator go mod (if needed)
>>>>>>> b36c20f (Update executable to luxd)
- [ ] Confirm goreleaser job has successfully generated binaries by checking the releases page
