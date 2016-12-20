package gocan

import (
	"fmt"
	"time"
)

//------ Package Variables ------//

const BusCap = 10
//registered nodes on the Bus
var nodes []*Transceiver
var Bus chan Frame


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
	transmit bool
}

/* Called by app, requests the transceiver to send
   a frame to the Bus */
func (t *Transceiver) Send(f Frame) {
	t.Tx<- f
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
			//stop to try to send it	
			t.waitingState = true
			t.transmit = false
			return
		} else {
			t.transmit = true //retransmit
		}
	}
	//TODO implement mask & filter logic
	t.Rx<- f
	//fmt.Println("<Transceiver> Received frame ", f) //debug
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
			//fmt.Println("<Transceiver> Waiting State") //debug
			t.sendingFrame = <-t.Tx
			t.waitingState = false
			t.transmit = true

		//SENDING STATE
		} else {
			if t.transmit { //TODO optimize this with bool channel
				t.Bus<- t.sendingFrame
				//fmt.Println("<Transceiver> Sent frame ", t.sendingFrame) //debug
				t.transmit = false
			}
		}
	}	
}


//------ Package Functions ------//
func init() {
}


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
	for i := range nodes {
		nodes[i].Filter(f)	
		//fmt.Println("<Bus> Broadcasted msg ", f) //debug
	}
}

/* To be run on separate goroutine. Runs the bus simulation */
func Simulate(Bus chan Frame) {

	for {
		f := <-Bus	
		time.Sleep(time.Millisecond)	

		//TODO arbitrate

		broadcast(f)
	}
}


/* Runs a example with timed nodes and logger */
func Example() {
	fmt.Println("GoCAN example program")

	//initialize
	bus := make(chan Frame, 10)
	timed := NewTimedNode(bus, 1000, 10)
	logger := NewLogger(bus, 20)

	//register
	RegisterNode(timed.T)
	RegisterNode(logger.T)

	//run
	go Simulate(bus)
	go timed.Start()
	go logger.Start()


	fmt.Scanln()
}