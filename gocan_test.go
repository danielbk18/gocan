package gocan

import(
	"testing"
	"os"
	"time"
)

var(
	bus chan *Frame
	f1 *Frame
	f2 *Frame
	t1 *Transceiver
	t2 *Transceiver
) 

func TestMain(m *testing.M) { 
    setup()

    retCode := m.Run()

    teardown()

    // call with result of m.Run()
    os.Exit(retCode)
}

func setup() {

	f1 = &Frame{Id: 1000}
	f2 = &Frame{Id: 2000}

	bus := make(chan *Frame, BusCap)
	t1 = &Transceiver{
		Tx : make(chan *Frame, BufferSize), 
		Rx : make(chan *Frame, BufferSize),
		Bus : bus,
		Id: 1,
	}
	t2 = &Transceiver{
		Tx : make(chan *Frame, BufferSize), 
		Rx : make(chan *Frame, BufferSize),
		Bus : bus,
		Id: 2,
	}

	RegisterNode(t1)
	RegisterNode(t2)

	go Simulate(bus)	
	go t1.Run()
	go t2.Run()

	time.Sleep(time.Millisecond * 100)

}

func teardown() {

}

func TestSend(t *testing.T) {
	t1.Send(f1)	

	time.Sleep(time.Millisecond * 100)
	
	if len(t1.Rx) != 0 {
		t.Errorf("len(t1.Rx) = %d", len(t1.Rx))
	}	

	if len(t2.Rx) != 1 {
		t.Errorf("len(t2.Rx) = %d", len(t2.Rx))
	}

	if len(bus) != 0 {
		t.Errorf("len(bus) = %d", len(bus))
	}
}