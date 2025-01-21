package main

import (
	"net/http"
	"os"
	"io"
	"mime"
	"context"
	"fmt"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

	const maxMemory = 1 << 30
	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse media type", err)
		return
	}

	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "File is not a valid mp4 video", nil)
		return
	}

	temp, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create temp file", err)
		return
	}
	defer os.Remove(temp.Name())
	defer temp.Close()

	if _, err := io.Copy(temp, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't copy file", err)
		return
	}

	ratio, err := getVideoAspectRatio(temp.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get aspect ratio", err)
		return
	}

	prefix := "other"
	if ratio == "16:9"{
		prefix = "landscape"
	}else if ratio == "9:16"{
		prefix = "portrait"
	}

	temp.Seek(0, io.SeekStart)

	output, err := processVideoForFastStart(temp.Name())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get fast video", err)
		return
	}

	tempFast, err := os.Open(output)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't read fast temp file", err)
		return
	}
	defer os.Remove(tempFast.Name())
	defer tempFast.Close()

	key := fmt.Sprintf("%s/%s", prefix, getAssetPath(mediaType))

	//fmt.Println(key)
	//return

	_, err = cfg.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &cfg.s3Bucket,
		Key: &key,
		Body: tempFast,
		ContentType: &mediaType,
	})
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't put file in bucket", err)
		return
	}


	//url := cfg.getAssetURL(assetPath)
	//url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)
	url := strings.Join([]string{cfg.s3Bucket, key}, ",")
	video.VideoURL = &url

	sVid, err := cfg.dbVideoToSignedVideo(video)
	if err != nil{
		respondWithError(w, http.StatusInternalServerError, "Couldn't sign video url", err)
		return
	}
	
	err = cfg.db.UpdateVideo(sVid)
	if err != nil {
		//delete(videoThumbnails, videoID)
		respondWithError(w, http.StatusBadRequest, "Unable to update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
