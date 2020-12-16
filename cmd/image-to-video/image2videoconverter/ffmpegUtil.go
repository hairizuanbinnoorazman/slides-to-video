package image2videoconverter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type ffprobeFormatted struct {
	Format audioDuration
}

type audioDuration struct {
	Duration string
}

func convertToUseAAC(filename, adjustedFilename string) error {
	// tempFilename := strings.Replace(filename, ".mp3", ".m4a", -1)
	cmd := exec.Command("ffmpeg", "-y", "-i", filename, "-c:a", "aac", adjustedFilename)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		log.Println(out.String())
		log.Println(stderr.String())
		return err
	}
	return nil
}

func getAudioDuration(filename string) (duration float32, err error) {
	cmd := exec.Command("ffprobe", "-i", filename, "-show_entries", "format=duration", "-v", "quiet", "-of", "json")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return 0.0, fmt.Errorf("Error %v %v", out.String(), stderr.String())
	}
	var ffprobe ffprobeFormatted
	rawAudioProbe := out.Bytes()
	json.Unmarshal(rawAudioProbe, &ffprobe)
	val, err := strconv.ParseFloat(ffprobe.Format.Duration, 32)
	if err != nil {
		return 0.0, fmt.Errorf("Unable to parse the following value. %v %v %v", val, ffprobe, rawAudioProbe)
	}
	return float32(val), nil
}

func generateSilentVideo(imageFilename string, duration float32, outputFile string) error {
	cmd := exec.Command("ffmpeg", "-r", "1/"+fmt.Sprintf("%f", duration), "-i", imageFilename, "-y", "-c:v", "libx264", "-vf", "fps=25", "-pix_fmt", "yuv420p", outputFile)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		log.Println(out.String())
		log.Println(stderr.String())
		return err
	}
	return nil
}

// addSilentAudio
// Creates a 1s silent audio track which is to be appended to the actual audio track
func addSilentAudio(filename, outputFilename string) error {
	silentFilename := "silent_" + outputFilename
	defer func() {
		os.Remove(silentFilename)
	}()
	silentCmd := exec.Command("ffmpeg", "-y", "-filter_complex", "aevalsrc=0", "-t", "1", silentFilename)
	cmd := exec.Command("ffmpeg", "-i", "concat:"+silentFilename+"|"+filename+"|"+silentFilename, "-y", "-c", "copy", outputFilename)
	var out bytes.Buffer
	var stderr bytes.Buffer
	silentCmd.Stdout = &out
	silentCmd.Stderr = &stderr
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := silentCmd.Run()
	if err != nil {
		log.Println(silentCmd.Args)
		log.Println(err)
		log.Println(out.String())
		log.Println(stderr.String())
		return err
	}
	err = cmd.Run()
	if err != nil {
		log.Println(cmd.Args)
		log.Println(err)
		log.Println(out.String())
		log.Println(stderr.String())
		return err
	}
	return nil
}

func muxSilentVideoAndAudio(silentVideoFilename, audioFilename, outputFilename string) error {
	cmd := exec.Command("ffmpeg", "-i", silentVideoFilename, "-y", "-i", audioFilename, "-c:v", "copy", "-c:a", "aac", outputFilename)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		log.Println(out.String())
		log.Println(stderr.String())
		return err
	}
	return nil
}
