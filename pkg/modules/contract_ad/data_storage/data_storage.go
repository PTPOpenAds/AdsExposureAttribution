/*
 * @Author: joyesjiang@tencent.com
 * @Date: 2021-01-27 17:18:25
 * @Last Modified by: xiaoyangma
 * @Last Modified time: 2021-01-28 19:14:55
 */

package hashleveldb

import (
	"fmt"
	"time"

	"github.com/golang/glog"

	"github.com/TencentAd/attribution/attribution/pkg/impression/kv/leveldb"
	"github.com/TencentAd/attribution/attribution/pkg/impression/kv/opt"
)

type item struct {
	Time  time.Time
	Value string
}

type HashLevelDB struct {
	hashDbMap map[string]*leveldb.LevelDb
	defaultDB *leveldb.LevelDb
}

func (s *HashLevelDB) Get(key string) (string, error) {
	dbKey := string(key[len(key)-1:])
	if pLevelDB, ok := s.hashDbMap[dbKey]; ok {
		return pLevelDB.Get(key)
	}
	fmt.Println("HashLevelDB-Get Failed, key out range:", key, " ", dbKey)
	glog.Infoln("HashLevelDB-Get Failed, key out range:", key, " ", dbKey)
	return s.defaultDB.Get(key)
}

func (s *HashLevelDB) Set(key string, value string) error {
	dbKey := string(key[len(key)-1:])
	if pLevelDB, ok := s.hashDbMap[dbKey]; ok {
		return pLevelDB.Set(key, value)
	}
	fmt.Println("HashLevelDB-Set Failed, key out range:", key, " ", dbKey)
	glog.Infoln("HashLevelDB-Set Failed, key out range:", key, " ", dbKey)
	return s.defaultDB.Set(key, value)
}

func New(option *opt.Option) (*HashLevelDB, error) {
	var hashLevelDB = HashLevelDB{
		hashDbMap: make(map[string]*leveldb.LevelDb),
	}
	keyList := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
	for _, value := range keyList {
		newOption := *option
		newOption.Address += "_" + value
		pLevelDB, err := leveldb.New(&newOption)
		if err != nil {
			fmt.Printf("NewHashLevelDB failed, value: %s, levelDB_addr: %s\n", value, newOption.Address)
			return nil, err
		}
		hashLevelDB.hashDbMap[value] = pLevelDB
		fmt.Printf("NewHashLevelDB Succ, value: %s, levelDB_addr: %s\n", value, newOption.Address)
		glog.Infof("NewHashLevelDB Succ, value: %s, levelDB_addr: %s\n", value, newOption.Address)
	}
	hashLevelDB.defaultDB, _ = leveldb.New(option)

	return &hashLevelDB, nil
}
