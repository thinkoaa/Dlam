package sub

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

var flag = fmt.Sprintf("Doraemon%d", time.Now().UnixMilli())
var TmpPort int

func StartAccept() {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	TmpPort = listener.Addr().(*net.TCPAddr).Port
	defer listener.Close()

	for {
		conn, _ := listener.Accept()
		go func() {
			defer conn.Close()
			conn.Write([]byte(flag))
		}()
	}
}

func Received(remoteAddr string, remotePort int) bool {
	address := "%s:%d"
	dialer := &net.Dialer{
		Timeout: 2 * time.Second,
	}
	conn, err := dialer.Dial("tcp", fmt.Sprintf(address, remoteAddr, remotePort))
	if err != nil {
		return false
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	if scanner.Scan() {
		return scanner.Text() == flag
	}
	return false
}
