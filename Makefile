OPERATOR_NAME := go-scimmer
IMAGE ?= hi-fi/$(OPERATOR_NAME)
ifeq ($(USE_JSON_OUTPUT), 1)
GOTEST_REPORT_FORMAT := -json
endif

.PHONY: clean deps test gofmt ensure run build build-image build-linux-amd64

clean:
	git clean -Xdf

deps:
	GO111MODULE=off go get -u golang.org/x/lint/golint

test:
	GO111MODULE=on go test ./... -v -coverprofile=gotest-coverage.out $(GOTEST_REPORT_FORMAT) > gotest-report.out && cat gotest-report.out || (cat gotest-report.out; exit 1)
	GO111MODULE=off golint -set_exit_status cmd/... pkg/... > golint-report.out && cat golint-report.out || (cat golint-report.out; exit 1)

ensure:
	GO111MODULE=on go mod tidy

run: build
	./bin/$(OPERATOR_NAME)

build-linux-amd64:
	rm -rf bin/linux/$(OPERATOR_NAME)
	GO111MODULE=on GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -i -o bin/linux/$(OPERATOR_NAME) ./cmd

build:
	rm -f bin/$(OPERATOR_NAME)
	GO111MODULE=on go build -v -i -o bin/$(OPERATOR_NAME) ./cmd

build-image: build-linux-amd64
	docker build -t $(IMAGE):latest .
