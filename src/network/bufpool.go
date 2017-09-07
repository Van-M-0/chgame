package network

import "fmt"

type Buffer struct {
	index 		int
	buf 		[]byte
}

type BufferPool struct {
	started 		bool
	count 			int
	bfList 			[]chan *Buffer
	indexes 		map[int]int
}

func NewBufferPool() *BufferPool {
	return &BufferPool{
		indexes: make(map[int]int),
	}
}

func (bf *BufferPool) Start() {
	bf.started = true

	type item struct {
		size 	int
		cap  	int
	}

	itemlist := []item {
		{
			64, 10000,
		},
		{
			128, 10000,
		},
		{
			512, 10000,
		},
		{
			1024, 10000,
		},
		{
			2*1024, 5000,	//27520
		},
		{
			4*1024, 1000,	//4096
		},
		{
			16*1024, 500, 	//8192
		},
		{
			32*1024, 200,	//6400
		},
		{
			64*1024, 100,	//6400
		},
		{
			128*1024, 100,	//12800
		},
		{
			256*1024, 100, //25600
		},
		{
			512*1024, 30,	//15360
		},
		{
			1024*1024, 10,	//10240
		},
	}
	bf.count = len(itemlist)
	bf.bfList = make([]chan *Buffer, bf.count)

	for k, i := range itemlist {
		bf.indexes[i.size/2] = k
		bf.bfList[k] = make(chan *Buffer, i.size)
		for n := 0; n < i.size; n++ {
			bf.bfList[k] <- &Buffer{
				index: k,
				buf: make([]byte, i.cap),
			}
		}
	}
}

func (bf *BufferPool) Stop() {

}

func (bf *BufferPool) Pop(size int) *Buffer {
	index := size / 2
	smallIndex := 64/2
	if index < smallIndex {
		index = smallIndex
	}
	if i, ok := bf.indexes[index]; ok {
		return <- bf.bfList[i]
	} else {
		fmt.Println("bufpool pop err ", size)
		return nil
	}
}

func (bf *BufferPool) Push(b *Buffer) {
	bf.bfList[b.index] <- b
}
