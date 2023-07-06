GO ?= go
SRC_PATHS ?= ./...

vet:
	$(GO) vet $(ARGS) $(SRC_PATHS)

unit_test:
	$(GO) test $(ARGS) $(SRC_PATHS)

integration_test:
	./internal/tests/integration_test.sh
