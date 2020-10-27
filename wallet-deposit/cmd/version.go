package cmd

const (
	VersionString = "v1.0"
	VersionTag    = ""
)

func Version() string {
	return VersionTag + "(" + VersionString + ")"
}
