package goserver

import "fmt"

func (s *GoServer) prepDLog() {
	s.Log(fmt.Sprintf("Prepping for Disk Logging"))
}
