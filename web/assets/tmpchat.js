const SEPARATOR = " • ";

const SignalingEvent = {
    "Entrance": 1,
    "Exit": 2,
    "NameChange": 3,
    "Clear": 4,
    "Welcome": 5,
    "RTCOffer": 6,
    "RTCAnswer": 7,
    "RTCICECandidate": 8
};

const newMessage = (type, content) => {
    return {
        "channel_name": unescape(window.location.pathname.substr(1)),
        "from_user": {"id": myUserId, "name": myName},
        "type": type,
        "content": content
    };
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
    }
    else {
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

const doEntrance = user => {
    if (user["id"] !== myUserId) {
        let tag = document.createElement("div");
        let name = document.createElement("span");
        name.className = user["id"];
        name.innerHTML = user["name"];
        tag.appendChild(name);
        tag.style.display = "inline";
        tag.innerHTML = SEPARATOR + tag.innerHTML;
        document.getElementById("namechange").appendChild(tag)
    }
};

const doExit = user => {
    if (user["id"] !== myUserId) {
        let element = document.getElementById("namechange").getElementsByClassName(user["id"])[0];
        element.parentElement.outerHTML = "";
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
    let toChange = document.getElementsByClassName(userId);
    for (let i=0; i<toChange.length; i++) {
        toChange[i].innerHTML = message["content"];
    }
    if (userId === myUserId) {
        myName = he.unescape(message["content"]);
        document.getElementById("myname").value = myName;
    }
};

const addNewDataChannel = member => {
    let dataChannel = rtcPeerConns[member["id"]]["conn"]
        .createDataChannel(unescape(window.location.pathname.substr(1)));
    dataChannel.onclose = () => {
        console.log(`dataChannel for ${member["id"]} has closed`);
        delete rtcPeerConns[member["id"]];
    };
    dataChannel.onopen = () => console.log(`dataChannel for ${member["id"]} has opened`);
    dataChannel.onmessage = event => write(JSON.parse(event.data));
    rtcPeerConns[member["id"]]["dataChannel"] = dataChannel;
};

const announce = message => {
    let announcement = document.createElement("div");
    announcement.className = "systemmessage";
    announcement.innerHTML = message.content;
    document.getElementById("messagelog").appendChild(announcement);
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

window.addEventListener("load", () => {
    const input = document.getElementById("messagetext");

    let ws = new WebSocket(`wss://${signalingHost}`);

    ws.onopen = () => {
        let nametag = document.createElement("span");
        nametag.className = myUserId;
        nametag.textContent = myName;
        let content = nametag.outerHTML + " joined";
        ws.send(JSON.stringify(newMessage(SignalingEvent.Entrance, content)));
    };

    window.onunload = window.onbeforeunload = () => {
        let nametag = document.createElement("span");
        nametag.className = myUserId;
        nametag.textContent = myName;
        let content = nametag.outerHTML + " left";
        ws.send(JSON.stringify(newMessage(SignalingEvent.Exit, content)));
        ws.close();
    };

    const addNewRTCPeerConn = (member, isLocal) => {
        let pc =  new RTCPeerConnection({
            iceServers: [{ urls: "stun:stun.l.google.com:19302" }]
        });
        pc.oniceconnectionstatechange = () => info(pc.iceConnectionState);
        pc.onicecandidate = event => {
            if (event.candidate) {
                let desc = btoa(JSON.stringify(event.candidate));
                let msg = newMessage(SignalingEvent.RTCICECandidate, desc);
                msg["to_user_id"] = member["id"];
                ws.send(JSON.stringify(msg));
            }
        };
        pc.onnegotiationneeded = () => pc.createOffer()
            .then(d => pc.setLocalDescription(d))
            .then(() => {
                if (isLocal) {
                    let desc = btoa(JSON.stringify(pc.localDescription));
                    let msg = newMessage(SignalingEvent.RTCOffer, desc);
                    msg["to_user_id"] = member["id"];
                    ws.send(JSON.stringify(msg));
                }
            })
            .catch(info);
        rtcPeerConns[member["id"]] = {
            "conn": pc,
        };
    };

    const answerRTCOffer = message => {
        rtcPeerConns[message["from_user"]["id"]]["conn"].ondatachannel = function (event) {
            event.channel.onmessage = e => write(JSON.parse(e.data));
            rtcPeerConns[message["from_user"]["id"]]["dataChannel"] = event.channel;
        };
        let offerDesc = JSON.parse(atob(message["content"]));
        let peerConn = rtcPeerConns[message["from_user"]["id"]]["conn"];
        peerConn.setRemoteDescription(new RTCSessionDescription(offerDesc))
            .then(() => peerConn.createAnswer())
            .then(answer => peerConn.setLocalDescription(answer))
            .then(() => {
                let desc = btoa(JSON.stringify(peerConn.localDescription));
                let response = newMessage(SignalingEvent.RTCAnswer, desc);
                response["to_user_id"] = message["from_user"]["id"];
                ws.send(JSON.stringify(response));
            })
            .catch(info);
    };

    ws.onmessage = event => {
        let message = JSON.parse(event.data);
        switch (message.type) {
            case SignalingEvent.Entrance:
                if (message["from_user"]["id"] !== myUserId) {
                    addNewRTCPeerConn(message["from_user"], false); // "remote" means "already here"
                }
                doEntrance(message["from_user"]);
                announce(message);
                break;
            case SignalingEvent.Exit:
                doExit(message["from_user"]);
                announce(message);
                break;
            case SignalingEvent.NameChange:
                doNameChange(message);
                break;
            case SignalingEvent.Clear:
                doClear();
                break;
            case SignalingEvent.Welcome:
                for (let i=0; i < message["content"].length; i++) {
                    let member = message["content"][i];
                    doEntrance(member);
                    if (member["id"] !== myUserId) {
                        addNewRTCPeerConn(member, true); // "local" means "newly arrived"
                        addNewDataChannel(member);
                    }
                }
                break;
            case SignalingEvent.RTCOffer:
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
                console.log(candidate);
                rtcPeerConns[message["from_user"]["id"]]["conn"]
                    .addIceCandidate(candidate)
                    .catch(info);
        }
    };

    ws.onerror = () => {
        ws.close();
    };

    input.onfocus = () => {
        let m = document.getElementById("messagelog");
        m.scrollTop = m.scrollHeight;
    };

    let doubleEnterTs; // Clear chat by double-pressing Enter with no text in the input field
    input.onkeypress = event => {
        if(event.key === "Enter" && !event.shiftKey) {
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

    document.getElementById("send").onclick = () => {
        if (!ws || input.value === "") {
            return false;
        }
        let msg = newMessage(0, input.value);
        write(msg);
        for (let id in rtcPeerConns) {
            let dc = rtcPeerConns[id]["dataChannel"];
            if (dc && dc.readyState === "open") {
                dc.send(JSON.stringify(msg));
            }
        }
        input.value = "";
        return false;
    };

    document.getElementById("namechange").onsubmit = () => {
        if (myName === "") {
            return false;
        }
        let message = newMessage(
            SignalingEvent.NameChange,
            he.escape(document.getElementById("myname").value)
        );
        doNameChange(message);
        if (!ws) {
            return false;
        }
        ws.send(JSON.stringify(message));
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
        ws.send(JSON.stringify(newMessage(SignalingEvent.Clear, null)));
        return false;
    };

    input.focus();
});