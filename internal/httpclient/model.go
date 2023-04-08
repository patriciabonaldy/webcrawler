package httpclient

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kennygrant/sanitize"
)

// Response is the representation of a HTTP response made by a Collector
type Response struct {
	URL string
	// StatusCode is the status code of the Response
	StatusCode int
	// Body is the content of the Response
	Body []byte
	// Headers contains the Response's HTTP headers
	Headers *http.Header
}

// Save writes response body to disk
func (r *Response) Save() error {
	folderName := strings.ReplaceAll(r.URL, "https://", "")

	if strings.HasSuffix(folderName, "/") {
		folderName = folderName[:len(folderName)-1]
	}

	err := os.MkdirAll(folderName, 0750)
	if err != nil {
		return err
	}

	fileName := getFileName(strings.TrimPrefix(r.URL, "/"))

	return os.WriteFile(fmt.Sprintf("%s/%s", folderName, fileName), r.Body, 0644)
}

func getFileName(fileName string) string {
	ext := filepath.Ext(fileName)

	cleanExt := sanitize.BaseName(ext)
	if cleanExt == "" {
		cleanExt = ".html"
	}

	return strings.Replace(fmt.Sprintf(
		"%s.%s",
		sanitize.BaseName(fileName[:len(fileName)-len(ext)]),
		cleanExt[1:],
	), "-", "_", -1)
}
