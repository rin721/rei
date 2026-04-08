package base

import "context"

// Metadata 描述应用或组件的基础身份信息。
type Metadata struct {
	Name    string
	Version string
	Stage   string
}

// Initializer 描述可初始化组件的最小契约。
type Initializer interface {
	Init(context.Context) error
}

// Runner 描述可运行组件的最小契约。
type Runner interface {
	Run(context.Context) error
}

// Shutdowner 描述可优雅关闭组件的最小契约。
type Shutdowner interface {
	Shutdown(context.Context) error
}
