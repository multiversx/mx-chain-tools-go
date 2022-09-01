package common

// FileInfo should provide basic information about a file
type FileInfo interface {
	Name() string
	IsDir() bool
}
