package room

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"golang.org/x/exp/rand"
	"rogchap.com/v8go"
)

const globalLobbyId = misc.RoomId("GlobalLobby")

func GetGlobalLobby(logger *slog.Logger) misc.RoomId {
	lobby, err := NewRoomWithId(logger, globalLobbyId, "Default Lobby", "matchmaker.js")
	if err != nil {
		logger.Error("Error creating default lobby", err)
		return ""
	}
	return lobby.Id
}

// Listen for messages in the room
type RoomListener interface {
	OnMessage(msg message.Message)
}

type Room struct {
	logger     *slog.Logger
	baseLogger *slog.Logger
	Id         misc.RoomId
	lock       sync.Mutex
	name       string
	script     string
	ctx        *v8go.Context
	style      lipgloss.Style
	listeners  map[misc.ListenerId]RoomListener
	consumer   *pubsub.Consumer
}

func (r *Room) HasListener(listenerId misc.ListenerId) bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	_, ok := r.listeners[listenerId]
	return ok
}

func (r *Room) JSTemplate(isolate *v8go.Isolate) *v8go.ObjectTemplate {
	// Create a new java object that represents a room
	objTemplate := v8go.NewObjectTemplate(isolate)
	objTemplate.Set("Id", r.Id)
	objTemplate.Set("Room", r)

	join := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		//id := misc.ListenerId(info.Args()[0].String())
		// TODO
		//conn := connection.FindConnectionById(id)
		//r.join(id, conn)
		return nil
	})
	objTemplate.Set("Join", join)

	leave := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		id := misc.ListenerId(info.Args()[0].String())
		r.leave(id)
		return nil
	})
	objTemplate.Set("Leave", leave)

	return objTemplate
}

var roomLock sync.Mutex
var rooms = map[misc.RoomId]*Room{}

func FindRoom(roomId misc.RoomId) *Room {
	roomLock.Lock()
	defer roomLock.Unlock()

	return rooms[roomId]
}

func createScriptEnvironmentForRoom(room *Room, adminScriptFilename string) (*v8go.Context, error) {
	// Create a new Isolate for sandboxed execution
	isolate := v8go.NewIsolate()

	data, err := os.ReadFile(adminScriptFilename)
	if err != nil {
		return nil, fmt.Errorf("Error Reading File: %w", err)
	}
	content := string(data)

	// Global object
	global := v8go.NewObjectTemplate(isolate)

	// create global endRoom() in JS context
	endRoom := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		EndRoom(room.Id)
		return nil
	})
	err = global.Set("endRoom", endRoom)
	if err != nil {
		return nil, fmt.Errorf("create endRoom function: %w", err)
	}

	// create global sendMsg in JS context
	sendMsg := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		jsonString, err := v8go.JSONStringify(info.Context(), info.Args()[0])
		if err != nil {
			fmt.Println(fmt.Errorf("create sendMsg function: %w", err))
			return nil
		}

		msg := message.NewMessageFromString(jsonString)
		msg.RoomId = room.Id
		msg.SenderId = misc.ListenerId(room.Id)
		room.sendMsg(msg)

		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("sendMsg", sendMsg)
	if err != nil {
		return nil, fmt.Errorf("create sendMsg function: %w", err)
	}

	// create global log in JS context
	log := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		msg := room.style.Render(fmt.Sprintf("%v", info.Args()))
		room.logger.Info(msg, "script", adminScriptFilename)
		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("log", log)
	if err != nil {
		return nil, fmt.Errorf("create log function: %w", err)
	}

	// create global NewRoom in JS context
	newRoom := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		name := info.Args()[0].String()
		script := info.Args()[1].String()

		newRoom, err := NewRoom(room.baseLogger, name, script)
		if err != nil {
			room.logger.Error(room.style.Render("newRoom"), "err", err)
			return nil
		}
		objTemplate := newRoom.JSTemplate(isolate)

		obj, err := objTemplate.NewInstance(info.Context())
		if err != nil {
			room.logger.Error(room.style.Render("NewObjectTemplate"), "err", err)
			return nil
		}

		return obj.Value // you can return a value back to the JS caller if required
	})
	err = global.Set("newRoom", newRoom)
	if err != nil {
		return nil, fmt.Errorf("create newRoom function: %w", err)
	}

	thisRoom := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// create global thisRoom in JS context
		objTemplate := room.JSTemplate(isolate)
		obj, err := objTemplate.NewInstance(room.ctx)
		if err != nil {
			room.logger.Error(room.style.Render(fmt.Sprintf("js objTemplate NewInstance: %w", err)))
			return nil
		}
		return obj.Value
	})
	err = global.Set("thisRoom", thisRoom)
	if err != nil {
		return nil, fmt.Errorf("create thisRoom function: %w", err)
	}

	ctx := v8go.NewContext(isolate, global) // new Context with the global Object set to our object template
	room.ctx = ctx

	_, err = ctx.RunScript(content, adminScriptFilename)

	if err != nil {
		return nil, fmt.Errorf("runScript(%s): %w", adminScriptFilename, err)
	}

	return ctx, nil
}

func hex() string {
	hexDigits := "0123456789ABCDEF"
	return string(hexDigits[rand.Intn(len(hexDigits))])
}

func newRandColor() lipgloss.Color {
	return lipgloss.Color("#" + hex() + hex() + hex() + hex() + hex() + hex())
}

func NewRoom(baseLogger *slog.Logger, name string, adminScript string) (*Room, error) {
	return NewRoomWithId(baseLogger, misc.RoomId("R"+misc.UUIDString()), name, adminScript)
}

func NewRoomWithId(baseLogger *slog.Logger, roomId misc.RoomId, name string, adminScript string) (*Room, error) {
	room := &Room{
		baseLogger: baseLogger,
		Id:         roomId,
		name:       name,
		script:     adminScript,
		listeners:  map[misc.ListenerId]RoomListener{},
	}
	room.style = lipgloss.NewStyle().
		Bold(true).
		Foreground(newRandColor())

	room.logger = baseLogger.With("roomId", room.Id, "name", name)
	ctx, err := createScriptEnvironmentForRoom(room, adminScript)
	if err != nil {
		return nil, fmt.Errorf("createScriptEnvironmentForRoom: %w", err)
	}
	room.ctx = ctx

	if !pubsub.TopicExists(string(room.Id)) {
		room.logger.Info("Creating topic")
		err = pubsub.CreateTopic(string(room.Id))
		if err != nil {
			return nil, err
		}
	} else {
		room.logger.Info("Topic already exists")
	}
	room.consumer = pubsub.NewConsumer(string(room.Id), room)

	roomLock.Lock()
	defer roomLock.Unlock()
	rooms[room.Id] = room

	go room.consumer.ProcessMessages()

	return room, nil
}

func (r *Room) OnMessageFromTopic(msg message.Message) {
	r.logger.Info("Room.OnMessageFromTopic", "msg", msg)
	// TODO
	// Let the room know we have joined
	r.callJSOnMessage(msg)
}

func EndRoom(roomId misc.RoomId) {
	roomLock.Lock()
	defer roomLock.Unlock()

	delete(rooms, roomId)

	pubsub.DeleteTopic(string(roomId))
}

func (r *Room) callJSOnMessage(msg message.Message) error {
	err := r.ctx.Global().Set("msg", msg.String())
	if err != nil {
		return err
	}
	cmd := "on" + msg.Cmd + "(JSON.parse(msg))"
	// fmt.Printf("@@@ %s : Running: %v\n", r.script, cmd)
	_, err = r.ctx.RunScript(cmd, "")
	if err != nil {
		// fmt.Println("@@@ %s : %v", r.script, err)
		return err
	}
	return nil
}

func (r *Room) sendMsg(msg message.Message) {
	pubsub.SendMessage(msg)
	/*
		if msg.ReceiverId != "" {
			l := r.listeners[msg.ReceiverId]
			if l != nil {
				l.OnMessage(msg)
			}
		} else {
			for _, l := range r.listeners {
				l.OnMessage(msg)
			}
		}
	*/
}
func (r *Room) onMessageFromTopic(msg message.Message) {
	if msg.ReceiverId != "" {
		l := r.listeners[msg.ReceiverId]
		if l != nil {
			l.OnMessage(msg)
		}
	} else {
		for _, l := range r.listeners {
			l.OnMessage(msg)
		}
	}
}

// func (r *Room) Join(id misc.ListenerId, listener RoomListener) error {
// 	r.lock.Lock()
// 	defer r.lock.Unlock()

// 	return r.join(id, listener)
// }

// func (r *Room) join(id misc.ListenerId, listener RoomListener) error {
// 	r.listeners[id] = listener

// 	joinMsg := message.NewMessage(r.Id, id, "", "Join", map[string]interface{}{})

// 	// Let the room know we have joined
// 	err := r.callJSOnMessage(joinMsg)
// 	if err != nil {
// 		delete(r.listeners, id)
// 		return err
// 	}

// 	r.sendMsg(joinMsg)
// 	return nil
// }

func (r *Room) Leave(id misc.ListenerId) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.leave(id)
}
func (r *Room) leave(id misc.ListenerId) error {

	leaveMsg := message.NewMessage(r.Id, id, "", "Leave", map[string]interface{}{})
	_ = r.callJSOnMessage(leaveMsg)

	r.sendMsg(leaveMsg)

	delete(r.listeners, id)

	if len(r.listeners) == 0 {
		_, err := r.ctx.RunScript("onRoomEmpty()", "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Room) Send(msg message.Message) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.send(msg)
}
func (r *Room) send(msg message.Message) error {
	for id, l := range r.listeners {
		if msg.ReceiverId == "" || msg.ReceiverId == id {
			l.OnMessage(msg)
		}
	}
	if msg.ReceiverId == "" || msg.ReceiverId == "room" {
		return r.callJSOnMessage(msg)
	}
	return nil
}

var consumersLock = sync.Mutex{}
var consumers = map[string]*pubsub.Consumer{}
var roomsByListener = map[misc.ListenerId][]misc.RoomId{}

func Join(roomId misc.RoomId, listenerId misc.ListenerId, handler pubsub.TopicMessageHandler) {
	joinMsg := message.NewMessage(roomId, listenerId, "", "Join", map[string]interface{}{})
	Send(joinMsg)

	consumerId := string(roomId + ":" + misc.RoomId(listenerId))
	consumersLock.Lock()
	consumer := pubsub.NewConsumer(string(roomId), handler)
	consumers[consumerId] = consumer
	list, ok := roomsByListener[listenerId]
	if !ok {
		list = []misc.RoomId{}
	}
	list = append(list, roomId)
	roomsByListener[listenerId] = list
	consumersLock.Unlock()

	go consumer.ProcessMessages()
}

func Leave(roomId misc.RoomId, listenerId misc.ListenerId) {
	joinMsg := message.NewMessage(roomId, listenerId, "", "Leave", map[string]interface{}{})
	Send(joinMsg)

	consumerId := string(roomId + ":" + misc.RoomId(listenerId))
	consumersLock.Lock()
	consumer := consumers[consumerId]
	delete(consumers, consumerId)
	consumersLock.Unlock()
	consumer.Close()
}

func LeaveAllRooms(listenerId misc.ListenerId) {
	fmt.Println("LeaveAllRooms", listenerId)
	consumersLock.Lock()
	list, ok := roomsByListener[listenerId]
	delete(roomsByListener, listenerId)
	consumersLock.Unlock()

	if ok {
		for _, roomId := range list {
			Leave(roomId, listenerId)
		}
	}
}

func RemoveAllRooms() {
	listeners := []misc.ListenerId{}

	consumersLock.Lock()
	for listenerId, _ := range roomsByListener {
		listeners = append(listeners, listenerId)
	}
	consumersLock.Unlock()

	for _, listenerId := range listeners {
		fmt.Println(4, listenerId)
		LeaveAllRooms(listenerId)
	}

	EndRoom(globalLobbyId)

}

func Send(msg message.Message) {
	pubsub.SendMessage(msg)
}
