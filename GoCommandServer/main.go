package main

/*
Simple socket command server
*/
import (
	"fmt"
	"github.com/JalfResi/GoCommandServer"
	"net"
	"os"
	"strconv"
)

func EchoCommand(c *gocommandserver.CommandServer, conn net.Conn, args []string) {
	for n := 1; n < len(args); n++ {
		_, err := conn.Write([]byte(fmt.Sprintf("%s\n", args[n])))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
	}
}

func AddCommand(c *gocommandserver.CommandServer, conn net.Conn, args []string) {
	var total int64 = 0
	for n := 1; n < len(args); n++ {
		n, err := strconv.ParseInt(args[n], 10, 32)
		if err == nil {
			total = total + n
		}
	}
	conn.Write([]byte(fmt.Sprintf("%d\n", total)))
}

func main() {
	s := gocommandserver.New(1, 0)
	s.HandleFunc("echo", EchoCommand)
	s.HandleFunc("add", AddCommand)
	s.ListenAndServe("127.0.0.1:1201")
}
