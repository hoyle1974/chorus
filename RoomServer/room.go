package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hoyle1974/chorus/distributed"
	"github.com/hoyle1974/chorus/message"
	"github.com/hoyle1974/chorus/misc"
	"github.com/hoyle1974/chorus/pubsub"
	"github.com/hoyle1974/chorus/store"
	"rogchap.com/v8go"
)

// This represents a Room that is running in a RoomServer
// For each room in the cluster only one of these exists.  If
// You try to create two with the same room Id only one will work and
// the other will wait for the other server to go away.
type Room struct {
	state    GlobalServerState
	logger   *slog.Logger
	roomId   misc.RoomId
	members  distributed.Set
	consumer *pubsub.Consumer
	ctx      *v8go.Context
}

func (r *Room) Destroy() {
	r.logger.Info("Deleting room")
	store.Del(r.roomId.RoomKey())
}

func GetRoom(state GlobalServerState, roomId misc.RoomId, adminScript string) *Room {
	state.Dist.Put(roomId.RoomKey(), adminScript, state.MachineLease)

	r := &Room{
		state:   state,
		roomId:  roomId,
		logger:  state.logger.With("roomId", roomId, "script", adminScript),
		members: state.Dist.BindSet(roomId.RoomMembershipKey(), state.MachineLease),
	}
	ctx, err := createScriptEnvironmentForRoom(r, adminScript)
	if err != nil {
		state.logger.Error("createScriptEnvironmentForRoom", "error", err)
		return nil
	}
	r.ctx = ctx

	r.consumer = pubsub.NewConsumer(roomId.Topic(), r)
	r.consumer.StartConsumer()
	time.Sleep(time.Duration(1) * time.Second)

	// Ask anyone in the room to respond
	msg := message.NewMessage(roomId, roomId.ListenerId(), "", "Ping", map[string]interface{}{})
	pubsub.SendMessage(msg)

	return r
}

func (r *Room) OnMessageFromTopic(msg message.Message) {
	r.logger.Info("Room.OnMessageFromTopic", "msg", msg)

	if msg.RoomId != r.roomId {
		r.logger.Error("Received an error for the wrong room", "targetRoomId", msg.RoomId, "roomId", r.roomId)
		return
	}

	if msg.Cmd == "Pong" {
		r.logger.Debug("Pong", "memberId", msg.SenderId)
		r.members.SAdd(string(msg.SenderId))
		store.AddMemberToSet(r.roomId.RoomMembershipKey(), string(msg.SenderId))
	}
	if msg.Cmd == "Join" {
		r.logger.Debug("Join", "memberId", msg.SenderId)
		r.members.SAdd(string(msg.SenderId))
		store.AddMemberToSet(r.roomId.RoomMembershipKey(), string(msg.SenderId))
	}
	if msg.Cmd == "Leave" {
		r.logger.Debug("Leave", "memberId", msg.SenderId)
		r.members.SRem(string(msg.SenderId))
		store.RemoveMemberFromSet(r.roomId.RoomMembershipKey(), string(msg.SenderId))
	}

	r.callJSOnMessage(msg)

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
		msg.RoomId = room.roomId
		msg.SenderId = room.roomId.ListenerId()
		//TODO room.sendMsg(msg)
		pubsub.SendMessage(msg)

		return nil // you can return a value back to the JS caller if required
	})
	err = global.Set("sendMsg", sendMsg)
	if err != nil {
		return nil, fmt.Errorf("create sendMsg function: %w", err)
	}

	// create global log in JS context
	log := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		// TODO msg := room.style.Render(fmt.Sprintf("%v", info.Args()))
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

		newRoom := GetRoom(room.state, misc.RoomId(name), script)
		objTemplate := newRoom.JSTemplate(isolate)

		obj, err := objTemplate.NewInstance(info.Context())
		if err != nil {
			room.logger.Error("NewObjectTemplate", "err", err)
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
			room.logger.Error(fmt.Sprintf("js objTemplate NewInstance: %w", err))
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
	objTemplate.Set("Id", r.roomId)
	objTemplate.Set("Room", r)

	join := v8go.NewFunctionTemplate(isolate, func(info *v8go.FunctionCallbackInfo) *v8go.Value {
		//id := misc.ListenerId(info.Args()[0].String())
		// TODO
		//conn := connection.FindConnectionById(id)
		//r.join(id, conn)
		//Join(r.Id, misc.ListenerId(id))
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
