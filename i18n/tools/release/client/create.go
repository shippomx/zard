package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

func (c *Client) CreateDownloads(path string) error {
	baseUrl := "https://api.bitbucket.org/2.0/repositories/%s/%s/downloads"
	url := fmt.Sprintf(baseUrl, c.Workspace, c.RepoSlug)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	baseName := filepath.Base(path)
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("files", baseName)
	if err != nil {
		return err
	}
	_, err = part.Write(content)
	if err != nil {
		return err
	}
	err = writer.Close()
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("request failed with error: %v,body: %s", err, string(body))
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	return err
}
