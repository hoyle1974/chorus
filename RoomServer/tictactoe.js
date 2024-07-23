log('tic tac toe!')

xUser = ""
oUser = ""


function setCharAt(str,index,chr) {
    if(index > str.length-1) return str;
    return str.substring(0,index) + chr + str.substring(index+1);
}

//board = "........."
board = "xx......."

ready=0
turn='x'


function boardAt(index) {
    return board.charAt(p)
}
function boardAtXY(x,y) {
    p=y * 3 + x
    return board.charAt(p)
}
function checkWin() {
    if (boardAtXY(0,0)!='.' && boardAtXY(0,0) === boardAtXY(1,0) && boardAtXY(0,0) == boardAtXY(2,0)) return true
    if (boardAtXY(0,1)!='.' && boardAtXY(0,1) === boardAtXY(1,1) && boardAtXY(0,1) == boardAtXY(2,1)) return true
    if (boardAtXY(0,2)!='.' && boardAtXY(0,2) === boardAtXY(1,2) && boardAtXY(0,2) == boardAtXY(2,2)) return true

    if (boardAtXY(0,0)!='.' && boardAtXY(0,0) === boardAtXY(0,1) && boardAtXY(0,0) == boardAtXY(0,2)) return true
    if (boardAtXY(1,0)!='.' && boardAtXY(1,0) === boardAtXY(1,1) && boardAtXY(1,0) == boardAtXY(1,2)) return true
    if (boardAtXY(2,0)!='.' && boardAtXY(2,0) === boardAtXY(2,1) && boardAtXY(2,0) == boardAtXY(2,2)) return true

    if (boardAtXY(0,0)!='.' && boardAtXY(0,0) === boardAtXY(1,1) && boardAtXY(0,0) == boardAtXY(2,2)) return true
    if (boardAtXY(2,0)!='.' && boardAtXY(2,0) === boardAtXY(1,1) && boardAtXY(2,0) == boardAtXY(0,2)) return true
    return false
}

function onMove(msg) {
    if ((turn=='x' && xUser === msg.SenderId) || (turn==='o' && oUser === msg.SenderId)) {
        x=Number(msg.Data.x)
        y=Number(msg.Data.y)
        p=y*3 + x
        if (p<0 || p>8) {
            sendMsg({ReceiverId:msg.SenderId, Cmd:"error",Data:{Msg:"position was not on board: " + p}})
            return
        }
        b = boardAt(p)
        if (b != '.') {
            sendMsg({ReceiverId:msg.SenderId, Cmd:"error",Data:{Msg:"illegal move, space not available: " + b}})
            return
        }

        board=setCharAt(board, p, turn)

        if (checkWin()) {
            sendMsg({Cmd:"win", Data:{Winner:turn}})
            turn=''
            thisRoom().Leave(oUser)
            thisRoom().Leave(xUser)
            return
        }

        if (turn=='x') {
            turn='o'
            sendMsg({ReceiverId:oUser, Cmd:"turn",Data:{Board:board}})
        } else {
            turn='x'
            sendMsg({ReceiverId:xUser, Cmd:"turn",Data:{Board:board}})
        }

    } else {
        sendMsg({ReceiverId:msg.SenderId, Cmd:"error", Data:{msg:"Not your turn"}})
    }
}

function onJoin(msg) {

    if (xUser === "") {
        xUser = msg.SenderId
        sendMsg({ReceiverId:xUser, Cmd:"x-user"})
        ready++
        if (ready==2) {
            sendMsg({ReceiverId:xUser, Cmd:"turn",Data:{Board:board}})
        }
        return
    }
    if (oUser === "") {
        oUser = msg.SenderId
        sendMsg({ReceiverId:oUser, Cmd:"o-user"})
        ready++
        if (ready==2) {
            sendMsg({ReceiverId:xUser, Cmd:"turn",Data:{Board:board}})
        }        
        return
    }


    log('room is full')
}

function onLeave(msg) {
    if (ready != 0) {
        sendMsg({Cmd:"endgame"})
        endRoom()
        ready=0
    }
}

function onRoomEmpty() {
    log("Room is empty, ending the room")
    endRoom()
}
  