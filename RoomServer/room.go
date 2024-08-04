package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/hoyle1974/chorus/db"
	"github.com/hoyle1974/chorus/dbx"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"rogchap.com/v8go"
)

// This represents a Room that is running in a RoomServer
// For each room in the cluster only one of these exists.  If
// You try to create two with the same room Id only one will work and
// the other will wait for the other server to go away.
type Room struct {
	state       GlobalServerState
	topic       misc.TopicId
	roomService *RoomService
	logger      *slog.Logger
	info        RoomInfo
	consumer    *pubsub.Consumer
	ctx         *v8go.Context
}

func (r *Room) AddMember(id misc.ConnectionId) {
	r.roomService.AddMember(r.info.RoomId, id)
}

func (r *Room) RemoveMember(id misc.ConnectionId) {
	r.roomService.RemoveMember(r.info.RoomId, id)
}

func (r *Room) Destroy() {
	r.logger.Info("Deleting room")
	r.roomService.DeleteRoom(r.info.RoomId)
}

func (r *Room) OnMessageFromTopic(m pubsub.Message) {
	msg := m.(*message.Message)

	r.logger.Info("Room.OnMessageFromTopic", "msg", msg)

	if msg.RoomId != r.info.RoomId {
		fmt.Println(msg)
		r.logger.Error("Received an error for the wrong room", "targetRoomId", msg.RoomId)
		return
	}

	if msg.Cmd == "Pong" {
		r.logger.Debug("Pong", "memberId", msg.SenderId)
		r.AddMember(misc.ConnectionId(msg.SenderId))
	}
	if msg.Cmd == "Join" {
		r.logger.Debug("Join", "memberId", msg.SenderId)
		r.AddMember(misc.ConnectionId(msg.SenderId))
	}
	if msg.Cmd == "Leave" {
		r.logger.Debug("Leave", "memberId", msg.SenderId)
		r.RemoveMember(misc.ConnectionId(msg.SenderId))
	}

	r.callJSOnMessage(msg)

}

func (r *Room) callJSOnMessage(msg *message.Message) error {
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

func createScriptEnvironmentForRoom(room *Room, adminScriptFilename string) (*v8go.Context, error) {
	// Create a new Isolate for sandboxed execution
	isolate := v8go.NewIsolate()

	data, err := os.ReadFile(adminScriptFilename)
	if err != nil {
		return nil, fmt.Errorf("error Reading File: %w", err)
	}
	content := string(data)

	// Global object
	global := v8go.NewObjectTemplate(isolate)

	// create global endRoom() in JS context
	endRoom := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// TODO EndRoom(room.RoomId)
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
		msg.RoomId = room.info.RoomId
		msg.SenderId = room.info.RoomId.ListenerId()
		//TODO room.sendMsg(msg)
		pubsub.SendMessage(&msg)

		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("sendMsg", sendMsg)
	if err != nil {
		return nil, fmt.Errorf("create sendMsg function: %w", err)
	}

	// create global log in JS context
	log := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		msg := fmt.Sprintf("%v", info.Args())
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

		roomInfo := RoomInfo{
			RoomId:          misc.RoomId(misc.UUIDString()),
			Name:            name,
			AdminScript:     script,
			DestroyOnOrphan: true,
		}

		newRoom, err := room.roomService.NewRoom(roomInfo)
		if err != nil {
			room.logger.Error("NewRoom", "error", err)
			return nil
		}
		objTemplate := newRoom.JSTemplate(isolate)

		obj, err := objTemplate.NewInstance(info.Context())
		if err != nil {
			room.logger.Error("NewInstance", "error", err)
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
			room.logger.Error("js objTemplate NewInstance", "error", err)
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

func (r *Room) JSTemplate(isolate *v8go.Isolate) *v8go.ObjectTemplate {
	// Create a new java object that represents a room
	objTemplate := v8go.NewObjectTemplate(isolate)
	objTemplate.Set("Id", r.info.RoomId)
	objTemplate.Set("Room", r)

	join := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		id := misc.ConnectionId(info.Args()[0].String())

		q := dbx.Dbx().Queries(db.New(dbx.GetConn()))
		mid := q.FindMachine(id)

		fmt.Println("Looked up ", id, " and found on ", mid)

		// What EUS is that client on?
		cmd := message.NewClientCmd(mid, id.ListenerId(), "ClientJoin", map[string]interface{}{"RoomId": r.info.RoomId})
		fmt.Println("Create ", cmd)
		pubsub.SendMessage(&cmd)

		return nil
	})
	objTemplate.Set("Join", join)

	leave := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		//id := misc.ListenerId(info.Args()[0].String())
		// TODO Leave(r.RoomId, misc.ListenerId(id))
		return nil
	})
	objTemplate.Set("Leave", leave)

	return objTemplate
}
