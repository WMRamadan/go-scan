package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	ip   string
	lock *semaphore.Weighted
}

var open_ports int = 0
var closed_ports int = 0

func Rlimit() int64 {
	var limit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &limit); err != nil {
		panic(err)
	}

	return int64(limit.Cur)
}

func ScanPort(ip string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
		} else {
			fmt.Println(port, "closed")
			closed_ports++
		}
		return
	}

	conn.Close()
	fmt.Println(port, "open")
	open_ports++
}

func (ps *PortScanner) Start(f, l int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for port := f; port <= l; port++ {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

func main() {
	var ip_arg string
	var start_port string
	var end_port string

	if len(os.Args[1:]) < 1 {
		fmt.Println("Please enter an IP, start port and end port to scan!")
		return
	}

	if len(os.Args[1:]) < 2 {
		fmt.Println("Please enter a start port and end port to scan!")
		return
	}

	if len(os.Args[1:]) < 3 {
		fmt.Println("Please enter an end port to scan!")
		return
	}

	for _, arg := range os.Args[1:2] {
		fmt.Println("Scanning IP: " + arg)
		ip_arg = arg
	}

	for _, p_arg := range os.Args[2:3] {
		fmt.Println("Scanning From port: " + p_arg)
		start_port = p_arg
	}

	for _, e_arg := range os.Args[3:4] {
		fmt.Println("Scanning To port: " + e_arg)
		end_port = e_arg
	}

	ps := &PortScanner{
		ip:   ip_arg,
		lock: semaphore.NewWeighted(Rlimit()),
	}

	start_port_int, _ := strconv.Atoi(start_port)
	end_port_int, _ := strconv.Atoi(end_port)

	ps.Start(start_port_int, end_port_int, 500*time.Millisecond)
	fmt.Println("Scanned IP: " + ip_arg)
	fmt.Println("Scanned From Port: " + start_port)
	fmt.Println("Scanned To Port: " + end_port)
	fmt.Println("Number of Open Ports: " + strconv.Itoa(open_ports))
	fmt.Println("Number of Closed Ports: " + strconv.Itoa(closed_ports))
}
