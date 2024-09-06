package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

func main() {
	version := os.Getenv("HUGO_VERSION")
	extended := os.Getenv("HUGO_EXTENDED") != "false"
	repo := "gohugoio/hugo"

	var versionFmt string
	if version == "" || version == "latest" {
		versionFmt = "latest"
	} else {
		// tags/v0.103.1
		versionFmt = "tags/v" + strings.TrimPrefix(version, "v")
	}

	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/releases/%s",
		repo,
		versionFmt,
	)

	body, err := request(apiURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	var m struct {
		Assets []struct {
			Browser_download_url string
		}
	}

	err = json.Unmarshal(body, &m)
	if err != nil {
		fmt.Println("Error with json:", err)
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "darwin" {
		goarch = "universal"
	}
	osarch := fmt.Sprintf("%s-%s", goos, goarch)

	var asset_url string
	for _, asset := range m.Assets {
		url := asset.Browser_download_url
		if strings.Contains(url, osarch) {
			// TODO: refactor
			if extended && strings.Contains(url, "extended") {
				asset_url = url
			} else if !extended && !strings.Contains(url, "extended") {
				asset_url = url
			}
		}
	}

	if asset_url == "" {
		fmt.Printf("Error: Asset not found on release %s for OS %s\n", version, osarch)
		return
	}

	tarball, err := downloadFile(asset_url)
	if err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}

	err = extractFileFromTarball(tarball)
	if err != nil {
		fmt.Println("Error uncompressing file:", err)
		return
	}
}

func downloadFile(url string) (string, error) {
	body, err := request(url)
	if err != nil {
		return "", err
	}

	url_parts := strings.Split(url, "/")
	filename := os.TempDir() + url_parts[len(url_parts)-1]

	err = os.WriteFile(filename, body, 0666)
	if err != nil {
		return "", err
	}

	fmt.Println("Downloaded Hugo archive to", filename)
	return filename, nil
}

func extractFileFromTarball(tarballPath string) error {
	executable_filename := "hugo"

	file, err := os.Open(tarballPath)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %w", err)
	}
	defer file.Close()

	// Create a gzip reader
	gzipReader, err := gzip.NewReader(file)

	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return fmt.Errorf("hugo executable not found in tarball")
		}
		if err != nil {
			return fmt.Errorf("failed to read tarball: %w", err)
		}

		if header.Name == executable_filename {
			outFile, err := os.OpenFile(executable_filename, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				return fmt.Errorf("failed to copy file content: %w", err)
			}

			return nil
		}
	}
}

func request(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error: Non-OK HTTP status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %s", err)
	}

	return body, nil
}
