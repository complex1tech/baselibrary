// Copyright 2024 Ivan Korobkov. All rights reserved.
// Use of this software is governed by the MIT License
// that can be found in the LICENSE file.

package ref

import "testing"

func BenchmarkR(b *testing.B) {
	r := NewNoop(1)
	b.SetParallelism(10)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r.Retain()
			r.Release()
		}
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops/s")
}

func BenchmarkConcurrentVar_Acquire(b *testing.B) {
	r := NewNoop(1)
	v := NewConcurrentVar[int]()
	v.SwapRetain(r)
	r.Release()

	b.SetParallelism(10)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r, ok := v.Acquire()
			if !ok {
				b.Fatal("acquire failed")
			}
			_ = r
		}
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops/s")
}

func BenchmarkConcurrentVar_Acquire_Release(b *testing.B) {
	r := NewNoop(1)
	v := NewConcurrentVar[int]()
	v.SwapRetain(r)
	r.Release()

	b.SetParallelism(10)

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			r, ok := v.Acquire()
			if !ok {
				b.Fatal("acquire failed")
			}
			r.Release()
		}
	})

	sec := b.Elapsed().Seconds()
	ops := float64(b.N) / sec
	b.ReportMetric(ops/1000_000, "mops/s")
}
