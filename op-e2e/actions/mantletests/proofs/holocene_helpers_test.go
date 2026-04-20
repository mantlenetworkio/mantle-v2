package proofs

type logExpectations struct {
	role   string
	filter string
	num    int
}
type expectations struct {
	safeHead uint64
	logs     []logExpectations
}

func sequencerOnce(filter string) []logExpectations {
	return []logExpectations{{filter: filter, role: "sequencer", num: 1}}
}
