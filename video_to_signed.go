package main

import(
	"strings"
	"errors"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"fmt"
)

func(cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error){
	split := strings.Split(*video.VideoURL, ",")
	if len(split) != 2{
		return database.Video{}, errors.New("No comma separated bucket and key")
	}
	bucket := split[0]
	key := split[1]

	signedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 5 * time.Minute)
	if err != nil{
		return database.Video{}, err
	}
	fmt.Println(signedURL)
	video.VideoURL = &signedURL
	return video, nil
}