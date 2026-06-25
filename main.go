package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/huagao/virtual-scanner/config"
	"github.com/huagao/virtual-scanner/handler"
	"github.com/huagao/virtual-scanner/scanner"
	"github.com/huagao/virtual-scanner/server"
	"github.com/huagao/virtual-scanner/store"
)

// Version 由构建时通过 ldflags 注入
var Version = "dev"

func main() {
	cfg := config.ParseFlags()

	if cfg.ShowVersion {
		fmt.Printf("华高扫描仪虚拟模拟器 v%s\n", Version)
		os.Exit(0)
	}

	// 配置日志
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	if !cfg.Verbose {
		log.SetFlags(log.Ldate | log.Ltime)
	}

	log.Println("═══════════════════════════════════════════")
	log.Println("  华高扫描仪虚拟模拟器 (Scanner Simulator)")
	log.Printf("  版本: %s\n", Version)
	log.Println("═══════════════════════════════════════════")

	// 创建全局状态
	st := store.New(cfg)

	// 创建虚拟扫描仪
	vScanner := scanner.New(st)

	// 创建 Handler 注册表
	registry := handler.NewRegistry(st, vScanner)

	// 启动 WebSocket 服务器
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("📡 WebSocket 服务器启动在 ws://0.0.0.0%s\n", addr)
	log.Printf("📁 图片目录: %s\n", cfg.ImageDir)
	log.Printf("🔧 设备名称: %s\n", cfg.DeviceName)
	log.Printf("⏱️  扫描间隔: %dms\n", cfg.ScanDelay)
	if cfg.ScanLoop {
		log.Println("🔄 循环扫描: 已启用")
	}
	if cfg.Verbose {
		log.Println("📝 详细日志: 已启用")
	}
	log.Println("───────────────────────────────────────────")
	log.Println("  等待客户端连接...")

	// 优雅退出处理
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("\n🛑 收到退出信号，正在关闭服务器...")
		os.Exit(0)
	}()

	if err := server.Start(addr, registry, st); err != nil {
		log.Fatalf("❌ 服务器错误: %v\n", err)
	}
}
