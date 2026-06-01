package scan

type StartCommand struct {
	UserID      int64
	ArchivePath string
	SizeBytes   int64
	ApiType     int

	Labels     []string
	EnableSAST bool
}
