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

	// 分离正面图和建立背面图的快速查找 map
	var frontFiles []string
	backFilesMap := make(map[string]string)
	for _, imgFile := range imageFiles {
		base := filepath.Base(imgFile)
		ext := filepath.Ext(base)
		nameWithoutExt := strings.TrimSuffix(base, ext)

		nameLower := strings.ToLower(nameWithoutExt)
		if strings.HasSuffix(nameLower, "_back") {
			// 背面图
			frontKey := nameWithoutExt[:len(nameWithoutExt)-5]
			backFilesMap[strings.ToLower(frontKey)] = imgFile
		} else {
			// 正面图
			frontFiles = append(frontFiles, imgFile)
		}
	}

	if len(frontFiles) == 0 {
		log.Printf("⚠️ 图片目录 %s 中没有找到任何正面图片文件 (不含 _back)", imageDir)
		session.SendEvent("scan_info", iden, map[string]interface{}{
			"is_error": false,
			"info":     fmt.Sprintf("图片目录 %s 中没有正面图片", imageDir),
		})
		session.SendEvent("scan_end", iden, nil)
		return
	}

	// 读取参数
	pageMode := vs.store.GetPageMode()
	scanMode := vs.store.GetScanMode()
	scanCount := vs.store.GetScanCount()
	scanDelay := time.Duration(vs.store.Cfg.ScanDelay) * time.Millisecond
	saveDir := vs.store.Cfg.SaveDir

	// 延续指针
	startOffset := vs.store.GetScanImageOffset()
	if startOffset < 0 {
		startOffset = 0
	}

	if vs.store.Cfg.ScanLoop {
		// 循环模式下，自动对偏移指针取模重置
		if len(frontFiles) > 0 {
			startOffset = startOffset % len(frontFiles)
		} else {
			startOffset = 0
		}
	} else {
		// 非循环模式下，如果当前偏移量大于或等于纸张总数，表明进纸器已无纸
		if startOffset >= len(frontFiles) {
			log.Printf("⚠️ 物理纸张已被全部扫完，进纸槽无纸 (当前偏移量: %d, 纸张总数: %d)", startOffset, len(frontFiles))
			session.SendEvent("scan_info", iden, map[string]interface{}{
				"is_error": true,
				"info":     "no paper in adf",
			})
			session.SendEvent("scan_end", iden, nil)
			return
		}
	}

	log.Printf("📷 找到 %d 张正面物理纸张，当前偏移量为 %d，开始模拟扫描 (PageMode: %s, ScanMode: %s, Count: %d)...", len(frontFiles), startOffset, pageMode, scanMode, scanCount)

	// 发送 scan_begin 事件
	session.SendEvent("scan_begin", iden, nil)

	scannedSheets := 0
	currIdx := startOffset
	finalOffset := startOffset

	for {
		// 检查退出条件（进纸数达到限制，仅在指定张数扫描模式下生效）
		if scanMode == "specified" && scanCount > 0 && scannedSheets >= scanCount {
			log.Printf("⏹️ 达到请求的扫描纸张数上限 (%d 张)，结束扫描", scanCount)
			break
		}

		// 检查物理图片是否全部扫完且未启用循环模式
		if currIdx >= len(frontFiles) && !vs.store.Cfg.ScanLoop {
			log.Println("ℹ️ 已扫完全部物理纸张，进纸槽已空")
			break
		}

		// 取出当前的正面图
		imgIndex := currIdx % len(frontFiles)
		imgFile := frontFiles[imgIndex]

		// 1. 扫描正面
		select {
		case <-ctx.Done():
			log.Println("🛑 扫描被中断")
			session.SendEvent("scan_info", iden, map[string]interface{}{
				"is_error": false,
				"info":     "扫描被用户中断",
			})
			session.SendEvent("scan_end", iden, nil)
			vs.store.SetScanImageOffset(currIdx)
			return
		default:
		}

		// 执行正面图扫描并发送
		if err := vs.scanSingleImageFile(session, iden, imgFile, saveDir, params, false); err != nil {
			log.Printf("⚠️ 扫描正面 %s 失败: %v", imgFile, err)
		}

		// 2. 扫描背面 (如果为双面模式)
		if pageMode == "duplex" {
			base := filepath.Base(imgFile)
			ext := filepath.Ext(base)
			nameWithoutExt := strings.TrimSuffix(base, ext)

			// 查找对应的背面图
			backFile, hasBack := backFilesMap[strings.ToLower(nameWithoutExt)]
			if hasBack {
				select {
				case <-ctx.Done():
					log.Println("🛑 扫描被中断")
					session.SendEvent("scan_info", iden, map[string]interface{}{
						"is_error": false,
						"info":     "扫描被用户中断",
					})
					session.SendEvent("scan_end", iden, nil)
					vs.store.SetScanImageOffset(currIdx)
					return
				case <-time.After(50 * time.Millisecond): // 极小间隔
				}

				if err := vs.scanSingleImageFile(session, iden, backFile, saveDir, params, true); err != nil {
					log.Printf("⚠️ 扫描背面 %s 失败: %v", backFile, err)
				}
			} else {
				log.Printf("ℹ️ 未找到 %s 对应的背面图，跳过背面扫描", imgFile)
			}
		}

		// 3. 本张纸扫描完毕，增加计数
		scannedSheets++
		currIdx++
		finalOffset = currIdx

		// 模拟扫描下一张纸的进纸延迟
		select {
		case <-ctx.Done():
			log.Println("🛑 扫描被中断")
			session.SendEvent("scan_info", iden, map[string]interface{}{
				"is_error": false,
				"info":     "扫描被用户中断",
			})
			session.SendEvent("scan_end", iden, nil)
			vs.store.SetScanImageOffset(finalOffset)
			return
		case <-time.After(scanDelay):
		}
	}

	// 保存更新后的物理纸张偏移指针
	vs.store.SetScanImageOffset(finalOffset)

	// 如果指定了 save_path_name，处理多页文件保存
	if params.SavePathName != "" {
		vs.saveMultiPageIfNeeded(session, iden, params.SavePathName)
	}

	log.Printf("✅ 扫描完成，共处理了 %d 张物理纸张", scannedSheets)
	session.SendEvent("scan_end", iden, nil)
}

// scanSingleImageFile 扫描并发送单张图片文件
func (vs *VirtualScanner) scanSingleImageFile(session Session, iden string, imgFile string, saveDir string, params ScanParams, isBack bool) error {
	data, err := os.ReadFile(imgFile)
	if err != nil {
		session.SendEvent("scan_info", iden, map[string]interface{}{
			"is_error": true,
			"info":     fmt.Sprintf("读取图片失败: %s", filepath.Base(imgFile)),
		})
		return err
	}

	format := detectFormat(imgFile)

	eventData := map[string]interface{}{
		"is_blank": false,
	}

	// 如果 local_save 为 true，保存到目录
	if params.LocalSave {
		baseName := filepath.Base(imgFile)
		if isBack {
			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			if !strings.HasSuffix(strings.ToLower(nameWithoutExt), "_back") {
				baseName = nameWithoutExt + "_back" + ext
			}
		}

		savePath := filepath.Join(saveDir, baseName)
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
		side := "Front"
		if isBack {
			side = "Back"
		}
		log.Printf("  📄 [%s] %s", side, filepath.Base(imgFile))
	}

	return nil
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
