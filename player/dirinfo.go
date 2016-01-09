package player

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type DirInfo struct {
	Files []string `json:"files"`
	Dirs  []string `json:"dirs"`
}

func isVideo(filename string) bool {
	filename = strings.ToLower(filename)
	return strings.HasSuffix(filename, ".avi") ||
		strings.HasSuffix(filename, ".asf") ||
		strings.HasSuffix(filename, ".flv") ||
		strings.HasSuffix(filename, ".mp4") ||
		strings.HasSuffix(filename, ".m4v") ||
		strings.HasSuffix(filename, ".mpg") ||
		strings.HasSuffix(filename, ".rm") ||
		strings.HasSuffix(filename, ".wmv") ||
		strings.HasSuffix(filename, ".mkv")
}

func (ms *MediaServer) handleDirInfo(w http.ResponseWriter, req *http.Request) {
	di := &DirInfo{
		Files: []string{},
		Dirs:  []string{},
	}

	if len(ms.Config.Dirs) > 1 && (req.URL.Path == "" || req.URL.Path == "/") {
		for _, d := range ms.Config.Dirs {
			di.Dirs = append(di.Dirs, d.Name)
		}
	} else {
		path, found := ms.checkFile(req.URL.Path)
		if !found {
			notFound(w)
			return
		}

		d, err := os.Open(path)
		if err != nil {
			notFound(w)
			return
		}
		fis, err := d.Readdir(0)
		if err != nil {
			notFound(w)
			return
		}

		for _, fi := range fis {
			if fi.IsDir() {
				di.Dirs = append(di.Dirs, fi.Name())
			} else if isVideo(fi.Name()) {
				di.Files = append(di.Files, fi.Name())
			}
		}
	}

	data, _ := json.Marshal(di)
	w.Header().Add("Content-type", "text/json")
	w.WriteHeader(200)
	w.Write(data)
}
