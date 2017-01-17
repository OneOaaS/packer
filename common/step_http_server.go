package common

import (
	"fmt"
	"net"
	"net/http"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step creates and runs the HTTP server that is serving files from the
// directory specified by the 'http_directory` configuration parameter in the
// template.
//
// Uses:
//   ui     packer.Ui
//
// Produces:
//   http_port int - The port the HTTP server started on.
type StepHTTPServer struct {
	HTTPDir  string
	HTTPPort uint

	l net.Listener
}

func (s *StepHTTPServer) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if s.HTTPDir == "" {
		state.Put("http_port", 0)
		return multistep.ActionContinue
	}

	var addr *net.TCPAddr
	var err error

	if s.HTTPPort == 0 {
		// Find an available TCP port for our HTTP server
		addr, err = net.ResolveTCPAddr("tcp", "0.0.0.0:0")
	} else {
		addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%d", s.HTTPPort))
	}
	if err != nil {
		err := fmt.Errorf("Error finding port to listen on: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.l, err = net.ListenTCP("tcp", addr)
	if err != nil {
		err := fmt.Errorf("Error listening on %s: %s", addr, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Starting HTTP server on %s", s.l.Addr()))

	// Start the HTTP server and run it in the background
	fileServer := http.FileServer(http.Dir(s.HTTPDir))
	server := &http.Server{Addr: s.l.Addr().String(), Handler: fileServer}
	go server.Serve(s.l)

	// Save the address into the state so it can be accessed in the future
	state.Put("http_port", uint(s.l.Addr().(*net.TCPAddr).Port))

	return multistep.ActionContinue
}

func (s *StepHTTPServer) Cleanup(multistep.StateBag) {
	if s.l != nil {
		// Close the listener so that the HTTP server stops
		s.l.Close()
	}
}
