// Package scanner 实现虚拟扫描仪：从目录读取图片，模拟扫描流程
package scanner

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/huagao/virtual-scanner/store"
)

// Session 接口，避免循环依赖
type Session interface {
	SendOK(funcName, iden string, extra map[string]interface{})
	SendError(funcName, iden string, ret int, errInfo string)
	SendEvent(funcName, iden string, extra map[string]interface{})
	SendMessage(msg interface{}) error
}

// VirtualScanner 虚拟扫描仪
type VirtualScanner struct {
	store *store.Store

	mu         sync.Mutex
	cancelFunc context.CancelFunc // 取消当前扫描
}

// New 创建虚拟扫描仪
func New(st *store.Store) *VirtualScanner {
	return &VirtualScanner{
		store: st,
	}
}

// IsScanning 返回是否正在扫描
func (vs *VirtualScanner) IsScanning() bool {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	return vs.cancelFunc != nil
}

// ScanParams 扫描参数
type ScanParams struct {
	BlankCheck  bool   `json:"blank_check"`
	LocalSave   bool   `json:"local_save"`
	GetBase64   bool   `json:"get_base64"`
	SavePathName string `json:"save_path_name"`
}

// StartScan 开始扫描
func (vs *VirtualScanner) StartScan(session Session, iden string, params ScanParams) {
	vs.mu.Lock()
	if vs.cancelFunc != nil {
		vs.mu.Unlock()
		session.SendError("start_scan", iden, -1, "device is already scanning")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	vs.cancelFunc = cancel
	vs.mu.Unlock()

	// 设置设备状态为扫描中
	vs.store.SetDeviceState(store.StateScanning)

	// 先返回成功响应
	session.SendOK("start_scan", iden, nil)

	// 在 goroutine 中执行扫描
	go vs.doScan(session, iden, params, ctx, cancel)
}

// StopScan 停止扫描
func (vs *VirtualScanner) StopScan(session Session, iden string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()

	if vs.cancelFunc == nil {
		session.SendError("stop_scan", iden, -1, "device is not scanning")
		return
	}

	vs.cancelFunc()
	session.SendOK("stop_scan", iden, nil)
}

// doScan 执行扫描流程
func (vs *VirtualScanner) doScan(session Session, iden string, params ScanParams, ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		vs.mu.Lock()
		vs.cancelFunc = nil
		vs.mu.Unlock()
		cancel()

		// 恢复设备状态
		vs.store.SetDeviceState(store.StateDeviceOpened)
	}()

	// 读取图片目录
	imageDir := vs.store.Cfg.ImageDir
	imageFiles, err := listImageFiles(imageDir)
	if err != nil {
		log.Printf("❌ 读取图片目录失败: %v", err)
		session.SendEvent("scan_info", iden, map[string]interface{}{
			"is_error": true,
			"info":     fmt.Sprintf("读取图片目录失败: %v", err),
		})
		session.SendEvent("scan_end", iden, nil)
		return
	}

	if len(imageFiles) == 0 {
		log.Printf("⚠️ 图片目录 %s 中没有找到图片文件", imageDir)
		session.SendEvent("scan_info", iden, map[string]interface{}{
			"is_error": false,
			"info":     fmt.Sprintf("图片目录 %s 中没有图片", imageDir),
		})
		session.SendEvent("scan_end", iden, nil)
		return
	}

	log.Printf("📷 找到 %d 张图片，开始模拟扫描...", len(imageFiles))

	// 发送 scan_begin 事件
	session.SendEvent("scan_begin", iden, nil)

	// 逐个扫描图片
	scanDelay := time.Duration(vs.store.Cfg.ScanDelay) * time.Millisecond
	saveDir := vs.store.Cfg.SaveDir

	for i, imgFile := range imageFiles {
		select {
		case <-ctx.Done():
			log.Println("🛑 扫描被中断")
			session.SendEvent("scan_info", iden, map[string]interface{}{
				"is_error": false,
				"info":     "扫描被用户中断",
			})
			session.SendEvent("scan_end", iden, nil)
			return
		default:
		}

		// 读取图片数据
		data, err := os.ReadFile(imgFile)
		if err != nil {
			log.Printf("⚠️ 读取图片 %s 失败: %v", imgFile, err)
			session.SendEvent("scan_info", iden, map[string]interface{}{
				"is_error": true,
				"info":     fmt.Sprintf("读取图片失败: %s", filepath.Base(imgFile)),
			})
			continue
		}

		// 检测格式
		format := detectFormat(imgFile)

		// 构建 scan_image 事件数据
		eventData := map[string]interface{}{
			"is_blank": false,
		}

		// 如果 local_save 为 true，保存到目录
		if params.LocalSave {
			savePath := filepath.Join(saveDir, filepath.Base(imgFile))
			if err := os.MkdirAll(saveDir, 0755); err == nil {
				if err := os.WriteFile(savePath, data, 0644); err == nil {
					eventData["image_path"] = savePath
					vs.store.MarkFileProtected(savePath)
				}
			}
		}

		// 如果 get_base64 为 true，附带 base64 数据
		if params.GetBase64 {
			eventData["image_base64"] = encodeBase64(data)
		}

		// 发送 scan_image 事件
		session.SendEvent("scan_image", iden, eventData)

		// 将图片添加到当前批号的图像列表（仅使用原始数据）
		vs.store.AddImage(&store.ImageRecord{
			FilePath: imgFile,
			Data:     data,
			Format:   format,
		})

		if vs.store.Cfg.Verbose {
			log.Printf("  📄 [%d/%d] %s", i+1, len(imageFiles), filepath.Base(imgFile))
		}

		// 模拟扫描延迟
		select {
		case <-ctx.Done():
			log.Println("🛑 扫描被中断")
			session.SendEvent("scan_info", iden, map[string]interface{}{
				"is_error": false,
				"info":     "扫描被用户中断",
			})
			session.SendEvent("scan_end", iden, nil)
			return
		case <-time.After(scanDelay):
		}
	}

	// 如果指定了 save_path_name，处理多页文件保存
	if params.SavePathName != "" {
		vs.saveMultiPageIfNeeded(session, iden, params.SavePathName)
	}

	log.Printf("✅ 扫描完成，共 %d 张图片", len(imageFiles))
	session.SendEvent("scan_end", iden, nil)
}

// saveMultiPageIfNeeded 如果需要，保存为多页文件
func (vs *VirtualScanner) saveMultiPageIfNeeded(session Session, iden, savePath string) {
	ext := strings.ToLower(filepath.Ext(savePath))
	if ext != ".pdf" && ext != ".tif" && ext != ".tiff" && ext != ".ofd" {
		return
	}

	images := vs.store.GetImageList()
	if len(images) == 0 {
		return
	}

	session.SendEvent("scan_info", iden, map[string]interface{}{
		"is_error": false,
		"info":     fmt.Sprintf("正在保存多页文件: %s", savePath),
	})
}

// listImageFiles 列出目录下所有图片文件（按文件名排序）
func listImageFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	exts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".bmp": true,
		".tif": true, ".tiff": true, ".gif": true, ".webp": true,
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if exts[ext] {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(files)
	return files, nil
}

// detectFormat 从文件扩展名检测图像格式
func detectFormat(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpg"
	case ".png":
		return "png"
	case ".bmp":
		return "bmp"
	case ".tif", ".tiff":
		return "tif"
	case ".gif":
		return "gif"
	case ".webp":
		return "webp"
	default:
		return "jpg"
	}
}
