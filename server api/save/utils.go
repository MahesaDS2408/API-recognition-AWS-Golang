package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-sdk-go/service/rekognition"
)

func uploadFile(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("plat_motor")
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}

	t := &rekognition.DetectTextInput{
		Image: &rekognition.Image{
			Bytes: fileBytes,
		},
	}
	if svc != nil {
		fmt.Println("Succses")
	}

	res, _ := svc.DetectText(t)

	w.Write([]byte(*(res.TextDetections[0].DetectedText)))
}

func uploadWajah(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(422), 422)
		return
	}

	r.ParseMultipartForm(10 << 20)

	uploadIn, _, err := r.FormFile("muka_masuk")
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	defer uploadIn.Close()

	uploadOut, _, err := r.FormFile("muka_keluar")
	if err != nil {
		http.Error(w, err.Error(), 422)
		return
	}
	defer uploadOut.Close()

	bufferIn, err := ioutil.ReadAll(uploadIn)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 422)
		return
	}
	bufferOut, err := ioutil.ReadAll(uploadOut)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 422)
		return
	}

	muka := rekognition.CompareFacesInput{
		SourceImage: &rekognition.Image{
			Bytes: bufferIn,
		},
		TargetImage: &rekognition.Image{
			Bytes: bufferOut,
		},
	}
	res, err := svc.CompareFaces(&muka)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), 422)
		return
	}

	if res != nil {
		switch akurasi := *(res.FaceMatches[0].Similarity); {
		case (akurasi > 85.0) && (akurasi < 100.0):
			w.Write([]byte("Sama"))
		default:
			w.Write([]byte("Tidak sama"))
		}
	}
}
