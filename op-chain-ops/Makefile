check-migration:
	go build -o ./bin/check-migration ./cmd/check-migration/main.go

check-migration-quick:
	go build -o ./bin/check-migration-quick ./cmd/check-migration/main.go

op-migrate:
	go build -o ./bin/op-migrate ./cmd/op-migrate/main.go

rollover:
	go build -o ./bin/rollover ./cmd/rollover/main.go

withdrawals:
	go build -o ./bin/withdrawals ./cmd/withdrawals/main.go

test:
	go test ./...

.PHONY: op-migrate rollover withdrawals test
