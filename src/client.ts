import { Dispatcher, showToast } from "./state";
import WebSocketRPC from "./websocket_rpc";
import BaseClient from "./base_client";

export default class Client extends BaseClient {
  private reconnectAlertID: number | null;

  constructor(url: string, dispatch: Dispatcher, token: string) {
    const rpc = new WebSocketRPC(url, { batching: false });
    super(rpc, dispatch);
    this.reconnectAlertID = null;
    rpc.addEventListener("open", () => {
      const showSuccess = this.reconnectAlertID != null;
      if (this.reconnectAlertID != null) {
        dispatch({ type: "removeToast", id: this.reconnectAlertID });
        this.reconnectAlertID = null;
      }
      rpc.request("auth", { token }).then(
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
          rpc.close();
        }
      );
    });
    rpc.addEventListener("close", () => {
      if (this.reconnectAlertID == null && !rpc.isClosed) {
        this.reconnectAlertID = showToast(
          dispatch,
          () => {
            const secs = Math.floor((rpc.getNextConnect() - Date.now()) / 1000);
            return secs <= 0
              ? "Reconnecting..."
              : `Lost connection. Reconnecting in ${secs}s`;
          },
          { severity: "error" },
          null
        );
      }
    });
  }
}
