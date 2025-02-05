package client

type Client struct {
	Workspace string
	RepoSlug  string
	Token     string
}

func NewClient(workspace, repo_slug, token string) *Client {
	return &Client{
		Workspace: workspace,
		RepoSlug:  repo_slug,
		Token:     token,
	}
}
