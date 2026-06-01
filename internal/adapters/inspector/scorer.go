package inspector

var (
	FileScore     = 10
	LineScore     = 1
	ManifestScore = 10_000
)

func score(
	langByLines map[string]int,
	langByFileCount map[string]int,
	manifestLangsDetected map[string]struct{},
) string {
	scores := make(map[string]int)

	for lang, lines := range langByLines {
		scores[lang] += lines * LineScore
	}

	for lang, files := range langByFileCount {
		scores[lang] += files * FileScore
	}

	for lang := range manifestLangsDetected {
		scores[lang] += ManifestScore
	}

	var targetLang string
	var bestScore int
	var bestHasManifest bool
	var bestLines int
	var bestFiles int

	for lang, currentScore := range scores {
		_, hasManifest := manifestLangsDetected[lang]
		lines := langByLines[lang]
		files := langByFileCount[lang]

		if targetLang == "" ||
			currentScore > bestScore ||
			currentScore == bestScore && hasManifest && !bestHasManifest ||
			currentScore == bestScore && hasManifest == bestHasManifest && lines > bestLines ||
			currentScore == bestScore && hasManifest == bestHasManifest && lines == bestLines && files > bestFiles {
			targetLang = lang
			bestScore = currentScore
			bestHasManifest = hasManifest
			bestLines = lines
			bestFiles = files
		}
	}
	return targetLang
}
