package cmd

const (
    VersionHash = "9430040"
    VersionTag  = ""
)

func Version() string {
	return VersionTag + "(" + VersionHash + ")"
}
