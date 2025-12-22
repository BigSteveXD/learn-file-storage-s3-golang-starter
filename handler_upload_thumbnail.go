package main

import (
	//"fmt"
	"net/http"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
	"io"
	//"log"
	//"encoding/base64"
	"os"
	//"path/filepath"
	//"strings"
	"mime"
	//"crypto/rand"
	//"encoding/base64"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
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

	//fmt.Println("uploading thumbnail for video", videoID, "by user", userID)



	//Parse form data
	const maxMemory = 10 << 20//bit shift num 10 left 20 times. 10MB
	r.ParseMultipartForm(maxMemory)

	//Get image data from the form
	//get file data and file headers. key is "thumbnail"
	fileData, fileHeader, err := r.FormFile("thumbnail")//file, header
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't get file data", err)
		return
	}
	defer fileData.Close()
	//get media type from files Content-Type header
	//mediaTypeVar := fileHeader.Header.Get("Content-Type")//*multipart.FileHeader
	/*
	if mediaTypeVar == "" {
		respondWithError(w, http.StatusBadRequest, "couldn't get Content-Type for thumbnail", nil)
		return
	}
	*/
	//check if file uploaded is the correct type
	mediaTypeVar, _, err := mime.ParseMediaType(fileHeader.Header.Get("Content-Type"))//params
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "error parsing media", err)
		return
	}
	if mediaTypeVar != "image/jpeg" && mediaTypeVar != "image/png" {
		respondWithError(w, http.StatusBadRequest, "wrong media type", err)
		return
	}

	//Read image data into byte slice
	/*
	imageData, err := io.ReadAll(fileData)//data
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't read file data", err)
		return
	}
	*/

	//Get video's metadata from SQLite db
	videoMetadata, err := cfg.db.GetVideo(videoID)//video
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get video", err)
		return
	}
	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "authenicated user and video owner don't match", nil)
		return
	}
	

	//Save thumbnail to global map
	//add thumbnail to global map using video's ID as key
	/*
	videoThumbnails[videoID] = thumbnail{
		data: imageData, 
		mediaType: mediaTypeVar,
	}
	*/
	//Update video metadata with new thumbnail URL, update record in db
	//ex http://localhost:<port>/api/thumbnails/{videoID}
	//url := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoID)

	//convert image data to base64 string
	//encodedVideoString := base64.StdEncoding.EncodeToString([]byte(imageData))
	//create data url
	//ex data:<media-type>;base64,<data>
	//url := fmt.Sprintf("data:%s;base64,%s", mediaTypeVar, encodedVideoString)

	//save to file
	///assets/<videoID>.<file_extension>
	//temp := videoID.String() + "." + strings.Replace(mediaTypeVar, "image/", "", -1)



	assetPath := getAssetPath(videoID, mediaTypeVar)//mediaType
	assetDiskPath := cfg.getAssetDiskPath(assetPath)


	//newFile, err := os.Create(filepath.Join(cfg.assetsRoot, temp))
	newFile, err := os.Create(assetDiskPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create file on server", err)
		return
	}
	defer newFile.Close()
	_, err = io.Copy(newFile, fileData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}
	
	url := cfg.getAssetURL(assetPath)

	videoMetadata.ThumbnailURL = &url
	err = cfg.db.UpdateVideo(videoMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't update video", err)
		return
	}

	//Respond with updated JSON of video's metadata
	respondWithJSON(w, http.StatusOK, videoMetadata)//200
}
