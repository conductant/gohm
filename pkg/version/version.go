package version

import (
	"flag"
	"fmt"
	"os"
)

var (
	flagVersion = flag.Bool("version", false, "Show version info")

	// These are bound by the -X key value ldflags option of the go compiler
	gitRepo        string
	gitBranch      string
	gitTag         string
	gitCommitHash  string
	buildTimestamp string
	buildNumber    string
	buildLabel     string
)

var (
	build_info = &Build{
		Timestamp: buildTimestamp,
		Label:     buildLabel,
		Number:    buildNumber,
		RepoUrl:   gitRepo,
		Tag:       gitTag,
		Branch:    gitBranch,
		Commit:    gitCommitHash,
	}
)

type Build struct {
	RepoUrl   string
	Branch    string
	Tag       string
	Commit    string
	Label     string
	Timestamp string
	Number    string
}

func BuildInfo() *Build {
	return build_info
}

func SetBuildInfo(b *Build) {
	build_info = b
}

func (this *Build) HandleFlag() {
	if *flagVersion {
		fmt.Println(this.Notice())
		os.Exit(0)
	}
}

func (this *Build) Notice() string {
	return fmt.Sprintf("%s: Version %s (%s), Build %s, Label %s. Built on %s.\n",
		os.Args[0], this.Tag, this.Commit, this.Number, this.Label, this.Timestamp)
}

func (this *Build) GetRepoUrl() string {
	return this.RepoUrl
}

func (this *Build) GetBranch() string {
	return this.Branch
}

func (this *Build) GetTag() string {
	return this.Tag
}

func (this *Build) GetCommitHash() string {
	return this.Commit
}

func (this *Build) GetTimestamp() string {
	return this.Timestamp
}

func (this *Build) GetNumber() string {
	return this.Number
}

func (this *Build) GetLabel() string {
	return this.Label
}
