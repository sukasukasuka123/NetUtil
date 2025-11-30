package idutil

import (
	"NetUtil/hashutil"
	"encoding/json"
	"errors"
	"fmt"
)

type ShardManager struct {
	FieldName string // 根据哪个字段分片
	ShardNum  int    // 分成多少片
}

func NewShardManager(field string, num int) *ShardManager {
	if num <= 0 {
		num = 1
	}
	return &ShardManager{
		FieldName: field,
		ShardNum:  num,
	}
}

// 输入 JSON 数据，根据字段计算属于哪一个分片
func (m *ShardManager) GetShardIndex(jsonBytes []byte) (int, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return -1, err
	}

	value, ok := obj[m.FieldName]
	if !ok {
		return -1, errors.New("field not found in JSON")
	}

	var key []byte

	switch v := value.(type) {
	case string:
		key = []byte(v)
	case float64:
		key = []byte(fmt.Sprintf("%f", v))
	default:
		key = []byte(fmt.Sprintf("%v", v))
	}

	return hashutil.FNVHashIndex(key, m.ShardNum), nil
}
