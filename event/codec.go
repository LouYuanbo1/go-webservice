package event

import "encoding/json"

// Codec 负责在 业务对象 与 字节流 之间转换，不绑定任何具体类型。
// 以 any 为边界：Marshal 接收任意值，Unmarshal 将字节流解码到任意目标。
type Codec interface {
	Marshal(v any) ([]byte, error)
	Unmarshal(data []byte, dst any) error
}

var _ Codec = JSONCodec{} // 确保 JSONCodec 实现了 Codec 接口

// JSONCodec 基于 encoding/json 的默认编解码器。
type JSONCodec struct{}

// Marshal 序列化 v 为 JSON。
func (JSONCodec) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Unmarshal 将 JSON 数据解码到 dst（通常为 *T）。
func (JSONCodec) Unmarshal(data []byte, dst any) error {
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, dst)
}
