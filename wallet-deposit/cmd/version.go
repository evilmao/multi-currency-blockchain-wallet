package cmd

const (
    VersionHash = "6491657"
    VersionTag  = ""
)

func Version() string {
	return VersionTag + "(" + VersionHash + ")"
}
