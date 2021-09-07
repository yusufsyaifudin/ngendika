package uid

type UID interface {
	NextID() (uint64, error)
}
