package utils

import (
	"os/exec"
	"strings"
)

func GetPrjPath() string {
	path_cmd := exec.Command("/bin/sh", "-c", "echo $GOPATH")
	path, _ := path_cmd.Output()
	path_str := strings.TrimSpace(string(path))
	return path_str + "/src/github.com/ufcg-lsd/arrebol-pb-worker/"
}
