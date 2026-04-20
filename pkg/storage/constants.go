package storage

// FSType 定义文件系统类型
type FSType string

const (
	// FSTypeOS 使用操作系统原生文件系统
	FSTypeOS FSType = "os"

	// FSTypeMemory 使用内存文件系统(用于测试)
	FSTypeMemory FSType = "memory"

	// FSTypeReadOnly 使用只读文件系统
	FSTypeReadOnly FSType = "readonly"

	// FSTypeBasePathFS 使用带基础路径的文件系统
	FSTypeBasePathFS FSType = "basepath"
)

// 默认配置值
const (
	// DefaultBasePath 默认基础路径
	DefaultBasePath = "."

	// DefaultFSType 默认文件系统类型
	DefaultFSType = FSTypeOS
)

// 文件监听事件类型
const (
	// WatchEventCreate 文件创建事件
	WatchEventCreate = "CREATE"

	// WatchEventWrite 文件写入事件
	WatchEventWrite = "WRITE"

	// WatchEventRemove 文件删除事件
	WatchEventRemove = "REMOVE"

	// WatchEventRename 文件重命名事件
	WatchEventRename = "RENAME"

	// WatchEventChmod 文件权限变更事件
	WatchEventChmod = "CHMOD"
)
