// Package handler - 四、图像业务接口 (35个)
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

// registerImageHandlers 注册图像业务相关处理器
func registerImageHandlers(reg *server.HandlerRegistry, st *store.Store, vs *scanner.VirtualScanner) {
	// 1. 获取批号列表
	reg.Register("get_batch_id_list", func(session *server.Session, iden string, raw json.RawMessage) {
		batchIDs := st.GetBatchIDList()
		session.SendOK("get_batch_id_list", iden, map[string]interface{}{
			"batch_id_list": batchIDs,
		})
	})

	// 2. 打开批号
	reg.Register("open_batch", func(session *server.Session, iden string, raw json.RawMessage) {
		batchID := getString(raw, "batch_id")
		if batchID == "" {
			batchID = "default"
		}
		if !st.HasBatch(batchID) {
			session.SendError("open_batch", iden, -1, "batch not found: "+batchID)
			return
		}
		if err := st.SetCurrentBatchID(batchID); err != nil {
			session.SendError("open_batch", iden, -1, err.Error())
			return
		}
		session.SendOK("open_batch", iden, nil)
	})

	// 3. 删除批号
	reg.Register("delete_batch", func(session *server.Session, iden string, raw json.RawMessage) {
		batchID := getString(raw, "batch_id")
		if err := st.DeleteBatch(batchID); err != nil {
			session.SendError("delete_batch", iden, -1, err.Error())
			return
		}
		session.SendOK("delete_batch", iden, nil)
	})

	// 4. 新建批号
	reg.Register("new_batch", func(session *server.Session, iden string, raw json.RawMessage) {
		batchID := getString(raw, "batch_id")
		id, err := st.NewBatch(batchID)
		if err != nil {
			session.SendError("new_batch", iden, -1, err.Error())
			return
		}
		session.SendOK("new_batch", iden, map[string]interface{}{
			"batch_id": id,
		})
	})

	// 5. 获取当前批号
	reg.Register("get_curr_batch_id", func(session *server.Session, iden string, raw json.RawMessage) {
		session.SendOK("get_curr_batch_id", iden, map[string]interface{}{
			"batch_id": st.GetCurrentBatchID(),
		})
	})

	// 6. 修改批号
	reg.Register("modify_batch_id", func(session *server.Session, iden string, raw json.RawMessage) {
		oldID := getString(raw, "batch_id")
		newID := getString(raw, "new_batch_id")
		if err := st.ModifyBatchID(oldID, newID); err != nil {
			session.SendError("modify_batch_id", iden, -1, err.Error())
			return
		}
		session.SendOK("modify_batch_id", iden, nil)
	})

	// 7. 绑定文件夹
	reg.Register("bind_folder", func(session *server.Session, iden string, raw json.RawMessage) {
		cfg := store.BindFolderConfig{
			Folder:    getString(raw, "folder"),
			NameMode:  getString(raw, "name_mode"),
			NameWidth: getInt(raw, "name_width", 1),
			NameBase:  getInt(raw, "name_base", 0),
		}
		st.BindFolder(cfg)
		session.SendOK("bind_folder", iden, nil)
	})

	// 8. 停止绑定文件夹
	reg.Register("stop_bind_folder", func(session *server.Session, iden string, raw json.RawMessage) {
		st.StopBindFolder()
		session.SendOK("stop_bind_folder", iden, nil)
	})

	// 9. 获取图像缩略图列表
	reg.Register("get_image_thumbnail_list", func(session *server.Session, iden string, raw json.RawMessage) {
		handleGetImageThumbnailList(session, iden, raw, st)
	})

	// 10. 获取图像数量
	reg.Register("get_image_count", func(session *server.Session, iden string, raw json.RawMessage) {
		session.SendOK("get_image_count", iden, map[string]interface{}{
			"image_count": st.GetImageCount(),
		})
	})

	// 11. 加载图像
	reg.Register("load_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleLoadImage(session, iden, raw, st)
	})

	// 12. 保存图像
	reg.Register("save_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleSaveImage(session, iden, raw, st)
	})

	// 13. 插入本地图像
	reg.Register("insert_local_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleInsertLocalImage(session, iden, raw, st)
	})

	// 14. 插入图像
	reg.Register("insert_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleInsertImage(session, iden, raw, st)
	})

	// 15. 修改图像标签
	reg.Register("modify_image_tag", func(session *server.Session, iden string, raw json.RawMessage) {
		handleModifyImageTag(session, iden, raw, st)
	})

	// 16. 删除图像
	reg.Register("delete_image", func(session *server.Session, iden string, raw json.RawMessage) {
		indices := getIntSlice(raw, "image_index_list")
		st.DeleteImages(indices)
		session.SendOK("delete_image", iden, nil)
	})

	// 17. 清空图像列表
	reg.Register("clear_image_list", func(session *server.Session, iden string, raw json.RawMessage) {
		st.ClearImages()
		session.SendOK("clear_image_list", iden, nil)
	})

	// 18. 修改图像
	reg.Register("modify_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleModifyImage(session, iden, raw, st)
	})

	// 19. 使用本地图像修改图像
	reg.Register("modify_image_by_local", func(session *server.Session, iden string, raw json.RawMessage) {
		handleModifyImageByLocal(session, iden, raw, st)
	})

	// 20. 移动图像
	reg.Register("move_image", func(session *server.Session, iden string, raw json.RawMessage) {
		indices := getIntSlice(raw, "image_index_list")
		mode := getString(raw, "mode")
		target := getInt(raw, "target", 0)
		st.MoveImages(indices, mode, target)
		session.SendOK("move_image", iden, nil)
	})

	// 21. 交换图像
	reg.Register("exchange_image", func(session *server.Session, iden string, raw json.RawMessage) {
		idx1 := getInt(raw, "image_index_1", 0)
		idx2 := getInt(raw, "image_index_2", 0)
		if err := st.ExchangeImages(idx1, idx2); err != nil {
			session.SendError("exchange_image", iden, -1, err.Error())
			return
		}
		session.SendOK("exchange_image", iden, nil)
	})

	// 22. 图像书籍排序
	reg.Register("image_book_sort", func(session *server.Session, iden string, raw json.RawMessage) {
		// 简化实现：不做实际排序，返回成功
		session.SendOK("image_book_sort", iden, nil)
	})

	// 23. 上传图像
	reg.Register("upload_image", func(session *server.Session, iden string, raw json.RawMessage) {
		// 简化实现：模拟上传成功
		session.SendOK("upload_image", iden, nil)
	})

	// 24. 合成图像
	reg.Register("merge_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleMergeImage(session, iden, raw, st)
	})

	// 25. 合成多页图像
	reg.Register("make_multi_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleMakeMultiImage(session, iden, raw, st)
	})

	// 26. 分割图像
	reg.Register("split_image", func(session *server.Session, iden string, raw json.RawMessage) {
		handleSplitImage(session, iden, raw, st)
	})

	// 27. 压缩图像
	reg.Register("make_zip_file", func(session *server.Session, iden string, raw json.RawMessage) {
		handleMakeZipFile(session, iden, raw, st)
	})

	// 28. 图像纠偏
	reg.Register("image_deskew", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_deskew", raw, st)
	})

	// 29. 图像添加水印
	reg.Register("image_add_watermark", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_add_watermark", raw, st)
	})

	// 30. 图像去污
	reg.Register("image_decontamination", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_decontamination", raw, st)
	})

	// 31. 图像方向校正
	reg.Register("image_direction_correct", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_direction_correct", raw, st)
	})

	// 32. 图像裁剪
	reg.Register("image_clip", func(session *server.Session, iden string, raw json.RawMessage) {
		handleImageClip(session, iden, raw, st)
	})

	// 33. 图像去底色
	reg.Register("image_fade_bkcolor", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_fade_bkcolor", raw, st)
	})

	// 34. 图像调颜色
	reg.Register("image_adjust_colors", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_adjust_colors", raw, st)
	})

	// 35. 图像二值化
	reg.Register("image_binarization", func(session *server.Session, iden string, raw json.RawMessage) {
		handlePassThroughImageOp(session, iden, "image_binarization", raw, st)
	})
}

// ─── 核心图像业务 handler 实现 ───

// handleGetImageThumbnailList 获取缩略图列表
func handleGetImageThumbnailList(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	images := st.GetImageList()
	thumbnailList := make([]map[string]interface{}, 0)

	for _, img := range images {
		var b64 string
		if img.Data != nil {
			// 生成缩略图
			decoded, _, err := imaging.DecodeImage(img.Data, img.Format)
			if err == nil {
				thumb, err := imaging.Thumbnail(decoded, 200)
				if err == nil {
					b64, _ = imaging.ImageToBase64(thumb)
				}
			}
		}
		if b64 == "" && img.Data != nil {
			// 缩略图生成失败，使用原图
			b64 = base64.StdEncoding.EncodeToString(img.Data)
		}

		tag := img.Tag
		if tag == "" {
			tag = filepath.Base(img.FilePath)
		}

		thumbnailList = append(thumbnailList, map[string]interface{}{
			"image_tag":    tag,
			"image_base64": b64,
		})
	}

	session.SendOK("get_image_thumbnail_list", iden, map[string]interface{}{
		"image_thumbnail_list": thumbnailList,
	})
}

// handleLoadImage 加载图像（按 index 返回 base64）
func handleLoadImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	img := st.GetImage(index)
	if img == nil {
		session.SendError("load_image", iden, -1, "image not found")
		return
	}

	var b64 string
	if img.Data != nil {
		b64 = base64.StdEncoding.EncodeToString(img.Data)
	}

	session.SendOK("load_image", iden, map[string]interface{}{
		"image_tag":    img.Tag,
		"image_base64": b64,
	})
}

// handleSaveImage 保存图像（按 index 保存到文件）
func handleSaveImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	img := st.GetImage(index)
	if img == nil || img.Data == nil {
		session.SendError("save_image", iden, -1, "image not found")
		return
	}

	saveDir := st.GetGlobalConfig().FileSavePath
	if saveDir == "" {
		saveDir = st.Cfg.SaveDir
	}
	os.MkdirAll(saveDir, 0755)

	format := img.Format
	if format == "" {
		format = "jpg"
	}
	filename := fmt.Sprintf("image_%04d.%s", index, format)
	savePath := filepath.Join(saveDir, filename)

	if err := os.WriteFile(savePath, img.Data, 0644); err != nil {
		session.SendError("save_image", iden, -1, err.Error())
		return
	}
	st.MarkFileProtected(savePath)

	session.SendOK("save_image", iden, map[string]interface{}{
		"image_path": savePath,
	})
}

// handleInsertLocalImage 插入本地图像
func handleInsertLocalImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	imagePath := getString(raw, "image_path")
	insertPos := getInt(raw, "insert_pos", -1)
	tag := getString(raw, "image_tag")

	data, err := os.ReadFile(imagePath)
	if err != nil {
		session.SendError("insert_local_image", iden, -1, fmt.Sprintf("读取文件失败: %v", err))
		return
	}

	format := detectFormatFromPath(imagePath)
	st.InsertImage(insertPos, &store.ImageRecord{
		FilePath: imagePath,
		Data:     data,
		Format:   format,
		Tag:      tag,
	})

	session.SendOK("insert_local_image", iden, nil)
}

// handleInsertImage 插入图像（base64 数据）
func handleInsertImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	b64 := getString(raw, "image_base64")
	insertPos := getInt(raw, "insert_pos", -1)
	tag := getString(raw, "image_tag")

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		session.SendError("insert_image", iden, -1, "invalid base64 data")
		return
	}

	// 检测格式
	format := "jpg"
	_, decodedFormat, err := imaging.DecodeImage(data, "")
	if err == nil {
		format = decodedFormat
	}

	st.InsertImage(insertPos, &store.ImageRecord{
		Data:   data,
		Format: format,
		Tag:    tag,
	})

	session.SendOK("insert_image", iden, nil)
}

// handleModifyImageTag 修改图像标签
func handleModifyImageTag(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	indices := getIntSlice(raw, "image_index_list")
	tags := getStringSlice(raw, "image_tag_list")

	updates := make(map[int]string)
	for i, idx := range indices {
		tag := ""
		if i < len(tags) {
			tag = tags[i]
		}
		updates[idx] = tag
	}
	st.ModifyImageTag(updates)
	session.SendOK("modify_image_tag", iden, nil)
}

// handleModifyImage 修改图像（用 base64 数据替换）
func handleModifyImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	b64 := getString(raw, "image_base64")

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		session.SendError("modify_image", iden, -1, "invalid base64 data")
		return
	}

	if err := st.ModifyImageData(index, data, "jpg"); err != nil {
		session.SendError("modify_image", iden, -1, err.Error())
		return
	}
	session.SendOK("modify_image", iden, nil)
}

// handleModifyImageByLocal 使用本地图像修改图像
func handleModifyImageByLocal(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	imagePath := getString(raw, "image_path")

	data, err := os.ReadFile(imagePath)
	if err != nil {
		session.SendError("modify_image_by_local", iden, -1, fmt.Sprintf("读取文件失败: %v", err))
		return
	}

	if err := st.ModifyImageData(index, data, detectFormatFromPath(imagePath)); err != nil {
		session.SendError("modify_image_by_local", iden, -1, err.Error())
		return
	}
	st.ModifyImageFilePath(index, imagePath)
	session.SendOK("modify_image_by_local", iden, nil)
}

// ─── 图像处理 handler ───

// handleMergeImage 合成图像（水平/垂直合并）
func handleMergeImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	indices := getIntSlice(raw, "image_index_list")
	mode := getString(raw, "mode")
	align := getString(raw, "align")
	interval := getInt(raw, "interval", 0)
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	if len(indices) < 2 {
		session.SendError("merge_image", iden, -1, "need at least 2 images")
		return
	}

	// 加载指定索引的图像
	var imgs []image.Image
	for _, idx := range indices {
		rec := st.GetImage(idx)
		if rec == nil || rec.Data == nil {
			session.SendError("merge_image", iden, -1, fmt.Sprintf("image %d not found", idx))
			return
		}
		img, _, err := imaging.DecodeImage(rec.Data, rec.Format)
		if err != nil {
			session.SendError("merge_image", iden, -1, fmt.Sprintf("decode image %d failed: %v", idx, err))
			return
		}
		imgs = append(imgs, img)
	}

	// 执行合并
	var result image.Image
	var err error
	switch mode {
	case "vert":
		result, err = imaging.MergeVert(imgs, align, interval)
	default: // horz
		result, err = imaging.MergeHorz(imgs, align, interval)
	}
	if err != nil {
		session.SendError("merge_image", iden, -1, err.Error())
		return
	}

	// 编码
	jpegData, _ := imaging.EncodeToJPEG(result, 80)
	extra := make(map[string]interface{})

	if localSave {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, "merged.jpg")
		os.WriteFile(savePath, jpegData, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 && jpegData != nil {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(jpegData)
	}

	session.SendOK("merge_image", iden, extra)
}

// handleMakeMultiImage 合成多页图像
func handleMakeMultiImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	indices := getIntSlice(raw, "image_index_list")
	format := getString(raw, "format")
	tiffCompression := getString(raw, "tiff_compression")
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	if len(indices) == 0 {
		session.SendError("make_multi_image", iden, -1, "no images specified")
		return
	}

	var imgs []image.Image
	for _, idx := range indices {
		rec := st.GetImage(idx)
		if rec == nil || rec.Data == nil {
			continue
		}
		img, _, err := imaging.DecodeImage(rec.Data, rec.Format)
		if err != nil {
			continue
		}
		imgs = append(imgs, img)
	}

	if len(imgs) == 0 {
		session.SendError("make_multi_image", iden, -1, "no valid images")
		return
	}

	// 编解码
	var outputData []byte
	var outputExt string
	if format == "tif" || format == "tiff" {
		outputData, _ = imaging.MakeMultiPageTIFF(imgs)
		outputExt = ".tif"
	} else {
		// pdf/ofd 简化：合并为一张图
		result, _ := imaging.MergeVert(imgs, "center", 0)
		outputData, _ = imaging.EncodeToJPEG(result, 80)
		outputExt = ".jpg"
	}

	if tiffCompression != "" {
		_ = tiffCompression
	}

	extra := make(map[string]interface{})

	if localSave && outputData != nil {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, "multi"+outputExt)
		os.WriteFile(savePath, outputData, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 && outputData != nil {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(outputData)
	}

	session.SendOK("make_multi_image", iden, extra)
}

// handleSplitImage 分割图像
func handleSplitImage(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	mode := getString(raw, "mode")
	location := getInt(raw, "location", 0)
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	rec := st.GetImage(index)
	if rec == nil || rec.Data == nil {
		session.SendError("split_image", iden, -1, "image not found")
		return
	}

	img, _, err := imaging.DecodeImage(rec.Data, rec.Format)
	if err != nil {
		session.SendError("split_image", iden, -1, fmt.Sprintf("decode failed: %v", err))
		return
	}

	var part1, part2 image.Image
	switch mode {
	case "vert":
		part1, part2, err = imaging.SplitVert(img, location)
	default: // horz
		part1, part2, err = imaging.SplitHorz(img, location)
	}
	if err != nil {
		session.SendError("split_image", iden, -1, err.Error())
		return
	}

	data1, _ := imaging.EncodeToJPEG(part1, 80)
	data2, _ := imaging.EncodeToJPEG(part2, 80)

	extra := make(map[string]interface{})

	if localSave {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)
		path1 := filepath.Join(saveDir, "split_1.jpg")
		path2 := filepath.Join(saveDir, "split_2.jpg")
		os.WriteFile(path1, data1, 0644)
		os.WriteFile(path2, data2, 0644)
		st.MarkFileProtected(path1)
		st.MarkFileProtected(path2)
		extra["image_path_list"] = []string{path1, path2}
	}
	if getBase64 {
		extra["image_base64_list"] = []string{
			base64.StdEncoding.EncodeToString(data1),
			base64.StdEncoding.EncodeToString(data2),
		}
	}

	session.SendOK("split_image", iden, extra)
}

// handleMakeZipFile 创建压缩文件
func handleMakeZipFile(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	indices := getIntSlice(raw, "image_index_list")
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	var imagesData [][]byte
	var names []string
	for _, idx := range indices {
		rec := st.GetImage(idx)
		if rec != nil && rec.Data != nil {
			imagesData = append(imagesData, rec.Data)
			names = append(names, fmt.Sprintf("image_%03d.jpg", idx))
		}
	}

	if len(imagesData) == 0 {
		session.SendError("make_zip_file", iden, -1, "no valid images")
		return
	}

	zipData, err := imaging.MakeZipFile(imagesData, names)
	if err != nil {
		session.SendError("make_zip_file", iden, -1, err.Error())
		return
	}

	extra := make(map[string]interface{})

	if localSave {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)
		zipPath := filepath.Join(saveDir, fmt.Sprintf("images_%s.zip", iden))
		os.WriteFile(zipPath, zipData, 0644)
		st.MarkFileProtected(zipPath)
		extra["zip_path"] = zipPath
	}
	if getBase64 {
		extra["zip_base64"] = base64.StdEncoding.EncodeToString(zipData)
	}

	session.SendOK("make_zip_file", iden, extra)
}

// handleImageClip 图像裁剪
func handleImageClip(session *server.Session, iden string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	x := getInt(raw, "x", 0)
	y := getInt(raw, "y", 0)
	w := getInt(raw, "width", 100)
	h := getInt(raw, "height", 100)
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	rec := st.GetImage(index)
	if rec == nil || rec.Data == nil {
		session.SendError("image_clip", iden, -1, "image not found")
		return
	}

	img, _, err := imaging.DecodeImage(rec.Data, rec.Format)
	if err != nil {
		session.SendError("image_clip", iden, -1, fmt.Sprintf("decode failed: %v", err))
		return
	}

	clipped, err := imaging.Clip(img, x, y, w, h)
	if err != nil {
		session.SendError("image_clip", iden, -1, err.Error())
		return
	}

	data, _ := imaging.EncodeToJPEG(clipped, 80)
	extra := make(map[string]interface{})

	if localSave {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, "clipped.jpg")
		os.WriteFile(savePath, data, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 && data != nil {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(data)
	}

	session.SendOK("image_clip", iden, extra)
}

// handlePassThroughImageOp 透传图像处理（返回原图）
func handlePassThroughImageOp(session *server.Session, iden string, funcName string, raw json.RawMessage, st *store.Store) {
	index := getInt(raw, "image_index", 0)
	localSave := getBool(raw, "local_save", true)
	getBase64 := getBool(raw, "get_base64", false)

	rec := st.GetImage(index)
	if rec == nil || rec.Data == nil {
		session.SendError(funcName, iden, -1, "image not found")
		return
	}

	extra := make(map[string]interface{})

	if localSave {
		saveDir := st.GetGlobalConfig().FileSavePath
		if saveDir == "" {
			saveDir = st.Cfg.SaveDir
		}
		os.MkdirAll(saveDir, 0755)
		savePath := filepath.Join(saveDir, fmt.Sprintf("%s_%s.jpg", funcName, iden))
		os.WriteFile(savePath, rec.Data, 0644)
		st.MarkFileProtected(savePath)
		extra["image_path"] = savePath
	}
	if getBase64 {
		extra["image_base64"] = base64.StdEncoding.EncodeToString(rec.Data)
	}

	session.SendOK(funcName, iden, extra)
}

// detectFormatFromPath 从路径检测图像格式
func detectFormatFromPath(path string) string {
	ext := filepath.Ext(path)
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
