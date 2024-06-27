package util

type MinHeapGenerics[T any] struct {
	data []T
	cmp  func(a, b any) bool
}

func (pq MinHeapGenerics[T]) up(j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !pq.less(j, i) {
			break
		}
		pq.swap(i, j)
		j = i
	}
}

func (pq MinHeapGenerics[T]) Top() T {
	return pq.data[0]
}

func (pq MinHeapGenerics[T]) Len() int { return len(pq.data) }

func (pq MinHeapGenerics[T]) Empty() bool { return len(pq.data) == 0 }

func (pq MinHeapGenerics[T]) less(i, j int) bool {
	return pq.cmp(pq.data[i], pq.data[j])
}

func (pq MinHeapGenerics[T]) swap(i, j int) {
	pq.data[i], pq.data[j] = pq.data[j], pq.data[i]
}

func (pq *MinHeapGenerics[T]) Push(x T) {
	pq.data = append(pq.data, x)

	pq.up(pq.Len() - 1)
}

func (pq *MinHeapGenerics[T]) down(i0, n int) bool {
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

func (pq *MinHeapGenerics[T]) Pop() T {
	n := pq.Len() - 1
	pq.swap(0, n)
	pq.down(0, n)
	return pq.pop()
}

func (pq *MinHeapGenerics[T]) pop() T {
	old := pq.data
	n := len(old)
	item := old[n-1]
	pq.data = old[0 : n-1]
	return item
}

func NewMinHeapGenerics[T any](data []T, cmp func(a, b any) bool) *MinHeapGenerics[T] {
	h := &MinHeapGenerics[T]{cmp: cmp}
	for _, v := range data {
		h.Push(v)
	}

	return h
}
