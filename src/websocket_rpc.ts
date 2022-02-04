import RPCHandler, { RPCOptions } from "./jsonrpc";
import ReconnectingWebSocket from "./reconnecting_websocket";

export default class WebSocketRPC {
  ws: ReconnectingWebSocket;
  rpc: RPCHandler;
  private tid: NodeJS.Timeout | null;

  constructor(url: string | URL, opts?: RPCOptions) {
    this.tid = null;
    this.rpc = new RPCHandler(opts);
    this.ws = new ReconnectingWebSocket(url);
    this.ws.addEventListener("message", (evt: MessageEvent) => {
      evt.data.text().then((text: string) => {
        this.rpc.receiveRawMessage(text);
        this.processData();
      });
    });
  }

  getNextConnect(): number {
    return this.ws.getNextConnect();
  }

  close(): void {
    this.ws.close();
  }

  addEventListener(type: any, listener: any, options?: any): void {
    this.ws.addEventListener(type, listener, options);
  }

  removeEventListener(type: any, listener: any, options?: any): void {
    this.ws.removeEventListener(type, listener, options);
  }

  get isConnected(): boolean {
    return this.ws.readyState === WebSocket.OPEN;
  }

  get isClosed(): boolean {
    return this.ws.isClosed;
  }

  waitForConnect(): Promise<void> {
    if (this.isConnected) {
      return Promise.resolve();
    } else {
      return new Promise((resolve) => {
        this.ws.addEventListener("open", resolve);
      });
    }
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

      if (this.ws.readyState === WebSocket.OPEN) {
        for (const message of this.rpc.flushOutgoing()) {
          this.ws.send(message);
        }
        if (this.rpc.hasPending) {
          this.processData(1);
        }
      } else {
        console.warn("Websocket is closed. Backing off...");
        this.processData(4 * after);
      }
    }, after);
  }
}
