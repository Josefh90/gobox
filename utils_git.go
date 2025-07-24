package gobox_utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ProgressHandler func(progress int, message string)

// GitFetchOptions defines options for fetching a GitHub repository
type GitFetchOptions struct {
	RepoURL         string
	Branch          string
	DestinationPath string
	Folders         []string
	PullOnlyFolders bool
	Overwrite       bool
	OnProgress      ProgressHandler
}

func defaultProgressHandler(progress int, message string) {
	fmt.Printf("[%d%%] %s\n", progress, message)
}

// FetchRepoFolders fetches and downloads specified folders from a GitHub repo
func FetchRepoFolders(opts GitFetchOptions) error {
	if opts.OnProgress == nil {
		opts.OnProgress = defaultProgressHandler
	}

	opts.OnProgress(1, "Starting fetch...")

	if opts.Branch == "" {
		opts.Branch = "main"
	}

	parts := strings.Split(strings.TrimPrefix(opts.RepoURL, "https://github.com/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid repo URL: %s", opts.RepoURL)
	}
	user, repo := parts[0], parts[1]

	opts.OnProgress(5, "Fetching folder list...")

	if len(opts.Folders) == 0 {
		allFolders, err := getTopLevelFolders(user, repo, opts.Branch, opts.PullOnlyFolders)
		if err != nil {
			return err
		}
		opts.Folders = allFolders
	}

	total := len(opts.Folders)
	for i, folder := range opts.Folders {
		msg := fmt.Sprintf("Downloading folder: %s", folder)
		pct := calcPercent(i, total)
		opts.OnProgress(pct, msg)

		err := downloadFolder(folder, user, repo, opts.Branch, opts.DestinationPath, opts.Overwrite)
		if err != nil {
			return err
		}
	}

	opts.OnProgress(100, "Download complete")
	return nil
}

// getTopLevelFolders queries GitHub for top-level folders of a repo
func getTopLevelFolders(user, repo, branch string, onlyDirs bool) ([]string, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/?ref=%s", user, repo, branch)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var contents []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&contents); err != nil {
		return nil, err
	}

	var results []string
	for _, item := range contents {
		if onlyDirs && item.Type != "dir" {
			continue
		}
		results = append(results, item.Name)
	}
	return results, nil
}

// downloadFolder downloads all files in a folder from GitHub
func downloadFolder(folder, user, repo, branch, destPath string, overwrite bool) error {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", user, repo, folder, branch)

	resp, err := http.Get(apiURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var files []struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return err
	}

	for _, file := range files {
		if file.Type != "file" {
			continue
		}

		targetPath := filepath.Join(destPath, file.Path)

		if !overwrite {
			if _, err := os.Stat(targetPath); err == nil {
				continue
			}
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return err
		}

		out, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer out.Close()

		res, err := http.Get(file.DownloadURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if _, err = io.Copy(out, res.Body); err != nil {
			return err
		}
	}
	return nil
}
