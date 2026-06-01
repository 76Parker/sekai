package scan

type InspectionInfo struct {
	FileName           string
	TargetLanguage     string
	ManifestPath       string
	DockerfilePath     string
	TargetLanguageSLOC int
	CompressedSize     int64
	UncompressedSize   int64
	SHA256             string
}
