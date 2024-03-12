package hikvision

type HikvisionClientOptioner func(o *hikvisionOptions)

func WithPoolSize(size int) HikvisionClientOptioner {
	return func(o *hikvisionOptions) {
		o.Poolsize = size
	}
}

type hikvisionOptions struct {
	Poolsize int `json:"poolSize"`
}
