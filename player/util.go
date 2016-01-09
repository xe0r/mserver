package player

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func splitAll(path string) []string {
	tmp := strings.Split(path, string(filepath.Separator))
	res := []string{}
	for _, s := range tmp {
		if len(s) > 0 {
			res = append(res, s)
		}
	}
	return res
}

func (ms *MediaServer) checkFile(path string) (fullpath string, found bool) {
	found = false

	path = filepath.Clean(path)

	var dir string
	if len(ms.Config.Dirs) == 1 {
		dir = ms.Config.Dirs[0].Path
	} else {
		elems := splitAll(path)
		if len(elems) == 0 {
			return
		}
		for _, d := range ms.Config.Dirs {
			if elems[0] == d.Name {
				dir = d.Path
				break
			}
		}
		path = filepath.Join(elems[1:]...)
	}
	if dir == "" {
		return
	}

	path = filepath.Join(dir, path)
	path, _ = filepath.Abs(path)

	if !strings.HasPrefix(path, dir) {
		return
	}

	_, err := os.Stat(path)
	if err != nil {
		return
	}

	fullpath = path
	found = true
	return
}

func notFound(w http.ResponseWriter) {
	w.Header().Add("Content-type", "text-plain")
	w.WriteHeader(404)
	fmt.Println(w, "Not found")
}
