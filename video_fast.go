package main

import(
	"os/exec"
)

func processVideoForFastStart(filepath string)(string, error){
	//outputPath := ".processing" + filepath
	outputPath := filepath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	err := cmd.Run()
	if err != nil{
		return "", err
	}

	return outputPath, nil
}