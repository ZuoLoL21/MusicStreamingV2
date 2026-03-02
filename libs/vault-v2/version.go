package vault_v2

import "sync/atomic"

var counter int32 = 0

func SetVersion(version int32) {
	atomic.StoreInt32(&counter, version)
}

func GetVersion() int32 {
	return atomic.LoadInt32(&counter)
}

func UpdateVersion(version int32, previous int32) bool {
	return atomic.CompareAndSwapInt32(&counter, previous, version)
}
