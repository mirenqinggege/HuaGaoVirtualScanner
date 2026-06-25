// Package config 处理 CLI 参数和运行配置
package config

import (
	"flag"
	"fmt"
	"os"
)

// Config 运行配置
type Config struct {
	Port       int    // WebSocket 监听端口
	ImageDir   string // 模拟扫描的图片目录
	DeviceName string // 虚拟设备名称
	DeviceSN   string // 虚拟设备序列号
	ScanDelay  int    // 每张图片扫描间隔（毫秒）
	ScanLoop   bool   // 是否循环扫描
	SaveDir    string // 文件保存根目录
	Verbose    bool   // 详细日志
	ShowVersion bool  // 显示版本号
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Port:       38999,
		ImageDir:   ".",
		DeviceName: "virtual_scanner",
		DeviceSN:   "SIM00000001",
		ScanDelay:  500,
		ScanLoop:   false,
		SaveDir:    "./scan_output",
		Verbose:    false,
		ShowVersion: false,
	}
}

// ParseFlags 解析命令行参数
func ParseFlags() *Config {
	cfg := DefaultConfig()

	flag.IntVar(&cfg.Port, "port", cfg.Port, "WebSocket 监听端口")
	flag.StringVar(&cfg.ImageDir, "image-dir", cfg.ImageDir, "模拟扫描的图片目录")
	flag.StringVar(&cfg.DeviceName, "device-name", cfg.DeviceName, "虚拟设备名称")
	flag.StringVar(&cfg.DeviceSN, "device-sn", cfg.DeviceSN, "虚拟设备序列号")
	flag.IntVar(&cfg.ScanDelay, "scan-delay", cfg.ScanDelay, "每张图片扫描间隔（毫秒）")
	flag.BoolVar(&cfg.ScanLoop, "scan-loop", cfg.ScanLoop, "循环扫描模式")
	flag.StringVar(&cfg.SaveDir, "save-dir", cfg.SaveDir, "文件保存根目录")
	flag.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "详细日志输出")
	flag.BoolVar(&cfg.ShowVersion, "version", cfg.ShowVersion, "显示版本号")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "华高扫描仪虚拟模拟器 (Scanner Simulator)\n\n")
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s --image-dir ./images --verbose\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --port 8080 --scan-delay 200\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --scan-loop --image-dir /path/to/images\n", os.Args[0])
	}

	flag.Parse()
	return cfg
}
