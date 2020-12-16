package videoconcater

import (
	"bytes"
	"fmt"
	"os/exec"
)

func combineVideo(videoListFile, combinedOutputVideoFile string) error {
	cmd := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", videoListFile, "-c", "copy", combinedOutputVideoFile)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Error: %v, Stdout: %v, Stderr: %v", err, out.String(), stderr.String())
	}
	return nil
}
