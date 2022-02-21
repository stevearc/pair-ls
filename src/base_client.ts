import { JsonRPC } from "./jsonrpc";
import { Dispatcher, View, File, ChangeTextRange, showToast } from "./state";

export default abstract class BaseClient {
  protected dispatch: Dispatcher;
  protected promises: { [filename: string]: Promise<void> };
  private rpc: JsonRPC;
  private last_file_fetch: string | null;

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
    this.last_file_fetch = null;
  }

  getFileLoadPromise(filename: string): Promise<void> | undefined {
    return this.promises[filename];
  }

  getText(filename: string): Promise<void> {
    if (filename === this.last_file_fetch) {
      return Promise.resolve();
    }
    if (this.promises[filename] != null) {
      return this.promises[filename];
    }
    const p = this.rpc.request<File>("getText", { filename }).then(
      (file) => {
        this.last_file_fetch = filename;
        delete this.promises[filename];
        const lines = file.lines ?? [];
        this.dispatch({ type: "setText", file_id: file.id, text: lines });
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
  private onInitialize({ view, files }: { view: View; files: File[] }) {
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
    id,
    language,
  }: {
    filename: string;
    id: number;
    language: string;
  }) {
    this.dispatch({
      type: "openFile",
      filename,
      id,
      language,
    });
  }

  // @ts-ignore
  private onCloseFile({ file_id }: { file_id: number }) {
    this.dispatch({
      type: "closeFile",
      file_id,
    });
  }

  // @ts-ignore
  private onTextReplaced({
    file_id,
    text,
  }: {
    file_id: number;
    text: string[];
  }) {
    this.dispatch({
      type: "setText",
      file_id,
      text,
    });
  }

  // @ts-ignore
  private onUpdateText({
    file_id,
    changes,
  }: {
    file_id: number;
    changes: ChangeTextRange[];
  }) {
    this.dispatch({
      type: "updateText",
      file_id,
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
