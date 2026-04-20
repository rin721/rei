package service

// IDProvider defines shared ID generation used by top-level seeding helpers.
type IDProvider interface {
	NextID() (int64, error)
}
