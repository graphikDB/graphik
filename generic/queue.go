package generic

// create a node that holds the graphs vertex as data
type queueNode struct {
	value interface{}
	next  *queueNode
}

// create a queue data structure
type Queue struct {
	head   *queueNode
	tail   *queueNode
	length int
}

func NewQueue() *Queue {
	return &Queue{}
}

// Enqueue adds a new node to the tail of the queue
func (q *Queue) Enqueue(value interface{}) {
	q.length++
	n := &queueNode{value: value}
	// if the queue is empty, set the head and tail as the node value
	if q.tail == nil {
		q.head = n
		q.tail = n
		return
	}
	q.tail.next = n
	q.tail = n
}

// Dequeue removes the head node from the queue and returns it
func (q *Queue) Dequeue() interface{} {
	n := q.head
	// return nil, if head is empty
	if n == nil {
		return nil
	}

	q.head = q.head.next

	// if there wasn't any next node, that
	// means the queue is empty, and the tail
	// should be set to nil
	if q.head == nil {
		q.tail = nil
	}
	q.length--
	return n.value
}

func (q *Queue) Len() int {
	return q.length
}
