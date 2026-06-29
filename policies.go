package cavy

type Policy int

const (
	LRU Policy = iota
	LFU
	FIFO
)
