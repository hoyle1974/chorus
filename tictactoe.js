log('tic tac toe!')

xUser = ""
oUser = ""


function setCharAt(str,index,chr) {
    if(index > str.length-1) return str;
    return str.substring(0,index) + chr + str.substring(index+1);
}

board = "........."

ready=0
turn='x'

function onMove(msg) {
    if ((turn=='x' && xUser === msg.SenderId) || (turn=='o' && oUser === msg.SenderId)) {
        x=msg.Data.x
        y=msg.Data.y
        p=y*3 +x
        board=setCharAt(board, p, turn)


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
            sendMsg({ReceiverId:xUser, Cmd:"turn"})
        }        
        return
    }


    log('room is ful')
}