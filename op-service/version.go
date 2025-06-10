package op_service

var (
	Version   = "v0.0.0"
	GitCommit = ""
	GitDate   = ""
	Meta      = "dev"
)

func DefaultFormatVersion() string {
	return FormatVersion(Version, GitCommit, GitDate, Meta)
}

func FormatVersion(version string, gitCommit string, gitDate string, meta string) string {
	v := version
	if gitCommit != "" {
		if len(gitCommit) >= 8 {
			v += "-" + gitCommit[:8]
		} else {
			v += "-" + gitCommit
		}
	}
	if gitDate != "" {
		v += "-" + gitDate
	}
	if meta != "" {
		v += "-" + meta
	}
	return v
}
