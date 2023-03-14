.SILENT :
.PHONY : vet test-pgx

WITH_ENV = env `cat .env 2>/dev/null | xargs`

GO=$(shell which go)
GOMOD=$(shell echo "$${GO111MODULE:-auto}")


vet:
	echo "Checking ./... , with GOMOD=$(GOMOD)"
	GO111MODULE=$(GOMOD) $(GO) vet ./...


test-oid: vet
	mkdir -p tests
	@$(WITH_ENV) GO111MODULE=$(GOMOD) $(GO) test -v -cover -coverprofile tests/cover_oid.out -count=1 ./models/oid
	@$(WITH_ENV) $(GO) tool cover -html=tests/cover_oid.out -o tests/cover_oid.out.html

test-pgx: vet
	mkdir -p tests
	@$(WITH_ENV) GO111MODULE=$(GOMOD) $(GO) test -v -cover -coverprofile tests/cover_pgx.out -count=1 ./stores/pgx
	@$(WITH_ENV) $(GO) tool cover -html=tests/cover_pgx.out -o tests/cover_pgx.out.html
