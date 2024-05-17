module github.com/ethereum-optimism/optimism/op-service/eigenda

go 1.19

replace github.com/ethereum/go-ethereum v1.11.6 => github.com/mantlenetworkio/op-geth v0.0.0-20240515065754-1acb19be4efb

require (
	github.com/Layr-Labs/eigenda/api v0.6.2
	github.com/ethereum/go-ethereum v1.11.6
	github.com/urfave/cli v1.22.15
	github.com/urfave/cli/v2 v2.27.2
	google.golang.org/grpc v1.64.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.4 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240312152122-5f08fbb34913 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240318140521-94a12d6c2237 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)
