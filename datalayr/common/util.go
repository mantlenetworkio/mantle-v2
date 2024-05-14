package common

func PrefixEnvVar(prefix, suffix string) string {
	return prefix + "_" + suffix
}
