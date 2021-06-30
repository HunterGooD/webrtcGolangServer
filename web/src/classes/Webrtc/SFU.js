import {
    EventEmitter
} from 'events';
import RTC from './RTC';

export default class SFU extends EventEmitter {

    constructor(userId, roomId) {
        super();
        this._rtc = new RTC();

        var sfuUrl = "ws://localhost:3000/ws?userId=" + userId + "&roomId=" + roomId;

        this.socket = new WebSocket(sfuUrl);

        this.socket.onopen = () => {
            console.log("WebSocket подключен..");
            this._onRoomConnect();
        };

        this.socket.onmessage = (e) => {
            var parseMessage = JSON.parse(e.data);

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

    join(userId, userName, roomId) {
        console.log('Join to [' + roomId + ']');
        this.userId = userId;
        this.userName = userName;
        this.roomId = roomId;

        let message = {
            'type': 'join',
            'data': {
                'userName': userName,
                'userId': userId,
                'roomId': roomId,
            }
        };
        this.send(message);

    }

    send = (data) => {
        this.socket.send(JSON.stringify(data));
    }


    publish() {
        this._createSender(this.userId);
    }

    async _createSender(pubid) {

        try {
            let sender = await this._rtc.createSender(pubid);
            this.sender = sender;

            sender.pc.onicecandidate = async () => {
                if (!sender.senderOffer) {
                    var offer = sender.pc.localDescription;
                    sender.senderOffer = true;
                    await this.publishToServer(offer, pubid);
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

    async publishToServer(offer, pubid) {
        let message = {
            'type': 'publish',
            'data': {
                'jsep': offer,
                'pubid': pubid,
                'userName': this.userName,
                'userId': this.userId,
                'roomId': this.roomId,
            }
        };
        this.send(message);
    }

    async onPublish(message) {
        if (this.sender && message['data']['userId'] == this.userId) {
            console.log('onPublish:::user Id:::' + message['data']['userId']);
            await this.sender.pc.setRemoteDescription(message['data']['jsep']);
        }

        if (message['data']['userId'] != this.userId) {
            console.log('onPublish:::user id:::' + message['data']['userId']);
            await this._onRtcCreateReceiver(message['data']['userId']);
        }

    }

    onUnpublish(meesage) {
        console.log('Покинул:' + meesage['data']['pubid']);
        this._rtc.closeReceiver(meesage['data']['pubid']);
    }

    async _onRtcCreateReceiver(pubid) {
        try {
            let receiver = await this._rtc.createReciver(pubid);

            receiver.pc.onicecandidate = async () => {
                if (!receiver.senderOffer) {
                    var offer = receiver.pc.localDescription;
                    receiver.senderOffer = true;
                    await this.subscribeFromServer(offer, pubid);
                }
            }

            let desc = await receiver.pc.createOffer();
            receiver.pc.setLocalDescription(desc);

        } catch (error) {
            console.log('onRtcCreateReceiver error =>' + error);
        }
    }


    async subscribeFromServer(offer, pubid) {
        let message = {
            'type': 'subscribe',
            'data': {
                'jsep': offer,
                'pubid': pubid,
                'userName': this.userName,
                'userId': this.userId,
                'roomId': this.roomId,
            }
        };
        this.send(message);
    }


    onSubscribe(message) {
        var receiver = this._rtc.getReceivers(message['data']['pubid']);
        if (receiver) {
            console.log('Id:' + message['data']['pubid']);
            receiver.pc.setRemoteDescription(message['data']['jsep']);
        } else {
            console.log('receiver == null');
        }
    }

}