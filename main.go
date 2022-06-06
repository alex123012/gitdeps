package main

import (
	"context"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/alex123012/dependency-bot/pkg/structs"
	"k8s.io/klog/v2"
)

var (
	HeaderName = "PRIVATE-TOKEN"
	token      = os.Getenv("TOKEN")
	GET        = "GET"
	POST       = "POST"
)

func main() {
	ctx := context.Background()
	tmp := structs.NewGitLab(os.Getenv("TOKEN_SWEED"), "https://gitlab.walli.com/")
	// go SystemStats(ctx)

	err := tmp.Run(ctx)
	if err != nil {
		klog.Errorln(err)
	}
}
func SystemStats(ctx context.Context) error {
	mem := &runtime.MemStats{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		for {
			cpu := runtime.NumCPU()
			log.Println("CPU:", cpu)

			rot := runtime.NumGoroutine()
			log.Println("Goroutine:", rot)

			// Byte
			runtime.ReadMemStats(mem)
			log.Println("Memory:", mem.Alloc/1024)

			time.Sleep(2 * time.Second)
			log.Println("-------")
		}
	}()
	<-ctx.Done()
	return nil
}
