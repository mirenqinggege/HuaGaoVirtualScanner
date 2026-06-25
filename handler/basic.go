// Package handler - 二、基础功能接口 (19个)
package handler

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"os"
	"path/filepath"

	"github.com/huagao/virtual-scanner/imaging"
	"github.com/huagao/virtual-scanner/scanner"
	"github.com/huagao/virtual-scanner/server"
	"github.com/huagao/virtual-scanner/store"
)

// registerBasicHandlers 注册基础功能处理器
func registerBasicHandlers(reg *server.HandlerRegistry, st *store.Store, vs *scanner.VirtualScanner) {
	// 1. 设置全局配置
	reg.Register("set_global_config", func(session *server.Session, iden string, raw json.RawMessage) {
		var cfg store.GlobalConfig
		if err := json.Unmarshal(raw, &cfg); err != nil {
			session.SendError("set_global_config", iden, -1, "invalid config")
			return
		}
		st.SetGlobalConfig(cfg)
		session.SendOK("set_global_config", iden, nil)
	})

	// 2. 获取全局配置
	reg.Register("get_global_config", func(session *server.Session, iden string, raw json.RawMessage) {
		cfg := st.GetGlobalConfig()
		session.SendOK("get_global_config", iden, map[string]interface{}{
			"file_save_path":         cfg.FileSavePath,
			"file_name_prefix":       cfg.FileNamePrefix,
			"file_name_mode":         cfg.FileNameMode,
			"image_format":           cfg.ImageFormat,
			"image_jpeg_quality":     cfg.ImageJpegQuality,
			"image_tiff_compression": cfg.ImageTiffCompression,
			"image_tiff_jpeg_quality": cfg.ImageTiffJpegQuality,
			"image_jp2_ratio":        cfg.ImageJp2Ratio,
		})
	})

	// 3. 加载本地图像
	reg.Register("load_local_image", func(session *server.Session, iden string, raw json.RawMessage) {
		imagePath := getString(raw, "image_path")
		data, err := os.ReadFile(imagePath)
		if err != nil {
			session.SendError("load_local_image", iden, -1, fmt.Sprintf("读取文件失败: %v", err))
			return
		}
		session.SendOK("load_local_image", iden, map[string]interface{}{
			"image_base64": base64.StdEncoding.EncodeToString(data),
		})
	})

	// 4. 保存本地图像
	reg.Register("save_local_image", func(session *server.Session, iden string, raw json.RawMessage) {
		b64 := getString(raw, "image_base64")
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			session.SendError("save_local_image", iden, -1, "invalid base64 data")
			return
		}

		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)

		// 检测格式并保持正确的扩展名
		format := getString(raw, "image_format")
		if format == "" {
			format = "jpg"
		}
		savePath := filepath.Join(saveDir, fmt.Sprintf("saved_%s.%s", iden, format))
		if err := os.WriteFile(savePath, data, 0644); err != nil {
			session.SendError("save_local_image", iden, -1, err.Error())
			return
		}
		st.MarkFileProtected(savePath)

		session.SendOK("save_local_image", iden, map[string]interface{}{
			"image_path": savePath,
		})
	})

	// 5. 删除本地文件（安全：只删除本项目生成的文件）
	reg.Register("delete_local_file", func(session *server.Session, iden string, raw json.RawMessage) {
		filePath := getString(raw, "file_path")
		if !st.IsFileProtected(filePath) {
			session.SendError("delete_local_file", iden, -1, "安全限制：只能删除本项目生成的文件")
			return
		}
		if err := os.Remove(filePath); err != nil {
			session.SendError("delete_local_file", iden, -1, err.Error())
			return
		}
		session.SendOK("delete_local_file", iden, nil)
	})

	// 6. 清空全局文件保存目录（安全：只删除本项目生成的文件）
	reg.Register("clear_global_file_save_path", func(session *server.Session, iden string, raw json.RawMessage) {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}

		entries, err := os.ReadDir(saveDir)
		if err != nil {
			session.SendError("clear_global_file_save_path", iden, -1, err.Error())
			return
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			path := filepath.Join(saveDir, entry.Name())
			if st.IsFileProtected(path) {
				os.Remove(path)
			}
		}
		session.SendOK("clear_global_file_save_path", iden, nil)
	})

	// 7. 上传本地文件
	reg.Register("upload_local_file", func(session *server.Session, iden string, raw json.RawMessage) {
		// 简化实现：模拟上传成功
		session.SendOK("upload_local_file", iden, nil)
	})

	// 8. 合成本地图像
	reg.Register("merge_local_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalMergeOrSplit(session, iden, "merge_local_image", raw, st, "merge")
	})

	// 9. 本地合成多页图像
	reg.Register("local_make_multi_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalMultiImage(session, iden, raw, st)
	})

	// 10. 垂直分割图像
	reg.Register("split_local_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalMergeOrSplit(session, iden, "split_local_image", raw, st, "split")
	})

	// 11. 本地生成压缩文件
	reg.Register("local_make_zip_file", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalZipFile(session, iden, raw, st)
	})

	// 12. 本地图像纠偏
	reg.Register("local_image_deskew", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_deskew", raw, st)
	})

	// 13. 本地图像添加水印
	reg.Register("local_image_add_watermark", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_add_watermark", raw, st)
	})

	// 14. 本地图像去污
	reg.Register("local_image_decontamination", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_decontamination", raw, st)
	})

	// 15. 本地图像方向校正
	reg.Register("local_image_direction_correct", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_direction_correct", raw, st)
	})

	// 16. 本地图像裁剪
	reg.Register("local_image_clip", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalImageClip(session, iden, raw, st)
	})

	// 17. 本地图像去底色
	reg.Register("local_image_fade_bkcolor", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_fade_bkcolor", raw, st)
	})

	// 18. 本地图像调颜色
	reg.Register("local_image_adjust_colors", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_adjust_colors", raw, st)
	})

	// 19. 本地图像二值化
	reg.Register("local_image_binarization", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLocalPassThrough(session, iden, "local_image_binarization", raw, st)
	})
}

// ─── 本地基础功能 handler 实现 ───

// handleLocalMergeOrSplit 处理本地合并/分割操作
func handleLocalMergeOrSplit(session *server.Session, iden string, funcName string, raw json.RawMessage, st *store.Store, opType string) {
	imagePaths := getStringSlice(raw, "image_path_list")
	mode := getString(raw, "mode")
	align := getString(raw, "align")
	interval := getInt(raw, "interval", 0)
	location := getInt(raw, "location", 0)
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	saveDir := st.GetGlobalConfig().FileSavePath
	if saveDir == "" {
		saveDir = st.Cfg.SaveDir
	}

	var outputData []byte
	var outputPaths []string
	var outputB64s []string

	switch opType {
	case "merge":
		if len(imagePaths) < 2 {
			session.SendError(funcName, iden, -1, "need at least 2 images")
			return
		}
		var imgs []image.Image
		for _, path := range imagePaths {
			img, _, err := imaging.LoadImage(path)
			if err != nil {
				session.SendError(funcName, iden, -1, fmt.Sprintf("load %s failed: %v", path, err))
				return
			}
			imgs = append(imgs, img)
		}
		var result image.Image
		var err error
		switch mode {
		case "vert":
			result, err = imaging.MergeVert(imgs, align, interval)
		default:
			result, err = imaging.MergeHorz(imgs, align, interval)
		}
		if err != nil {
			session.SendError(funcName, iden, -1, err.Error())
			return
		}
		outputData, _ = imaging.EncodeToJPEG(result, 80)
		if localSave {
			os.MkdirAll(saveDir, 0755)
			savePath := filepath.Join(saveDir, fmt.Sprintf("merged_%s.jpg", iden))
			os.WriteFile(savePath, outputData, 0644)
			st.MarkFileProtected(savePath)
			outputPaths = append(outputPaths, savePath)
		}
		if getBase64 && outputData != nil {
			outputB64s = append(outputB64s, base64.StdEncoding.EncodeToString(outputData))
		}

	case "split":
		if len(imagePaths) == 0 {
			session.SendError(funcName, iden, -1, "no image specified")
			return
		}
		img, _, err := imaging.LoadImage(imagePaths[0])
		if err != nil {
			session.SendError(funcName, iden, -1, fmt.Sprintf("load failed: %v", err))
			return
		}
		var part1, part2 image.Image
		switch mode {
		case "vert":
			part1, part2, err = imaging.SplitVert(img, location)
		default:
			part1, part2, err = imaging.SplitHorz(img, location)
		}
		if err != nil {
			session.SendError(funcName, iden, -1, err.Error())
			return
		}
		data1, _ := imaging.EncodeToJPEG(part1, 80)
		data2, _ := imaging.EncodeToJPEG(part2, 80)
		if localSave {
			os.MkdirAll(saveDir, 0755)
			p1 := filepath.Join(saveDir, fmt.Sprintf("split_%s_1.jpg", iden))
			p2 := filepath.Join(saveDir, fmt.Sprintf("split_%s_2.jpg", iden))
			os.WriteFile(p1, data1, 0644)
			os.WriteFile(p2, data2, 0644)
			st.MarkFileProtected(p1)
			st.MarkFileProtected(p2)
			outputPaths = append(outputPaths, p1, p2)
		}
		if getBase64 {
			outputB64s = append(outputB64s,
				base64.StdEncoding.EncodeToString(data1),
				base64.StdEncoding.EncodeToString(data2),
			)
		}
	}

	extra := make(map[string]interface{})
	if len(outputPaths) > 0 {
		if opType == "merge" {
			extra["image_path"] = outputPaths[0]
		} else {
			extra["image_path_list"] = outputPaths
		}
	}
	if len(outputB64s) > 0 {
		if opType == "merge" {
			extra["image_base64"] = outputB64s[0]
		} else {
			extra["image_base64_list"] = outputB64s
		}
	}

	session.SendOK(funcName, iden, extra)
}

// handleLocalMultiImage 处理本地合成多页图像
func handleLocalMultiImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	imagePaths := getStringSlice(raw, "image_path_list")
	format := getString(raw, "format")
	tiffCompression := getString(raw, "tiff_compression")
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	if len(imagePaths) == 0 {
		session.SendError("local_make_multi_image", iden, -1, "no images specified")
		return
	}

	var imgs []image.Image
	for _, path := range imagePaths {
		img, _, err := imaging.LoadImage(path)
		if err != nil {
			continue
		}
		imgs = append(imgs, img)
	}

	if len(imgs) == 0 {
		session.SendError("local_make_multi_image", iden, -1, "no valid images")
		return
	}

	var outputData []byte
	var outputExt string
	if format == "tif" || format == "tiff" {
		outputData, _ = imaging.MakeMultiPageTIFF(imgs)
		outputExt = ".tif"
		_ = tiffCompression
	} else {
		result, _ := imaging.MergeVert(imgs, "center", 0)
		outputData, _ = imaging.EncodeToJPEG(result, 80)
		outputExt = ".jpg"
	}

	extra := make(map[string]interface{})

	saveDir := st.GetGlobalConfig().FileSavePath
	if saveDir == "" {
		saveDir = st.Cfg.SaveDir
	}

	if localSave && outputData != nil {
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, "multi"+outputExt)
		os.WriteFile(savePath, outputData, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 && outputData != nil {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(outputData)
	}

	session.SendOK("local_make_multi_image", iden, extra)
}

// handleLocalZipFile 处理本地生成压缩文件
func handleLocalZipFile(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	filePaths := getStringSlice(raw, "file_path_list")
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)
	zipPath := getString(raw, "zip_path")

	var dataList [][]byte
	var names []string
	for _, path := range filePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		dataList = append(dataList, data)
		names = append(names, filepath.Base(path))
	}

	if len(dataList) == 0 {
		session.SendError("local_make_zip_file", iden, -1, "no valid files")
		return
	}

	zipData, err := imaging.MakeZipFile(dataList, names)
	if err != nil {
		session.SendError("local_make_zip_file", iden, -1, err.Error())
		return
	}

	extra := make(map[string]interface{})

	if localSave {
		if zipPath == "" {
			saveDir := st.GetGlobalConfig().FileSavePath
			if saveDir == "" {
				saveDir = st.Cfg.SaveDir
			}
			os.MkdirAll(saveDir, 0755)
			zipPath = filepath.Join(saveDir, fmt.Sprintf("files_%s.zip", iden))
		}
		os.WriteFile(zipPath, zipData, 0644)
		st.MarkFileProtected(zipPath)
		extra["zip_path"] = zipPath
	}
	if getBase64 {
		extra["zip_base64"] = base64.StdEncoding.EncodeToString(zipData)
	}

	session.SendOK("local_make_zip_file", iden, extra)
}

// handleLocalImageClip 处理本地图像裁剪
func handleLocalImageClip(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	imagePath := getString(raw, "image_path")
	x := getInt(raw, "x", 0)
	y := getInt(raw, "y", 0)
	w := getInt(raw, "width", 100)
	h := getInt(raw, "height", 100)
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	img, _, err := imaging.LoadImage(imagePath)
	if err != nil {
		session.SendError("local_image_clip", iden, -1, fmt.Sprintf("load failed: %v", err))
		return
	}

	clipped, err := imaging.Clip(img, x, y, w, h)
	if err != nil {
		session.SendError("local_image_clip", iden, -1, err.Error())
		return
	}

	data, _ := imaging.EncodeToJPEG(clipped, 80)
	extra := make(map[string]interface{})

	saveDir := st.GetGlobalConfig().FileSavePath
	if saveDir == "" {
		saveDir = st.Cfg.SaveDir
	}

	if localSave {
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, "clipped.jpg")
		os.WriteFile(savePath, data, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 && data != nil {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(data)
	}

	session.SendOK("local_image_clip", iden, extra)
}

// handleLocalPassThrough 本地图像处理透传（返回原图）
func handleLocalPassThrough(session *server.Session, iden string, funcName string, raw json.RawMessage, st *store.Store) {
	imagePath := getString(raw, "image_path")
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	data, err := os.ReadFile(imagePath)
	if err != nil {
		session.SendError(funcName, iden, -1, fmt.Sprintf("读取文件失败: %v", err))
		return
	}

	extra := make(map[string]interface{})

	saveDir := st.GetGlobalConfig().FileSavePath
	if saveDir == "" {
		saveDir = st.Cfg.SaveDir
	}

	if localSave {
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, fmt.Sprintf("%s_%s.jpg", funcName, iden))
		os.WriteFile(savePath, data, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(data)
	}

	session.SendOK(funcName, iden, extra)
}
