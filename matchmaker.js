var id= ""

log("------------------------")
log(JSON.stringify(thisRoom().Id))
log("------------------------")



function onJoin(msg) {
  if (id==="") {
    id = msg.SenderId
  } else {
    room = newRoom(id + " vs " + msg.SenderId, "tictactoe.js")
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
  