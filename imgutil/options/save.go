package options

type Save struct {
	storageDir string
	quality    int
}

func NewSave() *Save {
	return &Save{}
}

func (s *Save) GetStorageDir() string {
	return s.storageDir
}

func (s *Save) GetQuality() int {
	return s.quality
}

// 链式调用
func (s *Save) WithStorageDir(dir string) *Save {
	s.storageDir = dir
	return s
}

func (s *Save) WithQuality(quality int) *Save {
	s.quality = quality
	return s
}

type SaveOption func(*Save)

func WithStorageDirOption(dir string) SaveOption {
	return func(s *Save) {
		s.storageDir = dir
	}
}

func WithQualityOption(quality int) SaveOption {
	return func(s *Save) {
		s.quality = quality
	}
}

func NewSaveWithOptions(opts ...SaveOption) *Save {
	s := NewSave()
	for _, opt := range opts {
		opt(s)
	}
	return s
}
