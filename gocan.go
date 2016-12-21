package gocan

import (
	"fmt"
	"time"
)

//------ Package Variables ------//

const(
	BusCap = 100 
	NumNodes = 1
	BufferSize = 3
)

//registered nodes on the Bus
var nodes []*Transceiver


//------ Structs with Methods ------//
type Frame struct {
	Id int
	Rtr bool
	Dlc uint8
	Data [8]uint8
	TimeStamp time.Time
}

type Transceiver struct {
	mask uint32
	filter uint32
	Tx chan Frame
	Rx chan Frame
	Id int
	Bus chan Frame
	sendingFrame Frame	
	//state machine variables
	BusOff bool
	waitingState bool
	transmit chan bool //must be buffered
}

/* Called by app, requests the transceiver to send
   a frame to the Bus */
func (t *Transceiver) Send(f Frame) {
	select {
		case t.Tx<- f:
			fmt.Println("Transceiver<", t.Id, "> Put Frame on Tx")
		default:
			fmt.Println("Transceiver<", t.Id, "> DROPPED Frame on Send (Tx full", len(t.Tx), ")")

	}
}

/* Called by app, reads a received message. May block
   caller until there is a msg to read */
func (t *Transceiver) Receive() Frame {
	return <-t.Rx
}

/* Called by app, requests the number of msgs to be read
   on the received buffer. This should be called before the
   'receive' method to avoId blocking */
func (t *Transceiver) PendingMsgs() int {
	return len(t.Rx)
}

/* Sets the mask of the Transceiver */
func (t *Transceiver) SetMask(newMask uint32) {
	t.mask = newMask
}

/* Sets the filter of the Transceiver */
func (t *Transceiver) SetFilter(newFilter uint32) {
	t.filter = newFilter	
}

/* Called by the Bus simulation, handles a frame for the Transceiver
   to filter. If the msg passes the filter it is added to the RxBuffer.
   Also used to check if the incoming message was the last message sent,
   confirming that the msg sent was indeed transmitted. */
func (t *Transceiver) Filter(f Frame) {
	if !t.waitingState {
		if f.Id == t.sendingFrame.Id {
			//received message was the one sent, may
			//stop trying to send it	
			t.waitingState = true
			t.transmit<- false
			return
		} else {
			t.transmit<- true //retransmit
		}
	}
	//TODO implement mask & filter logic
	select { //select is used because BUS cannot block if Rx is full
		case t.Rx <- f:
			fmt.Println("<Transceiver", t.Id, "> Received frame ", f) //debug
		default:		
			fmt.Println("<Transceiver", t.Id, "> DROPPED frame(Rx Full) ", f) //debug
	}
	//t.Rx<- f
	//fmt.Println("<Transceiver", t.Id, "> Received frame ", f) //debug
}

/* Called by the Bus simulation, shuts off this transceiver prohibiting
   it to transfer new messages to the Bus */
func (t *Transceiver) shutFromBus() {
	t.BusOff = true
}

/* Runs the transceiver simulation logic,
   must be called as a new goroutine */
func (t *Transceiver) Run() {
	t.waitingState = true

	for !t.BusOff {
		//WAITING STATE
		if t.waitingState {
			fmt.Println("<Transceiver", t.Id, "> Waiting State, len(Tx) = ", len(t.Tx)) //debug
			t.sendingFrame = <-t.Tx
			fmt.Println("<Transceiver", t.Id, "> Removed frame from Tx")	
			t.waitingState = false
			t.transmit<- true		
		//SENDING STATE
		} else {
			//fmt.Println("<Transceiver", t.Id, "> Sending state") //debug
			if <-t.transmit { 
				t.Bus<- t.sendingFrame
				fmt.Println("<Transceiver", t.Id, "> Sent frame to bus", t.sendingFrame) //debug
			}
		}
	}	
}

func (t *Transceiver) debug() {
	fmt.Println(t)
}

//------ Package Functions ------//

/* Used to register a transceiver (as a node) in the Bus */
func RegisterNode(t *Transceiver) {
	//TODO check if node was already added
	nodes = append(nodes, t)
}

//TODO
func arbitrate() {

}

/* Broadcasts the frame to all nodes in the Bus */
func broadcast(f Frame)	{
	for _, t := range nodes {
		t.Filter(f)	
		//fmt.Println("<Bus> Broadcasted msg ", f) //debug
	}
}

/* To be run on separate goroutine. Runs the bus simulation */
func Simulate(Bus chan Frame) {
	fmt.Println("<> Bus Simulation Started with", len(nodes), "nodes")
	for {
		f := <-Bus	
		fmt.Println("<BUS> Retirada msg do chan bus")
		//DebugReport() //DEBUG
		time.Sleep(time.Millisecond)	

		//TODO arbitrate
		broadcast(f)
	}
}

func DebugReport() {
	fmt.Println("<<< DEBUG REPORT >>>")
	fmt.Println("<<< REGISTERED NODES : ", nodes)
	fmt.Println("<<< REGISTERED NODES : ", nodes)
	fmt.Println("<<< NODE STATUS")
	for _, n := range nodes {
		fmt.Println(n)
	}
}

/* Runs a example with timed nodes and logger */
func Example() {
	fmt.Println("GoCAN example")

	//initialize
	bus := make(chan Frame, BusCap)
	var timeds []*timed
	for i := 1; i <= NumNodes; i++ {
		timeds = append(timeds, NewTimedNode(bus, i*1000, i*10))
	} 
	logger := NewLogger(bus, 0)

	//register
	for _, t := range timeds {
		//Register the nodes' transceivers into the bus
		t2 := t //fresh variable copy
		RegisterNode(t2.T)
	}
	RegisterNode(logger.T)

	//run
	go Simulate(bus)
	go logger.Start()
	for _, t := range timeds {
		t2 := t //fresh variable copy
		go t2.Start()
	}

	//fmt.Scanln()
}
