mock_packages := internal,internal

dev_containers:
	docker compose stop || true
	docker compose rm -f || true
	docker compose up -d --wait --remove-orphans
	go install github.com/activatedio/go-healthchecks@v0.0.6
	go-healthchecks -c hc.yaml check
	go clean -testcache

generate_mocks:
	go mod vendor
	@for v in $(mock_packages) ; do \
		vpath=$$(echo $$v | cut -f1 -d,) ; \
		vname=$$(echo $$v | cut -f2 -d,) ; \
		echo "Generating mocks for $$vpath, $$vname"; \
		docker run --rm -v ${PWD}:/src -w /src/$$vpath --user $$(id -u):$$(id -g) vektra/mockery --name=.*  --with-expecter --outpkg mock_$$vname --output ./mock_$$vname ; \
  done
	rm -fr vendor
	go fmt ./...

test:
	go test ./...

