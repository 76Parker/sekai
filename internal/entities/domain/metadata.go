package domain

import "unicode/utf8"

var availableLangs = map[string]struct{}{
	"go": {},
}

type ArtifactMetadata struct {
	fileName         string
	targetLang       string
	schemaVersion    int
	sloc             int
	uncompressedSize int64
	compressedSize   int64
}

func (m ArtifactMetadata) FileName() string {
	return m.fileName
}

func (m ArtifactMetadata) TargetLang() string {
	return m.targetLang
}

func (m ArtifactMetadata) SchemaVersion() int {
	return m.schemaVersion
}

func (m ArtifactMetadata) SLOC() int {
	return m.sloc
}

func (m ArtifactMetadata) UncompressedSize() int64 {
	return m.uncompressedSize
}

func (m ArtifactMetadata) CompressedSize() int64 {
	return m.compressedSize
}

func NewArtifactMetadata(
	fileName, targetLang string,
	schemaVersion, sloc int,
	uncompressedSize, compressedSize int64,
) (ArtifactMetadata, error) {
	if _, ok := availableLangs[targetLang]; !ok {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "unsupported language for scanning")
	}
	if sloc <= 1 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "sloc must be greater than 1")
	}
	if uncompressedSize <= 0 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "uncompressedSize must be greater than 0")
	}
	if compressedSize <= 0 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "compressedSize must be greater than 0")
	}

	if utf8.RuneCountInString(fileName) < 1 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "fileName must be a valid UTF-8 string")
	}
	if utf8.RuneCountInString(fileName) > 30 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "fileName cannot be longer than 30 characters")
	}
	if utf8.RuneCountInString(targetLang) < 1 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "targetLang must be a valid UTF-8 string")
	}
	if utf8.RuneCountInString(targetLang) > 10 {
		return ArtifactMetadata{}, NewError("ArtifactMetadata", "targetLang cannot be longer than 10 characters")
	}
	return ArtifactMetadata{
		fileName:         fileName,
		targetLang:       targetLang,
		schemaVersion:    schemaVersion,
		sloc:             sloc,
		uncompressedSize: uncompressedSize,
		compressedSize:   compressedSize,
	}, nil
}

func ReconstituteArtifactMetadata(
	fileName, targetLang string,
	schemaVersion, sloc int,
	uncompressedSize, compressedSize int64,
) ArtifactMetadata {
	return ArtifactMetadata{
		fileName:         fileName,
		targetLang:       targetLang,
		schemaVersion:    schemaVersion,
		sloc:             sloc,
		uncompressedSize: uncompressedSize,
		compressedSize:   compressedSize,
	}
}
