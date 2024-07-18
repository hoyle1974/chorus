var id= ""


function onJoin(msg) {  
  log("onJoin")

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

function onLeave(msg) {
  log("onLeave")

  if (msg.SenderId === id) {
    log("first user left",msg.SenderId)
    id = ""
  }
}
  