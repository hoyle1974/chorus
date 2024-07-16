package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"rogchap.com/v8go"
	v8 "rogchap.com/v8go"
)

type RoomListener interface {
	OnMessage(msg Message)
}

type Room struct {
	id        string
	lock      sync.Mutex
	name      string
	ctx       *v8.Context
	listeners map[string]RoomListener
}

func NewRoom(name string, adminScript string) *Room {
	roomId := uuid.NewString()

	room := &Room{
		id:        roomId,
		name:      name,
		listeners: map[string]RoomListener{},
	}

	// Create a new Isolate for sandboxed execution
	isolate := v8.NewIsolate()

	data, err := os.ReadFile(adminScript)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}
	content := string(data)

	// Global object
	global := v8.NewObjectTemplate(isolate)

	// sendMsg
	sendMsg := v8.NewFunctionTemplate(isolate, func(info *v8.FunctionCallbackInfo) *v8.Value {
		jsonString, err := v8go.JSONStringify(info.Context(), info.Args()[0])
		if err != nil {
			fmt.Println("Error ", err)
			return nil
		}

		msg := NewMessageFromString(jsonString)
		msg.SenderId = roomId
		room.sendMsg(msg)

		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("sendMsg", sendMsg)
	if err != nil {
		fmt.Println("Set Error", err)
	}

	// log
	log := v8.NewFunctionTemplate(isolate, func(info *v8.FunctionCallbackInfo) *v8.Value {
		fmt.Printf("%v\n", info.Args())
		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("log", log)
	if err != nil {
		fmt.Println("Set Error", err)
	}

	// NewRoom
	newRoom := v8.NewFunctionTemplate(isolate, func(info *v8.FunctionCallbackInfo) *v8.Value {
		name := info.Args()[0].String()
		script := info.Args()[1].String()

		room := NewRoom(name, script)

		// Create a new java object that represents a room
		objTemplate := v8.NewObjectTemplate(isolate)
		objTemplate.Set("id", room.id)
		objTemplate.Set("room", room)

		join := v8.NewFunctionTemplate(isolate, func(info *v8.FunctionCallbackInfo) *v8.Value {
			id := info.Args()[0].String()
			conn := FindConnectionById(id)
			room.Join(id, conn)
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

	ctx := v8.NewContext(isolate, global) // new Context with the global Object set to our object template
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

	if msg.ReceiverId == "" {
		fmt.Println("Sending to all listeners")
		for _, l := range r.listeners {
			l.OnMessage(msg)
		}
	} else {
		fmt.Println("Sending just to ", msg.ReceiverId)
		l := r.listeners[msg.ReceiverId]
		if l == nil {
			fmt.Println("not found ", msg.ReceiverId)
		}
		l.OnMessage(msg)
	}
}

func (r *Room) Join(id string, listener RoomListener) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.listeners[id] = listener

	r.callJSOnMessage(NewMessage(id, "", "Join", map[string]string{}))
}

func (r *Room) Leave(id string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	delete(r.listeners, id)

	r.callJSOnMessage(NewMessage(id, "", "Leave", map[string]string{}))
}

func (r *Room) Send(msg Message) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, l := range r.listeners {
		l.OnMessage(msg)
	}
	r.callJSOnMessage(msg)

}
