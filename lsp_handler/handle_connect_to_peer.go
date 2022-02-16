package lsp_handler

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"github.com/pion/webrtc/v3"
	"github.com/sourcegraph/jsonrpc2"
)

type PeerConnectRequestParams struct {
	Token string `json:"token,omitempty"`
}

type PeerToken struct {
	Desc     *webrtc.SessionDescription `json:"desc"`
	ClientID string                     `json:"client_id,omitempty"`
}

func (h *LspHandler) handleConnectToPeer(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params PeerConnectRequestParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if params.Token == "" {
		// If no token was provided, initiate a WebRTC call
		desc, clientID, err := h.callRTCPeer()

		json, err := json.Marshal(PeerToken{
			Desc:     desc,
			ClientID: clientID,
		})
		if err != nil {
			return nil, err
		}

		token := base64.StdEncoding.EncodeToString(json)
		return struct {
			URL string `json:"url"`
		}{
			URL: h.createStaticUrl(token),
		}, nil
	} else {
		token, err := h.respondRTCPeer(params.Token)
		if err != nil {
			return nil, err
		}
		if token == "" {
			return nil, nil
		} else {
			return struct {
				Token string `json:"token"`
			}{
				Token: token,
			}, nil
		}
	}
}
