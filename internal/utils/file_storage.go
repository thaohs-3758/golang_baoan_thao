package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

var (
	ErrEmptyFileName  = errors.New("attachment file name is empty")
	ErrUnsafeFileName = errors.New("attachment file name is unsafe")
	ErrDisallowedMime = errors.New("attachment mime type not allowed")
	ErrPathEscape     = errors.New("attachment path escapes upload directory")
)

var allowedMime = map[string]struct{}{
	"application/pdf": {},
	"image/jpeg":      {},
	"image/png":       {},
}

var unsafeChars = regexp.MustCompile(`[^A-Za-z0-9._-]`)

type FileStorage interface {
	SaveApplicationFile(applicationID string, fh *multipart.FileHeader) (publicURL, sniffedMime string, savedSize int64, err error)
	RemoveApplicationDir(applicationID string) error
	RemoveFile(publicURL string) error
	ListApplicationDirs() ([]TempDirInfo, error)
}

type TempDirInfo struct {
	ApplicationID string
	CreatedAt     time.Time
}

type LocalDiskStorage struct {
	BaseDir      string
	PublicPrefix string
}

func NewLocalDiskStorage(baseDir, publicPrefix string) *LocalDiskStorage {
	return &LocalDiskStorage{BaseDir: baseDir, PublicPrefix: strings.TrimRight(publicPrefix, "/")}
}

func (s *LocalDiskStorage) SaveApplicationFile(applicationID string, fh *multipart.FileHeader) (string, string, int64, error) {
	safeName, err := sanitizeFileName(fh.Filename)
	if err != nil {
		return "", "", 0, err
	}

	src, err := fh.Open()
	if err != nil {
		return "", "", 0, err
	}
	defer src.Close()

	mime, err := mimetype.DetectReader(src)
	if err != nil {
		return "", "", 0, err
	}
	if _, ok := allowedMime[mime.String()]; !ok {
		return "", "", 0, ErrDisallowedMime
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", "", 0, err
	}

	dir := filepath.Join(s.BaseDir, "applications", applicationID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", "", 0, err
	}

	finalName := uniquifyFileName(safeName)
	dst := filepath.Join(dir, finalName)

	absDir, _ := filepath.Abs(dir)
	absDst, _ := filepath.Abs(dst)
	if !strings.HasPrefix(absDst, absDir+string(os.PathSeparator)) {
		return "", "", 0, ErrPathEscape
	}

	tmp := dst + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return "", "", 0, err
	}

	n, err := io.Copy(out, src)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		_ = os.Remove(tmp)
		return "", "", 0, err
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return "", "", 0, err
	}

	publicURL := fmt.Sprintf("%s/applications/%s/%s", s.PublicPrefix, applicationID, finalName)
	return publicURL, mime.String(), n, nil
}

func (s *LocalDiskStorage) RemoveApplicationDir(applicationID string) error {
	if applicationID == "" {
		return nil
	}
	return os.RemoveAll(filepath.Join(s.BaseDir, "applications", applicationID))
}

func (s *LocalDiskStorage) RemoveFile(publicURL string) error {
	if publicURL == "" {
		return nil
	}
	// publicURL is expected to contain the public prefix and a path like "/applications/<id>/<file>"
	prefix := strings.TrimRight(s.PublicPrefix, "/")
	var rel string
	if strings.HasPrefix(publicURL, prefix) {
		rel = strings.TrimPrefix(publicURL, prefix)
	} else {
		// fallback: try to find "/applications/" in the URL
		idx := strings.Index(publicURL, "/applications/")
		if idx == -1 {
			return nil
		}
		rel = publicURL[idx:]
	}
	rel = strings.TrimPrefix(rel, "/")
	// construct filesystem path under BaseDir
	path := filepath.Join(s.BaseDir, filepath.FromSlash(rel))
	// best-effort remove
	_ = os.Remove(path)
	return nil
}

func (s *LocalDiskStorage) ListApplicationDirs() ([]TempDirInfo, error) {
	root := filepath.Join(s.BaseDir, "applications")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	out := make([]TempDirInfo, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, TempDirInfo{
			ApplicationID: entry.Name(),
			CreatedAt:     info.ModTime(),
		})
	}
	return out, nil
}

func sanitizeFileName(raw string) (string, error) {
	base := filepath.Base(raw)
	if base == "" || base == "." || base == "/" {
		return "", ErrEmptyFileName
	}
	if strings.Contains(base, "..") || strings.HasPrefix(base, ".") {
		return "", ErrUnsafeFileName
	}
	cleaned := unsafeChars.ReplaceAllString(base, "_")
	if cleaned == "" {
		return "", ErrEmptyFileName
	}
	return cleaned, nil
}

func uniquifyFileName(name string) string {
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		panic("crypto/rand failure: " + err.Error())
	}
	hexSuffix := hex.EncodeToString(suffix)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	return stem + "-" + hexSuffix + ext
}
