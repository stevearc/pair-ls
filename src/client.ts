import { Dispatcher, View, FileMap, showToast } from "./state";
import WebSocketRPC from "./websocket_rpc";

export default class Client {
  private rpc: WebSocketRPC;
  private dispatch: Dispatcher;
  private promises: { [filename: string]: Promise<string[]> };
  private reconnectAlertID: number | null;

  constructor(url: string, dispatch: Dispatcher, token: string) {
    this.rpc = new WebSocketRPC(url, { batching: false });
    this.reconnectAlertID = null;
    this.rpc.addEventListener("open", () => {
      const showSuccess = this.reconnectAlertID != null;
      if (this.reconnectAlertID != null) {
        dispatch({ type: "removeToast", id: this.reconnectAlertID });
        this.reconnectAlertID = null;
      }
      this.rpc.request("auth", { token }).then(
        () => {
          if (showSuccess) {
            showToast(
              dispatch,
              "Connection restored",
              { severity: "success" },
              4000
            );
          }
        },
        () => {
          showToast(
            dispatch,
            "Authentication error. Please refresh",
            {
              severity: "error",
            },
            null
          );
          this.rpc.close();
        }
      );
    });
    this.rpc.addEventListener("close", () => {
      if (this.reconnectAlertID == null && !this.rpc.isClosed) {
        this.reconnectAlertID = showToast(
          dispatch,
          () => {
            const secs = Math.floor(
              (this.rpc.getNextConnect() - Date.now()) / 1000
            );
            return secs <= 0
              ? "Reconnecting..."
              : `Lost connection. Reconnecting in ${secs}s`;
          },
          { severity: "error" },
          null
        );
      }
    });
    for (const key of Object.getOwnPropertyNames(Client.prototype)) {
      if (/^on/.test(key)) {
        const method = key[2].toLowerCase() + key.slice(3);
        const callback = (this as any)[key];
        this.rpc.registerMethod(method, callback.bind(this));
      }
    }
    this.dispatch = dispatch;
    this.promises = {};
  }

  getText(filename: string): Promise<string[]> {
    if (this.promises[filename] != null) {
      return this.promises[filename];
    }
    const p = this.rpc
      .request<string[]>("getText", { filename })
      .then((text) => {
        delete this.promises[filename];
        this.dispatch({ type: "setText", filename, text });
        return text;
      });
    this.promises[filename] = p;
    return p;
  }

  // @ts-ignore
  private onInitialize({ view, files }: { view: View; files: FileMap }) {
    this.dispatch({
      type: "initialize",
      sync: {
        view,
        files,
      },
    });
  }

  // @ts-ignore
  private onOpenFile({
    filename,
    language,
  }: {
    filename: string;
    language: string;
  }) {
    this.dispatch({
      type: "openFile",
      filename,
      language,
    });
  }

  // @ts-ignore
  private onCloseFile({ filename }: { filename: string }) {
    this.dispatch({
      type: "closeFile",
      filename,
    });
  }

  // @ts-ignore
  private onTextReplaced({
    filename,
    text,
  }: {
    filename: string;
    text: string[];
  }) {
    this.dispatch({
      type: "setText",
      filename,
      text,
    });
  }

  // @ts-ignore
  private onUpdateView({ view }: { view: View }) {
    this.dispatch({
      type: "updateView",
      view,
    });
  }
}
