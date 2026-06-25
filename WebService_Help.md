# WebService 帮助文档

## 一、接口简介

1. 通信基于 websocket 协议，端口默认设置为 38999，可以通过配置文件修改。
2. 消息为 json 字符串形式，分为请求消息、响应消息和事件消息，通过 `func` 字段标识接口的功能。
3. 请求消息中携带的字段除 `func` 外，还包含 `iden` 和对应接口的参数。`iden` 由调用者用于标识本次请求。
4. 响应消息中携带的字段除 `func` 外，还包含 `ret` 字段，表示返回值。当 `ret` 不为 0 时，还有 `err_info` 字段，表示错误信息。还包含 `iden` 和对应响应的信息字段。`iden` 和对应的请求消息的 `iden` 相同。
5. 事件消息中携带的字段除 `func` 外，还包含 `iden` 和该事件携带的信息。`iden` 和对应的请求消息的 `iden` 相同。

## 二、基础功能接口

### 1. 设置全局配置

**请求：**
```json
{
    "func": "set_global_config",
    "iden": "56fc",
    // 文件保存
    "file_save_path": "C:\\",
    "file_name_prefix": "",
    "file_name_mode": "date_time",
    // 可选值：date_time, random, sn_date_time, "folder_time_img_order"
    // date_time: 图片直接以时间命名保存；
    // random: 图片以随机值命名保存；
    // sn_date_time: 用户自定义sn+时间命名
    // Folder_time_img_order: 用户自定义文件夹+时间命名+文件夹内，图片按顺序命名保存
    // 图像保存
    "image_format": "jpg",
    // 格式：支持 jpg, bmp, png, tif, pdf, ofd, ocr-pdf, ocr-ofd
    "image_jpeg_quality": 80,
    "image_tiff_compression": "none",
    // 可选值：none, lzw, jpeg
    "image_tiff_jpeg_quality": 80,
    "image_jp2_ratio": 10.0
}
```

**响应：**
```json
{
    "func": "set_global_config",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 2. 获取全局配置

**请求：**
```json
{
    "func": "get_global_config",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_global_config",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "file_save_path": "C:\\",
    "file_name_prefix": "",
    "file_name_mode": "date_time",
    "image_format": "jpg",
    "image_jpeg_quality": 80,
    "image_tiff_compression": "lzw",
    "image_tiff_jpeg_quality": 80,
    "image_jp2_ratio": 10.0
}
```

### 3. 加载本地图像

**请求：**
```json
{
    "func": "load_local_image",
    "iden": "56fc",
    "image_path": "C:\\1.jpg"
}
```

**响应：**
```json
{
    "func": "load_local_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_base64": "xxx"
}
```

### 4. 保存本地图像

**请求：**
```json
{
    "func": "save_local_image",
    "iden": "56fc",
    "image_base64": "xxx"
}
```

**响应：**
```json
{
    "func": "save_local_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg"
}
```

### 5. 删除本地文件（为了安全，只能删除本项目生成的文件）

**请求：**
```json
{
    "func": "delete_local_file",
    "iden": "56fc",
    "file_path": "C:\\1.jpg"
}
```

**响应：**
```json
{
    "func": "delete_local_file",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 6. 清空全局文件保存目录（为了安全，只能删除本项目生成的文件）

**请求：**
```json
{
    "func": "clear_global_file_save_path",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "clear_global_file_save_path",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 7. 上传本地文件

**请求：**
```json
{
    "func": "upload_local_file",
    "iden": "56fc",
    "file_path": "C:\\1.jpg",
    "remote_file_path": "/path/1.jpg",
    "upload_mode": "http",
    "http_host": "192.168.1.100",
    "http_port": 80,
    "http_path": "/upload.php",
    "ftp_user": "xxx",
    "ftp_password": "xxx",
    "ftp_host": "192.168.1.100",
    "ftp_port": 21
}
```

**响应：**
```json
{
    "func": "upload_local_file",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 8. 合成本地图像

**请求：**
```json
{
    "func": "merge_local_image",
    "iden": "56fc",
    "image_path_list": ["C:\\1.jpg", "C:\\2.jpg"],
    "mode": "horz",
    // 可选值：horz, vert
    "align": "top",
    // 可选值：top, bottom, center, left, right
    "interval": 0,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "merge_local_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\3.jpg",
    "image_base64": "xxx"
}
```

### 9. 本地合成多页图像

**请求：**
```json
{
    "func": "local_make_multi_image",
    "iden": "56fc",
    "image_path_list": ["C:\\1.jpg", "C:\\2.jpg"],
    "format": "tif",
    // 格式：支持 tif, pdf, ofd
    "tiff_compression": "none",
    // 可选值：none, lzw, jpeg
    "tiff_jpeg_quality": 80,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_make_multi_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\3.jpg",
    "image_base64": "xxx"
}
```

### 10. 垂直分割图像

**请求：**
```json
{
    "func": "split_local_image",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "mode": "horz",
    "location": 1000,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "split_local_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path_list": ["C:\\2.jpg", "C:\\3.jpg"],
    "image_base64_list": ["xxx", "xxx"]
}
```

### 11. 本地生成压缩文件

**请求：**
```json
{
    "func": "local_make_zip_file",
    "iden": "56fc",
    "file_path_list": ["C:\\1.jpg", "C:\\2.jpg"],
    "local_save": true,
    "get_base64": false,
    "zip_path": "C:\\1.zip"
}
```

**响应：**
```json
{
    "func": "local_make_zip_file",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "zip_path": "C:\\1.zip",
    "zip_base64": "xxx"
}
```

### 12. 本地图像纠偏

**请求：**
```json
{
    "func": "local_image_deskew",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_deskew",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 13. 本地图像添加水印

**请求：**
```json
{
    "func": "local_image_add_watermark",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "text": "123456",
    "text_color": "#000000",
    "text_opacity": 80,
    "text_pos": "left",
    // 可选值：left, right, top, bottom, left_top, right_top, left_bottom, right_bottom, center, location
    "margin_left": 10,
    "margin_top": 10,
    "margin_right": 10,
    "margin_bottom": 10,
    "location_x": 100,
    "location_y": 100,
    "font_name": "xxx",
    "font_size": 20,
    "font_bold": false,
    "font_underline": false,
    "font_italic": false,
    "font_strikeout": false,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_add_watermark",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 14. 本地图像去污

**请求：**
```json
{
    "func": "local_image_decontamination",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "mode": "inside",
    // 可选值：inside, outside
    "color": "#000000",
    "x": 0,
    "y": 0,
    "width": 100,
    "height": 100,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_add_watermark",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 15. 本地图像方向校正

**请求：**
```json
{
    "func": "local_image_direction_correct",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_direction_correct",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 16. 本地图像裁剪

**请求：**
```json
{
    "func": "local_image_clip",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "x": 0,
    "y": 0,
    "width": 100,
    "height": 100,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_clip",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 17. 本地图像去底色

**请求：**
```json
{
    "func": "local_image_fade_bkcolor",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_fade_bkcolor",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 18. 本地图像调颜色

**请求：**
```json
{
    "func": "local_image_adjust_colors",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "brightness": 0,
    "contrast": 0,
    "gamma": 1.0,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_adjust_colors",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 19. 本地图像二值化

**请求：**
```json
{
    "func": "local_image_binarization",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "local_image_binarization",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

## 三、图像采集接口

### 1. 设备初始化

**请求：**
```json
{
    "func": "init_device",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "init_device",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

**事件：**
```json
{
    "func": "device_arrive",
    "iden": "56fc",
    "device_name": "scanner"
}
```
```json
{
    "func": "device_remove",
    "iden": "56fc",
    "device_name": "scanner"
}
```

### 2. 设备反初始化（需初始化后调用）

**请求：**
```json
{
    "func": "deinit_device",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "deinit_device",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 3. 获取设备是否已初始化

**请求：**
```json
{
    "func": "is_device_init",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "is_device_init",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```
> `ret` 为 0 表示已初始化，非 0 表示未初始化。

### 4. 获取设备列表（需初始化后调用）

**请求：**
```json
{
    "func": "get_device_name_list",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_device_name_list",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "device_name_list": ["scanne_1", "scanne_2"]
}
```

### 5. 打开设备（需初始化后调用，每次只能打开一台设备）

**请求：**
```json
{
    "func": "open_device",
    "iden": "56fc",
    "device_name": "scanner"
}
```
> `device_name` 为空表示打开第一台设备。

**响应：**
```json
{
    "func": "open_device",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 6. 关闭设备（需打开设备后调用）

**请求：**
```json
{
    "func": "close_device",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "close_device",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 7. 获取设备序列号（需打开设备后调用）

**请求：**
```json
{
    "func": "get_device_sn",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_device_sn",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "sn": "12345678"
}
```

### 8. 获取设备固件版本号（需打开设备后调用）

**请求：**
```json
{
    "func": "get_device_fwversion",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_device_fwversion",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "fwversion": "12345678"
}
```

### 9. 设置设备参数（需打开设备且未开始扫描时调用）

调用时有些参数可能设置成功，有些参数可能设置失败，有一个参数设置失败时就会返回非 0 的错误码和对应的错误信息。

**请求：**
```json
{
    "func": "set_device_param",
    "iden": "56fc",
    "device_param": [
        {
            "name": "xxx",
            "value": "xxx"
        },
        {
            "name": "xxx",
            "value": 0
        },
        {
            "name": "xxx",
            "value": 0.0000
        },
        {
            "name": "xxx",
            "value": true
        }
    ]
}
```

**响应：**
```json
{
    "func": "set_device_param",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 10. 获取设备参数（需打开设备后调用）

**请求：**
```json
{
    "func": "get_device_param",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_device_param",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "device_param": [
        {
            "group_name": "xxx",
            "group_param": [
                {
                    "name": "xxx",
                    "value_type": "string",
                    "value": "xxx",
                    "range_type": "list",
                    "value_list": ["xxx", "xxx"]
                },
                {
                    "name": "xxx",
                    "value_type": "int",
                    "value": 100,
                    "range_type": "list",
                    "value_list": [100, 200]
                },
                {
                    "name": "xxx",
                    "value_type": "int",
                    "value": 20,
                    "range_type": "min_max",
                    "value_min": 10,
                    "value_max": 30
                },
                {
                    "name": "xxx",
                    "value_type": "bool",
                    "value": true
                }
            ]
        }
    ]
}
```

### 11. 重置设备参数（需打开设备且未开始扫描时调用）

**请求：**
```json
{
    "func": "reset_device_param",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "reset_device_param",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 12. 获取当前设备名称（需打开设备后调用）

**请求：**
```json
{
    "func": "get_curr_device_name",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_curr_device_name",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "device_name": "scanner"
}
```

### 13. 开始扫描（需打开设备后调用）

**请求：**
```json
{
    "func": "start_scan",
    "iden": "56fc",
    "blank_check": false,
    "local_save": true,
    "get_base64": false,
    "save_path_name": "C:\\1.pdf"
}
```
> 图像保存路径，支持 pdf、tif、ofd。默认为空，表示保存到全局目录。当 `local_save` 设置为 `true`、`get_base64` 设置为 `false` 时，`scan_image` 事件中的 `image_path` 和 `image_base64` 为空。收到 `scan_end` 事件后，表示图像已保存完成。

**响应：**
```json
{
    "func": "start_scan",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

**事件：**
```json
{
    "func": "scan_begin",
    "iden": "56fc"
}
```
```json
{
    "func": "scan_end",
    "iden": "56fc"
}
```
```json
{
    "func": "scan_info",
    "iden": "56fc",
    "is_error": false,
    "info": "startscanning"
}
```
```json
{
    "func": "scan_image",
    "iden": "56fc",
    "is_blank": false,
    "image_path": "C:\\1.jpg",
    "image_base64": "xxx"
}
```

### 14. 停止扫描（需开始扫描后调用）

**请求：**
```json
{
    "func": "stop_scan",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "stop_scan",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 15. 获取设备是否正在扫描

**请求：**
```json
{
    "func": "is_device_scanning",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "is_device_scanning",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```
> `ret` 为 0 表示正在扫描，非 0 表示未在扫描。

## 四、图像业务接口

### 1. 获取批号列表

**请求：**
```json
{
    "func": "get_batch_id_list",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_batch_id_list",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "batch_id_list": ["default", "cbf1", "efc2"]
}
```

### 2. 打开批号（打开后会自动创建 default 批号，无需手动创建；关闭后会自动打开 default 批号）

**请求：**
```json
{
    "func": "open_batch",
    "iden": "56fc",
    "batch_id": "default"
}
```

**响应：**
```json
{
    "func": "open_batch",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 3. 删除批号

**请求：**
```json
{
    "func": "delete_batch",
    "iden": "56fc",
    "batch_id": "default"
}
```
> 不能删除当前批号。

**响应：**
```json
{
    "func": "delete_batch",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 4. 新建批号

新建批号后不会自动打开该批号。

**请求：**
```json
{
    "func": "new_batch",
    "iden": "56fc",
    "batch_id": "001"
}
```
> `batch_id` 为空会自动生成值。

**响应：**
```json
{
    "func": "new_batch",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "batch_id": "001"
}
```

### 5. 获取当前批号

**请求：**
```json
{
    "func": "get_curr_batch_id",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_curr_batch_id",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "batch_id": "001"
}
```

### 6. 修改批号

**请求：**
```json
{
    "func": "modify_batch_id",
    "iden": "56fc",
    "batch_id": "001",
    "new_batch_id": "002"
}
```

**响应：**
```json
{
    "func": "modify_batch_id",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 7. 绑定文件夹

**请求：**
```json
{
    "func": "bind_folder",
    "iden": "56fc",
    "folder": "C:\\",
    "name_mode": "order",
    "name_width": 1,
    "name_base": 0
}
```

**响应：**
```json
{
    "func": "bind_folder",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 8. 停止绑定文件夹

**请求：**
```json
{
    "func": "stop_bind_folder",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "stop_bind_folder",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 9. 获取图像缩略图列表

**请求：**
```json
{
    "func": "get_image_thumbnail_list",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_image_thumbnail_list",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_thumbnail_list": [
        {"image_tag": "001", "image_base64": "xxx"}
    ]
}
```

### 10. 获取图像数量

**请求：**
```json
{
    "func": "get_image_count",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "get_image_count",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_count": 2
}
```

### 11. 加载图像

**请求：**
```json
{
    "func": "load_image",
    "iden": "56fc",
    "image_index": 0
}
```
> `image_index` 从 0 开始。

**响应：**
```json
{
    "func": "load_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_tag": "001",
    "image_base64": "xxx"
}
```

### 12. 保存图像

**请求：**
```json
{
    "func": "save_image",
    "iden": "56fc",
    "image_index": 0
}
```
> `image_index` 从 0 开始。

**响应：**
```json
{
    "func": "save_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg"
}
```

### 13. 插入本地图像

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "insert_local_image",
    "iden": "56fc",
    "image_path": "C:\\1.jpg",
    "insert_pos": -1,
    "image_tag": "001"
}
```
> `insert_pos` 为 -1 表示末尾。`image_tag` 标签，可为空。

**响应：**
```json
{
    "func": "insert_local_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 14. 插入图像

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "insert_image",
    "iden": "56fc",
    "image_base64": "xxx",
    "insert_pos": -1,
    "image_tag": "001"
}
```
> `insert_pos` 为 -1 表示末尾。`image_tag` 标签，可为空。

**响应：**
```json
{
    "func": "insert_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 15. 修改图像标签

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "modify_image_tag",
    "iden": "56fc",
    "image_index_list": [0, 1, 2],
    "image_tag_list": ["001", "002", "003"]
}
```

**响应：**
```json
{
    "func": "modify_image_tag",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 16. 删除图像

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "delete_image",
    "iden": "56fc",
    "image_index_list": [0, 1, 2]
}
```

**响应：**
```json
{
    "func": "delete_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 17. 清空图像列表

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "clear_image_list",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "clear_image_list",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 18. 修改图像

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "modify_image",
    "iden": "56fc",
    "image_index": 0,
    "image_base64": "xxx"
}
```

**响应：**
```json
{
    "func": "modify_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 19. 使用本地图像修改图像

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "modify_image_by_local",
    "iden": "56fc",
    "image_index": 0,
    "image_path": "C:\\1.jpg"
}
```

**响应：**
```json
{
    "func": "modify_image_by_local",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 20. 移动图像

需返回成功后前端才修改图像列表，对应 UI 修改。如果前端不知道如何修改 UI，可刷新图像列表。

**请求：**
```json
{
    "func": "move_image",
    "iden": "56fc",
    "image_index_list": [0, 1, 2],
    "mode": "pos",
    "target": 0
}
```
> `mode` 可选值：`pos`, `index`，默认为 `index`。

**响应：**
```json
{
    "func": "move_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 21. 交换图像

需返回成功后前端才修改图像列表，对应 UI 修改。

**请求：**
```json
{
    "func": "exchange_image",
    "iden": "56fc",
    "image_index_1": 0,
    "image_index_2": 1
}
```

**响应：**
```json
{
    "func": "exchange_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 22. 图像书籍排序

需返回成功后前端才修改图像列表，对应 UI 修改。如果前端不知道如何修改 UI，可刷新图像列表。

**请求：**
```json
{
    "func": "image_book_sort",
    "iden": "56fc"
}
```

**响应：**
```json
{
    "func": "image_book_sort",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 23. 上传图像

**请求：**
```json
{
    "func": "upload_image",
    "iden": "56fc",
    "image_index": 0,
    "remote_file_path": "/path/1.jpg",
    "upload_mode": "http",
    "http_host": "192.168.1.100",
    "http_port": 80,
    "http_path": "/upload.php",
    "ftp_user": "xxx",
    "ftp_password": "xxx",
    "ftp_host": "192.168.1.100",
    "ftp_port": 21
}
```

**响应：**
```json
{
    "func": "upload_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": ""
}
```

### 24. 合成图像

**请求：**
```json
{
    "func": "merge_image",
    "iden": "56fc",
    "image_index_list": [0, 1, 2],
    "mode": "horz",
    "align": "top",
    "interval": 0,
    "local_save": true,
    "get_base64": false
}
```
> `mode` 可选值：`horz`, `vert`
> `align` 可选值：`top`, `bottom`, `center`, `left`, `right`

**响应：**
```json
{
    "func": "merge_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg",
    "image_base64": "xxx"
}
```

### 25. 合成多页图像

**请求：**
```json
{
    "func": "make_multi_image",
    "iden": "56fc",
    "image_index_list": [0, 1, 2],
    "format": "tif",
    "tiff_compression": "none",
    "tiff_jpeg_quality": 80,
    "local_save": true,
    "get_base64": false
}
```
> `format` 可选值：`tif`, `pdf`, `ofd`
> `tiff_compression` 可选值：`none`, `lzw`, `jpeg`

**响应：**
```json
{
    "func": "make_multi_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg",
    "image_base64": "xxx"
}
```

### 26. 分割图像

**请求：**
```json
{
    "func": "split_image",
    "iden": "56fc",
    "image_index": 0,
    "mode": "horz",
    "location": 1000,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "split_image",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path_list": ["C:\\1.jpg", "C:\\2.jpg"],
    "image_base64_list": "xxx"
}
```

### 27. 压缩图像

**请求：**
```json
{
    "func": "make_zip_file",
    "iden": "56fc",
    "image_index_list": [0, 1, 2],
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "make_zip_file",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "zip_path": "C:\\1.zip",
    "zip_base64": "xxx"
}
```

### 28. 图像纠偏

**请求：**
```json
{
    "func": "image_deskew",
    "iden": "56fc",
    "image_index": 0,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "image_deskew",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg",
    "image_base64": "xxx"
}
```

### 29. 图像添加水印

**请求：**
```json
{
    "func": "image_add_watermark",
    "iden": "56fc",
    "image_index": 0,
    "text": "123456",
    "text_color": "#000000",
    "text_opacity": 80,
    "text_pos": "left",
    "margin_left": 10,
    "margin_top": 10,
    "margin_right": 10,
    "margin_bottom": 10,
    "location_x": 100,
    "location_y": 100,
    "font_name": "xxx",
    "font_size": 20,
    "font_bold": false,
    "font_underline": false,
    "font_italic": false,
    "font_strikeout": false,
    "local_save": true,
    "get_base64": false
}
```
> `text_pos` 可选值：`left`, `right`, `top`, `bottom`, `left_top`, `right_top`, `left_bottom`, `right_bottom`, `center`, `location`

**响应：**
```json
{
    "func": "image_add_watermark",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg",
    "image_base64": "xxx"
}
```

### 30. 图像去污

**请求：**
```json
{
    "func": "image_decontamination",
    "iden": "56fc",
    "image_index": 0,
    "mode": "inside",
    "color": "#000000",
    "x": 0,
    "y": 0,
    "width": 100,
    "height": 100,
    "local_save": true,
    "get_base64": false
}
```
> `mode` 可选值：`inside`, `outside`

**响应：**
```json
{
    "func": "image_decontamination",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 31. 图像方向校正

**请求：**
```json
{
    "func": "image_direction_correct",
    "iden": "56fc",
    "image_index": 0,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "image_direction_correct",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\1.jpg",
    "image_base64": "xxx"
}
```

### 32. 图像裁剪

**请求：**
```json
{
    "func": "image_clip",
    "iden": "56fc",
    "image_index": 0,
    "x": 0,
    "y": 0,
    "width": 100,
    "height": 100,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "image_clip",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 33. 图像去底色

**请求：**
```json
{
    "func": "image_fade_bkcolor",
    "iden": "56fc",
    "image_index": 0,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "image_fade_bkcolor",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 34. 图像调颜色

**请求：**
```json
{
    "func": "image_adjust_colors",
    "iden": "56fc",
    "image_index": 0,
    "brightness": 0,
    "contrast": 0,
    "gamma": 1.0,
    "local_save": true,
    "get_base64": false
}
```
> `brightness` 取值 -255 到 255 之间
> `contrast` 取值 -127 到 127 之间
> `gamma` 取值 0.1 到 5.0 之间

**响应：**
```json
{
    "func": "image_adjust_colors",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```

### 35. 图像二值化

**请求：**
```json
{
    "func": "image_binarization",
    "iden": "56fc",
    "image_index": 0,
    "local_save": true,
    "get_base64": false
}
```

**响应：**
```json
{
    "func": "image_binarization",
    "iden": "56fc",
    "ret": 0,
    "err_info": "",
    "image_path": "C:\\2.jpg",
    "image_base64": "xxx"
}
```
