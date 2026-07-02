package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/huagao/virtual-scanner/config"
	"github.com/huagao/virtual-scanner/store"
)

type MockSession struct {
	Events      []string
	ImageEvents []map[string]interface{}
}

func (m *MockSession) SendOK(funcName, iden string, extra map[string]interface{}) {}
func (m *MockSession) SendError(funcName, iden string, ret int, errInfo string) {}
func (m *MockSession) SendEvent(funcName, iden string, extra map[string]interface{}) {
	m.Events = append(m.Events, funcName)
	if funcName == "scan_image" {
		m.ImageEvents = append(m.ImageEvents, extra)
	}
}
func (m *MockSession) SendMessage(msg interface{}) error { return nil }

// TestStoreScanParams 测试 store 中配置解析的兼容性
func TestStoreScanParams(t *testing.T) {
	cfg := config.DefaultConfig()
	st := store.New(cfg)

	// 1. 测试 scan-mode (中下划线、数字、字符串兼容性)
	st.SetDeviceParams(map[string]interface{}{"scan-mode": "duplex"})
	if st.GetScanMode() != "duplex" {
		t.Errorf("expected duplex, got %s", st.GetScanMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan_mode": 1})
	if st.GetScanMode() != "duplex" {
		t.Errorf("expected duplex, got %s", st.GetScanMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan_mode": "0"})
	if st.GetScanMode() != "simplex" {
		t.Errorf("expected simplex, got %s", st.GetScanMode())
	}

	// 2. 测试 scan-count (中下划线、字符串/数字转换)
	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan-count": 5})
	if st.GetScanCount() != 5 {
		t.Errorf("expected 5, got %d", st.GetScanCount())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan_count": "10"})
	if st.GetScanCount() != 10 {
		t.Errorf("expected 10, got %d", st.GetScanCount())
	}
}

// TestSimplexAndDuplexScan 测试单双面逻辑与 _back 后缀匹配
func TestSimplexAndDuplexScan(t *testing.T) {
	tempDir := t.TempDir()
	
	// 创建测试图片
	// 001 为双面，有 001.jpg 和 001_back.jpg
	// 002 为单面，只有 002.jpg
	err := os.WriteFile(filepath.Join(tempDir, "001.jpg"), []byte("front1"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "001_back.jpg"), []byte("back1"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "002.jpg"), []byte("front2"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.ImageDir = tempDir
	cfg.ScanDelay = 1 // 快速测试
	st := store.New(cfg)
	vs := New(st)

	// 1. 单面扫描测试 (simplex)
	st.SetDeviceParams(map[string]interface{}{"scan-mode": "simplex", "scan-count": 0})
	session1 := &MockSession{}
	vs.StartScan(session1, "test1", ScanParams{LocalSave: false, GetBase64: false})
	
	// 等待扫描协程完成
	for vs.IsScanning() {
		// spin wait
	}

	// 应该包含 001 和 002 共两张正面图，事件有 scan_begin, 2x scan_image, scan_end
	imageCount1 := 0
	for _, ev := range session1.Events {
		if ev == "scan_image" {
			imageCount1++
		}
	}
	if imageCount1 != 2 {
		t.Errorf("expected 2 simplex scanned images, got %d", imageCount1)
	}

	// 2. 双面扫描测试 (duplex)
	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan-mode": "duplex", "scan-count": 0})
	session2 := &MockSession{}
	vs.StartScan(session2, "test2", ScanParams{LocalSave: false, GetBase64: false})

	for vs.IsScanning() {
	}

	// 001 有背面， 002 无背面。应该输出 001 正面、001 背面、002 正面，总共 3 张
	imageCount2 := 0
	for _, ev := range session2.Events {
		if ev == "scan_image" {
			imageCount2++
		}
	}
	if imageCount2 != 3 {
		t.Errorf("expected 3 duplex scanned images, got %d", imageCount2)
	}
}

// TestScanContinuationAndCountLimit 测试选项二的指针延续和进纸页数限制
func TestScanContinuationAndCountLimit(t *testing.T) {
	tempDir := t.TempDir()
	
	// 准备 3 张正面图片：001.jpg, 002.jpg, 003.jpg，其中 001 有背面 001_back.jpg
	_ = os.WriteFile(filepath.Join(tempDir, "001.jpg"), []byte("f1"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "001_back.jpg"), []byte("b1"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "002.jpg"), []byte("f2"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "003.jpg"), []byte("f3"), 0644)

	cfg := config.DefaultConfig()
	cfg.ImageDir = tempDir
	cfg.ScanDelay = 1
	st := store.New(cfg)
	vs := New(st)

	// 限制只能扫 1 张物理纸张 (scan-count = 1)，双面模式 (duplex)
	// 由于是方案 B（按物理进纸张数限制），应当扫完 001 的正面和背面，然后结束。
	st.SetDeviceParams(map[string]interface{}{
		"scan-mode":  "duplex",
		"scan-count": 1,
	})

	// 第一次扫描会话
	session1 := &MockSession{}
	vs.StartScan(session1, "session1", ScanParams{LocalSave: false, GetBase64: false})
	for vs.IsScanning() {
	}

	imageCount1 := 0
	for _, ev := range session1.Events {
		if ev == "scan_image" {
			imageCount1++
		}
	}
	// 方案B：扫完 1 张物理纸张的正面与背面，共 2 张图像
	if imageCount1 != 2 {
		t.Errorf("expected 2 images for 1 duplex sheet, got %d", imageCount1)
	}

	// 确认偏移量更新为 1
	if st.GetScanImageOffset() != 1 {
		t.Errorf("expected offset to advance to 1, got %d", st.GetScanImageOffset())
	}

	// 第二次扫描会话：还是扫 1 张纸，双面模式。应该扫第二张纸 (002.jpg，它没有背面)
	session2 := &MockSession{}
	vs.StartScan(session2, "session2", ScanParams{LocalSave: false, GetBase64: false})
	for vs.IsScanning() {
	}

	imageCount2 := 0
	for _, ev := range session2.Events {
		if ev == "scan_image" {
			imageCount2++
		}
	}
	// 002 没有背面，只输出 1 张正面
	if imageCount2 != 1 {
		t.Errorf("expected 1 image (no back file), got %d", imageCount2)
	}

	// 确认偏移量更新为 2
	if st.GetScanImageOffset() != 2 {
		t.Errorf("expected offset to advance to 2, got %d", st.GetScanImageOffset())
	}
}
