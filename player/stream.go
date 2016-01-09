package player

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
)

type StreamInfo struct {
	Index  int    `json:"index"`
	Type   string `json:"codec_type"`
	Codec  string `json:"codec_name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type FormatInfo struct {
	Filename string `json:"filename"`
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}

type FileInfo struct {
	Format  FormatInfo   `json:"format"`
	Streams []StreamInfo `json:"streams"`
}

func (ms *MediaServer) getFileInfo(path string) (fi FileInfo, err error) {
	cmd := exec.Command(ms.Config.Ffprobe,
		path,
		"-of", "json",
		"-v", "quiet",
		"-show_format",
		"-show_streams")
	so, _ := cmd.StdoutPipe()

	err = cmd.Start()
	if err != nil {
		return
	}

	data, err := ioutil.ReadAll(so)
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &fi)
	return
}

func (ms *MediaServer) handleVideo(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Got connection", req)

	path, found := ms.checkFile(req.URL.Path)
	if !found {
		notFound(w)
		return
	}

	args := []string{}
	q := req.URL.Query()
	ts := q.Get("ts")

	if len(ts) > 0 {
		v, err := strconv.ParseInt(ts, 10, 32)
		if err == nil {
			args = append(args, "-ss", strconv.FormatFloat(float64(v)/100.0, 'f', 2, 32))
		}
	}

	args = append(args,
		"-i", path)

	// Disable subtitles
	args = append(args,
		"-sn")

	if !ms.Config.UseMp4 {
		args = append(args,
			"-f", "webm",
			"-vcodec", "vp8",
			"-b:v", "1000000",
			"-deadline", "realtime",
			"-acodec", "libvorbis")
	} else {
		// frag_duration: microsecond per fragment
		args = append(args,
			"-f", "mp4",
			"-moov_size", "32768",
			"-movflags", "frag_keyframe",
			"-frag_size", "500000",
			"-vcodec", "libx264",
			"-acodec", "mp3")
	}

	args = append(args,
		"-")

	fmt.Println("args:", args)

	cmd := exec.Command(
		ms.Config.Ffmpeg,
		args...)

	so, _ := cmd.StdoutPipe()
	se, _ := cmd.StderrPipe()
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := se.Read(buf)
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Print(string(buf[:n]))
		}
	}()

	w.Header().Add("Content-type", "video/webm")
	w.Header().Add("Cache-control", "max-age=0")
	w.WriteHeader(200)

	buf := make([]byte, ms.Config.BufferSize)
	for {
		n, err := so.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		if n > 0 {
			_, err = w.Write(buf[:n])
			if err != nil {
				fmt.Println("Err writing", err)
				so.Close()
				break
			}
			//fmt.Println(n, "bytes written")
		}
	}

	cmd.Wait()
	fmt.Println("Done")
}

func (ms *MediaServer) handleInfo(w http.ResponseWriter, req *http.Request) {
	path, found := ms.checkFile(req.URL.Path)
	if !found {
		notFound(w)
		return
	}

	fmt.Println(req)
	w.Header().Add("Content-type", "text/json")
	w.WriteHeader(200)
	fi, err := ms.getFileInfo(path)
	if err != nil {
		fmt.Fprint(w, `{"error": true}`, err)
	} else {
		data, _ := json.Marshal(fi)
		w.Write(data)
	}
}
