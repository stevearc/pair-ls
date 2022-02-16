package lsp_handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"pair-ls/server"

	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func (h *LspHandler) connectToPeer(offer webrtc.SessionDescription) (*webrtc.PeerConnection, *webrtc.SessionDescription, error) {
	config := webrtc.Configuration{}
	peerConnection, err := h.rtc.NewPeerConnection(config)
	if err != nil {
		return nil, nil, err
	}
	peerConnection.SetRemoteDescription(offer)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, nil, err
	}
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return nil, nil, err
	}

	return peerConnection, &answer, nil
}

func (h *LspHandler) respondRTCPeer(callToken string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(callToken)
	if err != nil {
		return "", err
	}
	var data PeerToken
	if err := json.Unmarshal(decoded, &data); err != nil {
		return "", err
	}

	if data.Desc.Type == webrtc.SDPTypeOffer {
		// If we were given an offer, create and respond with an answer
		peerConnection, answer, err := h.connectToPeer(*data.Desc)
		if err != nil {
			return "", err
		}
		h.runPeerConnection(peerConnection, func() {})
		json, err := json.Marshal(answer)
		if err != nil {
			return "", err
		}

		return base64.StdEncoding.EncodeToString(json), nil
	} else {
		// If we were given an offer, find the corresponding offer connection and
		// complete it
		peerConnection := h.getConn(data.ClientID)
		if peerConnection == nil {
			return "", errors.New("No matching connection found")
		}
		peerConnection.SetRemoteDescription(*data.Desc)
		return "", nil
	}
}

func (h *LspHandler) callRTCPeer() (*webrtc.SessionDescription, string, error) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{
					"stun:openrelay.metered.ca:80",
					"stun:stun.l.google.com:19302",
					"stun:stun1.l.google.com:19302",
					"stun:stun2.l.google.com:19302",
					"stun:stun3.l.google.com:19302",
					"stun:stun4.l.google.com:19302",
				},
			},
			{
				URLs: []string{
					"turn:openrelay.metered.ca:80",
					"turn:openrelay.metered.ca:443",
					"turn:openrelay.metered.ca:443?transport=tcp",
					"turns:openrelay.metered.ca:443",
				},
				Username:   "openrelayproject",
				Credential: "openrelayproject",
			},
		},
	}
	clientID, err := createClientID()
	if err != nil {
		return nil, "", err
	}
	peerConnection, err := h.rtc.NewPeerConnection(config)
	if err != nil {
		return nil, "", err
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateFailed {
			h.showMessage("Failed to connect to peer", lsp.MTError)
		}
	})
	ordered := true
	_, err = peerConnection.CreateDataChannel("messaging-channel", &webrtc.DataChannelInit{Ordered: &ordered})
	if err != nil {
		return nil, "", err
	}
	offer, err := peerConnection.CreateOffer(&webrtc.OfferOptions{})
	if err != nil {
		return nil, "", err
	}
	err = peerConnection.SetLocalDescription(offer)
	if err != nil {
		return nil, "", err
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete

	h.setConn(clientID, peerConnection)
	h.runPeerConnection(peerConnection, func() {
		h.delConn(clientID)
	})
	return peerConnection.LocalDescription(), clientID, nil
}

func (h *LspHandler) runPeerConnection(peerConnection *webrtc.PeerConnection, closeCallback func()) {
	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnOpen(func() {
			raw, err := dc.Detach()
			if err != nil {
				h.logger.Println("Failed to detach data channel")
				peerConnection.Close()
				return
			}

			conn := jsonrpc2.NewConn(
				context.Background(),
				jsonrpc2.NewBufferedStream(raw, jsonrpc2.PlainObjectCodec{}),
				jsonrpc2.HandlerWithError(h.handlePeerRPC),
			)
			defer h.logger.Println("Closing peer connection")
			defer peerConnection.Close()
			conn.Notify(context.Background(), "initialize", server.InitializeClient{
				View:  h.state.GetView(),
				Files: h.state.GetFiles(),
			})
			var forward = server.GetForwardStateChangesCallback(h.logger, conn)
			h.state.Subscribe(forward)
			defer h.state.Unsubscribe(forward)
			<-conn.DisconnectNotify()
			h.logger.Println("Peer connection disconnected?")
		})
		dc.OnError(func(err error) {
			h.logger.Println("Error from DataChannel", err)
		})
		dc.OnClose(func() {
			h.logger.Println("DataChannel closed")
		})
	})
	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		h.logger.Printf("Peer Connection State has changed: %s\n", state.String())
		switch state {
		case webrtc.PeerConnectionStateConnected:
			h.showMessage("PairLS: Connected to peer", lsp.Info)
		case webrtc.PeerConnectionStateFailed:
			peerConnection.Close()
		case webrtc.PeerConnectionStateClosed:
			closeCallback()
			h.showMessage("PairLS: Peer connection closed", lsp.Info)
		}
	})
}
