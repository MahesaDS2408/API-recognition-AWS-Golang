package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/service/rekognition"
)

func rekamOrang(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}
	r.ParseMultipartForm(10 << 20)

	filePlatMotor, _, err := r.FormFile("plat_motor")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer filePlatMotor.Close()

	filePlatMotorBytes, err := ioutil.ReadAll(filePlatMotor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	fileMuka, _, err := r.FormFile("muka_masuk")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer fileMuka.Close()

	fileMukaKeluarBytes, err := ioutil.ReadAll(fileMuka)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	t := &rekognition.DetectTextInput{
		Image: &rekognition.Image{
			Bytes: filePlatMotorBytes,
		},
	}
	if svc != nil {
		fmt.Println("Succses meminta ke AWS")
	}

	res, _ := svc.DetectText(t)

	var platNomor = ""

	if len(res.TextDetections) != 0 {
		platNomor = *(res.TextDetections[0].DetectedText)
	} else {
		http.Error(w, "Nomor Plat tidak terdeteksi", http.StatusUnprocessableEntity)
		return
	}

	var namaFileMukaMasuk = platNomor + "_wajah"

	fMukaMasuk, err := ioutil.ReadFile("./berkas/" + namaFileMukaMasuk)
	if err != nil {
		http.Error(w, "Nomor Plat tidak terdeteksi", http.StatusUnprocessableEntity)
		return
	}

	cocok, err := bandingWajah(fMukaMasuk, fileMukaKeluarBytes, svc)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if cocok {
		w.Write([]byte("Sama"))
		os.Remove("./berkas/" + platNomor + "_plat")
		os.Remove("./berkas/" + platNomor + "_wajah")
	} else {
		w.Write([]byte("Tidak Sama"))
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func keluarPlat(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}
	r.ParseMultipartForm(10 << 20)

	filePlatMotor, _, err := r.FormFile("plat_motor")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer filePlatMotor.Close()

	filePlatMotorBytes, err := ioutil.ReadAll(filePlatMotor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	t := &rekognition.DetectTextInput{
		Image: &rekognition.Image{
			Bytes: filePlatMotorBytes,
		},
	}

	if svc != nil {
		fmt.Println("Succses meminta ke AWS")
	}

	res, _ := svc.DetectText(t)

	var platNomor = ""

	if len(res.TextDetections) != 0 {
		platNomor = *(res.TextDetections[0].DetectedText) //deteksi
	} else {
		http.Error(w, "Nomor Plat tidak terdeteksi", http.StatusUnprocessableEntity) //tidak ketideksi
		return
	}

	berkas, err := os.Open("./.lock-out")
	// cek lock
	if os.IsExist(err) {
		berkas.Write([]byte(platNomor))
		berkas.Sync()

		var namaFilePlatMotor = platNomor + "_plat"
		fPlatMotor, _ := os.Create("./berkas/" + namaFilePlatMotor)

		_, err = fPlatMotor.Write(filePlatMotorBytes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fPlatMotor.Sync()
	} else {
		berkas, _ := os.Create("./.lock")
		berkas.Write([]byte(platNomor))
		berkas.Sync()

		var namaFilePlatMotor = platNomor + "_plat"
		fPlatMotor, _ := os.Create("./berkas/" + namaFilePlatMotor)

		_, err = fPlatMotor.Write(filePlatMotorBytes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fPlatMotor.Sync()
	}

	_, err = os.Stat("./berkas/temp_keluar_wajah")
	if !os.IsNotExist(err) {

		var namaFileMukaMasuk = string(platNomor) + "_wajah"
		fMukaMasuk, err := ioutil.ReadFile("./berkas/" + namaFileMukaMasuk)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fMukaKeluar, err := ioutil.ReadFile("./berkas/temp_keluar_wajah")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		cocok, err := bandingWajah(fMukaMasuk, fMukaKeluar, svc)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if cocok {
			w.Write([]byte("Sama"))
			os.Remove("./berkas/" + string(platNomor) + "_plat")
			os.Remove("./berkas/" + string(platNomor) + "_wajah")
			os.Remove("./.lock-out")
			os.Remove("./berkas/temp_keluar_wajah")
		} else {
			w.Write([]byte("Tidak Sama"))
		}
	} else {
		fmt.Println("No Temp Wajah")
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func keluarMuka(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	fileMuka, _, err := r.FormFile("muka_keluar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer fileMuka.Close()

	fileMukaMasukBytes, err := ioutil.ReadAll(fileMuka)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err = os.Stat("./.lock-out")
	fmt.Println(err)
	if !os.IsNotExist(err) {
		nomorPlat, err := ioutil.ReadFile("./.lock-out")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var namaFileMukaMasuk = string(nomorPlat) + "_wajah" // A 001 TE_wajah
		fMukaMasuk, err := ioutil.ReadFile("./berkas/" + namaFileMukaMasuk)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fMukaKeluar, err := ioutil.ReadFile("./berkas/temp_keluar_wajah")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		cocok, err := bandingWajah(fMukaMasuk, fMukaKeluar, svc)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if cocok {
			w.Write([]byte("Sama"))
			os.Remove("./berkas/" + string(nomorPlat) + "_plat")
			os.Remove("./berkas/" + string(nomorPlat) + "_wajah")
			os.Remove("./.lock-out")
			os.Remove("./berkas/temp_keluar_wajah")
		} else {
			w.Write([]byte("Tidak Sama"))
		}

	} else {
		fmt.Println("No Lock")
		var namaFileMukaMasuk = "temp_keluar_wajah"
		fMukaMasuk, err := os.Create("./berkas/" + namaFileMukaMasuk)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = fMukaMasuk.Write(fileMukaMasukBytes)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fMukaMasuk.Sync()
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func rekamPlat(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}
	r.ParseMultipartForm(10 << 20)

	filePlatMotor, _, err := r.FormFile("plat_motor")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer filePlatMotor.Close()

	filePlatMotorBytes, err := ioutil.ReadAll(filePlatMotor)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	t := &rekognition.DetectTextInput{
		Image: &rekognition.Image{
			Bytes: filePlatMotorBytes,
		},
	}

	if svc != nil {
		fmt.Println("Succses meminta ke AWS")
	}

	res, _ := svc.DetectText(t)

	var platNomor = ""

	if len(res.TextDetections) != 0 {
		platNomor = *(res.TextDetections[0].DetectedText)
	} else {
		http.Error(w, "Nomor Plat tidak terdeteksi", http.StatusUnprocessableEntity)
		return
	}

	berkas, err := os.Open("./.lock")
	// cek lock
	if os.IsExist(err) {
		berkas.Write([]byte(platNomor))
		berkas.Sync()

		var namaFilePlatMotor = platNomor + "_plat"
		fPlatMotor, _ := os.Create("./berkas/" + namaFilePlatMotor)

		_, err = fPlatMotor.Write(filePlatMotorBytes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fPlatMotor.Sync()
	} else {
		berkas, _ := os.Create("./.lock")
		berkas.Write([]byte(platNomor))
		berkas.Sync()

		var namaFilePlatMotor = platNomor + "_plat"
		fPlatMotor, _ := os.Create("./berkas/" + namaFilePlatMotor)

		_, err = fPlatMotor.Write(filePlatMotorBytes)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fPlatMotor.Sync()
	}

	_, err = os.Stat("./berkas/temp_wajah")
	if !os.IsNotExist(err) {
		err = os.Rename("./berkas/temp_wajah", "./berkas/"+platNomor+"_wajah")
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
		os.Remove("./.lock")
	} else {
		fmt.Println("No Temp Wajah")
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func rekamMuka(w http.ResponseWriter, r *http.Request) {
	fileMuka, _, err := r.FormFile("muka_masuk")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer fileMuka.Close()

	fileMukaMasukBytes, err := ioutil.ReadAll(fileMuka)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	_, err = os.Stat("./.lock")
	fmt.Println(err)
	if !os.IsNotExist(err) {
		nomorPlat, err := ioutil.ReadFile("./.lock")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var namaFileMukaMasuk = string(nomorPlat) + "_wajah"
		fMukaMasuk, err := os.Create("./berkas/" + namaFileMukaMasuk)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = fMukaMasuk.Write(fileMukaMasukBytes)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fMukaMasuk.Sync()
		os.Remove("./.lock")
		os.Remove("./berkas/temp_wajah")
	} else {
		fmt.Println("No Lock")
		var namaFileMukaMasuk = "temp_wajah"
		fMukaMasuk, err := os.Create("./berkas/" + namaFileMukaMasuk)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		_, err = fMukaMasuk.Write(fileMukaMasukBytes)

		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fMukaMasuk.Sync()
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	svc, ok := r.Context().Value("aws_header").(*rekognition.Rekognition)
	if !ok {
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("plat_motor")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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
		http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
		return
	}

	r.ParseMultipartForm(10 << 20)

	uploadIn, _, err := r.FormFile("muka_masuk")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer uploadIn.Close()

	uploadOut, _, err := r.FormFile("muka_keluar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	defer uploadOut.Close()

	bufferIn, err := ioutil.ReadAll(uploadIn)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	bufferOut, err := ioutil.ReadAll(uploadOut)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

func bandingWajah(wajahMasuk []byte, wajahKeluar []byte, svc *rekognition.Rekognition) (bool, error) {
	muka := rekognition.CompareFacesInput{
		SourceImage: &rekognition.Image{
			Bytes: wajahMasuk,
		},
		TargetImage: &rekognition.Image{
			Bytes: wajahKeluar,
		},
	}

	res, err := svc.CompareFaces(&muka)
	if err != nil {
		fmt.Println("Line 352")
		return false, err
	}

	if res != nil {
		switch akurasi := *(res.FaceMatches[0].Similarity); {
		case (akurasi > 85.0) && (akurasi < 100.0):
			return true, nil
		default:
			return false, nil
		}
	}
	return false, errors.New("Res Empty")
}
