package imgutil

type Filter int

const (
	Lanczos Filter = iota
	CatmullRom
	MitchellNetravali
	Linear
	Box
	NearestNeighbor
)

type Config struct {
	Transform TransformConfig `mapstructure:"transform"` // 转换配置
	Save      SaveConfig      `mapstructure:"save"`      // 保存配置
}

type TransformConfig struct {
	Width  int    `mapstructure:"width"`  // 处理宽度
	Height int    `mapstructure:"height"` // 处理高度
	Filter Filter `mapstructure:"filter"` // 转换滤波器
}

type SaveConfig struct {
	Quality    int    `mapstructure:"quality"`     // 质量 (1-100) 适用于JPEG格式,可兼容PNG格式
	StorageDir string `mapstructure:"storage_dir"` // 保存目录
}
