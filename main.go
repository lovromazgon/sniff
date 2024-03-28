package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net"
	"time"
	"unsafe"
)

var (
	listen  = flag.String("in", ":9093", "The address to listen on for incoming requests.")
	forward = flag.String("out", "localhost:9092", "The address to forward requests to.")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *listen)
	if err != nil {
		panic("connection error:" + err.Error())
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept Error:", err)
			continue
		}
		go handle(conn)
	}
}

func handle(src net.Conn) {
	remoteAddr := src.RemoteAddr().String()

	defer src.Close()
	if src, ok := src.(*net.TCPConn); ok {
		src.SetNoDelay(false)
	}
	var dst net.Conn
	requestsPerConnection := 0
	for {
		src.SetDeadline(time.Now().Add(10 * time.Second)) // 10 second timeout

		if requestsPerConnection >= 50 {
			return
		}
		buf := make([]byte, 8192)
		n, err := src.Read(buf)
		if err != nil {
			log(remoteAddr, "ERR READ", err.Error())
			if dst != nil {
				dst.Close()
			}
			return
		}
		request := buf[:n]
		log(remoteAddr, "REQ", "\n"+hex.Dump(request))
		if dst == nil {
			log(remoteAddr, "DIAL", "Dialing real server ...")
			dst, err = net.DialTimeout("tcp", *forward, time.Second*10)
			if err != nil {
				log(remoteAddr, "ERR DIAL", err.Error())
				src.Write(str2bytes(err.Error()))
				return
			}
			go func() {
				defer dst.Close()
				dstConn := io.TeeReader(dst, &responseDumper{remoteAddr: remoteAddr})
				n, err := io.Copy(src, dstConn) // directly transfer the data from real server to client.
				if err != nil {
					log(remoteAddr, "ERR COPY", err.Error())
				} else {
					log(remoteAddr, "COPY", fmt.Sprintf("Copied %d bytes", n))
				}
			}()
		} else {
			// Reuse the connection
		}
		dst.SetDeadline(time.Now().Add(10 * time.Second)) // 10 second timeout
		log(remoteAddr, "WRITE", "Writing to real server ...")
		_, err = dst.Write(request)
		if err != nil {
			log(remoteAddr, "ERR WRITE", err.Error())
			src.Write(str2bytes(err.Error()))
			return
		}
		requestsPerConnection++
	}
}

func log(remoteAddr string, prefix string, message string) {
	fmt.Printf("%s [%s] %s: %s\n", remoteAddr, time.Now().Format("2006-01-02 15:04:05"), prefix, message)
}

func str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

type responseDumper struct {
	remoteAddr string
}

func (d *responseDumper) Write(p []byte) (n int, err error) {
	log(d.remoteAddr, "RESPONSE", "\n"+hex.Dump(p))
	return len(p), nil
}
