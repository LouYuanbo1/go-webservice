package config

type ImgUtilConfig struct {
	DefaultWidth      int    `mapstructure:"default_width"`       // 默认处理宽度
	DefaultHeight     int    `mapstructure:"default_height"`      // 默认处理高度
	DefaultQuality    int    `mapstructure:"default_quality"`     // 质量 (1-100) 适用于JPEG格式,可兼容PNG格式
	DefaultStorageDir string `mapstructure:"default_storage_dir"` // 存储目录
}
