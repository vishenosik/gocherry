package gocherry

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v2"
)

const (
	_yaml_ = iota
	_json_

	unset = "not configured"
)

// Build flags variables
var (
	BuildDate string = unset
	GitBranch string = unset
	GitCommit string = unset
	GoVersion string = unset
	GitTag    string = unset
)

type BuildInfo struct {
	BuildDate string `json:"build_date,omitempty" yaml:"build_date,omitempty"`
	GitBranch string `json:"git_branch,omitempty" yaml:"git_branch,omitempty"`
	GitCommit string `json:"git_commit,omitempty" yaml:"git_commit,omitempty"`
	GoVersion string `json:"go_version,omitempty" yaml:"go_version,omitempty"`
	GitTag    string `json:"git_tag,omitempty" yaml:"git_tag,omitempty"`
}

var buildInfo BuildInfo

func init() {
	buildInfo = BuildInfo{
		BuildDate: BuildDate,
		GitBranch: GitBranch,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
		GitTag:    GitTag,
	}
}

func BuildInfoYaml(writer io.Writer) {
	writeBuildInfo(writer, _yaml_)
}

func BuildInfoJson(writer io.Writer) {
	writeBuildInfo(writer, _json_)
}

func writeBuildInfo(writer io.Writer, format int) {
	var (
		buf []byte
		err error
	)

	switch format {
	case _yaml_:
		buf, err = yaml.Marshal(buildInfo)

	case _json_:
		buf, err = json.MarshalIndent(buildInfo, "", "  ")

	default:
		return
	}
	if err != nil {
		writer.Write([]byte(err.Error()))
		return
	}
	writer.Write(buf)
}
