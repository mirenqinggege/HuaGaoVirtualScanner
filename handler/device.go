// Package handler - 三、图像采集接口 (15个)
package handler

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/huagao/virtual-scanner/scanner"
	"github.com/huagao/virtual-scanner/server"
	"github.com/huagao/virtual-scanner/store"
)

// registerDeviceHandlers 注册设备相关处理器
func registerDeviceHandlers(reg *server.HandlerRegistry, st *store.Store, vs *scanner.VirtualScanner) {
	// 1. 设备初始化
	reg.Register("init_device", func(session *server.Session, iden string, raw json.RawMessage) {
		handleInitDevice(session, iden, raw, st)
	})

	// 2. 设备反初始化
	reg.Register("deinit_device", func(session *server.Session, iden string, raw json.RawMessage) {
		handleDeinitDevice(session, iden, raw, st)
	})

	// 3. 获取设备是否已初始化
	reg.Register("is_device_init", func(session *server.Session, iden string, raw json.RawMessage) {
		handleIsDeviceInit(session, iden, raw, st)
	})

	// 4. 获取设备列表
	reg.Register("get_device_name_list", func(session *server.Session, iden string, raw json.RawMessage) {
		handleGetDeviceNameList(session, iden, raw, st)
	})

	// 5. 打开设备
	reg.Register("open_device", func(session *server.Session, iden string, raw json.RawMessage) {
		handleOpenDevice(session, iden, raw, st)
	})

	// 6. 关闭设备
	reg.Register("close_device", func(session *server.Session, iden string, raw json.RawMessage) {
		handleCloseDevice(session, iden, raw, st)
	})

	// 7. 获取设备序列号
	reg.Register("get_device_sn", func(session *server.Session, iden string, raw json.RawMessage) {
		handleGetDeviceSN(session, iden, raw, st)
	})

	// 8. 获取设备固件版本号
	reg.Register("get_device_fwversion", func(session *server.Session, iden string, raw json.RawMessage) {
		handleGetDeviceFWVersion(session, iden, raw, st)
	})

	// 9. 设置设备参数
	reg.Register("set_device_param", func(session *server.Session, iden string, raw json.RawMessage) {
		handleSetDeviceParam(session, iden, raw, st)
	})

	// 10. 获取设备参数
	reg.Register("get_device_param", func(session *server.Session, iden string, raw json.RawMessage) {
		handleGetDeviceParam(session, iden, raw, st)
	})

	// 11. 重置设备参数
	reg.Register("reset_device_param", func(session *server.Session, iden string, raw json.RawMessage) {
		handleResetDeviceParam(session, iden, raw, st)
	})

	// 12. 获取当前设备名称
	reg.Register("get_curr_device_name", func(session *server.Session, iden string, raw json.RawMessage) {
		handleGetCurrDeviceName(session, iden, raw, st)
	})

	// 13. 开始扫描
	reg.Register("start_scan", func(session *server.Session, iden string, raw json.RawMessage) {
		handleStartScan(session, iden, raw, st, vs)
	})

	// 14. 停止扫描
	reg.Register("stop_scan", func(session *server.Session, iden string, raw json.RawMessage) {
		handleStopScan(session, iden, raw, st, vs)
	})

	// 15. 获取设备是否正在扫描
	reg.Register("is_device_scanning", func(session *server.Session, iden string, raw json.RawMessage) {
		handleIsDeviceScanning(session, iden, raw, st, vs)
	})
}

// 1. init_device - 设备初始化
func handleInitDevice(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	if st.GetDeviceState() != store.StateUninitialized {
		session.SendError("init_device", iden, -1, "device already initialized")
		return
	}

	st.SetDeviceState(store.StateInitialized)

	// 回复成功
	session.SendOK("init_device", iden, nil)

	// 模拟设备到达事件
	time.Sleep(100 * time.Millisecond)
	session.SendEvent("device_arrive", iden, map[string]interface{}{
		"device_name": st.Cfg.DeviceName,
	})
}

// 2. deinit_device - 设备反初始化
func handleDeinitDevice(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	state := st.GetDeviceState()
	if state == store.StateUninitialized {
		session.SendError("deinit_device", iden, -1, "device not initialized")
		return
	}
	if state == store.StateScanning {
		session.SendError("deinit_device", iden, -1, "device is scanning, stop first")
		return
	}

	deviceName := st.GetCurrDevice()
	st.SetDeviceState(store.StateUninitialized)
	st.SetCurrDevice("")

	session.SendOK("deinit_device", iden, nil)

	// 发送设备移除事件
	if deviceName != "" {
		session.SendEvent("device_remove", iden, map[string]interface{}{
			"device_name": deviceName,
		})
	}
}

// 3. is_device_init - 获取设备是否已初始化
func handleIsDeviceInit(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	state := st.GetDeviceState()
	if state == store.StateUninitialized {
		session.SendError("is_device_init", iden, -1, "device not initialized")
	} else {
		session.SendOK("is_device_init", iden, nil)
	}
}

// 4. get_device_name_list - 获取设备列表
func handleGetDeviceNameList(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	if st.GetDeviceState() == store.StateUninitialized {
		session.SendError("get_device_name_list", iden, -1, "device not initialized")
		return
	}

	session.SendOK("get_device_name_list", iden, map[string]interface{}{
		"device_name_list": []string{st.Cfg.DeviceName},
	})
}

// 5. open_device - 打开设备
func handleOpenDevice(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	state := st.GetDeviceState()
	if state == store.StateUninitialized {
		session.SendError("open_device", iden, -1, "device not initialized")
		return
	}
	if state == store.StateDeviceOpened {
		session.SendError("open_device", iden, -1, "device already opened")
		return
	}
	if state == store.StateScanning {
		session.SendError("open_device", iden, -1, "device is scanning")
		return
	}

	deviceName := getString(raw, "device_name")
	if deviceName == "" {
		deviceName = st.Cfg.DeviceName // 默认打开第一个设备
	}

	st.SetCurrDevice(deviceName)
	st.SetDeviceState(store.StateDeviceOpened)

	session.SendOK("open_device", iden, nil)
}

// 6. close_device - 关闭设备
func handleCloseDevice(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	state := st.GetDeviceState()
	if state != store.StateDeviceOpened {
		session.SendError("close_device", iden, -1, "device not opened")
		return
	}

	st.SetDeviceState(store.StateInitialized)
	st.SetCurrDevice("")

	session.SendOK("close_device", iden, nil)
}

// 7. get_device_sn - 获取设备序列号
func handleGetDeviceSN(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	if st.GetDeviceState() != store.StateDeviceOpened {
		session.SendError("get_device_sn", iden, -1, "device not opened")
		return
	}

	session.SendOK("get_device_sn", iden, map[string]interface{}{
		"sn": st.Cfg.DeviceSN,
	})
}

// 8. get_device_fwversion - 获取固件版本号
func handleGetDeviceFWVersion(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	if st.GetDeviceState() != store.StateDeviceOpened {
		session.SendError("get_device_fwversion", iden, -1, "device not opened")
		return
	}

	session.SendOK("get_device_fwversion", iden, map[string]interface{}{
		"fwversion": "1.0.0",
	})
}

// 9. set_device_param - 设置设备参数
func handleSetDeviceParam(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	state := st.GetDeviceState()
	if state != store.StateDeviceOpened {
		session.SendError("set_device_param", iden, -1, "device not opened")
		return
	}
	if state == store.StateScanning {
		session.SendError("set_device_param", iden, -1, "device is scanning")
		return
	}

	var req struct {
		DeviceParam []map[string]interface{} `json:"device_param"`
	}
	if err := json.Unmarshal(raw, &req); err != nil {
		session.SendError("set_device_param", iden, -1, "invalid params")
		return
	}

	// 合并参数
	params := make(map[string]interface{})
	for _, p := range req.DeviceParam {
		name, _ := p["name"].(string)
		if name != "" {
			params[name] = p["value"]
		}
	}
	st.SetDeviceParams(params)

	session.SendOK("set_device_param", iden, nil)
}

// 10. get_device_param - 获取设备参数
func handleGetDeviceParam(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	if st.GetDeviceState() == store.StateUninitialized {
		session.SendError("get_device_param", iden, -1, "device not opened")
		return
	}

	// 返回模拟的设备参数
	deviceParam := []map[string]interface{}{
		{
			"group_name": "基本设置",
			"group_param": []map[string]interface{}{
				{
					"name":       "resolution",
					"value_type": "int",
					"value":      300,
					"range_type": "list",
					"value_list": []int{100, 150, 200, 300, 400, 600},
				},
				{
					"name":       "color_mode",
					"value_type": "string",
					"value":      "color",
					"range_type": "list",
					"value_list": []string{"color", "gray", "black_white"},
				},
				{
					"name":       "paper_size",
					"value_type": "string",
					"value":      "A4",
					"range_type": "list",
					"value_list": []string{"A4", "A3", "A5", "Letter", "Legal"},
				},
				{
					"name":       "duplex",
					"value_type": "bool",
					"value":      false,
				},
				{
					"name":       "brightness",
					"value_type": "int",
					"value":      0,
					"range_type": "min_max",
					"value_min":  -100,
					"value_max":  100,
				},
				{
					"name":       "contrast",
					"value_type": "int",
					"value":      0,
					"range_type": "min_max",
					"value_min":  -100,
					"value_max":  100,
				},
				{
					"name":       "scan-mode",
					"value_type": "string",
					"value":      "simplex",
					"range_type": "list",
					"value_list": []string{"simplex", "duplex"},
				},
				{
					"name":       "scan-count",
					"value_type": "int",
					"value":      0,
					"range_type": "min_max",
					"value_min":  0,
					"value_max":  9999,
				},
			},
		},
		{
			"group_name": "高级设置",
			"group_param": []map[string]interface{}{
				{
					"name":       "auto_crop",
					"value_type": "bool",
					"value":      true,
				},
				{
					"name":       "auto_deskew",
					"value_type": "bool",
					"value":      false,
				},
				{
					"name":       "blank_threshold",
					"value_type": "int",
					"value":      10,
					"range_type": "min_max",
					"value_min":  0,
					"value_max":  100,
				},
			},
		},
	}

	// 合并用户自定义参数
	userParams := st.GetDeviceParams()
	for _, group := range deviceParam {
		groupParams := group["group_param"].([]map[string]interface{})
		for _, param := range groupParams {
			name := param["name"].(string)
			if v, ok := userParams[name]; ok {
				param["value"] = v
			} else {
				// 额外支持下划线与中划线兼容合并
				if name == "scan-mode" {
					if vAlt, okAlt := userParams["scan_mode"]; okAlt {
						param["value"] = vAlt
					}
				} else if name == "scan-count" {
					if vAlt, okAlt := userParams["scan_count"]; okAlt {
						param["value"] = vAlt
					}
				}
			}
		}
	}

	session.SendOK("get_device_param", iden, map[string]interface{}{
		"device_param": deviceParam,
	})
}

// 11. reset_device_param - 重置设备参数
func handleResetDeviceParam(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	state := st.GetDeviceState()
	if state != store.StateDeviceOpened {
		session.SendError("reset_device_param", iden, -1, "device not opened")
		return
	}
	if state == store.StateScanning {
		session.SendError("reset_device_param", iden, -1, "device is scanning")
		return
	}

	st.ResetDeviceParams()
	session.SendOK("reset_device_param", iden, nil)
}

// 12. get_curr_device_name - 获取当前设备名称
func handleGetCurrDeviceName(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	if st.GetDeviceState() != store.StateDeviceOpened {
		session.SendError("get_curr_device_name", iden, -1, "device not opened")
		return
	}

	session.SendOK("get_curr_device_name", iden, map[string]interface{}{
		"device_name": st.GetCurrDevice(),
	})
}

// 13. start_scan - 开始扫描
func handleStartScan(session *server.Session, iden string, raw json.RawMessage, st *store.Store, vs *scanner.VirtualScanner) {
	state := st.GetDeviceState()
	if state != store.StateDeviceOpened {
		session.SendError("start_scan", iden, -1, "device not opened")
		return
	}
	if state == store.StateScanning {
		session.SendError("start_scan", iden, -1, "device is already scanning")
		return
	}

	// 解析扫描参数
	params := scanner.ScanParams{
		BlankCheck:   getBool(raw, "blank_check", false),
		LocalSave:    getBool(raw, "local_save", true),
		GetBase64:    getBool(raw, "get_base64", false),
		SavePathName: getString(raw, "save_path_name"),
	}

	// 如果 local_save 为 true, 确保保存目录存在
	if params.LocalSave {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		if err := os.MkdirAll(saveDir, 0755); err != nil {
			session.SendError("start_scan", iden, -1, fmt.Sprintf("无法创建保存目录: %v", err))
			return
		}
	}

	// 启动扫描
	vs.StartScan(session, iden, params)
}

// 14. stop_scan - 停止扫描
func handleStopScan(session *server.Session, iden string, raw json.RawMessage, st *store.Store, vs *scanner.VirtualScanner) {
	vs.StopScan(session, iden)
}

// 15. is_device_scanning - 获取设备是否正在扫描
func handleIsDeviceScanning(session *server.Session, iden string, raw json.RawMessage, st *store.Store, vs *scanner.VirtualScanner) {
	if vs.IsScanning() || st.GetDeviceState() == store.StateScanning {
		session.SendOK("is_device_scanning", iden, nil)
	} else {
		session.SendError("is_device_scanning", iden, -1, "device is not scanning")
	}
}
