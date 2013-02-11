package gocommandserver

import (
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
)

const LOG_FORMAT = "%s %s %s"

type CommandFunc func(c *CommandServer, conn net.Conn, args []string)

type CommandServer struct {
	versionMajor int
	versionMinor int
	commands     map[string]CommandFunc
}

func New(versionMajor, versionMinor int) *CommandServer {
	var cmds = map[string]CommandFunc{
		"quit":    ExitCommand,
		"help":    HelpCommand,
		"version": VersionCommand,
	}
	return &CommandServer{
		versionMajor: versionMajor,
		versionMinor: versionMinor,
		commands:     cmds,
	}
}

func (c *CommandServer) HandleFunc(cmd string, f CommandFunc) {
	c.commands[cmd] = f
}

func (c *CommandServer) ListenAndServe(addr string) {
	tcpAddr, err := net.ResolveTCPAddr("ip4", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		// run as a goroutine
		go handleClient(conn, c)
	}
}

func handleClient(conn net.Conn, c *CommandServer) {
	// close connection on exit
	defer conn.Close()
	var buf [512]byte
	for {
		// read upto 512 bytes
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		// Split into an slice of strings
		fields := strings.Fields(string(buf[0:n]))
		if len(fields) == 0 {
			continue
		}
		log.Printf(LOG_FORMAT, conn.RemoteAddr(), fields[0], fields[1:])
		if cmd, ok := c.commands[fields[0]]; ok {
			cmd(c, conn, fields)
		} else {
			UnknownCommand(c, conn, fields)
		}
	}
}

func ExitCommand(c *CommandServer, conn net.Conn, args []string) {
	conn.Close()
}

// Add optional single argument: (MINOR|MAJOR)?
func VersionCommand(c *CommandServer, conn net.Conn, args []string) {
	msg := fmt.Sprintf("---\r\nversion: %d.%d\r\n\r\n", c.versionMajor, c.versionMinor)
	_, err := conn.Write([]byte(fmt.Sprintf("OK %d\r\n%s", len(msg), msg)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func HelpCommand(c *CommandServer, conn net.Conn, args []string) {
	mk := make([]string, len(c.commands))
	i := 0
	for k, _ := range c.commands {
		mk[i] = k
		i++
	}
	sort.Strings(mk)
	msg := fmt.Sprintf("---\r\n%s\r\n\r\n", strings.Join(mk, "\r\n"))
	_, err := conn.Write([]byte(fmt.Sprintf("OK %d\r\n%s", len(msg), msg)))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func UnknownCommand(c *CommandServer, conn net.Conn, args []string) {
	_, err := conn.Write([]byte("UNKNOWN_COMMAND\r\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
