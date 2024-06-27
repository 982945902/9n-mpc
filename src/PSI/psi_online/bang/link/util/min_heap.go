package util

type IntType interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint64 | ~float32 | ~float64
}

type MinHeap[T IntType] []T

func (pq MinHeap[T]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !pq.less(j, i) {
			break
		}
		pq.swap(i, j)
		j = i
	}
}

func (pq MinHeap[T]) Top() T {
	return pq[0]
}

func (pq MinHeap[T]) Len() int { return len(pq) }

func (pq MinHeap[T]) less(i, j int) bool {
	return pq[i] < pq[j]
}

func (pq MinHeap[T]) swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *MinHeap[T]) Push(x T) {
	*pq = append(*pq, x)

	pq.up(pq.Len() - 1)
}

func (pq *MinHeap[T]) down(i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && pq.less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !pq.less(j, i) {
			break
		}
		pq.swap(i, j)
		i = j
	}
	return i > i0
}

func (pq *MinHeap[T]) Pop() T {
	n := pq.Len() - 1
	pq.swap(0, n)
	pq.down(0, n)
	return pq.pop()
}

func (pq *MinHeap[T]) pop() T {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq *MinHeap[T]) Remove(elem T) {
	n := pq.Len() - 1

	for i, v := range *pq {
		if v == elem {
			pq.swap(n, i)
			if !pq.down(i, n) {
				pq.up(i)
			}
		}
	}

	pq.pop()
}

func NewMinHeap[T IntType](data []T) *MinHeap[T] {
	h := &MinHeap[T]{}
	for _, v := range data {
		h.Push(v)
	}

	return h
}
