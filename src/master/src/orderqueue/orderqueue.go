package orderqueue

type OrderType struct {
	To    string
	From  string
	Floor int
	Dir   int
	New   bool
}

type TimedOrder struct {
	Order OrderType
	timeStamp time.Time
}


type Queue struct {
	 list []OrderType
}

func (q Queue) Enqueue(o orderType) OrderQueue {
	q.list = append(q.list,o)
	return q
}

func (q Queue) Dequeue() (OrderQueue, OrderType, error) {
	l := len(q.list)
	if l == 0 {
		return q, OrderType{}, QueueError{"Queue is empty"}
	}
	o := q.list[0]
	q.list = q.list[1:]
	return q, o, nil
}

func NewTimedQueue() *Queue {
	l := make([]timedOrder)
	q := &Queue{l}
	return q
}

func NewQueue() *Queue {
	l := make([] OrderType)
	q := &Queue{l}
	return q
}


// Ikke sikker på om Find() er nødvendig

func (q Queue) Find(o timedOrder) (timedOrder, error) {
	for i, order := range(q.list) {
		if o.Floor == order.Floor && o.Dir == order.Dir {
			return i, order, nil
		}
	}
	return -1, OrderType{}, QueueError{"Could not find Order"}
}

type QueueError struct {
	err string
}

func (q *QueueError) Error() {
	return err
}

type OrderQueue interface {
	Enqueue(OrderType) OrderQueue
	Dequeue() (OrderQueue,OrderType,error)
}