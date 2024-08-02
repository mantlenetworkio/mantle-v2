package eigenda

type Metrics interface {
	RecordInterval(method string) func(error)
}
