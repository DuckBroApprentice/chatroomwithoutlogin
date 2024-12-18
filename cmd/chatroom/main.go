package main

import (
	"fmt"
	"log"
	"net/http"

	_ "net/http/pprof"

	"github.com/DuckBroApprentice/chatroom/global"
	"github.com/DuckBroApprentice/chatroom/server"
)

var (
	addr   = ":2022"
	banner = `
    ____              _____
   |    |    |   /\     |
   |    |____|  /  \    | 
   |    |    | /----\   |
   |____|    |/      \  |

Go語言編程之旅 -- 一起用Go做項目:ChatRoom,start on:%s

`
)

func init() {
	global.Init() //
}

func main() {
	fmt.Printf(banner, addr)

	server.RegisterHandle() //

	log.Fatal(http.ListenAndServe(addr, nil))
}
