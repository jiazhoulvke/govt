package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nfnt/resize"
)

var (
	ffmpegPath   string
	ffprobePath  string
	saveDir      string
	saveName     string
	maxWidth     int
	maxHeight    int
	startAt      int
	allowExts    = []string{}
	allowExtsStr string
)

func init() {
	flag.StringVar(&saveDir, "save_dir", "", "save picture at(if null then use video's path)")
	flag.StringVar(&saveName, "save_name", "", "picture name(if null then use video's name)")
	flag.IntVar(&maxWidth, "max_width", 1000, "picture max width")
	flag.IntVar(&maxWidth, "max_height", 1000, "picture max height")
	flag.IntVar(&startAt, "start_at", 10, "start at (second)")
	flag.StringVar(&allowExtsStr, "allow_exts", ".avi,.rmvb,.rm,.asf,.divx,.mpg,.mpeg,.mpe,.wmv,.mp4,.mkv,.vob,.mov", "video format")
}

//Ext get file ext
func Ext(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

//VideoInfo video info
type VideoInfo struct {
	Format struct {
		Duration float64 `json:"duration,string"`
		Size     float64 `json:"size,string"`
	} `json:"format"`
}

//GetVideoDuration get video duration (second)
func GetVideoDuration(filename string) (int, error) {
	cmd := exec.Command(ffprobePath, "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", filename)
	var stdOutput bytes.Buffer
	cmd.Stdout = &stdOutput
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("run ffprobe error:%v", err)
	}
	var videoInfo VideoInfo
	if err := json.Unmarshal(stdOutput.Bytes(), &videoInfo); err != nil {
		return 0, fmt.Errorf("parse json error:%v", err)
	}
	return int(videoInfo.Format.Duration), nil
}

//Screenshot get video screenshot
func Screenshot(filename string, postion int) ([]byte, error) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		return nil, err
	}
	tempFile.Close()
	tempFileName := tempFile.Name()
	cmd := exec.Command(ffmpegPath, "-ss", fmt.Sprint(postion), "-i", filename, "-y", "-r", "1", "-vframes", "1", "-an", "-vcodec", "mjpeg", "-f", "image2", tempFileName)
	var errOutput bytes.Buffer
	cmd.Stderr = &errOutput
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("get screenshot error:%v", err)
	}
	stat, err := os.Stat(tempFileName)
	if err != nil {
		return nil, fmt.Errorf("get temp file stat error:%v", err)
	}
	if stat.Size() == 0 {
		return nil, fmt.Errorf("temp file is null")
	}
	defer os.Remove(tempFileName)
	content, err := ioutil.ReadFile(tempFileName)
	return content, err
}

//CreateVideoThumbnail create video thumbnail
func CreateVideoThumbnail(filename string, saveDir string, saveName string, startAt int) error {
	duration, err := GetVideoDuration(filename)
	if err != nil {
		return err
	}
	var output []byte
	if duration < 10 {
		var err error
		pos := startAt
		if pos > duration {
			pos = duration / 2
		}
		output, err = Screenshot(filename, pos)
		if err != nil {
			return err
		}
	} else { //
		imgs := make([]image.Image, 0, 9)
		step := (duration - startAt) / 9
		pos := startAt

		var wg sync.WaitGroup
		chErr := make(chan error)
		go func() {
			wg.Add(9)
			for i := 0; i < 9; i++ {
				pos += step
				go func(wg *sync.WaitGroup, pos int) {
					content, err := Screenshot(filename, pos)
					if err != nil {
						chErr <- err
						wg.Done()
						return
					}
					img, err := jpeg.Decode(bytes.NewBuffer(content))
					if err != nil {
						chErr <- err
						wg.Done()
						return
					}
					imgs = append(imgs, img)
					wg.Done()
				}(&wg, pos)
			}
			wg.Wait()
			chErr <- nil
		}()
		select {
		case err := <-chErr:
			if err != nil {
				return err
			}
		}

		size := imgs[0].Bounds().Size()
		w, h := size.X, size.Y
		dstImage := image.NewCMYK(image.Rect(0, 0, w*3, h*3))
		for i := 0; i < 9; i++ {
			pX, pY := i%3*w, i/3*h
			draw.Draw(dstImage, imgs[i].Bounds().Add(image.Pt(pX, pY)), imgs[i], image.Pt(0, 0), draw.Src)
		}
		resizedImage := resize.Thumbnail(uint(maxWidth), uint(maxWidth), dstImage, resize.NearestNeighbor) //最快的算法
		b := bytes.NewBufferString("")
		if err := jpeg.Encode(b, resizedImage, nil); err != nil {
			return err
		}
		output = b.Bytes()
	}

	if saveDir == "" {
		saveDir = filepath.Dir(filename)
	}
	if saveName == "" {
		saveName = strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)) + ".jpg"
	}
	fullpath := filepath.Join(saveDir, saveName)
	if err := ioutil.WriteFile(fullpath, output, 0666); err != nil {
		return err
	}
	fmt.Println("generate thumbnail:", fullpath)
	return nil
}

func main() {
	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Println("video path is required")
		os.Exit(1)
	}
	allowExts = strings.Split(allowExtsStr, ",")
	var err error
	ffmpegPath, err = exec.LookPath("ffmpeg")
	if err != nil {
		fmt.Println("ffmpeg not found:", err)
		os.Exit(1)
	}
	ffprobePath, err = exec.LookPath("ffprobe")
	if err != nil {
		fmt.Println("ffmpeg not found:", err)
		os.Exit(1)
	}

	files := make([]string, 0)
	for _, p := range flag.Args() {
		info, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
			}
		}
		if info.IsDir() {
			filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if !IsVideoFile(path) {
					return nil
				}
				if info.IsDir() {
					return nil
				}
				files = append(files, path)
				return nil
			})
		} else {
			files = append(files, p)
		}
	}
	for _, filename := range files {
		err = CreateVideoThumbnail(filename, saveDir, saveName, startAt)
		if err != nil {
			fmt.Println("create thumbnail error:", err)
			os.Exit(1)
		}
	}
}

//IsVideoFile check file type
func IsVideoFile(filename string) bool {
	ext := Ext(filename)
	for _, allowExt := range allowExts {
		if allowExt == ext {
			return true
		}
	}
	return false
}
