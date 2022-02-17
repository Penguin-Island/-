package be

import (
	"net/http"
	"sync"
	"time"

	"github.com/Penguin-Island/ohatori/be/shiritori"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

const (
	ErrMsgTime   = "参加可能な時間ではありません"
	ErrMsgBadReq = "不正なリクエストです"
)

const (
	EventTypeNotifyWaitState = "notifyWaitState"
	EventTypeOnStart         = "onStart"
	EventTypeOnTick          = "onTick"
	EventTypeOnFailure       = "onFailure"
	EventTypeOnChangeTurn    = "onChangeTurn"
	EventTypeOnError         = "onError"
	EventSendAnswer          = "sendAnswer"
	EventConfirmRetry        = "confirmRetry"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  64,
	WriteBufferSize: 64,
	// already checked by middleware
	CheckOrigin: func(*http.Request) bool { return true },
}

const (
	InternalGameStarted = iota
	InternalTick
	InternalChangeTurn
	InternalSendWord
	InternalConfirmRetry
	InternalOnFailure
)

type InternalNotification struct {
	Type        int
	EmitterUser int
	Tick        struct {
		Remain       int
		TurnRemain   int
		WaitingRetry bool
	}
	ChangeTurn struct {
		PrevWord   string
		NextUserId int
	}
	SendWord struct {
		Word string
	}
}

type GameState struct {
	mu        sync.Mutex
	notifier  []chan InternalNotification
	users     []int
	isPending bool
	toHub     chan InternalNotification
}

type GameStates struct {
	games   map[string]*GameState
	gamesMu sync.Mutex
}

type EventPayload struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func getGroupId(userId int) (string, error) {
	return "bd1725fb-3c1b-48bc-8514-dbc160256874", nil
}

func getStartTimeForGroup(groupId string) time.Time {
	return time.Now().Add(time.Minute)
}

func isJoinableTime(startTime time.Time) bool {
	now := time.Now()
	return now.After(now.Add(-10*time.Minute)) && now.Before(now.Add(10*time.Minute))
}

func appendUser(users []int, userId int) []int {
	for _, uid := range users {
		if uid == userId {
			return users
		}
	}
	return append(users, userId)
}

func (s *GameState) notifyToEveryone(n InternalNotification) {
	s.mu.Lock()
	for _, c := range s.notifier {
		go func(c chan InternalNotification) { c <- n }(c)
	}
	defer s.mu.Unlock()
}

// ゲーム全体の進行を管理する
func manageGame(state *GameState) {
	time.Sleep(time.Second)
	remain := 300
	turnRemain := 20
	waitingRetry := false
	var retryConfirmed []bool
	noti := InternalNotification{}
	turnIndex := 0
	prevWord := "おはよう"
	noti.Type = InternalChangeTurn
	noti.ChangeTurn.PrevWord = prevWord
	noti.ChangeTurn.NextUserId = state.users[turnIndex]
	state.notifyToEveryone(noti)
	noti.Type = InternalTick
	noti.Tick.Remain = remain
	noti.Tick.TurnRemain = turnRemain
	state.notifyToEveryone(noti)
	ticker := time.NewTicker(time.Second)
	for remain >= 0 {
		select {
		case <-ticker.C:
			if waitingRetry {
				if turnRemain == 0 {
					noti.Type = InternalOnFailure
					state.notifyToEveryone(noti)
					return
				}
				turnRemain--
			} else {
				if turnRemain == 0 {
					waitingRetry = true
					turnRemain = 20
					retryConfirmed = make([]bool, len(state.users))
					break
				}
				remain--
				turnRemain--
			}
			noti.Type = InternalTick
			noti.Tick.Remain = remain
			noti.Tick.TurnRemain = turnRemain
			noti.Tick.WaitingRetry = waitingRetry
			state.notifyToEveryone(noti)
			break

		case noti := <-state.toHub:
			switch noti.Type {
			case InternalSendWord:
				if noti.EmitterUser != state.users[turnIndex] {
					log.Warn("Recieved from non-turn user")
					break
				}
				if shiritori.IsValidShiritori(prevWord, noti.SendWord.Word) {
					// 成功
					prevWord = noti.SendWord.Word
					turnIndex = (turnIndex + 1) % len(state.users)
					turnRemain = 20

					noti.Type = InternalChangeTurn
					noti.ChangeTurn.PrevWord = prevWord
					noti.ChangeTurn.NextUserId = state.users[turnIndex]
					state.notifyToEveryone(noti)
				} else {
					// 失敗
					waitingRetry = true
					turnRemain = 20
					retryConfirmed = make([]bool, len(state.users))
				}
				break

			case InternalConfirmRetry:
				if !waitingRetry {
					break
				}
				hasNotConfirmedUsers := false
				for i, u := range state.users {
					if u == noti.EmitterUser {
						retryConfirmed[i] = true
						break
					} else {
						if !retryConfirmed[i] {
							hasNotConfirmedUsers = true
						}
					}
				}
				if !hasNotConfirmedUsers {
					waitingRetry = false
				}
			}
			break
		}
	}
	ticker.Stop()

	// 成功
}

// グループのゲームに参加する
func (s *GameStates) joinGroup(groupId string, userId int) (chan InternalNotification, chan InternalNotification) {
	s.gamesMu.Lock()
	defer s.gamesMu.Unlock()
	state, ok := s.games[groupId]
	if !ok {
		state = &GameState{
			isPending: true,
			toHub:     make(chan InternalNotification),
		}
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	notifier := make(chan InternalNotification)
	state.notifier = append(state.notifier, notifier)

	state.users = appendUser(state.users, userId)
	if state.isPending {
		if len(state.users) == 2 {
			state.isPending = false
			s.games[groupId] = state

			for _, n := range state.notifier {
				if n != notifier {
					n <- InternalNotification{
						Type: InternalGameStarted,
					}
				}
			}
			go manageGame(state)
		}
	}
	s.games[groupId] = state

	return notifier, state.toHub
}

func (s *GameStates) unjoinGroup(groupId string, userId int, notifier chan InternalNotification, removeUser bool) {
	s.gamesMu.Lock()
	defer s.gamesMu.Unlock()
	state := s.games[groupId]
	state.mu.Lock()
	defer state.mu.Unlock()
	if removeUser {
		for i, u := range state.users {
			if u == userId {
				if len(state.users) > 0 {
					state.users[i] = state.users[len(state.users)-1]
				}
				state.users = state.users[:len(state.users)-1]
			}
		}
	}
	for i, n := range state.notifier {
		if n == notifier {
			if len(state.notifier) > 0 {
				state.notifier[i] = state.notifier[len(state.notifier)-1]
			}
			state.notifier = state.notifier[:len(state.notifier)-1]
		}
	}
}

func handleSocketConnection(app *App, c *gin.Context) {
	sess := sessions.Default(c)
	userId := sess.Get("user_id")
	if userId == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if _, ok := userId.(int); !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// HTTP接続をWebSocketにする
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	// 所属するグループのIDを取得
	groupId, err := getGroupId(userId.(int))
	if err != nil {
		log.Error(err)
		return
	}

	startTime := getStartTimeForGroup(groupId)

	// 参加可能な時間でなければエラーを返して終了
	if !isJoinableTime(startTime) {
		payload := EventPayload{
			Type: EventTypeOnError,
			Data: map[string]interface{}{
				"reason": ErrMsgTime,
			},
		}
		conn.WriteJSON(payload)
		return
	}

	// グループのゲームに参加する
	notificationChan, toHub := app.gameStates.joinGroup(groupId, userId.(int))

	// 他のユーザが来るまで待機する
	app.gameStates.gamesMu.Lock()
	isPending := app.gameStates.games[groupId].isPending
	app.gameStates.gamesMu.Unlock()

	finishChan := make(chan struct{})
	// 読む側 (イベントを hub にディスパッチするだけ)
	go func() {
		for {
			var ev EventPayload
			if err := conn.ReadJSON(&ev); err != nil {
				log.Error(err)
				goto disconnect
			}

			if isPending {
				continue
			}

			var intNoti InternalNotification
			intNoti.EmitterUser = userId.(int)
			switch ev.Type {
			case EventSendAnswer:
				word := ev.Data["word"]
				if _, ok := word.(string); !ok {
					conn.WriteJSON(EventPayload{
						Type: EventTypeOnError,
						Data: map[string]interface{}{
							"reason": ErrMsgBadReq,
						},
					})
					goto next
				}
				intNoti.Type = InternalSendWord
				intNoti.EmitterUser = userId.(int)
				intNoti.SendWord.Word = word.(string)
				toHub <- intNoti
				break

			case EventConfirmRetry:
				intNoti.Type = InternalConfirmRetry
				intNoti.EmitterUser = userId.(int)
				toHub <- intNoti
				break
			}
		next:
		}

	disconnect:
		close(finishChan)
	}()

	if isPending {
		timer := time.NewTimer(startTime.Add(10 * time.Minute).Sub(time.Now()))
		for {
			select {
			case <-timer.C:
				payload := EventPayload{
					Type: EventTypeOnError,
					Data: map[string]interface{}{
						"reason": ErrMsgTime,
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					app.gameStates.unjoinGroup(groupId, userId.(int), notificationChan, true)
					return
				}
				break

			case <-notificationChan:
				if !timer.Stop() {
					<-timer.C
				}
				isPending = false
				goto startGame

			case <-finishChan:
				app.gameStates.unjoinGroup(groupId, userId.(int), notificationChan, true)
				return
			}
		}
	}
startGame:

	payload := EventPayload{
		Type: EventTypeOnStart,
	}
	if err := conn.WriteJSON(payload); err != nil {
		log.Error(err)
		goto disconnect
	}

	for {
		select {
		case noti, ok := <-notificationChan:
			if !ok {
				goto disconnect
			}
			switch noti.Type {
			case InternalTick:
				payload := EventPayload{
					Type: EventTypeOnTick,
					Data: map[string]interface{}{
						"remainSec":     noti.Tick.Remain,
						"turnRemainSec": noti.Tick.TurnRemain,
						"finished":      noti.Tick.Remain == 0,
						"waitingRetry":  noti.Tick.WaitingRetry,
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}

				if noti.Tick.Remain == 0 {
					goto disconnect
				}
				break

			case InternalChangeTurn:
				payload := EventPayload{
					Type: EventTypeOnChangeTurn,
					Data: map[string]interface{}{
						"prevAnswer": noti.ChangeTurn.PrevWord,
						"yourTurn":   noti.ChangeTurn.NextUserId == userId.(int),
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
				break

			case InternalOnFailure:
				payload := EventPayload{
					Type: EventTypeOnFailure,
					Data: map[string]interface{}{},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
				break
			}
		case <-finishChan:
			goto disconnect
		}
	}

disconnect:
	app.gameStates.unjoinGroup(groupId, userId.(int), notificationChan, false)
}
