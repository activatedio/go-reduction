mock_packages := internal,internal

generate_mocks:
	find . -type f | grep mock_ | xargs rm || true
	go install github.com/vektra/mockery/v2@v2.46.3
	mockery

nix_shell:
	nix-shell default.nix --command $${SHELL}

test:
	go test ./...

