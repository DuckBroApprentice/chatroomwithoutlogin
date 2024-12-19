// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"log"
// 	"net"
// 	"strconv"
// 	"time"
// )

// type User struct {
// 	ID             int
// 	Addr           string
// 	EnterAt        time.Time
// 	MessageChannel chan string
// }

// type Message struct {
// 	OwnerID int
// 	Content string
// }

// var (
// 	//使用者到來or離開，透過對應channel進行登記
// 	enteringChannel = make(chan *User)
// 	leavingChannel  = make(chan *User)
// 	messageChannel  = make(chan Message, 8)
// )

// func main() {
// 	listener, err := net.Listen("tcp", ":2020")
// 	if err != nil {
// 		panic(err)
// 	}

// 	//broadcaster用於記錄聊天室使用者，並進行訊息廣播:
// 	//1.新使用者進來;2.使用者普通訊息;3.使用者離開
// 	go broadcaster()
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			log.Println(err)
// 			continue
// 		}

// 		go handleConn(conn)
// 	}
// }
// func broadcaster() {
// 	users := make(map[*User]struct{})

// 	for {
// 		select {
// 		case user := <-enteringChannel:
// 			//新使用者進入
// 			users[user] = struct{}{}
// 		case user := <-leavingChannel:
// 			//使用者離開
// 			delete(users, user)
// 			//避免goroutine洩露
// 			close(user.MessageChannel)
// 		case msg := <-messageChannel:
// 			//給所有線上使用者發送訊息
// 			for user := range users {
// 				if user.ID == msg.OwnerID {
// 					continue
// 				}
// 				user.MessageChannel <- msg.Content
// 			}
// 		}
// 	}
// }
// func handleConn(conn net.Conn) {
// 	defer conn.Close()

// 	//1.新使用者進來，建置該使用者的實例
// 	user := &User{
// 		ID:             GenUserID(),
// 		Addr:           conn.RemoteAddr().String(),
// 		EnterAt:        time.Now(),
// 		MessageChannel: make(chan string, 8),
// 	}
// 	//2.由於目前是在一個新的goroutine中進行讀取操作的，所以需要開一個goroutine用於
// 	//寫入操作。讀寫goroutine之間可以透過channel進行通訊
// 	go sendMessage(conn, user.MessageChannel)

// 	//3.給目前使用者發送歡迎資訊，向所有使用者告知新使用者到來
// 	user.MessageChannel <- "Welcome," + user.String()
// 	messageChannel <- "user:1`" + strconv.Itoa(user.ID) + "` has enter"
// 	//之後會改寫，目前p4-8會報錯(messageChannel  = make(chan Message, 8)通道類型不是string)

// 	//4.記錄到全域使用者清單中，避免用鎖
// 	var userActive = make(chan struct{})
// 	go func() {
// 		d := 5 * time.Minute
// 		timer := time.NewTimer(d)
// 		for {
// 			select {
// 			case <-timer.C:
// 				conn.Close()
// 			case <-userActive:
// 				timer.Reset(d)
// 			}
// 		}
// 	}()
// 	enteringChannel <- user

// 	//5.循環讀取使用者輸入
// 	input := bufio.NewScanner(conn)
// 	for input.Scan() {
// 		msg.Content <- strconv.Itoa(user.ID) + ":" + input.Text()
// 		messageChannel <- msg
// 	}
// 	//使用者活躍
// 	userActive <- struct{}{}

// 	if err := input.Err(); err != nil {
// 		log.Panicln("讀取錯誤:", err)
// 	}

// 	//6.使用者離開
// 	leavingChannel <- user
// 	messageChannel <- "user:`" + strconv.Itoa(user.ID) + "` has left"

// }

// // 2.用於給使用者發送訊息
//
//	func sendMessage(conn net.Conn, ch <-chan string) {
//		for msg := range ch {
//			fmt.Fprintln(conn, msg)
//		}
//	}
package main

import (
	"fmt"
	"time"
)

func main() {
	name := "ploaris"
	go func() {
		name = "ABC"
	}()
	fmt.Println("name is", name)

	time.Sleep(1e9)
}
