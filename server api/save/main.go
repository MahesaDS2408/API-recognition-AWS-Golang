package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AWSCredential struct {
	ACCESS_KEY        string
	SECRET_ACCESS_KEY string
	SESSION_TOKEN     string
}

func AWSHeaderParser(r *http.Request) (*rekognition.Rekognition, error) {
	accessKey := r.Header.Get("AWS-TOKEN")
	secretKey := r.Header.Get("AWS-SECRET")
	sessionToken := r.Header.Get("AWS-SESSION")

	if accessKey == "" || secretKey == "" {
		return nil, errors.New("Cek Header, kamu sudah memberikan token credensial?")
	}

	sess, e := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
		Credentials: credentials.NewStaticCredentials(
			accessKey,
			secretKey,
			sessionToken,
		)})

	return rekognition.New(sess), e
}

func AWSHCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header, err := AWSHeaderParser(r)
		if err != nil {
			http.Error(w, err.Error(), 422)
			return
		}
		ctx := context.WithValue(r.Context(), "aws_header", header)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CheckAWS() *session.Session {

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			os.Getenv("SESSION_TOKEN")),
	})

	if err == nil {
		return sess
	} else {
		return nil
	}

}

func main() {
	var config = CheckAWS()

	svc := rekognition.New(config)

	if len(os.Args) < 2 {
		fmt.Println("Tidak ada Perintah")
		return
	}

	switch os.Args[1] {
	case "getText":
		filename := flag.String("f", "", "The file to upload")

		// this will be placed with buffer upload
		target, err := ioutil.ReadFile(*filename)
		if err != nil {
			panic(err)
		}

		t := &rekognition.DetectTextInput{
			Image: &rekognition.Image{
				Bytes: target,
			},
		}
		if config != nil {
			fmt.Println("Succses")
		}

		res, err := svc.DetectText(t)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(res)
	case "detectMuka":
		face1 := flag.String("f1", "", "The file to upload")
		face2 := flag.String("f2", "", "The file to upload")

		t1, err := ioutil.ReadFile(*face1)
		if err != nil {
			panic(err)
		}
		t2, err := ioutil.ReadFile(*face2)
		if err != nil {
			panic(err)
		}

		muka := rekognition.CompareFacesInput{
			SourceImage: &rekognition.Image{
				Bytes: t1,
			},
			TargetImage: &rekognition.Image{
				Bytes: t2,
			},
		}
		res, _ := svc.CompareFaces(&muka)

		for _, v := range res.FaceMatches {
			fmt.Println(&v.Similarity)
		}
	case "online":
		r := chi.NewRouter()
		r.Use(middleware.Logger)
		r.Use(AWSHCtx)

		r.Post("/nomor_plat", uploadFile)

		r.Post("/cocokan_wajah", uploadWajah)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Okeh!"))
		})

		fmt.Println("Server running @ :" + os.Getenv("PORT"))
		http.ListenAndServe(":"+os.Getenv("PORT"), r)
	default:
		fmt.Println("Tidak ada Perintah")
		return
	}
}
