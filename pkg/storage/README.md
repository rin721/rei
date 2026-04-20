# Storage

强大的通用文件服务工具库,集成多个主流开源库,提供统一的文件操作接口。

## 功能特性

### 🎯 核心能力

- **抽象文件系统** - 基于 afero,支持 OS、内存、只读等多种文件系统
- **文件监听** - 基于 fsnotify,实时监控文件变化
- **高效复制** - 基于 otiai10/copy,快速复制文件和目录
- **MIME检测** - 基于 mimetype,精准识别文件类型
- **Excel处理** - 基于 excelize,读写和操作 Excel 文件
- **图片处理** - 基于 imaging,调整大小、裁剪、转码等

### ✨ 设计优势

- ✅ 统一接口,屏蔽底层实现细节
- ✅ 并发安全,所有方法支持并发调用
- ✅ 灵活配置,支持多种文件系统类型
- ✅ 易于测试,支持内存文件系统 mock
- ✅ 功能丰富,一站式文件处理解决方案

## 安装

```bash
go get github.com/rei0721/go-scaffold/pkg/storage
```

## 快速开始

```go
package main

import (
    "fmt"
    "log"

    "github.com/rei0721/go-scaffold/pkg/storage"
)

func main() {
    // 创建文件服务实例
    cfg := &storage.Config{
        FSType:   storage.FSTypeOS,
        BasePath: "./data",
    }

    fs, err := storage.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // 写入文件
    err = fs.WriteFile("hello.txt", []byte("Hello, World!"), 0644)
    if err != nil {
        log.Fatal(err)
    }

    // 读取文件
    data, err := fs.ReadFile("hello.txt")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(string(data)) // 输出: Hello, World!
}
```

## 使用示例

### 基础文件操作

```go
// 检查文件是否存在
exists, err := fs.Exists("test.txt")

// 创建目录
err = fs.MkdirAll("path/to/dir", 0755)

// 列出目录
files, err := fs.ListDir("path/to/dir")

// 获取文件大小
size, err := fs.FileSize("test.txt")

// 删除文件
err = fs.Remove("test.txt")

// 递归删除目录
err = fs.RemoveAll("path/to/dir")
```

### 文件复制

```go
// 复制单个文件
err = fs.Copy("source.txt", "dest.txt")

// 带选项复制
err = fs.Copy("source.txt", "dest.txt",
    storage.WithPreserveTimes(true),
    storage.WithSync(true),
)

// 复制目录
err = fs.CopyDir("./source_dir", "./dest_dir")

// 跳过特定文件
err = fs.CopyDir("./source", "./dest",
    storage.WithSkip(func(path string) bool {
        return strings.HasSuffix(path, ".tmp")
    }),
)
```

### MIME 类型检测

```go
// 从文件检测
mimeType, err := fs.DetectMIME("image.jpg")
fmt.Println(mimeType) // "image/jpeg"

// 从字节数据检测
data, _ := fs.ReadFile("document.pdf")
mimeType, err := fs.DetectMIMEFromBytes(data)
fmt.Println(mimeType) // "application/pdf"
```

### 文件监听

```go
// 启动监听
err = fs.Watch("./watch_dir", func(event storage.WatchEvent) {
    fmt.Printf("[%s] %s: %s (IsDir: %v)\n",
        event.Time.Format("15:04:05"),
        event.Op,
        event.Path,
        event.IsDir,
    )
})

// 监听单个文件
err = fs.Watch("config.yaml", func(event storage.WatchEvent) {
    if event.Op == storage.WatchEventWrite {
        fmt.Println("配置文件已更新,重新加载...")
    }
})

// 停止监听
err = fs.StopWatch("./watch_dir")

// 停止所有监听
fs.StopAllWatch()
```

### Excel 文件处理

```go
// 读取 Excel
rows, err := fs.ReadExcelSheet("data.xlsx", "Sheet1")
for i, row := range rows {
    fmt.Printf("Row %d: %v\n", i, row)
}

// 创建 Excel
file := fs.CreateExcel()
file.SetCellValue("Sheet1", "A1", "姓名")
file.SetCellValue("Sheet1", "B1", "年龄")
file.SetCellValue("Sheet1", "A2", "张三")
file.SetCellValue("Sheet1", "B2", 25)

// 保存 Excel
err = fs.SaveExcel(file, "output.xlsx")

// 高级操作
file, err := fs.OpenExcel("template.xlsx")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

// 设置样式
style, _ := file.NewStyle(&excelize.Style{
    Font: &excelize.Font{Bold: true, Size: 14},
    Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0EBF5"}, Pattern: 1},
})
file.SetCellStyle("Sheet1", "A1", "B1", style)

err = fs.SaveExcel(file, "styled.xlsx")
```

### 图片处理

```go
// 调整图片大小
err = fs.ResizeImage(
    "input.jpg",
    "output.jpg",
    800,  // 宽度 (0=按比例)
    600,  // 高度 (0=按比例)
    imaging.JPEG,
)

// 裁剪图片
rect := image.Rect(100, 100, 500, 500)
err = fs.CropImage(
    "input.jpg",
    "cropped.jpg",
    rect,
    imaging.PNG,
)

// 高级图片处理
img, err := fs.OpenImage("photo.jpg")
if err != nil {
    log.Fatal(err)
}

// 使用 imaging 库进行更多操作
img = imaging.Blur(img, 2.0)
img = imaging.Sharpen(img, 1.5)
img = imaging.AdjustBrightness(img, 10)

// 保存处理后的图片
err = fs.SaveImage(img, "processed.jpg", imaging.JPEG)
```

## 配置说明

| 字段            | 类型   | 默认值     | 说明                         |
| --------------- | ------ | ---------- | ---------------------------- |
| FSType          | FSType | `FSTypeOS` | 文件系统类型                 |
| BasePath        | string | `.`        | 基础路径 (basepath 类型使用) |
| EnableWatch     | bool   | `true`     | 是否启用文件监听             |
| WatchBufferSize | int    | `100`      | 监听事件缓冲区大小           |

### 文件系统类型

- `FSTypeOS` - 操作系统原生文件系统
- `FSTypeMemory` - 内存文件系统 (用于测试)
- `FSTypeReadOnly` - 只读文件系统
- `FSTypeBasePathFS` - 带基础路径的文件系统

### 环境变量

可通过环境变量覆盖配置:

```bash
export STORAGE_FS_TYPE=os
export STORAGE_BASE_PATH=/var/data
export STORAGE_ENABLE_WATCH=true
export STORAGE_WATCH_BUFFER_SIZE=200
```

## 接口文档

### FileService 接口

主接口,提供所有文件操作方法。

**基础操作:**

- `FileSystem() afero.Fs` - 获取底层文件系统
- `ReadFile(path string) ([]byte, error)` - 读取文件
- `WriteFile(path, data, perm) error` - 写入文件
- `Remove(path string) error` - 删除文件
- `RemoveAll(path string) error` - 删除目录
- `Exists(path string) (bool, error)` - 检查存在
- `MkdirAll(path, perm) error` - 创建目录
- `IsDir(path string) (bool, error)` - 判断是否目录
- `IsFile(path string) (bool, error)` - 判断是否文件
- `FileSize(path string) (int64, error)` - 获取文件大小
- `ListDir(path string) ([]os.FileInfo, error)` - 列出目录

**文件复制:**

- `Copy(src, dst, ...opts) error` - 复制文件
- `CopyDir(src, dst, ...opts) error` - 复制目录

**MIME检测:**

- `DetectMIME(path string) (string, error)` - 检测 MIME 类型
- `DetectMIMEFromBytes(data []byte) (string, error)` - 从字节检测

**文件监听:**

- `Watch(path, handler) error` - 监听文件/目录
- `StopWatch(path string) error` - 停止监听
- `StopAllWatch()` - 停止所有监听

**Excel操作:**

- `OpenExcel(path) (*excelize.File, error)` - 打开 Excel
- `CreateExcel() *excelize.File` - 创建 Excel
- `SaveExcel(file, path) error` - 保存 Excel
- `ReadExcelSheet(path, sheet) ([][]string, error)` - 读取工作表

**图片操作:**

- `OpenImage(path) (image.Image, error)` - 打开图片
- `SaveImage(img, path, format) error` - 保存图片
- `ResizeImage(src, dst, w, h, format) error` - 调整大小
- `CropImage(src, dst, rect, format) error` - 裁剪图片

**生命周期:**

- `Close() error` - 关闭服务
- `Reload(ctx, config) error` - 重载配置

## 最佳实践

### 1. 使用 defer 关闭资源

```go
fs, err := fileservice.New(cfg)
if err != nil {
    return err
}
defer fs.Close()
```

### 2. 错误处理

```go
exists, err := fs.Exists("file.txt")
if err != nil {
    // 处理错误
    return fmt.Errorf("failed to check file: %w", err)
}
if !exists {
    // 文件不存在
    return storage.ErrPathNotFound
}
```

### 3. 测试时使用内存文件系统

```go
func TestMyFunction(t *testing.T) {
    cfg := &storage.Config{
        FSType: storage.FSTypeMemory,
    }
    fs, _ := storage.New(cfg)
    defer fs.Close()

    // 进行测试...
}
```

### 4. 文件监听的资源管理

```go
// 确保停止监听
defer fs.StopWatch("./watch_dir")

// 或使用 context 控制
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    <-ctx.Done()
    fs.StopAllWatch()
}()
```

## 线程安全

所有 FileService 方法都是并发安全的,内部使用 `sync.RWMutex` 保护并发访问。

```go
// 可以安全地在多个 goroutine 中使用
for i := 0; i < 10; i++ {
    go func(n int) {
        data := fmt.Sprintf("data-%d", n)
        fs.WriteFile(fmt.Sprintf("file-%d.txt", n), []byte(data), 0644)
    }(i)
}
```

## 依赖库

- [github.com/spf13/afero](https://github.com/spf13/afero) - 抽象文件系统
- [github.com/fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) - 文件监听
- [github.com/otiai10/copy](https://github.com/otiai10/copy) - 文件复制
- [github.com/gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) - MIME检测
- [github.com/xuri/excelize](https://github.com/xuri/excelize) - Excel处理
- [github.com/disintegration/imaging](https://github.com/disintegration/imaging) - 图片处理

## License

MIT License
