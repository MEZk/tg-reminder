# Get the latest commit branch, hash, and date
TAG=$(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
BRANCH=$(if $(TAG),$(TAG),$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null))
HASH=$(shell git rev-parse --short=7 HEAD 2>/dev/null)
TIMESTAMP=$(shell git log -1 --format=%ct HEAD 2>/dev/null | xargs -I{} date -u -r {} +%Y%m%dT%H%M%S)
GIT_REV=$(shell printf "%s-%s-%s" "$(BRANCH)" "$(HASH)" "$(TIMESTAMP)")
REV=$(if $(filter --,$(GIT_REV)),latest,$(GIT_REV)) # fallback to latest if not in git repo

DB_FILE="var/tg-reminder/tg-reminder.db"

LOCAL_BIN=$(CURDIR)/bin
MOQ_BIN?=$(LOCAL_BIN)/moq
GOLANGCI_BIN?=$(LOCAL_BIN)/golangci-lint
GORELEASER?=$(LOCAL_BIN)/goreleaser

SHELL=/bin/bash
TEST_COVERAGE_THRESHOLD=90.0


docker:
	docker build -t mezk/tg-reminder .

race-test:
	go test -race -timeout=100s -count 1 ./...

build:
	mkdir -p bin
	cd cmd/bot && go build -ldflags "-X main.revision=$(REV) -s -w" -o ../../bin/tg-reminder.$(BRANCH)
	cp bin/tg-reminder.$(BRANCH) bin/tg-reminder

release:
	rm -f bin/release
	mkdir -p bin/release
	@echo release to bin/release
	${GORELEASER} --snapshot --clean
	ls -l bin/release

test:
	go clean -testcache
ifneq ($(CI),)
	echo 'Running tests in CI...'
	go test -v -timeout=100s -covermode=atomic -coverprofile=coverage.out ./...
else
	echo 'Running tests locally...'
	rm -f coverage.out coverage_no_mocks.out
	go test -timeout=100s -covermode=atomic -coverprofile=coverage.out ./...
endif
	grep -v "_mock.go" coverage.out | grep -v mocks > coverage_no_mocks.out ;\
	coverage=$$(go tool cover -func=coverage_no_mocks.out | grep total | grep -Eo '[0-9]+\.[0-9]+') ;\
	echo -e "\033[32mCurrent code coverage:       $$coverage %" ;\
	echo -e "\033[32mCode coverage threshold:     $(TEST_COVERAGE_THRESHOLD) %" ;\
	if [ $$(bc <<< "$$coverage < $(TEST_COVERAGE_THRESHOLD)") -eq 1 ]; then \
		echo -e "\033[31mLow test coverage: $$coverage < $(TEST_COVERAGE_THRESHOLD). Please add more tests!" ;\
		exit 1 ;\
	fi

new-db:
	rm -f $(DB_FILE)
	touch $(DB_FILE)

bin-deps:
	tmp=$$(mktemp -d) && cd $$tmp && pwd && go mod init temp && \
	GOBIN=$(LOCAL_BIN) go install github.com/matryer/moq@v0.3.4 && \
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1 && \
	GOBIN=$(LOCAL_BIN) go install github.com/goreleaser/goreleaser/v2@v2.1.0 && \
	rm -fr $$tmp

mocks:
	find . -type f -name "*_mock.go" -exec rm {} \;
	$(MOQ_BIN) --out internal/pkg/sender/mocks/botapi_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/sender BotAPI
	$(MOQ_BIN) --out internal/pkg/notifier/mocks/storage_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/notifier Storage
	$(MOQ_BIN) --out internal/pkg/notifier/mocks/sender_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/notifier BotResponseSender
	$(MOQ_BIN) --out internal/pkg/listener/mocks/botapi_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/listener BotAPI
	$(MOQ_BIN) --out internal/pkg/listener/mocks/updates_receiver_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/listener UpdateReceiver
	$(MOQ_BIN) --out internal/pkg/bot/mocks/storage_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/bot Storage
	$(MOQ_BIN) --out internal/pkg/bot/mocks/response_sender_mock.go --pkg mocks --skip-ensure --with-resets internal/pkg/bot ResponseSender

lint:
	$(GOLANGCI_BIN) run \
     --config=.golangci.yml \
     --sort-results \
     --max-issues-per-linter=1000 \
     --max-same-issues=1000 \
     ./...

.PHONY: docker race-test build test lint

