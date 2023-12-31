# nodeinfo

[![docs](https://pkg.go.dev/badge/github.com/rafaelespinoza/nodeinfo.svg)](https://pkg.go.dev/github.com/rafaelespinoza/nodeinfo)
[![tests](https://github.com/rafaelespinoza/nodeinfo/actions/workflows/tests.yaml/badge.svg)](https://github.com/rafaelespinoza/nodeinfo/actions/workflows/tests.yaml)

This project is an HTTP client for fetching [NodeInfo](https://nodeinfo.diaspora.software) data.
Source code (golang) is under `nodeinfo/`. Integration tests are at `internal/tests/`.

## tests

### unit tests

```sh
make unit_test

# pass in some flags to "go test"
make unit_test ARGS='-count 1 -coverprofile /tmp/cover.out'
```

### integration tests

Integration test requirements:
- A Unix-like platform
- [`jq`](https://jqlang.github.io/jq/)
- [`go`](https://go.dev/), version >= 1.20

Run integration tests:
```sh
make integration_test

# By default, it will use the first go binary in your PATH.
# Specify a path to another golang version like so:
GO=/path/to/other/golang/version/bin make integration_test
```

Some testdata is already committed to source control. Get new testdata and write to a temporary directory:
```sh
make fetch_integration_testdata
```
