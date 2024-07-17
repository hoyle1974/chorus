package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"rogchap.com/v8go"
)

type RoomListener interface {
	OnMessage(msg Message)
}

type Room struct {
	id        RoomId
	lock      sync.Mutex
	name      string
	ctx       *v8go.Context
	listeners map[ListenerId]RoomListener
}

var roomLock sync.Mutex
var rooms = map[RoomId]*Room{}

func FindRoom(roomId RoomId) *Room {
	roomLock.Lock()
	defer roomLock.Unlock()

	return rooms[roomId]
}

func NewRoom(name string, adminScript string) *Room {
	room := &Room{
		id:        RoomId("R" + uuid.NewString()),
		name:      name,
		listeners: map[ListenerId]RoomListener{},
	}
	roomLock.Lock()
	rooms[room.id] = room
	roomLock.Unlock()

	// Create a new Isolate for sandboxed execution
	isolate := v8go.NewIsolate()

	data, err := os.ReadFile(adminScript)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}
	content := string(data)

	// Global object
	global := v8go.NewObjectTemplate(isolate)

	// sendMsg
	sendMsg := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		jsonString, err := v8go.JSONStringify(info.Context(), info.Args()[0])
		if err != nil {
			fmt.Println("Error ", err)
			return nil
		}

		msg := NewMessageFromString(jsonString)
		msg.MsgId = MessageId(uuid.NewString())
		msg.RoomId = room.id
		msg.SenderId = ListenerId(room.id)
		room.sendMsg(msg)

		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("sendMsg", sendMsg)
	if err != nil {
		fmt.Println("Set Error", err)
	}

	// log
	log := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		fmt.Printf("%v\n", info.Args())
		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("log", log)
	if err != nil {
		fmt.Println("Set Error", err)
	}

	// NewRoom
	newRoom := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		name := info.Args()[0].String()
		script := info.Args()[1].String()

		newRoom := NewRoom(name, script)

		// Create a new java object that represents a room
		objTemplate := v8go.NewObjectTemplate(isolate)
		objTemplate.Set("id", newRoom.id)
		objTemplate.Set("room", newRoom)

		join := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
			id := ListenerId(info.Args()[0].String())
			conn := FindConnectionById(id)
			newRoom.Join(id, conn)
			return nil
		})
		objTemplate.Set("Join", join)

		obj, err := objTemplate.NewInstance(info.Context())
		if err != nil {
			fmt.Printf("NewObjectTemplate %v\n", err)
		}

		return obj.Value // you can return a value back to the JS caller if required
	})
	err = global.Set("NewRoom", newRoom)
	if err != nil {
		fmt.Println("Set Error", err)
	}

	ctx := v8go.NewContext(isolate, global) // new Context with the global Object set to our object template
	ctx.RunScript(content, adminScript)

	room.ctx = ctx

	return room
}

func (r *Room) callJSOnMessage(msg Message) {
	err := r.ctx.Global().Set("msg", msg.String())
	if err != nil {
		fmt.Println(err)
	}
	_, err = r.ctx.RunScript("on"+msg.Cmd+"(JSON.parse(msg))", "")
	if err != nil {
		fmt.Println(err)
	}
}

func (r *Room) sendMsg(msg Message) {
	fmt.Println("sendMsg:", msg)

	if msg.ReceiverId != "" {
		fmt.Println("Sending just to ", msg.ReceiverId)
		l := r.listeners[msg.ReceiverId]
		if l == nil {
			fmt.Println("not found ", msg.ReceiverId)
		}
		l.OnMessage(msg)
	} else {
		fmt.Println("Sending to all listeners")
		for _, l := range r.listeners {
			l.OnMessage(msg)
		}
	}
}

func (r *Room) Join(id ListenerId, listener RoomListener) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.listeners[id] = listener

	r.callJSOnMessage(NewMessage(r.id, id, "", "Join", map[string]string{}))
}

func (r *Room) Leave(id ListenerId) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.listeners, id)

	r.callJSOnMessage(NewMessage(r.id, id, "", "Leave", map[string]string{}))
}

func (r *Room) Send(msg Message) {

	r.lock.Lock()
	defer r.lock.Unlock()

	for id, l := range r.listeners {
		if msg.ReceiverId == "" || msg.ReceiverId == id {
			l.OnMessage(msg)
		}
	}
	if msg.ReceiverId == "" || msg.ReceiverId == "room" {
		r.callJSOnMessage(msg)
	}

}
