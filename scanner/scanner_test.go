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
	InfoEvents  []map[string]interface{}
}

func (m *MockSession) SendOK(funcName, iden string, extra map[string]interface{}) {}
func (m *MockSession) SendError(funcName, iden string, ret int, errInfo string) {}
func (m *MockSession) SendEvent(funcName, iden string, extra map[string]interface{}) {
	m.Events = append(m.Events, funcName)
	if funcName == "scan_image" {
		m.ImageEvents = append(m.ImageEvents, extra)
	}
	if funcName == "scan_info" {
		m.InfoEvents = append(m.InfoEvents, extra)
	}
}
func (m *MockSession) SendMessage(msg interface{}) error { return nil }

// TestStoreScanParams 测试 store 中配置解析的兼容性
func TestStoreScanParams(t *testing.T) {
	cfg := config.DefaultConfig()
	st := store.New(cfg)

	// 1. 测试 page (单双面兼容性)
	st.SetDeviceParams(map[string]interface{}{"page": "双面"})
	if st.GetPageMode() != "duplex" {
		t.Errorf("expected duplex, got %s", st.GetPageMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"page": "对折"})
	if st.GetPageMode() != "duplex" {
		t.Errorf("expected duplex, got %s", st.GetPageMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"page": 1})
	if st.GetPageMode() != "duplex" {
		t.Errorf("expected duplex, got %s", st.GetPageMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"page": "单面"})
	if st.GetPageMode() != "simplex" {
		t.Errorf("expected simplex, got %s", st.GetPageMode())
	}

	// 2. 测试 scan-mode (中下划线、中文/英文、字符串/数字兼容性)
	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan-mode": "扫描指定张数"})
	if st.GetScanMode() != "specified" {
		t.Errorf("expected specified, got %s", st.GetScanMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan_mode": "连续扫描"})
	if st.GetScanMode() != "continuous" {
		t.Errorf("expected continuous, got %s", st.GetScanMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan_mode": 1})
	if st.GetScanMode() != "specified" {
		t.Errorf("expected specified, got %s", st.GetScanMode())
	}

	st.ResetDeviceParams()
	st.SetDeviceParams(map[string]interface{}{"scan_mode": "0"})
	if st.GetScanMode() != "continuous" {
		t.Errorf("expected continuous, got %s", st.GetScanMode())
	}

	// 3. 测试 scan-count (中下划线、字符串/数字转换)
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
	st.SetDeviceParams(map[string]interface{}{
		"page":      "单面",
		"scan-mode": "连续扫描",
	})
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
	st.SetDeviceParams(map[string]interface{}{
		"page":      "双面",
		"scan-mode": "连续扫描",
	})
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

// TestScanContinuationAndCountLimit 测试指针延续和进纸页数限制
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
	st.SetDeviceParams(map[string]interface{}{
		"page":       "双面",
		"scan-mode":  "扫描指定张数",
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
	// 扫完 1 张物理纸张的正面与背面，共 2 张图像
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

// TestOutOfPaperErrorAndReset 测试无纸报错以及重置逻辑
func TestOutOfPaperErrorAndReset(t *testing.T) {
	tempDir := t.TempDir()

	// 准备 2 张物理纸张
	_ = os.WriteFile(filepath.Join(tempDir, "001.jpg"), []byte("f1"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "002.jpg"), []byte("f2"), 0644)

	cfg := config.DefaultConfig()
	cfg.ImageDir = tempDir
	cfg.ScanDelay = 1
	st := store.New(cfg)
	vs := New(st)

	// 设置为扫描指定张数，每次限制扫描 1 张
	st.SetDeviceParams(map[string]interface{}{
		"page":       "单面",
		"scan-mode":  "扫描指定张数",
		"scan-count": 1,
	})

	// 第一次：扫完 001.jpg
	session1 := &MockSession{}
	vs.StartScan(session1, "s1", ScanParams{LocalSave: false, GetBase64: false})
	for vs.IsScanning() {}
	if st.GetScanImageOffset() != 1 {
		t.Errorf("expected offset 1, got %d", st.GetScanImageOffset())
	}

	// 第二次：扫完 002.jpg
	session2 := &MockSession{}
	vs.StartScan(session2, "s2", ScanParams{LocalSave: false, GetBase64: false})
	for vs.IsScanning() {}
	if st.GetScanImageOffset() != 2 {
		t.Errorf("expected offset 2, got %d", st.GetScanImageOffset())
	}

	// 第三次：此时已经没有物理纸张了，应该触发“无纸报错”
	session3 := &MockSession{}
	vs.StartScan(session3, "s3", ScanParams{LocalSave: false, GetBase64: false})
	for vs.IsScanning() {}

	// 验证是否触发了无纸报错，并且没有产生任何 scan_image
	if len(session3.InfoEvents) == 0 {
		t.Fatalf("expected scan_info event, got none")
	}
	isError := session3.InfoEvents[0]["is_error"].(bool)
	infoStr := session3.InfoEvents[0]["info"].(string)
	if !isError || infoStr != "no paper in adf" {
		t.Errorf("expected 'no paper in adf' error, got is_error: %t, info: %s", isError, infoStr)
	}

	imageCount3 := 0
	for _, ev := range session3.Events {
		if ev == "scan_image" {
			imageCount3++
		}
	}
	if imageCount3 != 0 {
		t.Errorf("expected 0 scanned images, got %d", imageCount3)
	}

	// 4. 重置设备，验证偏移量重新归零
	st.SetCurrDevice("")
	if st.GetScanImageOffset() != 0 {
		t.Errorf("expected reset offset to 0, got %d", st.GetScanImageOffset())
	}
}
