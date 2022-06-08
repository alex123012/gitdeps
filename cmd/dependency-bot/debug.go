package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"

	"golang.org/x/sync/errgroup"
)

type RuntimeStats struct {
	CpuUsage          int    `json:"cpu_count"`
	GoroutinesRunning int    `json:"goroutine_count"`
	MemoryAllocated   string `json:"allocated_memory"`
	MemoryHeap        string `json:"heap_allocated"`
}

func Debug(ctx context.Context, port int) error {
	r := http.NewServeMux()
	svr := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: r}
	mem := &runtime.MemStats{}
	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(func() error {
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		r.HandleFunc("/debug/pprof/trace", pprof.Trace)

		r.HandleFunc("/debug/stats", func(w http.ResponseWriter, r *http.Request) {
			cpu := runtime.NumCPU()
			rot := runtime.NumGoroutine()
			runtime.ReadMemStats(mem)

			data := RuntimeStats{
				CpuUsage:          cpu,
				GoroutinesRunning: rot,
				MemoryAllocated:   fmt.Sprintf("%d KB", mem.Alloc/1024),
				MemoryHeap:        fmt.Sprintf("%d KB", mem.HeapAlloc/1024),
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(data)
		})

		return svr.ListenAndServe()
	})
	grp.Go(func() error {
		<-ctx.Done()
		return svr.Shutdown(ctx)
	})
	return grp.Wait()
}
