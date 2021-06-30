package media

import (
	"github.com/HunterGooD/webrtcGolangServer/internal/util"
	"github.com/pion/webrtc/v3"
)

var defaultPeerCfg = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.stunprotocol.org:3478"},
		},
	},
}

const (
	//Единица передачи мультимедиа состоит из 1400, разделенных на 7 пакетов. Количество пакетов RTP, необходимых для каждого обнаружения.
	averageRtpPacketsPerFrame = 7
)

type WebRTCEngine struct {
	cfg webrtc.Configuration

	mediaEngine webrtc.MediaEngine

	api *webrtc.API
}

func NewWebRTCEngine() *WebRTCEngine {
	urls := []string{} //conf.SFU.Ices//[]string{"stun:stun.stunprotocol.org:3478"};//conf.SFU.Ices

	w := &WebRTCEngine{
		mediaEngine: webrtc.MediaEngine{},
		cfg: webrtc.Configuration{
			SDPSemantics: webrtc.SDPSemanticsUnifiedPlanWithFallback,
			ICEServers: []webrtc.ICEServer{
				{
					URLs: urls,
				},
			},
		},
	}
	if err := w.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeVP8, ClockRate: 90000, Channels: 0, SDPFmtpLine: "", RTCPFeedback: nil},
		PayloadType:        96,
	}, webrtc.RTPCodecTypeVideo); err != nil {
		panic(err)
	}
	if err := w.mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 2, SDPFmtpLine: "minptime=10; useinbandfec=1", RTCPFeedback: nil},
		PayloadType:        111,
	}, webrtc.RTPCodecTypeAudio); err != nil {
		panic(err)
	}

	w.api = webrtc.NewAPI(webrtc.WithMediaEngine(&w.mediaEngine))
	return w
}

func (s WebRTCEngine) CreateSender(offer webrtc.SessionDescription, pc **webrtc.PeerConnection, addLocalTrack **webrtc.TrackLocalStaticRTP, stop chan int) (answer webrtc.SessionDescription, err error) {

	*pc, err = s.api.NewPeerConnection(s.cfg)
	util.Infof("WebRTCEngine.CreateSender pc=%p", *pc)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	if *addLocalTrack != nil {
		(*pc).AddTrack(*addLocalTrack)
		err = (*pc).SetRemoteDescription(offer)
		if err != nil {
			return webrtc.SessionDescription{}, err
		}
	}

	answer, err = (*pc).CreateAnswer(nil)
	err = (*pc).SetLocalDescription(answer)
	util.Infof("WebRTCEngine.CreateReceiver ok")
	return answer, err

}

func (s WebRTCEngine) CreateReceiver(offer webrtc.SessionDescription, pc **webrtc.PeerConnection, localTrack **webrtc.TrackLocalStaticRTP, stop chan int, pli chan int) (answer webrtc.SessionDescription, err error) {

	*pc, err = s.api.NewPeerConnection(s.cfg)
	util.Infof("WebRTCEngine.CreateReceiver pc=%p", *pc)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	_, err = (*pc).AddTransceiverFromKind(webrtc.RTPCodecTypeVideo)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	_, err = (*pc).AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	(*pc).OnTrack(func(t *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {

		localTrack, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
		if err != nil {
			panic(err)
		}

		buf := make([]byte, 1500)
		for {
			i, _, err := t.Read(buf)
			if err != nil {
				return
			}

			if _, err = localTrack.Write(buf[:i]); err != nil {
				return
			}
		}
		// //Обработка видео
		// if remoteTrack.Codec().MimeType == webrtc.MimeTypeVP8 ||
		// 	remoteTrack.Codec().MimeType == webrtc.MimeTypeVP9 ||
		// 	remoteTrack.Codec().MimeType == webrtc.MimeTypeH264 {
		// 	*videoTrack, err = webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: remoteTrack.Codec().MimeType},  remoteTrack.ID(), remoteTrack.StreamID())

		// 	go func() {
		// 		for {
		// 			select {
		// 			case <-pli:
		// 				(*pc).WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.}})
		// 			case <-stop:
		// 				return
		// 			}
		// 		}
		// 	}()

		// 	var pkt rtp.Depacketizer

		// 	if remoteTrack.Codec().MimeType == webrtc.MimeTypeVP8 {
		// 		pkt = &codecs.VP8Packet{}
		// 	} else if remoteTrack.Codec().MimeType == webrtc.MimeTypeVP9 {
		// 		pkt = &codecs.VP9Packet{}
		// 	} else if remoteTrack.Codec().MimeType == webrtc.MimeTypeH264 {
		// 		pkt = &codecs.H264Packet{}
		// 	}

		// 	builder := samplebuilder.New(averageRtpPacketsPerFrame*5, pkt, 16000)
		// 	for {
		// 		select {

		// 		case <-stop:
		// 			return
		// 		default:
		// 			rtp, _, err := remoteTrack.ReadRTP()
		// 			if err != nil {
		// 				if err == io.EOF {
		// 					return
		// 				}
		// 				util.Errorf(err.Error())
		// 			}

		// 			builder.Push(rtp)
		// 			for s := builder.Pop(); s != nil; s = builder.Pop() {
		// 				if err := (*videoTrack).WriteSample(*s); err != nil && err != io.ErrClosedPipe {
		// 					util.Errorf(err.Error())
		// 				}
		// 			}
		// 		}
		// 	}
		// } else {
		// 	*audioTrack, err = webrtc.NewTrackLocalStaticSample(remoteTrack.PayloadType(), remoteTrack.SSRC(), "audio", remoteTrack.Label())

		// 	rtpBuf := make([]byte, 1400)
		// 	for {
		// 		select {
		// 		case <-stop:
		// 			return
		// 		default:
		// 			i, _, err := remoteTrack.Read(rtpBuf)
		// 			if err == nil {
		// 				(*audioTrack).WriteSample(media.Sample{Data: rtpBuf[:i]})
		// 			} else {
		// 				util.Infof(err.Error())
		// 			}
		// 		}
		// 	}
		// }
	})

	err = (*pc).SetRemoteDescription(offer)
	if err != nil {
		return webrtc.SessionDescription{}, err
	}

	answer, err = (*pc).CreateAnswer(nil)
	err = (*pc).SetLocalDescription(answer)
	util.Infof("WebRTCEngine.CreateReceiver ok")
	return answer, err

}
