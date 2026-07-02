// Package store 管理全局状态：设备状态、配置、批号、图像列表
package store

import (
	"crypto/rand"
	"fmt"
	"sync"

	"github.com/huagao/virtual-scanner/config"
)

// DeviceState 设备状态枚举
type DeviceState int

const (
	StateUninitialized DeviceState = iota
	StateInitialized
	StateDeviceOpened
	StateScanning
)

func (s DeviceState) String() string {
	switch s {
	case StateUninitialized:
		return "uninitialized"
	case StateInitialized:
		return "initialized"
	case StateDeviceOpened:
		return "device_opened"
	case StateScanning:
		return "scanning"
	default:
		return "unknown"
	}
}

// GlobalConfig 文件保存/图像相关的全局配置
type GlobalConfig struct {
	FileSavePath     string  `json:"file_save_path"`
	FileNamePrefix   string  `json:"file_name_prefix"`
	FileNameMode     string  `json:"file_name_mode"`
	ImageFormat      string  `json:"image_format"`
	ImageJpegQuality int     `json:"image_jpeg_quality"`
	ImageTiffCompression string `json:"image_tiff_compression"`
	ImageTiffJpegQuality int  `json:"image_tiff_jpeg_quality"`
	ImageJp2Ratio    float64 `json:"image_jp2_ratio"`
}

// BindFolderConfig 文件夹绑定配置
type BindFolderConfig struct {
	Folder    string `json:"folder"`
	NameMode  string `json:"name_mode"`
	NameWidth int    `json:"name_width"`
	NameBase  int    `json:"name_base"`
}

// ImageRecord 图像记录
type ImageRecord struct {
	Index    int    // 在列表中的位置
	Tag      string // 标签
	FilePath string // 源文件或保存路径
	Data     []byte // 内存中的图像数据
	Format   string // 图像格式 (jpg, png, etc.)
}

// Batch 批号
type Batch struct {
	ID         string
	Images     []*ImageRecord
	BindFolder *BindFolderConfig
}

// Store 全局状态容器（线程安全）
type Store struct {
	mu sync.RWMutex

	Cfg *config.Config

	// 全局文件配置
	GlobalCfg GlobalConfig

	// 设备状态
	deviceState     DeviceState
	currDevice      string // 当前打开的设备名
	deviceParams    map[string]interface{}
	scanImageOffset int // 当前待扫描正面图片的索引偏移量

	// 批号管理
	batches   map[string]*Batch
	currBatch string // 当前批号 ID

	// 受保护的文件列表（仅允许删除本项目生成的文件）
	protectedFiles map[string]bool
}

// New 创建全局状态
func New(cfg *config.Config) *Store {
	s := &Store{
		Cfg:     cfg,
		GlobalCfg: GlobalConfig{
			FileSavePath:         cfg.SaveDir,
			FileNamePrefix:       "",
			FileNameMode:         "date_time",
			ImageFormat:          "jpg",
			ImageJpegQuality:     80,
			ImageTiffCompression: "none",
			ImageTiffJpegQuality: 80,
			ImageJp2Ratio:        10.0,
		},
		deviceState:     StateUninitialized,
		deviceParams:    make(map[string]interface{}),
		batches:         make(map[string]*Batch),
		currBatch:       "",
		protectedFiles:  make(map[string]bool),
		scanImageOffset: 0,
	}

	// 创建默认批号
	defaultBatch := &Batch{
		ID:     "default",
		Images: make([]*ImageRecord, 0),
	}
	s.batches["default"] = defaultBatch
	s.currBatch = "default"

	return s
}

// ─── 设备状态操作 ───

// GetDeviceState 获取设备状态
func (s *Store) GetDeviceState() DeviceState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deviceState
}

// SetDeviceState 设置设备状态
func (s *Store) SetDeviceState(state DeviceState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceState = state
}

// GetCurrDevice 获取当前打开的设备名
func (s *Store) GetCurrDevice() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currDevice
}

// SetCurrDevice 设置当前设备名
func (s *Store) SetCurrDevice(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.currDevice = name
	s.scanImageOffset = 0 // 重置偏移量
}

// GetDeviceParams 获取设备参数
func (s *Store) GetDeviceParams() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	params := make(map[string]interface{})
	for k, v := range s.deviceParams {
		params[k] = v
	}
	return params
}

// SetDeviceParams 批量设置设备参数
func (s *Store) SetDeviceParams(params map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range params {
		s.deviceParams[k] = v
	}
}

// ResetDeviceParams 重置设备参数为默认值
func (s *Store) ResetDeviceParams() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deviceParams = make(map[string]interface{})
	s.scanImageOffset = 0 // 重置偏移量
}

// GetScanImageOffset 获取当前扫描的图片偏移量
func (s *Store) GetScanImageOffset() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.scanImageOffset
}

// SetScanImageOffset 设置当前扫描的图片偏移量
func (s *Store) SetScanImageOffset(offset int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scanImageOffset = offset
}

// GetScanMode 获取当前 scan-mode（兼容 scan_mode）
func (s *Store) GetScanMode() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var modeVal interface{}
	var ok bool
	if modeVal, ok = s.deviceParams["scan-mode"]; !ok {
		if modeVal, ok = s.deviceParams["scan_mode"]; !ok {
			return "simplex"
		}
	}

	switch val := modeVal.(type) {
	case string:
		if val == "1" {
			return "duplex"
		}
		if val == "0" {
			return "simplex"
		}
		return val
	case int:
		if val == 1 {
			return "duplex"
		}
		return "simplex"
	case float64:
		if int(val) == 1 {
			return "duplex"
		}
		return "simplex"
	default:
		return "simplex"
	}
}

// GetScanCount 获取当前 scan-count（兼容 scan_count）
func (s *Store) GetScanCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var countVal interface{}
	var ok bool
	if countVal, ok = s.deviceParams["scan-count"]; !ok {
		if countVal, ok = s.deviceParams["scan_count"]; !ok {
			return 0
		}
	}

	switch val := countVal.(type) {
	case int:
		return val
	case float64:
		return int(val)
	case int64:
		return int(val)
	case string:
		var i int
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i
		}
		return 0
	default:
		return 0
	}
}

// ─── 全局配置操作 ───

// SetGlobalConfig 设置全局配置
func (s *Store) SetGlobalConfig(cfg GlobalConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.GlobalCfg = cfg
}

// GetGlobalConfig 获取全局配置
func (s *Store) GetGlobalConfig() GlobalConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.GlobalCfg
}

// ─── 批号操作 ───

// GetBatchIDList 获取所有批号列表
func (s *Store) GetBatchIDList() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.batches))
	for id := range s.batches {
		ids = append(ids, id)
	}
	return ids
}

// HasBatch 检查批号是否存在
func (s *Store) HasBatch(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.batches[id]
	return ok
}

// GetCurrentBatchID 获取当前批号
func (s *Store) GetCurrentBatchID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currBatch
}

// SetCurrentBatchID 切换当前批号
func (s *Store) SetCurrentBatchID(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.batches[id]; !ok {
		return fmt.Errorf("batch %s not found", id)
	}
	s.currBatch = id
	return nil
}

// NewBatch 创建新批号（不会自动打开）
func (s *Store) NewBatch(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id == "" {
		id = s.generateID()
	}
	if _, ok := s.batches[id]; ok {
		return "", fmt.Errorf("batch %s already exists", id)
	}
	s.batches[id] = &Batch{
		ID:     id,
		Images: make([]*ImageRecord, 0),
	}
	return id, nil
}

// DeleteBatch 删除批号
func (s *Store) DeleteBatch(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id == s.currBatch {
		return fmt.Errorf("cannot delete current batch")
	}
	if _, ok := s.batches[id]; !ok {
		return fmt.Errorf("batch %s not found", id)
	}
	delete(s.batches, id)
	return nil
}

// ModifyBatchID 修改批号名称
func (s *Store) ModifyBatchID(oldID, newID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch, ok := s.batches[oldID]
	if !ok {
		return fmt.Errorf("batch %s not found", oldID)
	}
	if _, exists := s.batches[newID]; exists {
		return fmt.Errorf("batch %s already exists", newID)
	}
	delete(s.batches, oldID)
	batch.ID = newID
	s.batches[newID] = batch

	if s.currBatch == oldID {
		s.currBatch = newID
	}
	return nil
}

// GetCurrentBatch 获取当前批号对象
func (s *Store) GetCurrentBatch() *Batch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.batches[s.currBatch]
}

// ─── 图像列表操作 ───

// GetImageCount 获取当前批号下的图像数量
func (s *Store) GetImageCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return 0
	}
	return len(batch.Images)
}

// GetImage 按 index 获取图像记录
func (s *Store) GetImage(index int) *ImageRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	batch := s.batches[s.currBatch]
	if batch == nil || index < 0 || index >= len(batch.Images) {
		return nil
	}
	return batch.Images[index]
}

// GetImageList 获取当前批号的所有图像记录
func (s *Store) GetImageList() []*ImageRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return nil
	}
	result := make([]*ImageRecord, len(batch.Images))
	copy(result, batch.Images)
	return result
}

// AddImage 追加图像到当前批号末尾
func (s *Store) AddImage(record *ImageRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return
	}
	record.Index = len(batch.Images)
	batch.Images = append(batch.Images, record)
}

// InsertImage 在指定位置插入图像
func (s *Store) InsertImage(pos int, record *ImageRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return
	}

	n := len(batch.Images)
	if pos < 0 || pos > n {
		pos = n // -1 或其他无效值表示末尾
	}

	batch.Images = append(batch.Images, nil)
	copy(batch.Images[pos+1:], batch.Images[pos:])
	batch.Images[pos] = record

	// 重新索引
	for i := range batch.Images {
		batch.Images[i].Index = i
	}
}

// DeleteImages 删除指定索引的图像
func (s *Store) DeleteImages(indices []int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return
	}

	// 构建删除集合
	delSet := make(map[int]bool)
	for _, idx := range indices {
		delSet[idx] = true
	}

	newImages := make([]*ImageRecord, 0)
	for i, img := range batch.Images {
		if !delSet[i] {
			newImages = append(newImages, img)
		}
	}

	// 重新索引
	for i := range newImages {
		newImages[i].Index = i
	}
	batch.Images = newImages
}

// ClearImages 清空当前批号的图像列表
func (s *Store) ClearImages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return
	}
	batch.Images = make([]*ImageRecord, 0)
}

// ModifyImageTag 修改图像标签
func (s *Store) ModifyImageTag(updates map[int]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return
	}
	for idx, tag := range updates {
		if idx >= 0 && idx < len(batch.Images) {
			batch.Images[idx].Tag = tag
		}
	}
}

// ModifyImageData 修改指定索引的图像数据
func (s *Store) ModifyImageData(index int, data []byte, format string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return fmt.Errorf("no current batch")
	}
	if index < 0 || index >= len(batch.Images) {
		return fmt.Errorf("index out of range")
	}
	batch.Images[index].Data = data
	batch.Images[index].Format = format
	return nil
}

// ModifyImageFilePath 修改指定索引的图像文件路径
func (s *Store) ModifyImageFilePath(index int, path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return fmt.Errorf("no current batch")
	}
	if index < 0 || index >= len(batch.Images) {
		return fmt.Errorf("index out of range")
	}
	batch.Images[index].FilePath = path
	return nil
}

// MoveImages 移动图像位置
// mode: "pos" 表示移动到 target 之前, "index" 表示移动到 target 索引位置
func (s *Store) MoveImages(indices []int, mode string, target int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil || len(indices) == 0 {
		return
	}

	n := len(batch.Images)
	// 收集要移动的图像
	moveSet := make(map[int]bool)
	var moveImgs []*ImageRecord
	for _, idx := range indices {
		if idx >= 0 && idx < n && !moveSet[idx] {
			moveSet[idx] = true
			moveImgs = append(moveImgs, batch.Images[idx])
		}
	}
	if len(moveImgs) == 0 {
		return
	}

	// 从原列表中移除
	var remaining []*ImageRecord
	for i, img := range batch.Images {
		if !moveSet[i] {
			remaining = append(remaining, img)
		}
	}

	// 计算插入位置
	insertPos := target
	if insertPos < 0 || insertPos > len(remaining) {
		insertPos = len(remaining)
	}

	// 插入
	newImages := make([]*ImageRecord, 0, n)
	newImages = append(newImages, remaining[:insertPos]...)
	newImages = append(newImages, moveImgs...)
	newImages = append(newImages, remaining[insertPos:]...)

	// 重新索引
	for i := range newImages {
		newImages[i].Index = i
	}
	batch.Images = newImages
}

// ExchangeImages 交换两个位置的图像
func (s *Store) ExchangeImages(idx1, idx2 int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch == nil {
		return fmt.Errorf("no current batch")
	}
	n := len(batch.Images)
	if idx1 < 0 || idx1 >= n || idx2 < 0 || idx2 >= n {
		return fmt.Errorf("index out of range")
	}
	batch.Images[idx1], batch.Images[idx2] = batch.Images[idx2], batch.Images[idx1]
	batch.Images[idx1].Index = idx1
	batch.Images[idx2].Index = idx2
	return nil
}

// BindFolder 绑定文件夹
func (s *Store) BindFolder(cfg BindFolderConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch != nil {
		batch.BindFolder = &cfg
	}
}

// StopBindFolder 停止绑定文件夹
func (s *Store) StopBindFolder() {
	s.mu.Lock()
	defer s.mu.Unlock()
	batch := s.batches[s.currBatch]
	if batch != nil {
		batch.BindFolder = nil
	}
}

// GetBindFolder 获取当前绑定文件夹配置
func (s *Store) GetBindFolder() *BindFolderConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	batch := s.batches[s.currBatch]
	if batch != nil {
		return batch.BindFolder
	}
	return nil
}

// ─── 受保护文件操作 ───

// MarkFileProtected 标记文件为受保护（本项目生成的文件）
func (s *Store) MarkFileProtected(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.protectedFiles[path] = true
}

// IsFileProtected 检查文件是否受保护
func (s *Store) IsFileProtected(path string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.protectedFiles[path]
}

// ─── 工具函数 ───

// generateID 生成随机 4 位十六进制 ID
func (s *Store) generateID() string {
	b := make([]byte, 2)
	rand.Read(b)
	return fmt.Sprintf("%04x", b)
}
