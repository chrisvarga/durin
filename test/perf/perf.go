package token

import (
    "net"
    "log"
    "fmt"
)

func PerfDurin() {
    conn, err := net.Dial("tcp", "127.0.0.1:8045")
    if err != nil {
        log.Fatal("Failed to connect to Durin: ", err)
    }
    for i := 0; i < 10000; i++ {
       fmt.Fprintf(conn, "set %d bar\n", i)
    }
}

func PerfErebor() {
    conn, err := net.Dial("tcp", "localhost:8044")
    if err != nil {
        log.Fatal("Failed to connect to Erebor")
    }
    for i := 0; i < 10000; i++ {
       fmt.Fprintf(conn, "set %d bar\n", i)
    }
}

func PerfRedis() {
    conn, err := net.Dial("tcp", "localhost:6379")
    if err != nil {
        log.Fatal("Failed to connect to Redis")
    }
    for i := 0; i < 10000; i++ {
       fmt.Fprintf(conn, "set %d bar\n", i)
    }
}
