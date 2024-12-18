package server

import (
	"net/http"

	"github.com/DuckBroApprentice/chatroom/logic"
)

func RegisterHandle() {
	// 廣播消息處理
	go logic.Broadcaster.Start()

	http.HandleFunc("/", homeHandleFunc)
	http.HandleFunc("/user_list", userListHandleFunc)
	http.HandleFunc("/ws", WebSocketHandleFunc)
}
