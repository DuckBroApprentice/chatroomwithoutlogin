package logic

import (
	"expvar"
	"fmt"
	"log"

	"github.com/DuckBroApprentice/chatroom/global"
)

func init() {
	expvar.Publish("message_queue", expvar.Func(calcMessageQueueLen))
}

func calcMessageQueueLen() interface{} {
	fmt.Println("===len=:", len(Broadcaster.messageChannel))
	return len(Broadcaster.messageChannel)
}

// broadcaster廣播器
type broadcaster struct {
	// 所有聊天室用戶
	users map[string]*User

	// 所有 channel 統一管理，可以避免外部亂用

	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	// 判斷該昵稱用戶是否可進入聊天室(重複與否):true 能，false 不能
	checkUserChannel      chan string
	checkUserCanInChannel chan bool

	// 獲取用戶列表
	requestUsersChannel chan struct{}
	usersChannel        chan []*User
}

var Broadcaster = &broadcaster{
	users: make(map[string]*User),

	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, global.MessageQueueLen),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),

	requestUsersChannel: make(chan struct{}),
	usersChannel:        make(chan []*User),
}

// Start啟動廣播器
// 需要在一個新 goroutine 中運行，因為它不會返回
func (b *broadcaster) Start() {
	for {
		select {
		case user := <-b.enteringChannel:
			// 新用戶進入
			b.users[user.NickName] = user

			OfflineProcessor.Send(user)
		case user := <-b.leavingChannel:
			// 用戶離開
			delete(b.users, user.NickName)
			// 避免 goroutine 洩露
			user.CloseMessageChannel()
		case msg := <-b.messageChannel:
			// 給所有線上用戶發送訊息
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				user.MessageChannel <- msg
			}
			OfflineProcessor.Save(msg)
		case nickname := <-b.checkUserChannel:
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}
		case <-b.requestUsersChannel:
			userList := make([]*User, 0, len(b.users))
			for _, user := range b.users {
				userList = append(userList, user)
			}

			b.usersChannel <- userList
		}
	}
}

func (b *broadcaster) UserEntering(u *User) {
	b.enteringChannel <- u
}

func (b *broadcaster) UserLeaving(u *User) {
	b.leavingChannel <- u
}

func (b *broadcaster) Broadcast(msg *Message) {
	if len(b.messageChannel) >= global.MessageQueueLen {
		log.Println("broadcast queue 滿了")
	}
	b.messageChannel <- msg
}

func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserChannel <- nickname

	return <-b.checkUserCanInChannel
}

func (b *broadcaster) GetUserList() []*User {
	b.requestUsersChannel <- struct{}{}
	return <-b.usersChannel
}
