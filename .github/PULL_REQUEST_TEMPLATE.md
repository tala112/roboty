## Summary

<!-- Describe the change and why it's needed -->

## Type of Change

- [ ] Bug fix (non-breaking)
- [ ] New feature (non-breaking)
- [ ] Breaking change
- [ ] Refactor
- [ ] Documentation
- [ ] CI / Infra

## Safety Verification

<!-- Check ALL that apply -->

### Pre-Kill Safety
- [ ] Whitelist (`internal/modes/whitelist.json`) updated for any new process protections
- [ ] `systemCritical` (`internal/modes/safety.go`) updated for any new process protections
- [ ] `Whitelist sync test` (`TestCritical_WhitelistSyncWithSystemCritical`) passes
- [ ] Whitelist entries use normalized names (no `.exe` suffix, lowercase)

### Process & System Safety
- [ ] No new code can kill system processes (`explorer`, `dwm`, `svchost`, etc.)
- [ ] Kill-loop detection is engaged for any new blocking paths
- [ ] Ancestor process protection covers the new code path
- [ ] Self-protection (`roboty`, `wails`) verified

### Proxy Safety
- [ ] IPv6 localhost (`[::1]:port`) regression tested
- [ ] No proxy loop via environment variables (transport.Proxy = nil)
- [ ] CONNECT tunnel timeout (60s) is applied

### Testing
- [ ] All `TestCritical_*` tests pass
- [ ] All `TestChaos_*` tests pass
- [ ] Fuzz tests pass (NormalizeKillExec, IsAlwaysAllowed, IsAllowed)
- [ ] `go vet ./internal/modes/` clean
- [ ] `go test -race` passes (requires CGO_ENABLED=1 in CI)

### Pre-release (if applicable)
- [ ] VM test matrix executed: [ ] Win11 24H2 [ ] Win10 22H2 [ ] Ubuntu 24.04 Wayland [ ] macOS Sequoia
- [ ] `whitelist.json` and `safety.go systemCritical` are in sync (automated check)
- [ ] Fuzz tests run for minimum 30s per target
- [ ] Kill-loop detector verified with rapid restart simulation

## Related Issues

<!-- Link any related issues -->
