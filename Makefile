GO ?= go
SRC_PATHS ?= ./...

vet:
	$(GO) vet $(ARGS) $(SRC_PATHS)

unit_test:
	$(GO) test -v $(ARGS) $(SRC_PATHS)

integration_test:
	# This part requires golang v1.20 or higher.
	GO=$(GO) ./internal/tests/integration_test.sh $(ARGS)

fetch_integration_testdata:
	./internal/tests/fetch_testdata.sh $(ARGS)
