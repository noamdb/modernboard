package media

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	filesFolder      = "files/"
	thumbnailsFolder = "thumbnails/"
)

var validTypes = map[string]string{"image/jpeg": "jpg", "image/png": "png", "video/webm": "webm", "image/gif": "gif"}

func SaveFile(r io.Reader, path string, maxSize int64) (string, string, error) {
	buf := bufio.NewReader(r)
	sniff, _ := buf.Peek(512)
	contentType := http.DetectContentType(sniff)
	ext, ok := validTypes[contentType]
	if !ok {
		return "", "", errors.New("invalid file type")
	}
	f, err := ioutil.TempFile(path,
		fmt.Sprintf("%s*.%s", time.Now().Format("20060102"), ext))
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	lmt := io.MultiReader(buf, io.LimitReader(r, maxSize-511))
	written, err := io.Copy(f, lmt)
	if err != nil && err != io.EOF {
		return "", "", err
	}
	if written > maxSize {
		return "", "", errors.New("surpassed file size limit")
	}
	return filepath.Base(f.Name()), contentType, err
}

func DeleteFile(path string) error {
	if f, err := os.Stat(path); os.IsNotExist(err) || f.IsDir() {
		return nil
	}

	if err := os.Remove(path); err != nil {
		log.WithFields(log.Fields{
			"event": "delete file",
			"error": err,
		}).Error("could not delete file")
		return err
	}
	return nil
}

func CreateThumbnail(contentType string, fileName string, size int) (string, error) {
	thumbnailName := strings.SplitAfter(fileName, ".")[0] + "webp"
	thumbnailPath := getThumbnailPath(thumbnailName)
	var cmd *exec.Cmd
	if strings.HasPrefix(contentType, "image") {
		// [0] takes the first frame. usefull for gifs
		cmd = exec.Command(`gm`, `convert`,
			`-size`, fmt.Sprintf(`%dx%d`, size, size), getFilePath(fileName)+"[0]",
			`-resize`, fmt.Sprintf(`%dx%d`, size, size), `+profile`, `"*"`, thumbnailPath)
	} else {
		cmd = exec.Command(`ffmpeg`, `-i`, getFilePath(fileName), `-vframes`, `1`, `-vf`,
			fmt.Sprintf(`scale=w=%d:h=%d:force_original_aspect_ratio=decrease`, size, size),
			thumbnailPath, `-loglevel`, `error`, `-y`)
	}
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"event": "create thumbnail",
			"error": err,
		}).Error("could not run ", string(stdout.Bytes()), string(stderr.Bytes()))
		return thumbnailName, err
	}
	if errStr := string(stderr.Bytes()); errStr != "" {
		log.WithFields(log.Fields{
			"event": "create thumbnail",
			"error": errStr,
		}).Error("stdErr is not empty")
		return thumbnailName, errors.New(errStr)
	}
	return thumbnailName, nil
}

func HandleFile(r io.Reader, thumbnailSize int) (string, string, error) {
	fileName, ct, err := SaveFile(r, viper.GetString("static_path")+filesFolder,
		viper.GetInt64("max_image_size_mb")<<20)
	if err != nil {
		DeleteFile(getFilePath(fileName))
		return "", "", err
	}

	thumbnailName, err := CreateThumbnail(ct, fileName, thumbnailSize)
	if err != nil {
		DeleteFileAndThumbnail(fileName, thumbnailName)
		return "", "", err
	}

	return fileName, thumbnailName, nil
}

// we don't user this function because it take long time and consumes CPU
// convert GIF to WebM
func convertGIF(gifName string) (string, error) {
	gifPath := getFilePath(gifName)
	webmName := strings.SplitAfter(gifName, ".")[0] + "webm"
	var cmd *exec.Cmd
	// lower crf value => better quality
	cmd = exec.Command(`ffmpeg`, `-i`, gifPath, `-c:v`, `libvpx-vp9`, `-crf`,
		`30`, `-b:v`, `0`, getFilePath(webmName), `-loglevel`, `error`, `-y`)
	defer DeleteFile(gifPath)
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		log.WithFields(log.Fields{
			"event": "convert GIF",
			"error": err,
		}).Error("could not run ", string(stdout.Bytes()), string(stderr.Bytes()))
		return webmName, err
	}
	if errStr := string(stderr.Bytes()); errStr != "" {
		log.WithFields(log.Fields{
			"event": "convert GIF",
			"error": errStr,
		}).Error("stdErr is not empty")
		return webmName, errors.New(errStr)
	}
	return webmName, nil
}

func DeleteFileAndThumbnail(file string, thumbnail string) {
	DeleteFile(getFilePath(file))
	DeleteFile(getThumbnailPath(thumbnail))
}

// func HandleFile(r io.Reader, thumbnailSize int) (string, error) {
// 	fileName, err := HandleImage(r, thumbnailSize)
// 	return fileName, err
// }

func getFilePath(file string) string {
	return viper.GetString("static_path") + filesFolder + file
}
func getThumbnailPath(thumbnail string) string {
	return viper.GetString("static_path") + thumbnailsFolder + thumbnail
}
