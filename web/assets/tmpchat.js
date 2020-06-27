const SEPARATOR = " • ";

const SignalingEvent = {
    "Entrance": 0,
    "RTCOffer": 1,
    "RTCAnswer": 2,
    "RTCICECandidate": 3,
    "TURNCredRequest": 4,
    "TURNCredResponse": 5
};

const TmpchatEvent = {
    "Message": 0,
    "Clear": 1,
    "NameChange": 2,
    "Exit": 3
};

const newMessage = (type, content) => {
    return {
        "channel_name": unescape(window.location.pathname.substr(1)),
        "from_user": {"id": myUserId, "name": myName},
        "type": type,
        "content": content
    };
};

const broadcast = message => {
    for (let id in rtcPeerConns) {
        let dc = rtcPeerConns[id]["dataChannel"];
        if (dc && dc.readyState === "open") {
            dc.send(JSON.stringify(message));
        }
    }
};

const nameTag = (message, isFromMe) => {
    let tag = document.createElement("div");
    let name = document.createElement("span");
    name.className = message["from_user"]["id"];
    name.innerHTML = message["from_user"]["name"];
    tag.appendChild(name);
    if (isFromMe) {
        tag.className = "myname";
        tag.innerHTML = SEPARATOR + tag.innerHTML;
    } else {
        tag.className = "theirname";
        tag.innerHTML = tag.innerHTML + SEPARATOR;
    }
    return tag;
};

const shouldStackMsg = (message, lastMsgElement) => {
    if (message["type"] !== 0 || !lastMsgElement) {
        return false;
    }
    if (lastMsgElement.className === "systemmessage") { // only stack user messages
        return false;
    }
    let lastMsgUserId = lastMsgElement.firstElementChild.firstElementChild.className;
    return message["from_user"]["id"] === lastMsgUserId;
};

const announceEntrance = user => {
    let nametag = document.createElement("span");
    nametag.className = user["id"];
    nametag.textContent = user["name"];
    announce(nametag.outerHTML + " joined");
};

const announceExit = user => {
    let nametag = document.createElement("span");
    nametag.className = user["id"];
    nametag.textContent = user["name"];
    announce(nametag.outerHTML + " left");
};

const announce = announcementHTML => {
    let announcement = document.createElement("div");
    announcement.className = "systemmessage";
    announcement.innerHTML = announcementHTML;
    document.getElementById("messagelog").appendChild(announcement);
};

const appendToRoll = user => {
    userNames[user["id"]] = user["name"];
    let tag = document.createElement("div");
    let name = document.createElement("span");
    name.className = user["id"];
    name.innerHTML = user["name"];
    tag.appendChild(name);
    tag.style.display = "inline";
    tag.innerHTML = SEPARATOR + tag.innerHTML;
    document.getElementById("namechange").appendChild(tag);
};

const doExit = user => {
    if (user["id"] !== myUserId) {
        let element = document.getElementById("namechange").getElementsByClassName(user["id"])[0];
        element.parentElement.outerHTML = "";
        announceExit(user);
    }
};

const doClear = () => {
    let node = document.getElementById("messagelog");
    while (node.firstChild) {
        node.removeChild(node.firstChild);
    }
    document.getElementById("messagetext").focus();
};

const doNameChange = message => {
    let userId = message["from_user"]["id"];
    let newName = message["content"];
    userNames[userId] = newName;
    let toChange = document.getElementsByClassName(userId);
    for (let i = 0; i < toChange.length; i++) {
        toChange[i].innerHTML = newName;
    }
    if (userId === myUserId) {
        myName = he.unescape(newName);
        document.getElementById("myname").value = myName;
    }
};

const newNameIsOk = newName =>
    !(newName === "" || newName === myName || userNames.hasOwnProperty(newName));

const addNewDataChannel = member => {
    let dataChannel = rtcPeerConns[member["id"]]["conn"]
        .createDataChannel(unescape(window.location.pathname.substr(1)));
    dataChannel.onclose = () => {
        console.log(`dataChannel for ${member["id"]} has closed`);
        delete rtcPeerConns[member["id"]];
    };
    dataChannel.onopen = () => rtcPeerConns[member["id"]]["dataChannel"] = dataChannel;
    dataChannel.onmessage = event => handleTmpchatEvent(event);
    rtcPeerConns[member["id"]]["dataChannel"] = dataChannel;
};

const handleTmpchatEvent = event => {
    let message = JSON.parse(event.data);
    switch (message.type) {
        case TmpchatEvent.Message:
            write(message);
            break;
        case TmpchatEvent.Clear:
            doClear();
            break;
        case TmpchatEvent.NameChange:
            doNameChange(message);
            break;
        case TmpchatEvent.Exit:
            doExit(message["from_user"]);
            break;
    }
};

const write = message => {
    let messageLog = document.getElementById("messagelog");
    const lastMsgElement = messageLog.lastElementChild;
    if (shouldStackMsg(message, lastMsgElement)) {
        let currentText = lastMsgElement.getElementsByClassName("chatmessage")[0];
        currentText.textContent += "\n" + message["content"];
    } else {
        let isFromMe = message["from_user"]["id"] === myUserId;
        let name = nameTag(message, isFromMe);
        let msg = document.createElement("div");
        msg.className = isFromMe ? "mymessage" : "theirmessage";
        let pre = document.createElement("pre");
        pre.className = "chatmessage";
        pre.textContent = message["content"];
        msg.appendChild(pre);
        msg.insertAdjacentHTML("afterbegin", name.outerHTML);
        messageLog.appendChild(msg);
    }
    if (document.activeElement === document.getElementById("messagetext")) {
        messageLog.scrollTop = messageLog.scrollHeight;
    }
};

const info = txt => {
    document.getElementById("info").innerText = txt;
};

const rtcPeerConns = {};
const userNames = {};

let ws = new WebSocket(`${signalingURL}`);

ws.sendMessage = message => ws.send(JSON.stringify(message));

ws.onopen = () => {
    ws.sendMessage(newMessage(SignalingEvent.TURNCredRequest, null));
};

window.onunload = window.onbeforeunload = () => {
    broadcast(newMessage(TmpchatEvent.Exit), null);
    ws.close();
};

const addNewRTCPeerConn = (turnCreds, member, isLocal) => {
    // isLocal: true if we're already in the chat and adding an
    // RTCPeerConnection for a new arrival. false if we're adding
    // RTCPeerConnections for existing members because we're new.
    let pc = new RTCPeerConnection({
        iceServers: [{
            urls: "turn:turn.tmpch.at:3478",
            username: turnCreds["username"],
            credential: turnCreds["credential"]
        }]
    });
    pc.oniceconnectionstatechange = () => info(pc.iceConnectionState);
    pc.onicecandidate = event => {
        if (event.candidate !== null) {
            let desc = btoa(JSON.stringify(event.candidate));
            let msg = newMessage(SignalingEvent.RTCICECandidate, desc);
            msg["to_user_id"] = member["id"];
            ws.sendMessage(msg);
        }
    };
    pc.onnegotiationneeded = () => pc.createOffer()
        .then(d => pc.setLocalDescription(d))
        .then(() => {
            if (isLocal) {
                let desc = btoa(JSON.stringify(pc.localDescription));
                let msg = newMessage(SignalingEvent.RTCOffer, desc);
                msg["to_user_id"] = member["id"];
                ws.sendMessage(msg);
            }
        })
        .catch(info);
    pc.ondatachannel = function (event) {
        event.channel.onopen = () => rtcPeerConns[member["id"]]["dataChannel"] = event.channel;
        event.channel.onmessage = event => handleTmpchatEvent(event);
    };
    rtcPeerConns[member["id"]] = {
        "conn": pc,
    };
};

const answerRTCOffer = message => {
    let offerDesc = JSON.parse(atob(message["content"]));
    rtcPeerConns.add(message["from_user"], false);
    let peerConn = rtcPeerConns[message["from_user"]["id"]]["conn"];
    peerConn.setRemoteDescription(new RTCSessionDescription(offerDesc))
        .then(() => peerConn.createAnswer())
        .then(answer => peerConn.setLocalDescription(answer))
        .then(() => {
            let desc = btoa(JSON.stringify(peerConn.localDescription));
            let response = newMessage(SignalingEvent.RTCAnswer, desc);
            response["to_user_id"] = message["from_user"]["id"];
            ws.sendMessage(response);
        })
        .catch(info);
};

ws.onmessage = event => {
    let message = JSON.parse(event.data);
    switch (message.type) {
        case SignalingEvent.Entrance:
            let member = message["content"];
            if (member["id"] !== myUserId) {
                rtcPeerConns.add(member, true);
                addNewDataChannel(member);
                appendToRoll(member);
            }
            announceEntrance(member);
            break;
        case SignalingEvent.RTCOffer:
            appendToRoll(message["from_user"]);
            answerRTCOffer(message);
            break;
        case SignalingEvent.RTCAnswer:
            let answerDesc = JSON.parse(atob(message["content"]));
            rtcPeerConns[message["from_user"]["id"]]["conn"]
                .setRemoteDescription(new RTCSessionDescription(answerDesc))
                .catch(info);
            break;
        case SignalingEvent.RTCICECandidate:
            let candidate = JSON.parse(atob(message["content"]));
            rtcPeerConns[message["from_user"]["id"]]["conn"]
                .addIceCandidate(candidate)
                .catch(info);
            break;
        case SignalingEvent.TURNCredResponse:
            rtcPeerConns.add = (member, isLocal) => addNewRTCPeerConn(message["content"], member, isLocal)
    }
};

ws.onerror = () => {
    ws.close();
};

window.onload = () => {
    const input = document.getElementById("messagetext");

    document.getElementById("send").onclick = () => {
        if (input.value === "") {
            return false;
        }
        let msg = newMessage(TmpchatEvent.Message, input.value);
        write(msg);
        broadcast(msg);
        input.value = "";
        return false;
    };

    document.getElementById("namechange").onsubmit = () => {
        let newName = he.escape(document.getElementById("myname").value);
        if (!newNameIsOk(newName)) {
            return false;
        }
        let message = newMessage(TmpchatEvent.NameChange, newName);
        doNameChange(message);
        broadcast(message);
        input.focus();
        return false;
    };

    document.getElementById("myname").onfocus = () => {
        document.getElementById("myname").value = "";
        return false;
    };

    document.getElementById("myname").onblur = () => {
        document.getElementById("myname").value = myName;
        return false;
    };

    document.getElementById("clear").onclick = () => {
        if (!ws) {
            doClear();
            return false;
        }
        doClear();
        broadcast(newMessage(TmpchatEvent.Clear, null));
        return false;
    };

    input.onfocus = () => {
        let m = document.getElementById("messagelog");
        m.scrollTop = m.scrollHeight;
    };

    let doubleEnterTs; // Clear chat by double-pressing Enter with no text in the input field
    input.onkeypress = event => {
        if (event.key === "Enter" && !event.shiftKey) {
            if (input.value === "" && (Date.now() - doubleEnterTs < 150)) {
                document.getElementById("clear").click();
                return false;
            }
            if (input.value === "") {
                doubleEnterTs = Date.now();
                return false;
            }
            document.getElementById("send").click();
            return false;
        }
    };

    input.focus();
}