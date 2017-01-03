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
	t1.Reset()
	t2.Reset()
	t3.Reset()
	bus.Clean()
}

func setup() {
	f1 = &Frame{Id: 0x10}
	f2 = &Frame{Id: 0x20}
	f3 = &Frame{Id: 0x15}

	bus:= NewBus("Test Bus")

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
	cleanup()

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

}

func TestFiltering(t *testing.T) {
	cleanup()
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

}

func TestArbitration(t *testing.T) {
	cleanup()

	t1.Send(f2) //id 0x20
	t2.Send(f1) //id 0x10

	time.Sleep(time.Millisecond * 1000)

	if len(t3.Rx) != 2 {
		t.Errorf("T3 should have received all frames")
		fmt.Println("len(t3.Rx) = ", len(t3.Rx))
	}

	lastFrame := t3.Receive()

	if lastFrame.Id != 0x10 {
		t.Errorf("T3 should have received ID 0x10 first")
	}

}

func TestTraffic(t *testing.T) {
	cleanup()

	//t1 and t2 will store no messages
	t1.Mask = 0xFFFFFFFF;
	t1.Filter = 0x0;
	t2.Mask = 0xFFFFFFFF;
	t2.Filter = 0x0;

	finish1 := false
	finish2 := false

	var log []*Frame

	done := make(chan bool, 1)	

	//log until t1 and t2 are finished
	go func() {
		for !(finish1 && finish2) {
			f := t3.Receive()
			log = append(log, f)
		}
		done<- true
	}()

	numMsg := 10

	go func() {
		for i := 0; i < numMsg; i++ {
			t1.Send(f1)
		}
		finish1 = true
	}()

	go func() {
		for i := 0; i < numMsg; i++ {
			t2.Send(f2)		
		}
		finish2 = true
	}()

	<-done//wait until its finished

	if len(log) != 2*numMsg {
		t.Errorf("Log should have all messages sent")
		fmt.Println("len(log) = ", len(log))
	}

	count1 := 0
	count2 := 0

	for _, f := range log {
		switch f.Id {
		case 0x10:
			count1++
		case 0x20:
			count2++
		default:
			t.Errorf("Only Ids 0x10 and 0x20 should be present")
		}
	}

	if count1 != numMsg {
		t.Errorf("Should have 'numMsg' frames")
	}

	if count2 != numMsg {
		t.Errorf("Should have 'numMsg' frames")
	}

}


