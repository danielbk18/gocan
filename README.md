## Synopsis

A CAN (Controller Area Network) Bus simulation in GO.

## Code Example

This is a simple Timed Node implementation to send frames at a fixed frequency.

```go
type timed struct {
	T *Transceiver
	periodMs uint32
}

/* Returns a new TimedNode which sends random data messages every 'timeMs'
   as soon the Start() method is called */
func NewTimedNode(bus chan *Frame, timeMs uint32, id int) *timed {
	t := &Transceiver{
		Tx : make(chan *Frame, BufferSize), 
		Rx : make(chan *Frame, BufferSize),
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

/* Starts the Timed Node Simulation, must be run in a separate goroutine */
func (node *timed) Start() {
	//start node transceiver
	go node.T.Run()
	fmt.Println("... Timed Node <", node.T.Id, "> Started")

	ticker := time.NewTicker(time.Millisecond * time.Duration(node.periodMs))

	for tick := range ticker.C {
		//fmt.Println("tick", node.periodMs) //DEBUG
		node.T.Send(&Frame{Id: node.periodMs, TimeStamp: tick, Data: RandomData()}) 
		//fmt.Println("Node <", node.periodMs, "> sending Frame") //DEBUG
	}
}
```
This Example Program shows how to start (NumNodes) timed Nodes with a Logger and Run the Simulation

```go
func Example() {
	fmt.Println("GoCAN example")

	//initialize
	bus := make(chan *Frame, BusCap)
	var timeds []*timed
	for i := 1; i <= NumNodes; i++ {
		timeds = append(timeds, NewTimedNode(bus, uint32(i*1000), i*10))
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

}
```

## Installation

Just clone this rep into you src folder (see https://golang.org/doc/code.html for go workspace setup) and run go install.

```bash
cd $GOPATH/src
git clone https://github.com/danielbk18/gocan
cd gocan
go install
```

After this just import the pkg "gocan" into your project to use it:

```go
package main

import(
 "fmt"
 "gocan"
 )

func main() {
	gocan.Example()
	fmt.Scanln()
}
```