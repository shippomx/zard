package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) ListDownload() (map[string]interface{}, error) {
	baseUrl := "https://api.bitbucket.org/2.0/repositories/%s/%s/downloads"

	url := fmt.Sprintf(baseUrl, c.Workspace, c.RepoSlug)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+c.Token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}
	resultMap := make(map[string]interface{})
	err = json.Unmarshal(body, &resultMap)
	if err != nil {
		return nil, err
	}
	return resultMap, nil
}
