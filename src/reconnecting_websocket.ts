const BACKOFF_START = 1000;

export default class ReconnectingWebSocket implements WebSocket {
  CLOSED = WebSocket.CLOSED;
  CLOSING = WebSocket.CLOSING;
  CONNECTING = WebSocket.CONNECTING;
  OPEN = WebSocket.OPEN;
  private _url: string | URL;
  private _protocols: undefined | string | string[];
  private ws: WebSocket;
  private closed: boolean;
  private listeners: { [type: string]: Array<any> };
  private backoff: number = BACKOFF_START;
  private nextConnect: number | null;

  constructor(url: string | URL, protocols?: string | string[]) {
    this._url = url;
    this._protocols = protocols;
    this.listeners = {};
    this.closed = false;
    this.nextConnect = null;
    this.ws = this.connect();
  }

  get isClosed(): boolean {
    return this.closed;
  }

  private connect(): WebSocket {
    const ws = new WebSocket(this._url, this._protocols);
    ws.addEventListener("open", () => {
      this.backoff = BACKOFF_START;
      this.nextConnect = null;
    });
    ws.addEventListener("close", (evt) => {
      if (!this.closed) {
        this.nextConnect = Date.now() + this.backoff;
        setTimeout(() => this.connect(), this.backoff);
        this.backoff *= 2;
      } else if (this.listeners.finalClose != null) {
        for (const listener of this.listeners.finalClose) {
          listener(evt);
        }
      }
    });
    for (const key in this.listeners) {
      if (key !== "finalClose") {
        for (const listener of this.listeners[key]) {
          ws.addEventListener(key, listener);
        }
      }
    }
    this.ws = ws;
    return ws;
  }

  getNextConnect(): number {
    if (this.readyState != WebSocket.CONNECTING || this.nextConnect == null) {
      return 0;
    }
    return this.nextConnect;
  }

  close(code?: number, reason?: string): void {
    this.closed = true;
    this.ws.close(code, reason);
  }
  send(data: string | ArrayBufferLike | Blob | ArrayBufferView): void {
    this.ws.send(data);
  }

  addEventListener(type: any, listener: any, _options?: any): void {
    let listeners = this.listeners[type];
    if (listeners == null) {
      listeners = [];
      this.listeners[type] = listeners;
    }
    listeners.push(listener);
    this.ws.addEventListener(type, listener);
  }
  removeEventListener(type: any, listener: any, _options?: any): void {
    const listeners = this.listeners[type];
    if (listeners != null) {
      for (let i = 0; i < listeners.length; i++) {
        const l = listeners[i];
        if (l === listener) {
          listeners.splice(i, 1);
          break;
        }
      }
    }
    this.ws.removeEventListener(type, listener);
  }
  dispatchEvent(event: Event): boolean {
    return this.ws.dispatchEvent(event);
  }

  get binaryType(): BinaryType {
    return this.ws.binaryType;
  }

  get bufferedAmount(): number {
    return this.ws.bufferedAmount;
  }

  get extensions(): string {
    return this.ws.extensions;
  }

  get protocol(): string {
    return this.ws.protocol;
  }

  get readyState(): number {
    const state = this.ws.readyState;
    if (
      !closed &&
      (state === WebSocket.CLOSING || state === WebSocket.CLOSED)
    ) {
      return WebSocket.CONNECTING;
    }
    return state;
  }

  get url(): string {
    return this.ws.url;
  }

  set onclose(cb: (this: WebSocket, ev: CloseEvent) => any) {
    this.addEventListener("close", cb);
  }
  set onerror(cb: (this: WebSocket, ev: Event) => any) {
    this.addEventListener("error", cb);
  }
  set onmessage(cb: (this: WebSocket, ev: MessageEvent<any>) => any) {
    this.addEventListener("message", cb);
  }
  set onopen(cb: (this: WebSocket, ev: Event) => any) {
    this.addEventListener("open", cb);
  }
}
