package util

import (
	"fmt"
	"os/exec"
	"strings"
)

func Int64Ptr(i int64) *int64 {
	return &i
}

func BoolPtr(b bool) *bool {
	return &b
}

func RunCmdDetailed(cmd *exec.Cmd) error {
	var outbuf, errbuf strings.Builder
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("exec error: %w\nstdout:\n%s\n\nstderr:\n%s", err, outbuf.String(), errbuf.String())
	}
	return nil
}
