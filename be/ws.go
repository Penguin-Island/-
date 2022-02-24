package be

import (
	"errors"
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
	ErrMsgTime        = "参加可能な時間ではありません"
	ErrMsgBadReq      = "不正なリクエストです"
	ErrMsgServerError = "サーバーでエラーが発生しました"
)

const (
	EventTypeNotifyWaitState = "notifyWaitState"
	EventTypeOnStart         = "onStart"
	EventTypeOnTick          = "onTick"
	EventTypeOnFailure       = "onFailure"
	EventTypeOnChangeTurn    = "onChangeTurn"
	EventTypeOnError         = "onError"
	EventTypeSendAnswer      = "sendAnswer"
	EventTypeConfirmContinue = "confirmContinue"
	EventTypeOnInput         = "onInput"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  64,
	WriteBufferSize: 64,
	// already checked by middleware
	CheckOrigin: func(*http.Request) bool { return true },
}

type IEJoinMember struct {
	Channel chan InternalNotification
}
type IEUnjoinMember struct {
	Channel chan InternalNotification
}
type IETick struct {
	Remain          int
	TurnRemain      int
	WaitingContinue bool
	FailingUser     uint
}
type IEChangeTurn struct {
	PrevWord   string
	NextUserId uint
}
type IESendWord struct {
	Word string
}
type IEConfirmContinue struct{}
type IEStart struct{}
type IEFailure struct{}
type IEError struct {
	Reason string
}
type IEInput struct {
	Value string
}

type InternalNotification struct {
	EmitterUser uint
	Payload     interface{}
}

type GameStates struct {
	communicators map[uint]chan InternalNotification
	gamesMu       sync.Mutex
}

type EventPayload struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

func getGroupId(app *App, userId uint) (uint, error) {
	var user Member
	if err := app.db.First(&user, userId).Error; err != nil {
		return 0, err
	}

	if user.GroupId == 0 {
		return 0, errors.New("no group")
	}

	return user.GroupId, nil
}

func getStartTimeForGroup(app *App, groupId uint) (*time.Time, error) {
	var group Group
	if err := app.db.First(&group, groupId).Error; err != nil {
		return nil, err
	}

	timeStr := group.WakeUpTime

	savedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := time.Date(now.Year(), now.Month(), now.Day(), savedTime.Hour(), savedTime.Minute(), 0, 0, time.UTC)
	if now.After(result.Add(6 * time.Minute)) {
		result = result.Add(24 * time.Hour)
	}

	return &result, nil
}

func isJoinableTime(startTime *time.Time) bool {
	now := time.Now()
	return now.After(startTime.Add(-10*time.Minute)) && now.Before(startTime.Add(10*time.Minute))
}

func appendUser(users []uint, userId uint) []uint {
	for _, uid := range users {
		if uid == userId {
			return users
		}
	}
	return append(users, userId)
}

func notifyToEveryone(n InternalNotification, comminucators []chan InternalNotification) {
	for _, c := range comminucators {
		go func(c chan InternalNotification) {
			defer func() {
				if err := recover(); err != nil {
					log.Error(err)
				}
			}()
			c <- n
		}(c)
	}
}

func areAllMembersJoined(app *App, users []uint, groupId uint) (bool, error) {
	var members []Member
	if err := app.db.Find(&members, "group_id = ?", groupId).Error; err != nil {
		return false, err
	}

	for _, memb := range members {
		ok := false
		for _, joined := range users {
			if memb.ID == joined {
				ok = true
				break
			}
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

// ゲーム全体の進行を管理する
func manageGame(app *App, s *GameStates, groupId uint, startTime *time.Time, toHub chan InternalNotification) {
	remain := 300
	turnRemain := 20
	continueRemain := 30
	waitingContinue := false
	users := make([]uint, 0)
	userFailCount := make(map[uint]int)
	communicators := make([]chan InternalNotification, 0)
	noti := InternalNotification{}
	turnIndex := 0
	prevWord := "おはよう"
	gameStarted := false
	lastTickInfo := IETick{}
	lastChangeTurnInfo := IEChangeTurn{}
	ticker := time.NewTicker(11 * time.Minute)
	startTimer := time.NewTimer(startTime.Add(6 * time.Minute).Sub(time.Now()))
	for remain >= 0 {
		select {
		case <-startTimer.C:
			// 全員揃わなかった為失敗
			noti.Payload = IEFailure{}
			notifyToEveryone(noti, communicators)
			goto deleteCommunicator

		case <-ticker.C:
			if waitingContinue {
				continueRemain--
				lastTickInfo.Remain = remain
				lastTickInfo.TurnRemain = continueRemain
				lastTickInfo.WaitingContinue = true
				noti.Payload = lastTickInfo
				notifyToEveryone(noti, communicators)
				if continueRemain == 0 {
					waitingContinue = false
					userFailCount[users[turnIndex]]++
				}
				break
			}
			remain--
			turnRemain--
			lastTickInfo.Remain = remain
			lastTickInfo.TurnRemain = turnRemain
			lastTickInfo.WaitingContinue = false
			lastTickInfo.FailingUser = users[turnIndex]
			noti.Payload = lastTickInfo
			notifyToEveryone(noti, communicators)
			if turnRemain == 0 {
				if userFailCount[users[turnIndex]] < 1 {
					// 1回目の失敗
					waitingContinue = true
					continueRemain = 30
					userFailCount[users[turnIndex]]++
				} else {
					// 2回目以降の失敗
					turnIndex = (turnIndex + 1) % len(users)
					turnRemain = 20
					lastChangeTurnInfo.PrevWord = prevWord
					lastChangeTurnInfo.NextUserId = users[turnIndex]
					noti.Payload = lastChangeTurnInfo
					notifyToEveryone(noti, communicators)
				}
			}
			break

		case noti := <-toHub:
			switch payload := noti.Payload.(type) {
			case IEJoinMember:
				// 開始後に新しいユーザーが参加する状況は起こり得ない (全員集まらないとゲームが始まらないため)

				users = appendUser(users, noti.EmitterUser)
				communicators = append(communicators, payload.Channel)

				if gameStarted {
					go func() {
						noti.Payload = IEStart{}
						payload.Channel <- noti

						noti.Payload = lastTickInfo
						payload.Channel <- noti

						noti.Payload = lastChangeTurnInfo
						payload.Channel <- noti
					}()
				} else {
					if joined, err := areAllMembersJoined(app, users, groupId); err != nil {
						noti.Payload = IEError{
							Reason: ErrMsgServerError,
						}
						notifyToEveryone(noti, communicators)

						goto deleteCommunicator
					} else if joined {
						gameStarted = true
						ticker.Stop()
						if !startTimer.Stop() {
							<-startTimer.C
						}
						ticker = time.NewTicker(time.Second)

						noti.Payload = IEStart{}
						notifyToEveryone(noti, communicators)

						lastChangeTurnInfo.PrevWord = prevWord
						lastChangeTurnInfo.NextUserId = users[turnIndex]
						noti.Payload = lastChangeTurnInfo
						notifyToEveryone(noti, communicators)

						noti.Payload = IETick{
							Remain:     remain,
							TurnRemain: turnRemain,
						}
						notifyToEveryone(noti, communicators)
					}
				}
				break

			case IEUnjoinMember:
				if !gameStarted {
					for i, u := range users {
						if u == noti.EmitterUser {
							users[i] = users[len(users)-1]
							users = users[:len(users)-1]
							break
						}
					}
					// 誰もいないので終わらせて良い
					if len(users) == 0 {
						goto deleteCommunicator
					}
				}
				for i, c := range communicators {
					if c == payload.Channel {
						close(communicators[i])
						communicators[i] = communicators[len(communicators)-1]
						communicators = communicators[:len(communicators)-1]
						break
					}
				}
				break

			case IESendWord:
				if waitingContinue {
					log.Error("Recieved word while waiting continue")
					break
				}
				if noti.EmitterUser != users[turnIndex] {
					log.Error("Recieved word from non-turn user")
					break
				}
				if shiritori.IsValidShiritori(prevWord, payload.Word) {
					// 成功
					prevWord = payload.Word
					turnIndex = (turnIndex + 1) % len(users)
					turnRemain = 20
					lastChangeTurnInfo.PrevWord = prevWord
					lastChangeTurnInfo.NextUserId = users[turnIndex]
					noti.Payload = lastChangeTurnInfo
					notifyToEveryone(noti, communicators)
				} else {
					// 失敗
					if userFailCount[noti.EmitterUser] < 1 {
						// 1回目の失敗 (コンティニューできる)
						waitingContinue = true
						continueRemain = 30
					} else {
						// 2回目以降の失敗 (コンティニューできない)
						turnIndex = (turnIndex + 1) % len(users)
						turnRemain = 20
						lastChangeTurnInfo.PrevWord = prevWord
						lastChangeTurnInfo.NextUserId = users[turnIndex]
						noti.Payload = lastChangeTurnInfo
						notifyToEveryone(noti, communicators)
					}
					userFailCount[noti.EmitterUser]++
				}
				break

			case IEConfirmContinue:
				if !waitingContinue || noti.EmitterUser != users[turnIndex] {
					break
				}
				turnRemain += 10
				waitingContinue = false
				break

			case IEInput:
				if waitingContinue || noti.EmitterUser != users[turnIndex] {
					break
				}
				notifyToEveryone(noti, communicators)
				break
			}
			break
		}
	}
	ticker.Stop()
	startTimer.Stop()

	log.Info(len(users))
	for _, u := range users {
		log.WithField("userId", u).WithField("failCount", userFailCount[u]).Info()
		if err := recordStat(app, u, userFailCount[u] < 2); err != nil {
			log.Error(err)
		}
	}
	// 成功

deleteCommunicator:
	for _, u := range users {
		invalidateStatCache(app, u)
	}

	log.Info("Deleting communicator")

	for _, c := range communicators {
		close(c)
	}
	close(toHub)

	s.gamesMu.Lock()
	defer s.gamesMu.Unlock()
	delete(s.communicators, groupId)
}

// ゲームに接続する
func (s *GameStates) joinGame(app *App, startTime *time.Time, groupId uint, userId uint) (chan InternalNotification, chan InternalNotification) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	s.gamesMu.Lock()
	defer s.gamesMu.Unlock()
	toHub, ok := s.communicators[groupId]
	if !ok {
		toHub = make(chan InternalNotification)
		s.communicators[groupId] = toHub
		go manageGame(app, s, groupId, startTime, toHub)
	}

	notifier := make(chan InternalNotification)
	noti := InternalNotification{}
	noti.EmitterUser = userId
	noti.Payload = IEJoinMember{
		Channel: notifier,
	}
	toHub <- noti

	return notifier, toHub
}

// ゲームから切断する
func (s *GameStates) unjoinGame(userId uint, notifier chan InternalNotification, toHub chan InternalNotification) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	noti := InternalNotification{}
	noti.EmitterUser = userId
	noti.Payload = IEUnjoinMember{
		Channel: notifier,
	}
	toHub <- noti
}

func handleSocketConnection(app *App, c *gin.Context) {
	sess := sessions.Default(c)
	iUserId := sess.Get("user_id")
	if iUserId == nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	} else if _, ok := iUserId.(uint); !ok {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	userId := iUserId.(uint)

	// HTTP接続をWebSocketにする
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	// 所属するグループのIDを取得
	groupId, err := getGroupId(app, userId)
	if err != nil {
		log.Error(err)
		return
	}

	startTime, err := getStartTimeForGroup(app, groupId)
	if err != nil {
		log.Error(err)
		return
	}

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
	notificationChan, toHub := app.gameStates.joinGame(app, startTime, groupId, userId)

	finishChan := make(chan struct{})
	// 読む側 (イベントを hub にディスパッチするだけ)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error(err)
			}
		}()

		for {
			var ev EventPayload
			if err := conn.ReadJSON(&ev); err != nil {
				log.Error(err)
				goto disconnect
			}

			var intNoti InternalNotification
			intNoti.EmitterUser = userId
			switch ev.Type {
			case EventTypeSendAnswer:
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
				intNoti.EmitterUser = userId
				intNoti.Payload = IESendWord{
					Word: word.(string),
				}
				toHub <- intNoti
				break

			case EventTypeConfirmContinue:
				intNoti.EmitterUser = userId
				intNoti.Payload = IEConfirmContinue{}
				toHub <- intNoti
				break

			case EventTypeOnInput:
				intNoti.EmitterUser = userId
				value := ev.Data["value"]
				if _, ok := value.(string); !ok {
					conn.WriteJSON(EventPayload{
						Type: EventTypeOnError,
						Data: map[string]interface{}{
							"reason": ErrMsgBadReq,
						},
					})
					goto next
				}
				intNoti.Payload = IEInput{
					Value: value.(string),
				}
				toHub <- intNoti
				break
			}
		next:
		}

	disconnect:
		close(finishChan)
	}()

	for {
		select {
		case noti, ok := <-notificationChan:
			if !ok {
				goto disconnect
			}
			switch data := noti.Payload.(type) {
			case IETick:
				payload := EventPayload{
					Type: EventTypeOnTick,
					Data: map[string]interface{}{
						"remainSec":       data.Remain,
						"turnRemainSec":   data.TurnRemain,
						"finished":        data.Remain == 0,
						"waitingContinue": data.WaitingContinue,
						"yourFailure":     data.FailingUser == userId,
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}

				if data.Remain == 0 {
					goto disconnect
				}
				break

			case IEChangeTurn:
				payload := EventPayload{
					Type: EventTypeOnChangeTurn,
					Data: map[string]interface{}{
						"prevAnswer": data.PrevWord,
						"yourTurn":   data.NextUserId == userId,
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
				break

			case IEStart:
				payload := EventPayload{
					Type: EventTypeOnStart,
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
				break

			case IEFailure:
				log.Info("failure")
				payload := EventPayload{
					Type: EventTypeOnFailure,
					Data: map[string]interface{}{},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
				break

			case IEError:
				payload := EventPayload{
					Type: EventTypeOnError,
					Data: map[string]interface{}{
						"reason": data.Reason,
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
				break

			case IEInput:
				payload := EventPayload{
					Type: EventTypeOnInput,
					Data: map[string]interface{}{
						"value": data.Value,
					},
				}
				if err := conn.WriteJSON(payload); err != nil {
					log.Error(err)
					goto disconnect
				}
			}
		case <-finishChan:
			goto disconnect
		}
	}

disconnect:
	app.gameStates.unjoinGame(userId, notificationChan, toHub)
}
