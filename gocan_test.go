package gocan

import(
	"testing"
	"os"
	"time"
	"fmt"
)

var(
	bus *Bus
	f1  *Frame
	f2  *Frame
	t1  *Transceiver
	t2  *Transceiver
) 

func TestMain(m *testing.M) { 
    setup()

    retCode := m.Run()


    // call with result of m.Run()
    os.Exit(retCode)
}

func setup() {
	fmt.Println("... TEST SETUP ...")
	f1 = &Frame{Id: 1000}
	f2 = &Frame{Id: 2000}

	bus = &Bus{Name: "Bus1",
	           C: make(chan *Frame, BusCap)}

	t1 := NewTransceiver(bus, 1)
	t2 := NewTransceiver(bus, 2)

	bus.RegisterNode(t1)
	bus.RegisterNode(t2)

	go bus.Simulate()	
	go t1.Run()
	go t2.Run()

	time.Sleep(time.Millisecond * 100)

}

func TestSend(t *testing.T) {
	t1.Send(f1)	

	bus.C <- f1

	time.Sleep(time.Millisecond * 1000)


	if len(t1.Rx) != 0 {
		t.Errorf("len(t1.Rx) = %d", len(t1.Rx))
	}	

	if len(t2.Rx) != 1 {
		t.Errorf("len(t2.Rx) = %d", len(t2.Rx))
	}

	if len(bus.C) != 0 {
		t.Errorf("len(bus) = %d", len(bus.C))
	}
}