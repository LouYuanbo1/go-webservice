package config

type ImgUtilConfig struct {
	Width      int    `mapstructure:"width"`       // 默认处理宽度
	Height     int    `mapstructure:"height"`      // 默认处理高度
	Quality    int    `mapstructure:"quality"`     // 质量 (1-100) 适用于JPEG格式,可兼容PNG格式
	StorageDir string `mapstructure:"storage_dir"` // 存储目录
}
