package inspector

import (
	"archive/zip"
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sekai/internal/usecase/scan"
	"strings"
)

const (
	GoManifest     = "go.mod"
	JavaManifest   = "pom.xml"
	PythonManifest = "requirements.txt"
)

var manifestLangs = map[string]string{
	"go.mod":           "go",
	"pom.xml":          "java",
	"requirements.txt": "python",
	"package.json":     "js",
}

var langs map[string]string = map[string]string{
	".go":   "go",
	".py":   "python",
	".js":   "js",
	".ts":   "ts",
	".tsx":  "tsx",
	".php":  "php",
	".java": "java",
}

type ZipInspector struct {
	maxUncompressedSize int64
}

func NewZipInspector(maxUncompressedSize int64) *ZipInspector {
	return &ZipInspector{maxUncompressedSize: maxUncompressedSize}
}

func (z *ZipInspector) Inspect(ctx context.Context, archivePath string, sizeBytes int64) (scan.InspectionInfo, error) {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return scan.InspectionInfo{}, err
	}
	defer zipReader.Close()

	archiveDir := filepath.Dir(archivePath)
	manifestDir, err := os.MkdirTemp(archiveDir, "manifests")
	if err != nil {
		return scan.InspectionInfo{}, err
	}

	var uncompressedSize int64
	var dockerfilePath string

	langByLines := make(map[string]int)
	langByFileCount := make(map[string]int)
	manifestLangsDetected := make(map[string]struct{})

	manifestPaths := make(map[string]string)
	fileCount := 0

	hash := sha256.New()

	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fileCount++
		uncompressedSize += file.FileInfo().Size()
		if uncompressedSize > z.maxUncompressedSize {
			return scan.InspectionInfo{}, errors.New("uncompressed size exceeds max limit")
		}
		if manifestLang, ok := manifestLangs[path.Base(file.Name)]; ok {
			manifestPath, err := extractZipFileToDir(file, manifestDir)
			if err != nil {
				return scan.InspectionInfo{}, err
			}
			manifestPaths[manifestLang] = manifestPath
			manifestLangsDetected[manifestLang] = struct{}{}
		}

		if path.Base(file.Name) == "Dockerfile" {
			dockerPath, err := extractZipFileToDir(file, manifestDir)
			if err != nil {
				return scan.InspectionInfo{}, err
			}
			dockerfilePath = dockerPath
		}

		ext := path.Ext(file.Name)
		lang, ok := langs[ext]
		if !ok {
			continue
		}
		rc, err := file.Open()
		if err != nil {
			continue
		}

		fileHash := sha256.New()
		if _, err := io.Copy(fileHash, rc); err != nil {
			rc.Close()
			continue
		}
		rc.Close()
		fileHashBytes := fileHash.Sum(nil)
		hash.Write(fileHashBytes)

		lines, err := countZipFileLines(file)
		if err != nil {
			return scan.InspectionInfo{}, err
		}
		langByLines[lang] += int(lines)
		langByFileCount[lang]++
	}
	targetLang := score(langByLines, langByFileCount, manifestLangsDetected)
	inspectionInfo := scan.InspectionInfo{
		FileName:           filepath.Base(archivePath),
		TargetLanguage:     targetLang,
		ManifestPath:       manifestPaths[targetLang],
		DockerfilePath:     dockerfilePath,
		TargetLanguageSLOC: langByLines[targetLang],
		CompressedSize:     sizeBytes,
		UncompressedSize:   uncompressedSize,
		SHA256:             hex.EncodeToString(hash.Sum(nil)),
	}
	return inspectionInfo, nil
}

func countLines(r io.Reader) (int64, error) {
	reader := bufio.NewReader(r)

	var lines int64
	var hasData bool

	for {
		part, err := reader.ReadString('\n')
		if len(part) > 0 {
			hasData = true
		}

		if err == nil {
			lines++
			continue
		}

		if errors.Is(err, io.EOF) {
			if hasData && len(part) > 0 {
				lines++
			}
			break
		}

		return 0, err
	}

	return lines, nil
}
func countZipFileLines(file *zip.File) (int64, error) {

	rc, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()
	return countLines(rc)

}

func extractZipFileToDir(file *zip.File, dir string) (string, error) {
	if file.FileInfo().IsDir() {
		return "", nil
	}
	// ZIP paths use "/" always.
	cleanName := path.Clean(file.Name)

	if cleanName == "." ||
		cleanName == ".." ||
		path.IsAbs(cleanName) ||
		strings.HasPrefix(cleanName, "../") {
		return "", fmt.Errorf("unsafe zip path: %s", file.Name)
	}

	dstPath := filepath.Join(dir, filepath.FromSlash(cleanName))

	// Защита от path traversal после filepath.Join.
	dstClean := filepath.Clean(dstPath)
	dirClean := filepath.Clean(dir)

	if !strings.HasPrefix(dstClean, dirClean+string(os.PathSeparator)) && dstClean != dirClean {
		return "", fmt.Errorf("zip path escapes target dir: %s", file.Name)
	}

	if err := os.MkdirAll(filepath.Dir(dstClean), 0o755); err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstClean, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return dstClean, nil
}
