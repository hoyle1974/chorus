var id= ""


function onJoin(msg) {  
  log("Hello from onJoin")
  if (id==="") {
    log("first user joined",msg.SenderId)
    id = msg.SenderId
  } else {
    log("second user joined",msg.SenderId)

    room = newRoom(id + " vs " + msg.SenderId, "tictactoe.js")
    room.Join(id)
    room.Join(msg.SenderId)

    log("new room created and joined")
    id = ""
  }
}

function onSay(msg) {
  sendMsg({ReceiverId:"", Cmd:"say", Data:{"From":msg.SenderId, "Msg":msg.Data.Msg}})
}

function onLeave(msg) {
  if (msg.SenderId === id) {
    log("first user left",msg.SenderId)
    id = ""
  }
}

function onRoomEmpty() {
}


  