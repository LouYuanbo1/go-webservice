package breaker

import (
	"math/rand/v2"
	"sync"
)

// ========== 概率判断器 ==========

type proba struct {
	r  *rand.Rand
	mu sync.Mutex
}

func newProba() *proba {
	return &proba{
		// v2 版本：无需手动种子，自动安全初始化
		r: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
}

// TrueOnProba 按概率 prob 返回 true
func (p *proba) TrueOnProba(prob float64) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	// v2 提供更高精度、更高性能的浮点随机数
	return p.r.Float64() < prob
}
