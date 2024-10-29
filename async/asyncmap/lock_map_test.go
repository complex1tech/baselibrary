// Copyright 2023 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package asyncmap

import (
	"testing"
	"time"
	"unsafe"

	"github.com/basecomplextech/baselibrary/async"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLockMap__should_lock_key(t *testing.T) {
	m := newLockMap[int]()
	key := 123
	value := 0

	lock := m.Get(key)
	defer lock.Free()

	<-lock.Lock()

	done := make(chan struct{})
	go func() {
		defer close(done)

		lock := m.Get(key)
		defer lock.Free()

		<-lock.Lock()
		defer lock.Unlock()

		value = 3
	}()

	time.Sleep(10 * time.Millisecond)
	value = 2
	lock.Unlock()

	<-done
	assert.Equal(t, 3, value)
}

func TestLockMap__should_retain_key_lock(t *testing.T) {
	m := newLockMap[int]()
	key := 123

	lock := m.Get(key).(*keyLock[int])
	defer lock.Free()

	shard := m.shard(key)
	item, ok := shard.getNoRetain(key)
	require.True(t, ok)
	assert.Same(t, lock.item, item)
	assert.Equal(t, int32(1), item.refs)
}

func TestLockMap__should_retain_key_lock_when_already_locked(t *testing.T) {
	m := newLockMap[int]()
	key := 123

	lock := m.Get(key)
	defer lock.Free()

	<-lock.Lock()
	defer lock.Unlock()

	go func() {
		lock := m.Get(key)
		defer lock.Free()

		<-lock.Lock()
		lock.Unlock()
	}()

	time.Sleep(10 * time.Millisecond)
	shard := m.shard(key)

	item, ok := shard.getNoRetain(key)
	require.True(t, ok)
	assert.Equal(t, int32(2), item.refs)
}

// Lock

func TestLockMap_Lock__should_acquire_locked_key(t *testing.T) {
	m := newLockMap[int]()
	key := 123
	ctx := async.NoContext()

	lock, st := m.Lock(ctx, key)
	if !st.OK() {
		t.Fatal(st)
	}
	lock.Free()

	shard := m.shard(key)
	_, ok := shard.getNoRetain(key)
	assert.False(t, ok)
}

// Free

func TestKeyLock_Free__should_release_delete_key_lock(t *testing.T) {
	m := newLockMap[int]()
	key := 123

	lock := m.Get(key)
	lock.Free()

	shard := m.shard(key)

	_, ok := shard.getNoRetain(key)
	assert.False(t, ok)
}

// Shard

func TestLockShard__size_must_be_256(t *testing.T) {
	size := unsafe.Sizeof(lockShard[int]{})
	assert.Equal(t, 256, int(size))
}