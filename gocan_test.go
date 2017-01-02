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
	f3  *Frame
	t1  *Transceiver
	t2  *Transceiver
	t3  *Transceiver
) 

const delay = 5


func TestMain(m *testing.M) { 
    setup()

    retCode := m.Run()


    // call with result of m.Run()
    os.Exit(retCode)
}

/* ----- Internal funcs ----- */
func wait() {
	time.Sleep(time.Millisecond * delay)
}

func cleanup() {
	for i := 0; i < len(t1.Rx); i++ {
		t1.Receive()
	}

	for i := 0; i < len(t2.Rx); i++ {
		t2.Receive()
	}

	for i := 0; i < len(t3.Rx); i++ {
		t3.Receive()
	}

}

func setup() {
	f1 = &Frame{Id: 0x10}
	f2 = &Frame{Id: 0x20}
	f3 = &Frame{Id: 0x15}

	bus = &Bus{Name: "Bus1",
	           C: make(chan *Frame, BusCap)}

	t1 = NewTransceiver(bus, 1)
	t2 = NewTransceiver(bus, 2)
	t3 = NewTransceiver(bus, 3) 

	go bus.Simulate()	
	go t1.Run()
	go t2.Run()
	go t3.Run()

	time.Sleep(time.Millisecond * delay)

}

/* ------ TESTS ------ */
func TestSend(t *testing.T) {
	t1.Send(f1)	
	wait()

	if len(t1.Rx) != 0 {
		t.Errorf("T1 should not receive anything. len(t1.Rx) = %d", len(t1.Rx))
	}	

	if len(t2.Rx) != 1 {
		t.Errorf("T2 should receive the Frame. len(t2.Rx) = %d", len(t2.Rx))
	}

	if len(bus.C) != 0 {
		t.Errorf("Bus should not keep the frame. len(bus) = %d", len(bus.C))
	}

	cleanup()
}

func TestFiltering(t *testing.T) {
	//this combination of Filter and Mask allows t2 to receive anything from 0x10 to 0x1F 
	t2.Mask = 0xFFFFFFF0
	t2.Filter = 0x10 

	t1.Send(f1) //id 0x10
	wait()
	if len(t2.Rx) != 1 {
		fmt.Println(len(t2.Rx))
		t.Errorf("T2 should receive this frame")
	}	

	t1.Send(f3) //id 0x15
	wait()
	if len(t2.Rx) != 2 {
		t.Errorf("T2 should receive this frame")
	}	

	t1.Send(f2) //id 0x20
	wait()
	if len(t2.Rx) != 2 {
		t.Errorf("T2 should not receive this frame")
	}	

	if len(t1.Rx) != 0 {
		t.Errorf("T1 shouldn't receive any frame")
	}

	t2.Mask = 0x0;
	t2.Filter = 0x0;
	cleanup()
}

func TestArbitration(t *testing.T) {
	t1.Send(f2) //id 0x20
	t1.Send(f3) //id 0x15
	t2.Send(f1) //id 0x10

	wait()

	if len(t3.Rx) != 3 {
		t.Errorf("T3 should have received all frames")
		fmt.Println("len(t3.Rx) = ", len(t3.Rx))
	}

	lastFrame := t3.Receive()

	if lastFrame.Id != 0x20 {
		t.Errorf("T3 should have received ID 0x10 first")
	}


}


