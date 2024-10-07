package oscpixels

import (
	"fmt"
	"github.com/hypebeast/go-osc/osc"
	"os"
)

// Start some threads which will read and parse incoming OSC messages in the background.
// Return a channel which emits pointers to osc.Message structs.
// "addr" should be the hostname and port to listen on, e.g. "127.0.0.1:8765".
func GetOSCMessageStream(addr string) chan *osc.Message {
	oscMessageChan := make(chan *osc.Message, 500)

	go oscEventReader(addr, oscMessageChan)

	return oscMessageChan
}

func oscEventReader(addr string, outCh chan *osc.Message) {

	d := osc.NewStandardDispatcher()
	d.AddMsgHandler("*", func(msg *osc.Message) {
// 		osc.PrintMessage(msg)
		outCh <- msg
	})
	//d.AddMsgHandler("/tempo/setBPM", func(msg *osc.Message) {
	//	outCh <- msg
	//	osc.PrintMessage(msg)
	//})
	server := &osc.Server{
		Addr:       addr,
		Dispatcher: d,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
