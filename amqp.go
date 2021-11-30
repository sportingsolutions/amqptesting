package main

/*Important
I have modfiy the heartbeater on streadway/amqp to print send and receive frames
if you compile this yourself you won't get those printed out unless you modify the libs on
your own GOPATH
*/
import (
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	scheme   = flag.String("scheme", "amqp", "AMQP scheme")
	username = flag.String("username", "guest", "AMQP username")
	password = flag.String("password", "guest", "AMQP password")
	vhost    = flag.String("vhost", "/", "AMQP vhost")
	hostname = flag.String("hostname", "localhost", "AMQP hostname")
	port     = flag.String("port", "5672", "AMQP port")
	interval = flag.String("interval", "10", "AMQP port")
)

func init() {
	flag.Parse()
}
func main() {
	scheme := *scheme
	username := *username
	password := *password
	hostname := *hostname
	vhost := *vhost
	port := *port
	interval, _ := strconv.Atoi(*interval)

	s := fmt.Sprintf("%s://%s:%s@%s:%s/%s", scheme, username, password, hostname, port, vhost)
	fmt.Println(s)
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
			fmt.Printf("[Disconnected] to:%s://%s/%s started at:%s and died at:%s\n", scheme, hostname, vhost, sct, ct)
		}
	}

}
