/*
GOCAN PKG

Simulates a CAN Network.

A basic simulation with involves three main GoRoutines:
- Bus Simulation Routine   -> Simulate(bus chan Frame) 
- Transceiver Routine      -> Run() 
- Node application Routine -> Start() (see nodes.go )

.Every node has an embedded Transceiver (see nodes.go)
.Every transceiver has a Rx and Tx channel, which are buffers of const size "BufferSize"

There are 3 main channels into play:
- Transceiver.Rx (chan Frame)
- Transceiver.Tx (chan Frame)
- Bus.C          (chan Frame)

The insertion/removal of Frames from these Channels are done by the following GoRoutines:

Transceiver    ( Transceiver.Run()     )  <- Tx  <- Node App       ( Transceiver.Send()      )
Bus Simulation ( Bus.Simulate()        )  <- Bus <- Transceiver    ( Transceiver.Run()       )
Node App       ( Transceiver.Receive() )  <- Rx  <- Bus Simulation ( Transceiver.FilterMsg() )

A simple usage program can be seen in the Example() func.
*/
package gocan

import (
	"fmt"
	"time"
	"math/rand"
	"errors"
)

//------ Package Variables ------//

var(
	BusCap     = 100 //Capacity of Bus channel
	NumNodes   = 10 //num of Timed Nodes in the Example Simulation
	BufferSize = 3  //size of Rx and Tx buffers
	ArbtDelay  = 3  //arbitration delay in ms
)


//------ Structs with its Methods ------//
type Frame struct {
	Id uint32
	Rtr bool
	Dlc uint8
	Data uint64
	TimeStamp time.Time
}

type Transceiver struct {
	Mask uint32
	Filter uint32
	Tx chan *Frame
	Rx chan *Frame
	Id int
	Bus chan *Frame
	sendingFrame *Frame	
	//state machine variables
	BusOff bool
	waitingState bool
	transmit chan bool //must be buffered(1)
}

type Bus struct {
	Name  string
	Nodes []*Transceiver
	C     chan *Frame
}

/* Formats the Frame printing */
func (f *Frame) String() string {
	return fmt.Sprintf("<ID: %d, RTR: %t, DLC: %d, Data: %X, TimeStamp: %v", 
		f.Id, f.Rtr, f.Dlc, f.Data, f.TimeStamp.Format("3:04:05"))
}

/* Called by app, requests the transceiver to send
   a frame to the Bus, may block until transceiver buffer has space */
func (t *Transceiver) Send(f *Frame) {
	t.Tx<- f
}

/* Non-Blocking Send. Called by app, requests the transceiver to send
   a frame to the Bus, if transceiver buffer is full 
   returns an error but doesnt block */
func (t *Transceiver) NBSend(f *Frame) error {
	select {
		case t.Tx<- f:
			//fmt.Println("Transceiver<", t.Id, "> Put Frame on Tx")
			return nil
		default:
			fmt.Printf("Transceiver<%d> DROPPED Frame on Send (Tx full %d)", t.Id, len(t.Tx))
			return errors.New("Transceiver DROPPED Frame on Send (Tx full")

	}
}

/* Called by app, reads a received message. May block
   caller until there is a msg to read */
func (t *Transceiver) Receive() *Frame {
	return <-t.Rx
}

/* Non-Blocking Receive. Called by app, reads a received message. If Rx Buffer is
   empty returns an error */
func (t *Transceiver) NBReceive() (*Frame, error) {
	var f *Frame
	select {
		case f = <-t.Rx:
			return f, nil
		default:
			return nil, errors.New("Transceiver Rx Buffer was empty")
	}
}

/* Called by the Bus simulation, handles a frame to the Transceiver
   to filter. If the msg passes the filter it is added to the RxBuffer.
   Also used to check if the incoming message was the last message sent,
   confirming that the msg sent was indeed transmitted. */
func (t *Transceiver) FilterMsg(f *Frame) error {
	if !t.waitingState {
		if f.Id == t.sendingFrame.Id {
			//received message was the one sent, may
			//stop trying to send it	
			t.waitingState = true
			t.transmit<- false
			return nil
		} else {
			t.transmit<- true //retransmit
			fmt.Println("Transceiver <", t.Id, "> retransmitting Frame", t.sendingFrame) //DEBUG
		}
	}
	

	maskedId := f.Id & t.Mask;
	if maskedId == t.Filter {
		select { //select is used because BUS Simulation cannot block if Rx is full
			case t.Rx <- f:
				//fmt.Println("<Transceiver", t.Id, "> Received frame ", f) //debug
				return nil
			default:		
				fmt.Println("<Transceiver", t.Id, "> DROPPED frame on FILTER (Rx Full) ", f) //debug
				return errors.New("<Transceiver> DROPPED frame on FILTER (Rx Full)")
		}
	}

	return nil
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
			//fmt.Println("<Transceiver", t.Id, "> Waiting State, len(Tx) = ", len(t.Tx)) //debug
			t.sendingFrame = <-t.Tx
			t.waitingState = false
			t.transmit<- true		
		//SENDING STATE
		} else {
			//fmt.Println("<Transceiver", t.Id, "> Sending state") //debug
			if <-t.transmit { 
				t.Bus<- t.sendingFrame
			}
		}
	}	
}

/* Sets the transceiver to its initial state */
func (t *Transceiver) Reset() {
	t.BusOff = true;

	size := len(t.Rx)
	for i:=0; i < size; i++ {
		<-t.Rx
	}

	size = len(t.Tx)
	for i:=0; i < size; i++ {
		<-t.Tx
	}

	select {
	case <-t.transmit:
	default:
	}

	t.waitingState = true
	t.Filter = 0x00
	t.Mask = 0x00
	t.sendingFrame = nil

	t.BusOff = false
}

/* Used to register a transceiver (as a node) in the Bus */
func (bus *Bus) RegisterNode(t *Transceiver) {
	//TODO check if node was already added
	bus.Nodes = append(bus.Nodes, t)
}

/* To be run on separate goroutine. Runs the bus simulation */
func (bus *Bus) Simulate() {
	fmt.Println("<> Bus Simulation Started with", len(bus.Nodes), "nodes")
	for {
		f := <-bus.C

		//Sleeps to allow other nodes to input messages
		time.Sleep(time.Millisecond * time.Duration(ArbtDelay))
		var winner *Frame

		//if another node put a frame in Bus during delay time, arbitrate
		if size := len(bus.C); size > 0 {
			winner = bus.arbitrate(f, size)
		} else {
			winner = f
		}

		bus.broadcast(winner)
	}
}

/* Arbitrates with (size)nodes on Bus, winner is the one with Lowest Id,
   others are discarded and will be sent again by the Transceiver simulation */ 
func (bus *Bus) arbitrate(f *Frame, size int) *Frame {
	//fmt.Println("Arbitrating with size", size + 1) //+1 is the first frame removed
	winner := f
	for i := 0; i < size; i++ {
		frame := <-bus.C
		if frame.Id < winner.Id {
			winner = frame
		}
	}
	//fmt.Println("winner = ", winner) //DEBUG
	return winner
}

/* Broadcasts the frame to all nodes in the Bus */
func (bus *Bus) broadcast(f *Frame)	{
	for _, t := range bus.Nodes{
		t.FilterMsg(f)	
	}
	//fmt.Println("<Bus> Broadcasted msg ", f) //debug
}

/** Cleans up the bus **/
func (bus *Bus) Clean() {
	size := len(bus.C)
	for i:=0; i < size; i++ {
		<-bus.C
	}
}


//------ Package Functions ------//
/* Generates random data for frames*/
func RandomData() uint64 {
	var data uint64
	data = ( uint64(rand.Uint32()) << 32 ) | uint64(rand.Uint32())
	return data
}

func NewTransceiver(bus *Bus, id int) *Transceiver {
	t := &Transceiver{
		Tx : make(chan *Frame, BufferSize), 
		Rx : make(chan *Frame, BufferSize),
		Bus : bus.C,
		Id: id,
		transmit: make(chan bool, 1),
	}

	bus.RegisterNode(t)
	return t
}

func NewBus(name string) *Bus {
	bus := &Bus{
		Name: name,
		C: make(chan *Frame, BusCap)
	}
	return bus
}


/* Runs a example with (NumNodes)timed nodes and a logger */
func Example() {
	fmt.Println("GoCAN example")

	//initialize
	bus := NewBus("Bus1")
	var timeds []*timed
	for i := 1; i <= NumNodes; i++ {
		node := NewTimedNode(bus, uint32(i*1000), i*10))
		timeds = append(timeds, node)
	} 
	logger := NewLogger(bus, 0)

	//run
	go bus.Simulate()
	go logger.Start()
	for _, t := range timeds {
		t2 := t //fresh variable copy
		go t2.Start()
	}

	//fmt.Scanln()
}
