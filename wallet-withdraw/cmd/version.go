package cmd

const (
    VersionHash = "6de988e"
    VersionTag  = ""
)

func Version() string {
	return VersionTag + "(" + VersionHash + ")"
}
