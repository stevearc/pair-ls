import * as React from "react";
import { Dispatcher, showToast } from "./state";
import BaseClient from "./base_client";
import RTCRPC from "./rtc_rpc";
const { useState, useEffect } = React;

export type ClientState =
  | "new"
  | "no_session" // Could not find an editing session with that token
  | "connecting" // Going through the RTC connection process
  | "connected" // All good
  | "failed" // Failed to complete the RTC connection process
  | "closed"; // Connection is dead. Must refresh.
type CallOffer = {
  desc: RTCSessionDescriptionInit;
  client_id: string;
};
type StateChangeCallback = (status: ClientState) => void;

export default class RTCClient extends BaseClient {
  private rtc: RTCRPC;
  private state: ClientState;
  private statusCallbacks: StateChangeCallback[];
  private disconnectAlertID: number | null;

  constructor(dispatch: Dispatcher) {
    const rpc = new RTCRPC(undefined, { batching: false });
    super(rpc, dispatch);
    this.rtc = rpc;
    this.disconnectAlertID = null;
    this.state = "new";
    this.statusCallbacks = [];
    this.rtc.onConnectionStateChange((state) => {
      if (this.disconnectAlertID != null) {
        this.dispatch({ type: "removeToast", id: this.disconnectAlertID });
        this.disconnectAlertID = null;
      }
      switch (state) {
        case "new":
          this.setState("new");
          break;
        case "connecting":
          this.setState("connecting");
          break;
        case "connected":
          this.setState("connected");
          break;
        case "closed":
          this.setState("closed");
          break;
        case "disconnected":
          this.disconnectAlertID = showToast(
            this.dispatch,
            "Disconnected from editor",
            { severity: "error" },
            null
          );
          break;
        case "failed":
          this.setState("failed");
          break;
        default:
          let v: never = state;
          console.error("Unknown state", v);
          break;
      }
    });
  }

  async respondToCall(offerToken: string): Promise<string> {
    const offer: CallOffer = JSON.parse(window.atob(offerToken));
    const answer = await this.rtc.respondToCall(offer.desc);
    const answerToken = {
      desc: answer,
      client_id: offer.client_id,
    };
    return window.btoa(JSON.stringify(answerToken));
  }

  async startCall(): Promise<string> {
    this.setState("connecting");
    try {
      await this.rtc.createLocalOffer();
      await this.rtc.waitForIceComplete();
    } catch (e) {
      this.setState("failed");
      throw e;
    }
    const ld = this.rtc.localDescription;
    return ld == null ? "" : window.btoa(JSON.stringify({ desc: ld.toJSON() }));
  }

  setAnswer(answerToken: string): Promise<void> {
    const answer: RTCSessionDescriptionInit = JSON.parse(
      window.atob(answerToken)
    );
    return this.rtc.setAnswer(answer);
  }

  connectWithSignalToken(token: string): Promise<string> {
    if (token === "") {
      this.setState("no_session");
      return Promise.reject("Token is empty");
    }
    this.setState("connecting");
    return this.rtc.connect(token).then(null, (e) => {
      this.setState("no_session");
      return e;
    });
  }

  useStatus(): ClientState {
    const [status, setStatus] = useState(this.state);
    useEffect(() => {
      return this.onStatusChange(setStatus);
    }, [this, setStatus]);
    return status;
  }

  private onStatusChange(callback: StateChangeCallback): () => void {
    this.statusCallbacks.push(callback);
    return () => {
      for (let i = 0; i < this.statusCallbacks.length; i++) {
        if (this.statusCallbacks[i] === callback) {
          this.statusCallbacks.splice(i, 1);
          break;
        }
      }
    };
  }

  private setState(newState: ClientState) {
    if (this.state === newState) {
      return;
    }
    this.state = newState;
    for (const cb of this.statusCallbacks) {
      cb(newState);
    }
  }
}
