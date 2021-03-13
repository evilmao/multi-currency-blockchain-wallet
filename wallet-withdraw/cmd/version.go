package cmd

const (
    VersionHash = "1732672"
    VersionTag  = ""
)

func Version() string {
	return VersionTag + "(" + VersionHash + ")"
}
