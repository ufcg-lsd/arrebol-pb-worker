package utils

import (
	"log"
	"os/exec"
	"strings"
)

func GetPrjPath() string {
	path_cmd := exec.Command("/bin/sh", "-c", "echo $GOPATH")
	path, err := path_cmd.Output()

	if err != nil {
		log.Fatal("Unable to fetch path_cmd output")
	}

	path_str := strings.TrimSpace(string(path))
	return path_str + "/src/github.com/ufcg-lsd/arrebol-pb-worker/"
}
