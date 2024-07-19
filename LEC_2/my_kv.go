// 一个简单的分布式键值存储系统（RPC服务器/客户端）
package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
)

//
// Common RPC request/reply definitions
//

// 用于 Put 请求，包含键和值
type PutArgs struct {
	Key   string
	Value string
}

// 用于 Put 请求的响应，暂时没有字段
type PutReply struct{}

// 用于 Get 请求，包含键
type GetArgs struct {
	Key string
}

// 用于 Get 请求的响应，包含值
type GetReply struct {
	Value string
}

//
// Client：客户端用于与RPC服务器通信
//

// 返回与RPC服务器的连接
func connect() *rpc.Client {
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return client
}

// 向服务器发送 Get 请求并返回对应的值
func get(key string) string {
	client := connect()
	args := GetArgs{key}
	reply := GetReply{}
	err := client.Call("KV.Get", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	client.Close()
	return reply.Value
}

// 向服务器发送 Put 请求并将键值对存储到服务器
func put(key string, val string) {
	client := connect()
	args := PutArgs{key, val}
	reply := PutReply{}
	err := client.Call("KV.Put", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	client.Close()
}

//
// Server：服务器用于处理客户端的Put和Get请求
//

type KV struct {
	mu   sync.Mutex
	data map[string]string
}

func server() {
	kv := &KV{data: map[string]string{}}
	rpcs := rpc.NewServer()
	rpcs.Register(kv)
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err == nil {
				go rpcs.ServeConn(conn)
			} else {
				break
			}
		}
		l.Close()
	}()
}

func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	reply.Value = kv.data[args.Key]

	return nil
}

func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[args.Key] = args.Value

	return nil
}

//
// main
//

func main() {
	server()

	put("subject", "6.5840")
	fmt.Printf("Put(subject, 6.5840) done\n")
	fmt.Printf("get(subject) -> %s\n", get("subject"))
}
