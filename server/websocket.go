package server

import (
	"log"
	"net/http"

	"github.com/DuckBroApprentice/chatroom/logic"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	//"nhooyr.io/websocket"
	//"nhooyr.io/websocket/wsjson"
)

func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	/*
		Accept 從客戶端接受 WebSocket 握手，並將連線升級到 WebSocket。
		如果 Origin 網域與主機不同，Accept 將拒絕握手，除非設定了 InsecureSkipVerify 選項（透過第三個參數 AcceptOptions 設定）。
		換句話說，預設情況下，它不允許跨來源請求。如果發生錯誤，Accept 將始終寫入適當的回應
	*/
	/*Accept:
	return newConn(connConfig{
			subprotocol:    w.Header().Get("Sec-WebSocket-Protocol"),
			rwc:            netConn,
			client:         false,
			copts:          copts,
			flateThreshold: opts.CompressionThreshold,

			br: brw.Reader,
			bw: brw.Writer,
		}), nil
	*/
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}

	// 1.新用戶進來，建構該用戶的實例
	token := req.FormValue("token")
	nickname := req.FormValue("nickname")
	if l := len(nickname); l < 2 || l > 20 {
		log.Println("nickname illegal: ", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("非法昵稱,昵稱長度:2-20"))
		conn.Close(websocket.StatusUnsupportedData, "nickname illegal!")
		return
	}
	if !logic.Broadcaster.CanEnterRoom(nickname) {
		log.Println("昵稱已經存在:", nickname)
		wsjson.Write(req.Context(), conn, logic.NewErrorMessage("該昵稱已經存在!"))
		conn.Close(websocket.StatusUnsupportedData, "nickname exists!")
		return
	}

	userHasToken := logic.NewUser(conn, token, nickname, req.RemoteAddr)

	// 2. 開啟給用戶發送訊息的 goroutine
	go userHasToken.SendMessage(req.Context())

	// 3. 給當前用戶發送歡迎訊息
	userHasToken.MessageChannel <- logic.NewWelcomeMessage(userHasToken)

	// 避免 token 洩露
	tmpUser := *userHasToken
	user := &tmpUser
	user.Token = ""

	// 給所有用戶告知新用戶到來
	msg := logic.NewUserEnterMessage(user)
	logic.Broadcaster.Broadcast(msg)

	// 4. 將該用戶加入廣播器的用清單中
	logic.Broadcaster.UserEntering(user)
	log.Println("user:", nickname, "joins chat")

	// 5. 接收用戶訊息
	err = user.ReceiveMessage(req.Context())

	// 6. 用戶離開
	logic.Broadcaster.UserLeaving(user)
	msg = logic.NewUserLeaveMessage(user)
	logic.Broadcaster.Broadcast(msg)
	log.Println("user:", nickname, "leaves chat")

	// 根據讀取時的錯誤執行不同的 Close
	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error:", err)
		conn.Close(websocket.StatusInternalError, "Read from client error")
	}
}
