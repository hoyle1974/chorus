var id= ""


function onJoin(msg) {
  if (id==="") {
    id = msg.SenderId
  } else {
    room = NewRoom(id + " vs " + msg.SenderId, "tictactoe.js")
    room.Join(id)
    room.Join(msg.SenderId)
    id = ""
  }
}

function onLeave(msg) {
  if (msg.SenderID === id) {
    id = ""
  }
}
  