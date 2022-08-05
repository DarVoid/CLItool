package CLItool

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

type GitHubResponse struct {
	DefaultBranch string `json:"default_branch"`
}

func VerifyBranchName(username string, reponame string, branchName string) bool {
	response, err := http.Get("https://api.github.com/repos/" + username + "/" + reponame + "/branches" + branchName)
	if err != nil {
		check(err)
	}
	if response.StatusCode == http.StatusNotFound {
		return false
	}
	return true
}

func GetMainBranchName(username string, reponame string) (string, error) {
	response, err := http.Get("https://api.github.com/repos/" + username + "/" + reponame)
	if err != nil {
		return "", err
	}

	data, _ := ioutil.ReadAll(response.Body)
	// bodyStr := string(data)
	var obj GitHubResponse
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return "", err
	}

	return obj.DefaultBranch, nil
}

func InitRepo(path string) error {
	gitBin, _ := exec.LookPath("git")

	cmd := &exec.Cmd{
		Path:   gitBin,
		Args:   []string{gitBin, "init", path},
		Stdout: os.Stdout,
		Stdin:  os.Stdin,
	}

	err := cmd.Run()
	return err
}

func CheckIsGitInstalled() bool {
	binPath, _ := exec.LookPath("git")
	if binPath != "" {
		return true
	}
	return false
}
