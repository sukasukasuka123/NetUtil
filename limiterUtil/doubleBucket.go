package limiterUtil

type DualBucket struct {
	Stable *Bucket
	Burst  *Bucket
}

func NewDualBucket(stableCap, stableRate, burstCap, burstRate float64) *DualBucket {
	return &DualBucket{
		Stable: NewBucket(stableCap, stableRate),
		Burst:  NewBucket(burstCap, burstRate),
	}
}

// TryTake 优先稳定桶，稳定桶不足时再尝试突发桶
// 返回 true 表示成功拿到 token
func (d *DualBucket) TryTake() bool {
	// 先试 Stable
	if d.Stable.TryTake(1.0) {
		return true
	}
	// 再试 Burst
	if d.Burst.TryTake(1.0) {
		return true
	}
	return false
}

// Status 返回当前两个桶的 token 状态
func (d *DualBucket) Status() (stable float64, burst float64) {
	return d.Stable.Tokens(), d.Burst.Tokens()
}
