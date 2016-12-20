package gocan

import(
	"time"
	"fmt"
)

/* ------ STRUCTS ------ */

type timed struct {
	T *Transceiver
	periodMs int
}

type logger struct {
	T *Transceiver
	Log []Frame
}

/* ------ METHODS ------ */

/* --- timed methods --- */

func NewTimedNode(bus chan Frame, timeMs int, id int) timed {
	t := &Transceiver{
		Tx : make(chan Frame, 3), 
		Rx : make(chan Frame, 3),
		Bus : bus,
		Id: id,
	}
	
	node := timed{
		T : t,
		periodMs : timeMs,
	}

	return node 
}

func (node *timed) Start() {
	//start node transceiver
	go node.T.Run()

	ticker := time.NewTicker(time.Millisecond * time.Duration(node.periodMs))

	for tick := range ticker.C {
		node.T.Send(Frame{Id: node.T.Id*10, TimeStamp: tick})
	}
}

/* --- logger methods --- */

func NewLogger(bus chan Frame, id int) logger {
	t := &Transceiver{
		Tx : make(chan Frame, 3), 
		Rx : make(chan Frame, 3),
		Bus : bus,
		Id: id,
	}
	
	node := logger{
		T : t,
	}		

	return node
}

func (node *logger) Start() {
	//start node transceiver
	go node.T.Run()

	for {
		f := node.T.Receive()
		//append(node.Log, f)
		fmt.Println("<Logger> ", f)
	}
}
