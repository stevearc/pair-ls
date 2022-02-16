import RPCHandler, { RPCOptions, JsonRPC } from "./jsonrpc";
import { post } from "./api";

export default class RTCRPC implements JsonRPC {
  private rpc: RPCHandler;
  private conn: RTCPeerConnection;
  private chan: RTCDataChannel;
  private token: string;
  private clientId: string;
  private tid: NodeJS.Timeout | null;
  private pendingCandidates: RTCIceCandidate[];

  constructor(init?: RTCConfiguration, opts?: RPCOptions) {
    this.token = "";
    this.clientId = "";
    this.pendingCandidates = [];
    if (init == null) {
      init = {
        iceServers: [
          {
            urls: [
              "stun:openrelay.metered.ca:80",
              "stun:stun.l.google.com:19302",
              "stun:stun1.l.google.com:19302",
              "stun:stun2.l.google.com:19302",
              "stun:stun3.l.google.com:19302",
              "stun:stun4.l.google.com:19302",
            ],
          },
          {
            urls: [
              "turn:openrelay.metered.ca:80",
              "turn:openrelay.metered.ca:443",
              "turn:openrelay.metered.ca:443?transport=tcp",
              "turns:openrelay.metered.ca:443",
            ],
            username: "openrelayproject",
            credential: "openrelayproject",
          },
        ],
      };
    }
    this.tid = null;
    this.rpc = new RPCHandler(opts);
    this.conn = new RTCPeerConnection(init);
    this.conn.addEventListener("icecandidate", async (e) => {
      if (e.candidate != null) {
        if (this.clientId === "") {
          this.pendingCandidates.push(e.candidate);
        } else {
          post("ice", {
            token: this.token,
            client_id: this.clientId,
            candidate: e.candidate.toJSON(),
          });
        }
      }
    });
    this.conn.addEventListener("iceconnectionstatechange", () => {
      if (this.conn.iceConnectionState === "failed") {
        this.conn.restartIce();
      }
    });

    const dataChannelParams = { ordered: true };
    this.chan = this.conn.createDataChannel(
      "messaging-channel",
      dataChannelParams
    );
    this.chan.binaryType = "arraybuffer";
    this.chan.addEventListener("open", () => {
      if (this.tid != null) {
        clearTimeout(this.tid);
        this.tid = null;
      }
    });
    this.chan.addEventListener("close", () => {
      if (this.tid != null) {
        clearTimeout(this.tid);
        this.tid = null;
      }
    });
    this.chan.addEventListener("error", (e) => {
      console.error("Data channel error", e);
    });
    this.chan.addEventListener("message", (event) => {
      this.rpc.receiveRawData(event.data);
      this.processData();
    });
  }

  async createLocalOffer(): Promise<void> {
    const localOffer = await this.conn.createOffer();
    await this.conn.setLocalDescription(localOffer);
  }

  async connect(token: string): Promise<void> {
    this.token = token;
    const localOffer = await this.conn.createOffer();
    await this.conn.setLocalDescription(localOffer);
    const response = await post<{
      answer: RTCSessionDescriptionInit;
      client_id: string;
    }>("call", {
      token,
      offer: localOffer,
    });
    this.clientId = response.client_id;
    for (const candidate of this.pendingCandidates) {
      await post("ice", {
        token: this.token,
        client_id: this.clientId,
        candidate: candidate.toJSON(),
      });
    }
    this.pendingCandidates.length = 0;
    const sd = new RTCSessionDescription(response.answer);
    await this.conn.setRemoteDescription(sd);
  }

  onConnectionStateChange(
    callback: (state: RTCPeerConnectionState) => void
  ): () => void {
    const cb = () => {
      callback(this.conn.connectionState);
    };
    this.conn.addEventListener("connectionstatechange", cb);
    return () => {
      this.conn.removeEventListener("connectionstatechange", cb);
    };
  }

  get localDescription(): RTCSessionDescription | null {
    return this.conn.localDescription;
  }

  get isConnected(): boolean {
    return (
      this.conn.connectionState === "connected" &&
      this.chan.readyState === "open"
    );
  }

  waitForIceComplete(): Promise<void> {
    return new Promise((resolve, reject) => {
      const cb = () => {
        if (this.conn.iceGatheringState === "complete") {
          resolve();
          this.conn.removeEventListener("icegatheringstatechange", cb);
        } else if (this.conn.iceGatheringState === "new") {
          reject("ICE gathering failed");
          this.conn.removeEventListener("icegatheringstatechange", cb);
        }
      };
      this.conn.addEventListener("icegatheringstatechange", cb);
    });
  }

  setAnswer(answer: RTCSessionDescriptionInit): Promise<void> {
    if (answer.type !== "answer") {
      return Promise.reject("Invalid RTC answer");
    }
    return this.conn.setRemoteDescription(answer);
  }

  async respondToCall(
    offer: RTCSessionDescriptionInit
  ): Promise<RTCSessionDescriptionInit> {
    await this.conn.setRemoteDescription(offer);
    const answer = await this.conn.createAnswer();
    await this.conn.setLocalDescription(answer);
    return answer;
  }

  request<T>(
    method: string,
    params: any = null,
    timeout: number = 30000
  ): Promise<T> {
    const promise = this.rpc.request<T>(method, params, timeout);
    this.processData();
    return promise;
  }

  notify(method: string, params: any = null) {
    this.rpc.notify(method, params);
    this.processData();
  }

  registerMethod(method: string, callback: (...args: any[]) => any) {
    this.rpc.registerMethod(method, callback);
  }

  private processData(after: number = 10) {
    if (this.tid != null) {
      return;
    }
    this.tid = setTimeout(() => {
      this.tid = null;
      this.rpc.flushIncoming();

      if (this.isConnected) {
        for (const message of this.rpc.flushOutgoing()) {
          this.chan.send(message);
        }
        if (this.rpc.hasPending) {
          this.processData(1);
        }
      } else {
        console.warn("WebRTC channel is closed. Backing off...");
        this.processData(4 * after);
      }
    }, after);
  }
}
