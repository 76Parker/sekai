package contracts

type ArtifactMetadataV1 struct {
	FileName         string `json:"file_name"`
	TargetLang       string `json:"target_language"`
	SchemaVersion    int    `json:"schema_version"`
	SLOC             int    `json:"sloc"`
	UncompressedSize int64  `json:"uncompressed_size"`
	CompressedSize   int64  `json:"compressed_size"`
}
