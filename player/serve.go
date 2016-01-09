package player

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DirConfig struct {
	Path string `json:"path"`
	Name string `json:"name"`
}
type Config struct {
	Port       int         `json:"port"`
	UseMp4     bool        `json:"use_mp4"`
	BufferSize int         `json:"buffer_size"`
	Dirs       []DirConfig `json:"dirs"`
	Ffmpeg     string      `json:"ffmpeg"`
	Ffprobe    string      `json:"ffprobe"`
}

type MediaServer struct {
	Config Config
	srv    *http.Server
}

func (ms *MediaServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, "/video") {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/video")
		ms.handleVideo(w, req)
		return
	}
	if strings.HasPrefix(req.URL.Path, "/info") {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/info")
		ms.handleInfo(w, req)
		return
	}
	if strings.HasPrefix(req.URL.Path, "/dirinfo") {
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/dirinfo")
		ms.handleDirInfo(w, req)
		return
	}
	fmt.Println("root", req)
	if req.URL.Path == "/" {
		http.ServeFile(w, req, "index.html")
		return
	} else if req.URL.Path == "/player.js" {
		http.ServeFile(w, req, "player.js")
		return
	}
	w.Header().Add("Content-type", "text/plain")
	w.WriteHeader(404)
	fmt.Fprintln(w, "Not found")
}

func (ms *MediaServer) Serve() error {
	var err error
	if len(ms.Config.Dirs) == 0 {
		return errors.New("No dirs to serve")
	}
	for i, _ := range ms.Config.Dirs {
		ms.Config.Dirs[i].Path, err = filepath.Abs(ms.Config.Dirs[i].Path)
		if err != nil {
			return err
		}
		fi, err := os.Stat(ms.Config.Dirs[i].Path)
		if err != nil {
			return err
		}
		if !fi.IsDir() {
			return errors.New("Media directory not found")
		}
	}

	if ms.Config.BufferSize == 0 {
		ms.Config.BufferSize = 64 * 1024
	}

	if ms.Config.Port == 0 {
		ms.Config.Port = 7700
	}

	if ms.Config.Ffmpeg == "" {
		ms.Config.Ffmpeg = "ffmpeg"
	}

	if ms.Config.Ffprobe == "" {
		ms.Config.Ffmpeg = "ffprobe"
	}

	ms.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", ms.Config.Port),
		Handler: ms,
	}

	return ms.srv.ListenAndServe()
}
