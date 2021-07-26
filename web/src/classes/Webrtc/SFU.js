import {
    EventEmitter
} from 'events';
import RTC from './RTC';

export default class SFU extends EventEmitter {

    constructor(userID, roomId) {
        super();
        this._rtc = new RTC();

        var sfuUrl = "ws://localhost:3000/ws?uid=" + userID + "&roomId=" + roomId;

        this.socket = new WebSocket(sfuUrl);

        this.socket.onopen = () => {
            console.log("WebSocket подключен..");
            this._onRoomConnect();
        };

        this.socket.onmessage = (e) => {
            var parseMessage = JSON.parse(e.data);
            console.log(`get message  ${e.data}`);
            switch (parseMessage.type) {
                case 'joinRoom':
                    console.log(parseMessage);
                    break;
                case 'onJoinRoom':
                    console.log(parseMessage);
                    break;
                case 'onPublish':
                    this.onPublish(parseMessage);
                    break;
                case 'onUnpublish':
                    this.onUnpublish(parseMessage);
                    break;
                case 'onSubscribe':
                    this.onSubscribe(parseMessage);
                    break;
                case 'heartPackage':
                    console.log("heartPackage:::");
                    break;
                default:
                    console.error('Неизвестное сообщение', parseMessage);
            }
        };

        this.socket.onerror = (e) => {
            console.log('onerror::' + e);
        };

        this.socket.onclose = (e) => {
            console.log('onclose::' + e);
        };
    }


    _onRoomConnect = () => {
        console.log('onRoomConnect');

        this._rtc.on('localstream', (id, stream) => {
            this.emit('addLocalStream', id, stream);
        })

        this._rtc.on('addstream', (id, stream) => {
            this.emit('addRemoteStream', id, stream);
        })

        this._rtc.on('removestream', (id, stream) => {
            this.emit('removeRemoteStream', id, stream);
        })

        this.emit('connect');
    }

    join(userID, userName, roomId) {
        console.log('Join to [' + roomId + ']');
        this.userID = userID;
        this.userName = userName;
        this.roomId = roomId;

        let message = {
            'type': 'join',
            'data': {
                'userName': userName,
                'userID': userID,
                'roomId': roomId,
            }
        };
        this.send(message);

    }

    send (data) {
        this.socket.send(JSON.stringify(data));
    }


    publish() {
        this._createSender(this.userID);
    }

    async _createSender(pubID) {

        try {
            let sender = await this._rtc.createSender(pubID);
            this.sender = sender;

            sender.pc.onicecandidate = async () => {
                if (!sender.senderOffer) {
                    var offer = sender.pc.localDescription;
                    sender.senderOffer = true;
                    await this.publishToServer(offer, pubID);
                }
            }
            let desc = await sender.pc.createOffer({
                offerToReceiveVideo: false,
                offerToReceiveAudio: false
            })
            sender.pc.setLocalDescription(desc);

        } catch (error) {
            console.log('onCreateSender error =>' + error);
        }

    }

    async publishToServer(offer, pubID) {
        let message = {
            'type': 'publish',
            'data': {
                'jsep': offer,
                'pubID': pubID,
                'userName': this.userName,
                'userID': this.userID,
                'roomId': this.roomId,
            }
        };
        this.send(message);
    }

    async onPublish(message) {
        if (this.sender && message['data']['userID'] == this.userID) {
            console.log('onPublish:::user Id:::' + message['data']['userID']);
            await this.sender.pc.setRemoteDescription(message['data']['jsep']);
        }

        if (message['data']['userID'] != this.userID) {
            if (typeof message.data === "string") {
                message.data = JSON.parse(message.data)
            }
            console.log('onPublish:::user id:::' + message['data']['userID']);
            await this._onRtcCreateReceiver(message['data']['userID']);
        }

    }

    onUnpublish(meesage) {
        console.log('Покинул:' + meesage['data']['pubID']);
        this._rtc.closeReceiver(meesage['data']['pubID']);
    }

    async _onRtcCreateReceiver(pubID) {
        try {
            let receiver = await this._rtc.createReciver(pubID);

            receiver.pc.onicecandidate = async () => {
                if (!receiver.senderOffer) {
                    var offer = receiver.pc.localDescription;
                    receiver.senderOffer = true;
                    await this.subscribeFromServer(offer, pubID);
                }
            }

            let desc = await receiver.pc.createOffer();
            receiver.pc.setLocalDescription(desc);

        } catch (error) {
            console.log('onRtcCreateReceiver error =>' + error);
        }
    }


    async subscribeFromServer(offer, pubID) {
        console.log(pubID);
        let message = {
            'type': 'subscribe',
            'data': {
                'jsep': offer,
                'pubID': pubID,
                'userName': this.userName,
                'userID': this.userID,
                'roomId': this.roomId,
            }
        };
        this.send(message);
    }


    onSubscribe(message) {
        var receiver = this._rtc.getReceivers(message['data']['pubID']);
        if (receiver) {
            console.log('Id:' + message['data']['pubID']);
            receiver.pc.setRemoteDescription(message['data']['jsep']);
        } else {
            console.log('receiver == null');
        }
    }

}