package task_manager

// Queue 使用 chan 实现的简单队列
type Queue struct {
	ch chan string
}

func newQueue() *Queue {
	return &Queue{ch: make(chan string, 10240)}
}

func (q *Queue) Push(val string) {
	q.ch <- val
}

func (q *Queue) Pop() string {
	return <- q.ch
}
