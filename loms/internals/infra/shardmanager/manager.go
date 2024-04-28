package shardmanager

import (
	"errors"
	"fmt"
	"github.com/spaolacci/murmur3"
	"route256.ozon.ru/project/loms/internals/infra/db"
)

const MaxShards = 1000

var (
	ErrShardIndexOutOfRange = errors.New("shard index is out of range")
)

type ShardKey string
type ShardIndex int

type ShardFn func(ShardKey) ShardIndex

type Manager struct {
	fn     ShardFn
	shards []db.Pool
}

func GetShardFn(shardsCnt int) ShardFn {
	return func(key ShardKey) ShardIndex {
		hasher := murmur3.New32()
		defer hasher.Reset()

		_, _ = hasher.Write([]byte(key))
		return ShardIndex(hasher.Sum32() % uint32(shardsCnt))
	}
}

func GenerateUniqId(prevId int64, index ShardIndex) int64 {
	return prevId + MaxShards + int64(index)
}

func New(fn ShardFn, shards []db.Pool) *Manager {
	return &Manager{
		fn:     fn,
		shards: shards,
	}
}

func (m *Manager) GetShardIndexByKey(key ShardKey) ShardIndex {
	return m.fn(key)
}

func (m *Manager) GetShardIndexFromId(id int64) ShardIndex {
	return ShardIndex(id % 1000)
}

func (m *Manager) Get(key ShardKey) (db.Pool, ShardIndex, error) {
	index := m.GetShardIndexByKey(key)
	res, err := m.Pick(index)
	return res, index, err
}

func (m *Manager) Pick(index ShardIndex) (db.Pool, error) {
	if int(index) < len(m.shards) {
		return m.shards[index], nil
	}
	return nil, fmt.Errorf("%w: given index=%d, len=%d", ErrShardIndexOutOfRange, index, len(m.shards))
}

func (m *Manager) GetByOrderId(id int64) (db.Pool, error) {
	index := m.GetShardIndexFromId(id)
	return m.Pick(index)
}
