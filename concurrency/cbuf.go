package main

// SeqBuf is a sequential buffer.
// It does not support conncurrent access to the buffer.
type SeqBuf struct {
	buffer []interface{}
}

// NewSeqBuf creates and initializes a SeqBuf object, and returns a pointer to it.
func NewSeqBuf() *SeqBuf {
	sb := &SeqBuf{make([]interface{}, 0)}
	return sb
}

// Insert inserts a value into the SeqBuf.
func (sb *SeqBuf) Insert(val interface{}) {
	sb.buffer = append(sb.buffer, val)
}

// Remove removes a value from the SeqBuf.
func (sb *SeqBuf) Remove() interface{} {
	val := sb.buffer[0]
	sb.buffer = sb.buffer[1:]
	return val
}

// Flush flushes all values from the SeqBuf.
func (sb *SeqBuf) Flush() {
	sb.buffer = sb.buffer[0:0]
}

// Empty tests whether the SeqBuf is empty.
func (sb *SeqBuf) Empty() bool {
	return len(sb.buffer) == 0
}

// ConcBuf is a concurrent buffer implemented with the director pattern:
// the director() goroutine decides which calling method to give access to
// by "passing a token" either through readchan (for Remove) or opchan (for all
// other operations), and waits for the method to return the token through the
// ackchan.
type ConcBuf struct {
	sb       *SeqBuf
	readchan chan int
	opchan   chan int
	ackchan  chan int
}

// NewConcBuf creates and initializes a ConcBuf object, and returns a pointer to it.
func NewConcBuf() *ConcBuf {
	cb := new(ConcBuf)  // allocating memory for cb
	cb.sb = NewSeqBuf() // instantiating sb
	cb.readchan = make(chan int)
	cb.opchan = make(chan int)
	cb.ackchan = make(chan int)
	go cb.director()
	return cb
}

func (cb *ConcBuf) director() {
	for {
		if cb.sb.Empty() {
			cb.opchan <- 1
		} else {
			select {
			case cb.opchan <- 1:
			case cb.readchan <- 1:
			}
		}
		<-cb.ackchan
	}
}

// Insert inserts a value into ConcBuf.
// This operation is non-blocking.
func (cb *ConcBuf) Insert(val interface{}) {
	<-cb.opchan
	cb.sb.Insert(val)
	cb.ackchan <- 1
}

// Remove removes a value from ConcBuf.
// This operation is blocked if the ConcBuf is empty.
func (cb *ConcBuf) Remove() interface{} {
	<-cb.readchan
	val := cb.sb.Remove()
	cb.ackchan <- 1
	return val
}

// Flush flushes all values from ConcBuf.
// This operation is non-blocking.
func (cb *ConcBuf) Flush() {
	<-cb.opchan
	cb.sb.Flush()
	cb.ackchan <- 1
}

// Empty tests whether ConcBuf is empty.
// This operation is non-blocking.
func (cb *ConcBuf) Empty() {
	<-cb.opchan
	cb.sb.Empty()
	cb.ackchan <- 1
}

/* Buf: concurrent buffer implemented using the client/server pattern:
the runServer goroutine actually does all the work on the buffer,
and each method "calls runServer" to do work on its behalf by submitting
a request to the opchan (if it's a non-blocking operation) or the readchan
(if it's a blocking operation). Note: the only blocking operation in this case
is Remove. */

// Enumeration encoding the types of the different buffer methods
// iota is called the constant generator: it assigns the integer 0 to doinsert,
// 1 to doremove, 2 to doflush, 3 to doempty
const (
	doinsert = iota
	doremove
	doflush
	doempty
)

// request encapsulates the data sent by each method to runServer
// so it can do the work for the method
type request struct {
	op        int              // doinsert, doremove, etc.
	val       interface{}      // input value for Insert
	replychan chan interface{} // output of Remove and Empty, or server is done
}

// Buf is a concurrent buffer.
type Buf struct {
	sb       *SeqBuf
	opchan   chan *request
	readchan chan *request
}

// NewBuf creates and initializes a Buf object, and returns a pointer to it.
func NewBuf() *Buf {
	buf := new(Buf)
	buf.sb = NewSeqBuf()
	buf.opchan = make(chan *request)
	buf.readchan = make(chan *request)
	go buf.runServer()
	return buf
}

func (buf *Buf) runServer() {
	for {
		var r *request
		if buf.sb.Empty() {
			r = <-buf.opchan
		} else {
			select {
			case r = <-buf.opchan:
			case r = <-buf.readchan:
			}
		}
		switch r.op {
		case doinsert:
			buf.sb.Insert(r.val)
			r.replychan <- 1
		case doremove:
			r.replychan <- buf.sb.Remove()
		case doflush:
			buf.sb.Flush()
			r.replychan <- 1
		case doempty:
			r.replychan <- buf.sb.Empty()
		}
	}
}

// Insert inserts val into buf.
func (buf *Buf) Insert(val interface{}) {
	r := &request{doinsert, val, make(chan interface{})}
	buf.opchan <- r
	<-r.replychan // wait for operation to be done
}

// Remove removes an item from buf.
func (buf *Buf) Remove() interface{} {
	r := &request{doremove, nil, make(chan interface{})}
	buf.readchan <- r
	return <-r.replychan
}

// Flush removes all items from buf.
func (buf *Buf) Flush() {
	r := &request{doflush, nil, make(chan interface{})}
	buf.opchan <- r
	<-r.replychan
}

// Empty tests whether buf is empty.
func (buf *Buf) Empty() bool {
	r := &request{doempty, nil, make(chan interface{})}
	buf.opchan <- r
	reply := <-r.replychan
	// if reply is of type bool, empty is reply cast to bool
	if empty, ok := reply.(bool); ok {
		return empty
	}
	return true // line will not be read if implemented correctly
}
