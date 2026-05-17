.PHONY: test test-unit test-safety test-integration test-e2e test-all \
        test-chaos test-stress test-security test-race test-fuzz \
        build lint vet check coverage

# Quick tests (run on every save)
test-unit:
	go test -v -count=1 -short -run Test[A-Z] ./internal/...

# Safety-critical tests (run on every commit)
test-safety:
	go test -v -count=1 -run TestCritical ./internal/modes/
	go test -v -count=1 -run 'TestKillSafetyVerifier|TestWhitelist|TestNormalize' ./internal/modes/

# Whitelist sync check (safety.go systemCritical vs whitelist.json)
test-sync:
	go test -v -count=1 -run TestCritical_WhitelistSyncWithSystemCritical ./internal/modes/

# Integration tests (DI fakes + in-memory)
test-integration:
	go test -v -count=1 -race -run TestIntegration ./test/integration/...

# End-to-end tests (in VM only)
test-e2e:
	ROBOTY_E2E=1 ROBOTY_SAFE_MODE=true go test -v -count=1 -run TestE2E ./test/e2e/...

# Chaos + failure injection
test-chaos:
	go test -v -count=1 -race -run TestChaos ./test/chaos/...

# Stress tests
test-stress:
	go test -v -count=1 -race -run TestStress ./test/stress/...

# Security tests
test-security:
	go test -v -count=1 -run TestSecurity ./test/security/...
	govulncheck ./...
	gosec -quiet ./...

# Race detector (requires CGO_ENABLED=1; ensure MinGW on Windows)
test-race:
	@echo "NOTE: CGO_ENABLED=1 is required for -race. Set it in CI or env."
	go test -race -count=1 -short ./internal/...

# Fuzz tests (minimum 30s per target)
test-fuzz:
	go test -fuzz=FuzzNormalizeKillExec -fuzztime=30s ./internal/modes/
	go test -fuzz=FuzzIsAlwaysAllowed -fuzztime=30s ./internal/modes/
	go test -fuzz=FuzzIsAllowed -fuzztime=30s ./internal/modes/

# VM test matrix (requires Vagrant)
test-vm:
	@echo "=== VM Test Matrix (P0) ==="
	@echo "  Windows 11 24H2:    vagrant up win11"
	@echo "  Windows 10 22H2:    vagrant up win10"
	@echo "  Ubuntu 24.04:       vagrant up ubuntu"
	@echo "  macOS (no Vagrant): run on macos-latest GH runner"

# Full test suite
test-all: test-unit test-safety test-sync test-integration test-race test-fuzz test-security

# Build
build:
	go build -v -buildmode=pie -ldflags="-s -w" -o build/roboty .

# Lint + Vet
lint:
	staticcheck ./...
	golangci-lint run ./...

vet:
	go vet ./...

check: vet lint test-sync test-safety test-unit

# Coverage
coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./internal/...
	go tool cover -html=coverage.out -o coverage.html
