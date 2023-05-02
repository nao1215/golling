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
	latestGoVersion = "1.20.4"
	golangPath      = "/usr/local/go"
)

// golangTarballChecksums return key=taraball name , value=sha256 checksum
func golangTarballChecksums() map[string]string {
	return map[string]string{
		"go1.20.4.darwin-amd64.tar.gz":  "242b099b5b9bd9c5d4d25c041216bc75abcdf8e0541aec975eeabcbce61ad47f",
		"go1.20.4.darwin-arm64.tar.gz":  "61bd4f7f2d209e2a6a7ce17787fc5fea52fb11cc9efb3d8471187a8b39ce0dc9",
		"go1.20.4.linux-386.tar.gz":     "5dfa3db9433ef6a2d3803169fb4bd2f4505414881516eb9972d76ab2e22335a7",
		"go1.20.4.linux-amd64.tar.gz":   "698ef3243972a51ddb4028e4a1ac63dc6d60821bf18e59a807e051fee0a385bd",
		"go1.20.4.linux-arm64.tar.gz":   "105889992ee4b1d40c7c108555222ca70ae43fccb42e20fbf1eebb822f5e72c6",
		"go1.20.4.linux-armv6l.tar.gz":  "0b75ca23061a9996840111f5f19092a1bdbc42ec1ae25237ed2eec1c838bd819",
		"go1.20.4.freebsd-386.tar.gz":   "66b1646786304553c5f3208d4b5ed4b3f227293728bfa5c7b5d9a3c5fa1312cb",
		"go1.20.4.freebsd-amd64.tar.gz": "24ee729660372fb408c34a11284daa6e17e43e1db4a5bee2a96b4b6d291edfc5",
		"go1.20.4.linux-ppc64le.tar.gz": "8c6f44b96c2719c90eebabe2dd866f9c39538648f7897a212cac448587e9a408",
		"go1.20.4.linux-s390x.tar.gz":   "57f999a4e605b1dfa4e7e58c7dbae47d370ea240879edba8001ab33c9a963ebf",
	}
}

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
