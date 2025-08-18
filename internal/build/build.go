package build

var (
	xBuildCommit  = "dev"
	xBuildVersion = "dev"
)

func GetCommit() string {
	return xBuildCommit
}

func GetVersion() string {
	return xBuildVersion
}
