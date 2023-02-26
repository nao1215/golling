package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/nao1215/gorky/file"
	"github.com/spf13/cobra"
)

const (
	latestGoVersion = "1.20.1"
	golangPath      = "/usr/local/go"
)

type updateOption struct {
	force bool
}

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "golling update (or install) golang to the latest version.",
		Long: `golling updates golang to the latest version in /usr/local/go.

golling start update if golang is not up to date. By default, golling
checks /usr/local/go. If golang is not on the system, golling install the
latest golang in /usr/local/go. `,
		Example: "  sudo golling update",
		RunE:    update,
	}

	cmd.Flags().BoolP("force", "f", false, "force update")

	return cmd
}

// isWindows check whether runtime is windosw or not.
func isWindows() bool {
	return runtime.GOOS == "windows"
}

var errNoNeedToUpdateGo = errors.New("no need to update golang")

// update update /usr/local/go.
func update(cmd *cobra.Command, args []string) error {
	opt, err := newUpdateOption(cmd)
	if err != nil {
		return fmt.Errorf("%s: %w", "failed to parse option", err)
	}

	root, err := hasRootPrivirage()
	if err != nil {
		return fmt.Errorf("%s: %w", "can not get user information", err)
	}
	if !root {
		return errors.New("you must have root privileges")
	}

	if !opt.force {
		if err := compareCurrentVerAndLatestVer(); err != nil {
			if errors.Is(err, errNoNeedToUpdateGo) {
				fmt.Printf("current go version is equal to or newer than version %s\n", latestGoVersion)
				return nil
			}
			return fmt.Errorf("%s: %w", "go version check error", err)
		}
	}

	fmt.Printf("download %s at current directory\n", tarballName())
	if err := fetchGolangTarball(tarballName()); err != nil {
		return fmt.Errorf("%s %s: %w", "can not download", tarballName(), err)
	}

	if err := compareChecksum(tarballName()); err != nil {
		return fmt.Errorf("%s: %w", "failed to compare checksum", err)
	}

	fmt.Printf("backup original %s as %s\n", golangPath, golangBackupPath())
	if err := renameIfDirExists(golangPath, golangBackupPath()); err != nil {
		return fmt.Errorf("%s%s: %w", "failed to backup old ", golangPath, err)
	}

	fmt.Printf("start extract %s at %s\n", tarballName(), golangPath)
	if err := extractTarball(tarballName(), "/usr/local"); err != nil {
		fmt.Printf("failed to extract %s\n", tarballName())
		fmt.Printf("start restore %s from backup\n", golangPath)
		if err := recovery(golangPath, golangBackupPath()); err != nil {
			return fmt.Errorf("%s: %w", "!!! failed to restore !!! golang may not be available", err)
		}
		return errors.New("success to restore from backup")
	}

	fmt.Printf("delete backup (%s)\n", golangBackupPath())
	if err := os.RemoveAll(golangBackupPath()); err != nil {
		return fmt.Errorf("%s %s: %w", "failed to delete", golangBackupPath(), err)
	}

	fmt.Printf("delete %s\n", tarballName())
	if err := os.RemoveAll(tarballName()); err != nil {
		return fmt.Errorf("%s %s: %w", "failed to delete", tarballName(), err)
	}

	fmt.Println("")
	fmt.Printf("success to update golang (version %s)\n", latestGoVersion)
	return nil
}

func golangBackupPath() string {
	return golangPath + ".backup"
}

func newUpdateOption(cmd *cobra.Command) (*updateOption, error) {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return nil, err
	}
	return &updateOption{
		force: force,
	}, nil
}

func compareCurrentVerAndLatestVer() error {
	if _, err := exec.LookPath("/usr/local/go/bin/go"); err != nil {
		return nil // this system does not install golang. So, install it.
	}

	currentVer, err := getCurrentGoSemanticVer()
	if err != nil {
		return err
	}

	current, err := semver.NewVersion(currentVer)
	if err != nil {
		return err
	}

	latest, err := semver.NewVersion(latestGoVersion)
	if err != nil {
		return err
	}

	fmt.Printf("current=%s, latest=%s\n", currentVer, latestGoVersion)
	if current.Equal(latest) || current.GreaterThan(latest) {
		return errNoNeedToUpdateGo
	}
	return nil
}

func getCurrentGoSemanticVer() (string, error) {
	cmd := exec.Command("/usr/local/go/bin/go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// extract version (e.g. go1.2.1)
	verStr := strings.TrimSpace(string(bytes.Split(output, []byte(" "))[2]))
	return strings.Replace(verStr, "go", "", 1), nil
}

func hasRootPrivirage() (bool, error) {
	u, err := user.Current()
	if err != nil {
		return false, err
	}
	if u.Uid == "0" {
		return true, nil
	}
	return false, nil
}

func tarballName() string {
	return fmt.Sprintf("go%s.%s-%s.tar.gz", latestGoVersion, runtime.GOOS, runtime.GOARCH)
}

// golangTarballChecksums return key=taraball name , value=sha256 checksum
func golangTarballChecksums() map[string]string {
	return map[string]string{
		"go1.20.1.darwin-amd64.tar.gz":  "a300a45e801ab459f3008aae5bb9efbe9a6de9bcd12388f5ca9bbd14f70236de",
		"go1.20.1.darwin-arm64.tar.gz":  "f1a8e06c7f1ba1c008313577f3f58132eb166a41ceb95ce6e9af30bc5a3efca4",
		"go1.20.1.linux-386.tar.gz":     "3a7345036ebd92455b653e4b4f6aaf4f7e1f91f4ced33b23d7059159cec5f4d7",
		"go1.20.1.linux-amd64.tar.gz":   "000a5b1fca4f75895f78befeb2eecf10bfff3c428597f3f1e69133b63b911b02",
		"go1.20.1.linux-arm64.tar.gz":   "5e5e2926733595e6f3c5b5ad1089afac11c1490351855e87849d0e7702b1ec2e",
		"go1.20.1.linux-armv6l.tar.gz":  "e4edc05558ab3657ba3dddb909209463cee38df9c1996893dd08cde274915003",
		"go1.20.1.freebsd-386.tar.gz":   "57d80349dc4fbf692f8cd85a5971f97841aedafcf211e367e59d3ae812292660",
		"go1.20.1.freebsd-amd64.tar.gz": "6e124d54d5850a15fdb15754f782986f06af23c5ddb6690849417b9c74f05f98",
		"go1.20.1.linux-ppc64le.tar.gz": "85cfd4b89b48c94030783b6e9e619e35557862358b846064636361421d0b0c52",
		"go1.20.1.linux-s390x.tar.gz":   "ba3a14381ed4538216dec3ea72b35731750597edd851cece1eb120edf7d60149",
	}
}

// fetchGolangTarball download latest golang
func fetchGolangTarball(tarballName string) error {
	url := fmt.Sprintf("https://go.dev/dl/%s", tarballName)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	tarball, err := os.Create(tarballName)
	if err != nil {
		return err
	}
	defer tarball.Close()

	progress := NewProgress(resp.ContentLength)
	defer progress.Finish()

	_, err = io.Copy(tarball, io.TeeReader(resp.Body, progress))
	if err != nil {
		return err
	}
	return nil
}

// compareChecksum compare the "sha256 checksum of the downloaded tarball" with the "expected value"
func compareChecksum(tarballName string) error {
	checksumMap := golangTarballChecksums()
	expectSha256, ok := checksumMap[tarballName]
	if !ok {
		return errors.New("checksum (expected value) of downloaded go file not found")
	}

	data, err := os.ReadFile(tarballName)
	if err != nil {
		return err
	}
	sha256checksum := sha256.Sum256(data)
	gotSha256 := fmt.Sprintf("%x", sha256checksum)

	fmt.Println("[compare sha256 checksum]")
	fmt.Printf(" expect: %s\n", expectSha256)
	fmt.Printf(" got   : %s\n", gotSha256)
	fmt.Println("")

	if expectSha256 != gotSha256 {
		return errors.New("sha256 checksum does not match")
	}
	return nil
}

// renameOldGoDir rename /usr/local/go to /usr/local/go.backup
func renameIfDirExists(oldDir, newDir string) error {
	if file.IsDir(oldDir) {
		if err := os.Rename(oldDir, newDir); err != nil {
			return err
		}
	}
	return nil
}

// extractTarball extract tarball
func extractTarball(tarballPath, targetPath string) error {
	file, err := os.Open(tarballPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // end of extract
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetPath, header.Name)
		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}

		createFile := func() error {
			file, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
			return nil
		}
		if err := createFile(); err != nil {
			return err
		}
	}
	return nil
}

// recovery restore /usr/local/go from backup if update fails
func recovery(targetPath, backupPath string) error {
	if file.IsDir(targetPath) {
		if err := os.RemoveAll(targetPath); err != nil {
			return err
		}
	}

	if err := renameIfDirExists(backupPath, targetPath); err != nil {
		return err
	}
	return nil
}

// Progress is download tarball prgoress
type Progress struct {
	// Total is total byte
	Total int64
	// Current is downloaded byte
	Current int64
}

func NewProgress(total int64) *Progress {
	return &Progress{
		Total:   total,
		Current: 0,
	}
}

func (p *Progress) Write(data []byte) (n int, err error) {
	n = len(data)
	p.Current += int64(n)
	p.Show()
	return
}

func (p *Progress) Show() {
	if p.Total == 0 {
		return
	}
	fmt.Printf("\rDownloading...%d/%d kB (%d%%)", p.Current/1000, p.Total/1000, (p.Current*100)/p.Total)
}

func (p *Progress) Finish() {
	fmt.Println()
}
