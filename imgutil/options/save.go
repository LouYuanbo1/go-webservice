package options

type Save struct {
	StorageDir string
	Quality    int
}

type SaveOption func(*Save)

func WithStorageDir(dir string) SaveOption {
	return func(s *Save) {
		s.StorageDir = dir
	}
}

func WithQuality(quality int) SaveOption {
	return func(s *Save) {
		s.Quality = quality
	}
}
