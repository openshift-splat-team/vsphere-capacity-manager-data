package generation

// Asset Defines an asset that will be persisted to disk
type Asset struct {
	// Asset content to write to disk
	Asset any

	// FileName name of the file asset to persist to disk
	FileName string
}
