package alloc

import (
	"math"
	"testing"
	"time"
	"unsafe"
)

func BenchmarkArena_AllocInt64(b *testing.B) {
	a := newArena()
	size := unsafe.Sizeof(int64(0))

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(size))

	t0 := time.Now()

	var v *int64
	for i := 0; i < b.N; i++ {
		v = ArenaAlloc[int64](a)
	}

	*v = math.MaxInt64
	if *v != math.MaxInt64 {
		b.Fatal()
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	capacity := a.size / (1024 * 1024)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(capacity), "cap,mb")
}

func BenchmarkArena_AllocStruct(b *testing.B) {
	type Struct struct {
		Int8  int8
		Int16 int16
		Int32 int32
		Int64 int64
	}

	a := newArena()
	size := unsafe.Sizeof(Struct{})

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(size))

	t0 := time.Now()

	var s *Struct
	for i := 0; i < b.N; i++ {
		s = ArenaAlloc[Struct](a)
		s.Int64 = math.MaxInt64
		if s.Int64 != math.MaxInt64 {
			b.Fatal()
		}
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	capacity := a.size / (1024 * 1024)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(capacity), "cap,mb")
}

func BenchmarkArena_AllocBytes(b *testing.B) {
	a := newArena()
	size := 16

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(size))

	t0 := time.Now()

	var v []byte
	for i := 0; i < b.N; i++ {
		v = a.Bytes(size)
		if len(v) != size {
			b.Fatal()
		}
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	capacity := a.size / (1024 * 1024)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(capacity), "cap,mb")
}

func BenchmarkArena_AllocSlice(b *testing.B) {
	a := newArena()
	n := 4
	size := n * 4

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(size))

	t0 := time.Now()

	var v []int32
	for i := 0; i < b.N; i++ {
		v = ArenaSlice[int32](a, n)
		if len(v) != 4 {
			b.Fatal()
		}
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	capacity := a.size / (1024 * 1024)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(capacity), "cap,mb")
}

func BenchmarkArena_Alloc(b *testing.B) {
	a := newArena()
	size := 8

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(size))

	t0 := time.Now()

	var v unsafe.Pointer
	for i := 0; i < b.N; i++ {
		v = a.alloc(size)
		if uintptr(v) == 0 {
			b.Fatal()
		}
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	capacity := a.size / (1024 * 1024)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(capacity), "cap,mb")
}

func BenchmarkArenaFreeList_Get_Put(b *testing.B) {
	a := newArena()
	list := newArenaFreeList[int64](a)
	size := 8

	b.ResetTimer()
	b.ReportAllocs()
	b.SetBytes(int64(size))

	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		v := list.Get()
		list.Put(v)
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	capacity := a.size / (1024 * 1024)

	b.ReportMetric(ops, "ops")
	b.ReportMetric(float64(size), "size")
	b.ReportMetric(float64(capacity), "cap,mb")
}

func BenchmarkGetBlockClass(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	t0 := time.Now()

	for i := 0; i < b.N; i++ {
		cls := getBlockClass(maxBlockSize)
		if cls != len(blockClassSizes)-1 {
			b.Fatal()
		}
	}

	sec := time.Since(t0).Seconds()
	ops := float64(b.N) / float64(sec)
	b.ReportMetric(ops, "ops")
}
