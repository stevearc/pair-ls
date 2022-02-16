import { JsonRPC } from "./jsonrpc";
import { Dispatcher, View, FileMap, ChangeTextRange, showToast } from "./state";

export default abstract class BaseClient {
  protected dispatch: Dispatcher;
  protected promises: { [filename: string]: Promise<string[]> };
  private rpc: JsonRPC;

  constructor(rpc: JsonRPC, dispatch: Dispatcher) {
    this.rpc = rpc;
    for (const key of Object.getOwnPropertyNames(BaseClient.prototype)) {
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
    const p = this.rpc.request<string[]>("getText", { filename }).then(
      (text) => {
        delete this.promises[filename];
        this.dispatch({ type: "setText", filename, text });
        return text;
      },
      (e) => {
        delete this.promises[filename];
        showToast(this.dispatch, `Error fetching file text: ${e}`, {
          severity: "error",
        });
        return e;
      }
    );
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
  private onUpdateText({
    filename,
    changes,
  }: {
    filename: string;
    changes: ChangeTextRange[];
  }) {
    this.dispatch({
      type: "updateText",
      filename,
      changes,
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
