package main

/*Important
I have modfiy the heartbeater on streadway/amqp to print send and receive frames
if you compile this yourself you won't get those printed out unless you modify the libs on
your own GOPATH
*/
import (
	"fmt"
	"github.com/streadway/amqp"
	"os"
	"strconv"
	"syscall"
	"golang.org/x/sys/unix"
	"time"
	"strings"
	"net"
)

func main() {
	scheme := os.Args[1]
	username := os.Args[2]
	password := os.Args[3]
	hostname := os.Args[4]
	vhost := os.Args[5]
	port := os.Args[6]
	// default interval
	interval := 10
	var err error
	if len(os.Args) == 8 {
		interval, err = strconv.Atoi(os.Args[7])
		if err != nil {
			panic("timeout needs to be able to convert to int")
		}
	}
	s := fmt.Sprintf("%s://%s:%s@%s:%s/%s", scheme, username, password, hostname, port, vhost)
	var config amqp.Config

	config.Heartbeat = time.Duration(interval) * time.Second
	config.Dial = func(network, addr string) (net.Conn, error) {

		raddr, err := net.ResolveIPAddr("ip", strings.Split(addr, ":")[0])
		if err != nil {
			panic(err)
		}
		port, err := strconv.Atoi(strings.Split(addr, ":")[1])
		if err != nil {
			panic(err)
		}
		tcpaddr := net.TCPAddr{raddr.IP, port, ""}
		tcp, err := net.DialTCP("tcp", nil, &tcpaddr)
		if err != nil {
			panic(err)
		}
		ff, _ := tcp.File()
		err = syscall.SetsockoptInt(int(ff.Fd()), unix.SOL_SOCKET, unix.SO_REUSEPORT, 0)

		return tcp, nil
	}
	conn, err := amqp.DialConfig(s, config)
	if err != nil {
		fmt.Printf("Failed Initializing Broker Connection to %s\n", hostname)
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		fmt.Println(err)
	}
	defer ch.Close()

	if err != nil {
		fmt.Println(err)
	}

	msgs, err := ch.Consume(
		"test-queue",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range msgs {
			fmt.Printf("Recieved Message: %s\n", d.Body)
		}
	}()

	fmt.Printf("Connected to %s://%s:%s/%s\n", scheme, hostname, port, vhost)
	t := time.Now()
	st := fmt.Sprintf("%d/%d/%d:%d:%d:%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	for {
		time.Sleep(time.Second)
		ct := time.Now()
		sct := fmt.Sprintf("%d/%d/%d:%d:%d:%d", ct.Year(), ct.Month(), ct.Day(), ct.Hour(), ct.Minute(), ct.Second())
		if conn.IsClosed() == false {
			fmt.Printf("[Connected] to:%s://%s/%s started at:%s and still connected at:%s\n", scheme, hostname, vhost, st, sct)
		} else {
			fmt.Printf("[Disconnected] to:%s://%s/%s started at:%s and died at:%s\n", scheme, hostname, vhost, sct,ct)
		}
	}

}
