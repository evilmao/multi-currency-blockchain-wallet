package cmd

const (
    VersionHash = "f8aaaa5"
    VersionTag  = ""
)

func Version() string {
	return VersionTag + "(" + VersionHash + ")"
}
