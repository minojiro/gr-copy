package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Image struct {
	FileName string
	Dir      string
	SrcURI   string
}

func fetchImageList() ([]Image, error) {
	type ApiRes struct {
		Dirs []struct {
			Name  string   `json:"name"`
			Files []string `json:"files"`
		} `json:"dirs"`
	}
	BASE_URI := "http://192.168.0.1"
	PHOTOS_URI := BASE_URI + "/v1/photos"

	imageList := []Image{}

	// fetch api
	httpClient := http.Client{Timeout: 5 * time.Second}
	res, err := httpClient.Get(PHOTOS_URI)
	if err != nil {
		return imageList, err
	}
	defer res.Body.Close()

	// read and decode JSON
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return imageList, err
	}
	apiRes := ApiRes{}
	if err := json.Unmarshal(body, &apiRes); err != nil {
		return imageList, err
	}

	for _, dir := range apiRes.Dirs {
		for _, fileName := range dir.Files {
			imageList = append(imageList, Image{
				Dir:      dir.Name,
				FileName: fileName,
				SrcURI:   PHOTOS_URI + "/" + dir.Name + "/" + fileName,
			})
		}
	}
	return imageList, nil
}

func downloadImages(imageList []Image, distDir string) error {
	imageListLen := len(imageList)
	fmt.Printf("copying to: %s\n", distDir)

	for i, image := range imageList {
		os.MkdirAll(distDir+"/"+image.Dir, os.ModePerm)

		distPath := distDir + "/" + image.Dir + "/" + image.FileName
		label := fmt.Sprintf(
			"%d/%d %s", i+1, imageListLen, image.Dir+"/"+image.FileName)
		if _, err := os.Stat(distPath); !os.IsNotExist(err) {
			continue
		}

		req, _ := http.NewRequest("GET", image.SrcURI, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		f, _ := os.OpenFile(distPath, os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			label,
		)
		io.Copy(io.MultiWriter(f, bar), resp.Body)
	}

	fmt.Println("done 🎉")
	return nil
}

func main() {

	imageList, err := fetchImageList()
	if err != nil {
		panic(err)
	}

	curDir, _ := os.Getwd()
	dist := curDir + "/dist"
	if err := downloadImages(imageList, dist); err != nil {
		panic(err)
	}
}
