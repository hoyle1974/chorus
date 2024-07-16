log('tic tac toe!')

xUser = ""
oUser = ""

map = "........."

ready=0
turn='x'

function onJoin(msg) {

    if (xUser === "") {
        xUser = msg.SenderId
        sendMsg({ReceiverId:xUser, Cmd:"x-user"})
        ready++
        if (ready==2) {
            sendMsg({ReceiverId:xUser, Cmd:"turn"})
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