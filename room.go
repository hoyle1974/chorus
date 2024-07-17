package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/lithammer/shortuuid/v4"
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

func (r *Room) JSTemplate(isolate *v8go.Isolate) *v8go.ObjectTemplate {
	// Create a new java object that represents a room
	objTemplate := v8go.NewObjectTemplate(isolate)
	objTemplate.Set("Id", r.id)
	objTemplate.Set("Room", r)

	join := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		id := ListenerId(info.Args()[0].String())
		conn := FindConnectionById(id)
		r.join(id, conn)
		return nil
	})
	objTemplate.Set("Join", join)

	leave := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		id := ListenerId(info.Args()[0].String())
		r.leave(id)
		return nil
	})
	objTemplate.Set("Leave", leave)

	return objTemplate
}

var roomLock sync.Mutex
var rooms = map[RoomId]*Room{}

func FindRoom(roomId RoomId) *Room {
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

	// create global sendMsg in JS context
	sendMsg := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		jsonString, err := v8go.JSONStringify(info.Context(), info.Args()[0])
		if err != nil {
			fmt.Println(fmt.Errorf("create sendMsg function: %w", err))
			return nil
		}

		msg := NewMessageFromString(jsonString)
		msg.MsgId = MessageId(shortuuid.New())
		msg.RoomId = room.id
		msg.SenderId = ListenerId(room.id)
		room.sendMsg(msg)

		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("sendMsg", sendMsg)
	if err != nil {
		fmt.Println("Set Error", err)
		return nil, fmt.Errorf("create sendMsg function: %w", err)
	}

	// create global log in JS context
	log := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		fmt.Printf("%v\n", info.Args())
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

		newRoom, err := NewRoom(name, script)
		if err != nil {
			fmt.Printf("newRoom %v\n", err)
			return nil
		}
		objTemplate := newRoom.JSTemplate(isolate)

		obj, err := objTemplate.NewInstance(info.Context())
		if err != nil {
			fmt.Printf("NewObjectTemplate %v\n", err)
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
			fmt.Println(fmt.Errorf("js objTemplate NewInstance: %w", err))
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

func NewRoom(name string, adminScript string) (*Room, error) {
	room := &Room{
		id:        RoomId("R" + shortuuid.New()),
		name:      name,
		listeners: map[ListenerId]RoomListener{},
	}
	ctx, err := createScriptEnvironmentForRoom(room, adminScript)
	if err != nil {
		return nil, fmt.Errorf("createScriptEnvironmentForRoom: %w", err)
	}
	room.ctx = ctx

	roomLock.Lock()
	rooms[room.id] = room
	defer roomLock.Unlock()

	return room, nil
}

func (r *Room) callJSOnMessage(msg Message) error {
	err := r.ctx.Global().Set("msg", msg.String())
	if err != nil {
		fmt.Println("*** ", err)
		return err
	}
	cmd := "if (typeof on" + msg.Cmd + " == \"on" + msg.Cmd + "\") on" + msg.Cmd + "(JSON.parse(msg))"
	fmt.Println("Running: ", cmd)
	_, err = r.ctx.RunScript(cmd, "")
	if err != nil {
		fmt.Println("*** ", err)
		return err
	}
	return nil
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

func (r *Room) Join(id ListenerId, listener RoomListener) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.join(id, listener)
}
func (r *Room) join(id ListenerId, listener RoomListener) error {
	r.listeners[id] = listener

	joinMsg := NewMessage(r.id, id, "", "Join", map[string]string{})

	// Let the room know we have joined
	err := r.callJSOnMessage(joinMsg)
	if err != nil {
		delete(r.listeners, id)
		return err
	}

	r.sendMsg(joinMsg)
	return nil
}

func (r *Room) Leave(id ListenerId) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.leave(id)
}
func (r *Room) leave(id ListenerId) error {
	delete(r.listeners, id)

	leaveMsg := NewMessage(r.id, id, "", "Leave", map[string]string{})
	err := r.callJSOnMessage(leaveMsg)
	if err != nil {
		delete(r.listeners, id)
		return err
	}

	r.sendMsg(leaveMsg)
	return nil
}

func (r *Room) Send(msg Message) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.send(msg)
}
func (r *Room) send(msg Message) error {
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
