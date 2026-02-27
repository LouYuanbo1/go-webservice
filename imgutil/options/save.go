package options

import "github.com/LouYuanbo1/go-webservice/imgutil/config"

type save struct {
	storageDir string
	quality    int
}

func (s *save) GetStorageDir() string {
	return s.storageDir
}

func (s *save) GetQuality() int {
	return s.quality
}

func NewSave() *save {
	return &save{}
}

func NewSaveWithOptions(opts ...SaveOption) *save {
	s := NewSave()
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func NewSaveByConfig(config *config.SaveConfig) *save {
	return &save{
		storageDir: config.StorageDir,
		quality:    config.Quality,
	}
}

func SaveBuilder(cfg *config.SaveConfig, opts ...SaveOption) *save {
	s := NewSaveByConfig(cfg)
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// 链式调用
func (s *save) WithStorageDir(dir string) *save {
	s.storageDir = dir
	return s
}

func (s *save) WithQuality(quality int) *save {
	s.quality = quality
	return s
}

type SaveOption func(*save)

func WithStorageDir(dir string) SaveOption {
	return func(s *save) {
		s.storageDir = dir
	}
}

func WithQuality(quality int) SaveOption {
	return func(s *save) {
		s.quality = quality
	}
}
