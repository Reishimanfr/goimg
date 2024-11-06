package files

import (
	"container/heap"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
)

type PartialFileRecord struct {
	Path      string
	ExpiresAt int64
}

type ExpiryHeap []PartialFileRecord

type ExpiryManager struct {
	heap      ExpiryHeap
	heapMutex sync.Mutex
	log       *zap.Logger
}

func (h ExpiryHeap) Len() int           { return len(h) }
func (h ExpiryHeap) Less(i, j int) bool { return h[i].ExpiresAt < h[j].ExpiresAt }
func (h ExpiryHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *ExpiryHeap) Push(x interface{}) {
	*h = append(*h, x.(PartialFileRecord))
}

func (h *ExpiryHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func NewExpiryManager(logger *zap.Logger) *ExpiryManager {
	h := &ExpiryHeap{}
	heap.Init(h)

	return &ExpiryManager{heap: *h, log: logger}
}

func (m *ExpiryManager) AddFile(path string, expiresAt int64) {
	m.heapMutex.Lock()
	defer m.heapMutex.Unlock()

	heap.Push(&m.heap, PartialFileRecord{
		Path:      path,
		ExpiresAt: expiresAt,
	})
}

func (m *ExpiryManager) StartCleaner() {
	m.deleteExpiredFiles()

	go func() {
		for {
			m.heapMutex.Lock()

			if len(m.heap) == 0 {
				m.heapMutex.Unlock()
				time.Sleep(time.Minute)
				continue
			}

			nextExpiry := time.Unix(m.heap[0].ExpiresAt, 0)
			m.heapMutex.Unlock()

			time.Sleep(time.Until(nextExpiry))

			m.deleteExpiredFiles()
		}
	}()
}

func (m *ExpiryManager) deleteExpiredFiles() {
	m.heapMutex.Lock()
	defer m.heapMutex.Unlock()

	now := time.Now().Unix()

	for len(m.heap) > 0 && m.heap[0].ExpiresAt <= now {
		expiredFile := heap.Pop(&m.heap).(PartialFileRecord)
		// TODO: implement a function to delete files from different services
		err := os.Remove(expiredFile.Path)
		if err != nil {
			m.log.Error("Failed to delete expired file", zap.String("FilePath", expiredFile.Path), zap.Error(err))
		}
	}
}
