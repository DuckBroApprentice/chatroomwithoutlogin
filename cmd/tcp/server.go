package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":2020")
	if err != nil {
		panic(err)
	}

	go broadcaster()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		go handleConn(conn)
	}
}

type User struct {
	ID             int
	Addr           string
	EnterAt        time.Time
	MessageChannel chan string
}

func (u *User) String() string {
	return u.Addr + ", UID:" + strconv.Itoa(u.ID) + ", Enter At:" +
		u.EnterAt.Format("2006-01-02 15:04:05+8000")
}

// 給用戶發送的消息
type Message struct {
	OwnerID int
	Content string
}

var (
	// 新用戶到來，通過該channel進行登記
	enteringChannel = make(chan *User)
	// 用戶離開，通過該channel進行登記
	leavingChannel = make(chan *User)
	// 廣播專用的用戶普通消息channel,緩沖是盡可能避免出現堵塞
	messageChannel = make(chan Message, 8)
)

// broadcaster 用於記錄聊天室用戶，並進行消息廣播:
// 1. 新用戶進來；2. 用戶普通消息；3. 用戶離開
func broadcaster() {
	users := make(map[*User]struct{})

	for {
		select {
		case user := <-enteringChannel:
			// 新用戶進入
			users[user] = struct{}{}
		case user := <-leavingChannel:
			// 用戶離開
			delete(users, user)
			// 避免goroutin洩露
			close(user.MessageChannel)
		case msg := <-messageChannel:
			// 給所有在線用戶發送消息
			for user := range users {
				if user.ID == msg.OwnerID {
					continue
				}
				user.MessageChannel <- msg.Content
			}
		}
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// 1. 新用戶進來，建構該用戶的實例
	user := &User{
		ID:             GenUserID(),
		Addr:           conn.RemoteAddr().String(),
		EnterAt:        time.Now(),
		MessageChannel: make(chan string, 8),
	}
	/*
		 2. 當前在一個新的 goroutine 中，用來進行讀取操作，因此需要開一個 goroutine 用於寫入操作
			讀寫 goroutine 之間可以透過 channel 進行通信
	*/
	go sendMessage(conn, user.MessageChannel)

	// 3. 給當前用戶發送歡迎訊息；給所有用戶告知新用戶到來
	user.MessageChannel <- "Welcome, " + user.String()
	msg := Message{
		OwnerID: user.ID,
		Content: "user:`" + strconv.Itoa(user.ID) + "` has enter",
	}
	messageChannel <- msg

	// 4. 將該記錄到全域的用戶清單中，避免用鎖
	enteringChannel <- user

	// 控制超時用戶踢出
	var userActive = make(chan struct{})
	go func() {
		d := 1 * time.Minute
		timer := time.NewTimer(d)
		for {
			select {
			case <-timer.C:
				conn.Close()
			case <-userActive:
				timer.Reset(d)
			}
		}
	}()

	// 5. 循環讀取用戶的輸入
	input := bufio.NewScanner(conn)
	for input.Scan() {
		msg.Content = strconv.Itoa(user.ID) + ":" + input.Text()
		messageChannel <- msg

		// 用户活躍
		userActive <- struct{}{}
	}

	if err := input.Err(); err != nil {
		log.Println("讀取錯誤:", err)
	}

	// 6. 用户離開
	leavingChannel <- user
	msg.Content = "user:`" + strconv.Itoa(user.ID) + "` has left"
	messageChannel <- msg
}

func sendMessage(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

// 生成用户 ID
var (
	globalID int
	idLocker sync.Mutex
)

func GenUserID() int {
	idLocker.Lock()
	defer idLocker.Unlock()

	globalID++
	return globalID
}
