package types

// Validator 描述具备自校验能力的对象。
type Validator interface {
	Validate() error
}

// Namer 描述可暴露稳定名称的对象。
type Namer interface {
	Name() string
}
