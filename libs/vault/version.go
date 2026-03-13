package vault

import "sync/atomic"

// counter is the global key version counter for Vault Transit keys.
var counter int32 = 0

// SetVersion sets the global key version counter.
//
// This is used to track the current version of the Transit key in Vault.
// Uses atomic operations for thread-safety.
func SetVersion(version int32) {
	atomic.StoreInt32(&counter, version)
}

// GetVersion returns the current global key version counter.
//
// This returns the version that was set during application startup or key rotation.
// Uses atomic operations for thread-safety.
func GetVersion() int32 {
	return atomic.LoadInt32(&counter)
}

// UpdateVersion atomically updates the global key version if the current version matches the previous version.
//
// This is used during key rotation to safely update the version counter.
// Returns true if the update was successful, false if the version has already been updated.
func UpdateVersion(version int32, previous int32) bool {
	return atomic.CompareAndSwapInt32(&counter, previous, version)
}
