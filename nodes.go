package gocan

import(
	"time"
	"fmt"
)

/* ------ STRUCTS ------ */

type timed struct {
	T *Transceiver
	periodMs uint32
}

type logger struct {
	T *Transceiver
	Log []Frame
}

/* ------ METHODS ------ */

/* --- timed methods --- */

func NewTimedNode(bus chan Frame, timeMs uint32, id int) *timed {
	t := &Transceiver{
		Tx : make(chan Frame, BufferSize), 
		Rx : make(chan Frame, BufferSize),
		Bus : bus,
		Id: id,
		transmit: make(chan bool, 1),
		Mask: 0xFFFFFFFF,
	}
	
	node := &timed{
		T : t,
		periodMs : timeMs,
	}

	return node 
}

func (node *timed) Start() {
	//start node transceiver
	go node.T.Run()
	fmt.Println("... Timed Node <", node.T.Id, "> Started")

	ticker := time.NewTicker(time.Millisecond * time.Duration(node.periodMs))

	for tick := range ticker.C {
		//fmt.Println("tick", node.periodMs) //DEBUG
		node.T.Send(Frame{Id: node.periodMs, TimeStamp: tick, Data: RandomData()}) 
		//fmt.Println("Node <", node.periodMs, "> sending Frame") //DEBUG
	}
}

/* --- logger methods --- */

func NewLogger(bus chan Frame, id int) *logger {
	t := &Transceiver{
		Tx : make(chan Frame, BufferSize), 
		Rx : make(chan Frame, BufferSize),
		Bus : bus,
		Id: id,
	}
	
	node := &logger{
		T : t,
	}		

	return node
}

func (node *logger) Start() {
	//start node transceiver
	go node.T.Run()
	fmt.Println("... Logger Started")

	for {
		f := node.T.Receive()
		//append(node.Log, f)
		fmt.Println("<Logger> ", f)
	}
}
