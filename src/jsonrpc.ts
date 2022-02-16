type RPCError = {
  code: number;
  message: string;
  data?: any;
};
type RPCResponse =
  | {
      jsonrpc: "2.0";
      id: number | string | null;
      result: any;
    }
  | {
      jsonrpc: "2.0";
      id: number | string | null;
      error: RPCError;
    };
type RPCRequest = {
  jsonrpc: "2.0";
  method: string;
  params?: Object | Array<any>;
  id?: number | string | null;
};

type PromiseHandle<T> = {
  tid: NodeJS.Timeout;
  reject: (err: RPCError) => void;
  resolve: (value: T) => void;
};

const ERR_GENERAL = 1;
const ERR_TIMEOUT = 2;
export type RPCOptions = {
  batching?: boolean;
};

export interface JsonRPC {
  request<T>(method: string, params: any, timeout?: number): Promise<T>;
  notify(method: string, params?: any): void;
  registerMethod(method: string, callback: (...args: any[]) => any): void;
}

export default class RPCHandler implements JsonRPC {
  private nextID: number;
  private outgoingRequests: RPCRequest[];
  private incomingResponses: RPCResponse[];
  private incomingRequests: RPCRequest[];
  private outgoingResponses: RPCResponse[];
  private pendingRequests: { [id: string]: PromiseHandle<any> };
  private methods: { [method: string]: (...args: any[]) => any };
  private batching: boolean;
  private pendingData: string;

  constructor(opts?: RPCOptions) {
    this.nextID = 1;
    this.outgoingRequests = [];
    this.incomingResponses = [];
    this.incomingRequests = [];
    this.outgoingResponses = [];
    this.pendingRequests = {};
    this.methods = {};
    this.batching = opts?.batching ?? true;
    this.pendingData = "";
  }

  get hasPending(): boolean {
    return (
      this.outgoingRequests.length > 0 ||
      this.outgoingResponses.length > 0 ||
      this.incomingRequests.length > 0 ||
      this.incomingResponses.length > 0
    );
  }

  /**
   * Register a server RPC method
   */
  registerMethod(method: string, callback: (...args: any[]) => any) {
    this.methods[method] = callback;
  }

  /**
   * Send a request FROM this client TO the server
   */
  request<T>(
    method: string,
    params: any = null,
    timeout: number = 30000
  ): Promise<T> {
    const id = this.nextID++;
    const tid = setTimeout(() => {
      const handle = this.pendingRequests[id];
      if (handle != null) {
        handle.reject({ code: ERR_TIMEOUT, message: "Timeout" });
      }
      delete this.pendingRequests[id];
    }, timeout);
    const promise: Promise<T> = new Promise((resolve, reject) => {
      this.pendingRequests[id] = { resolve, reject, tid };
    });
    const request: RPCRequest = {
      jsonrpc: "2.0",
      method,
      id,
      params,
    };
    this.outgoingRequests.push(request);
    return promise;
  }

  /**
   * Send a notification FROM this client TO the server
   */
  notify(method: string, params: any = null) {
    const notification: RPCRequest = {
      jsonrpc: "2.0",
      method,
      params,
    };
    this.outgoingRequests.push(notification);
  }

  /**
   * Receive a response TO this client, FROM the server
   */
  receiveResponse(response: RPCResponse) {
    this.incomingResponses.push(response);
  }

  /**
   * Receive a raw string response TO this client, FROM the server
   */
  receiveRawResponse(strResponse: string) {
    this.receiveParsedResponse(JSON.parse(strResponse));
  }

  private receiveParsedResponse(response: any) {
    if (typeof response !== "object") {
      throw new Error(`Response must be an object: ${response}`);
    }
    if (response.jsonrpc !== "2.0") {
      throw new Error(`Response must be Json RPC 2.0: ${response}`);
    }
    if (!response.hasOwnProperty("id")) {
      throw new Error(`Response must have an id field: ${response}`);
    }
    if (
      !response.hasOwnProperty("result") &&
      !response.hasOwnProperty("error")
    ) {
      throw new Error(
        `Response must have a result or error field: ${response}`
      );
    }
    this.receiveResponse(response);
  }

  /**
   * Receive a request TO this server, FROM a client
   */
  receiveRequest(request: RPCRequest) {
    this.incomingRequests.push(request);
  }

  private respondError(code: number, message: string, data: any = null) {
    const err: RPCError = { code, message };
    if (data != null) {
      err.data = data;
    }
    this.outgoingResponses.push({
      jsonrpc: "2.0",
      id: null,
      error: err,
    });
  }

  /**
   * Receive a raw string request TO this server, FROM a client
   */
  receiveRawRequest(requestStr: string) {
    let request;
    try {
      request = JSON.parse(requestStr);
    } catch (error) {
      return this.respondError(-32700, "Parse error");
    }
    this.receiveParsedRequest(request);
  }

  receiveRawData(message: ArrayBuffer) {
    const str = this.pendingData + new TextDecoder().decode(message);
    this.pendingData = "";
    // TODO this is a sloppy way to do stream JSON decoding. It relies on the assumption that we will at some point receive a message that completes all previous messages and starts no new ones. For our particular application (low-ish traffic) this seems to be good enough for now.
    try {
      this.receiveRawMessage(str);
    } catch {
      this.pendingData = str;
      console.warn("Couldn't process data", str);
    }
  }

  /**
   * Receive a raw string that may be either a request or a response
   */
  receiveRawMessage(message: string) {
    let data;
    try {
      data = JSON.parse(message);
    } catch (err) {
      console.error("Error parsing message:", message);
      throw err;
    }
    if (typeof data !== "object") {
      throw new Error(`Message is not a JSON object: ${message}`);
    }
    if (Array.isArray(data)) {
      for (const m of data) {
        this.receiveParsedMessage(m);
      }
    } else {
      this.receiveParsedMessage(data);
    }
  }

  private receiveParsedMessage(data: any) {
    if (hasOwnProperty(data, "method")) {
      this.receiveParsedRequest(data);
    } else {
      this.receiveParsedResponse(data);
    }
  }

  private receiveParsedRequest(request: any) {
    if (typeof request !== "object") {
      this.respondError(
        -32600,
        "Invalid Request",
        "Request is not of type 'object'"
      );
    } else if (request.jsonrpc !== "2.0") {
      this.respondError(
        -32600,
        "Invalid Request",
        "Request jsonrpc is not 2.0"
      );
    } else if (request.hasOwnProperty("id") && request.id === null) {
      this.respondError(
        -32600,
        "Invalid Request",
        "Request must have a non-null id field"
      );
    } else if (
      !request.hasOwnProperty("method") ||
      typeof request.method !== "string"
    ) {
      this.respondError(
        -32600,
        "Invalid Request",
        "Request must have a method field"
      );
    } else if (
      request.hasOwnProperty("params") &&
      typeof request.params !== "object"
    ) {
      this.respondError(
        -32600,
        "Invalid Request",
        "Request params must be an object or array"
      );
    } else {
      this.receiveRequest(request);
    }
  }

  /**
   * Process incoming requests and responses
   */
  flushIncoming() {
    // Responses
    const responses = this.incomingResponses;
    this.incomingResponses = [];
    for (const response of responses) {
      if (response.id == null) {
        console.error(
          "Response object missing id. Request must have been invalid",
          response
        );
        continue;
      }
      const handle = this.pendingRequests[response.id];
      clearTimeout(handle.tid);
      if (hasOwnProperty(response, "result")) {
        handle.resolve(response.result);
      } else {
        handle.reject(response.error);
      }
      delete this.pendingRequests[response.id];
    }
    // Requests
    const requests = this.incomingRequests;
    this.incomingRequests = [];
    for (const request of requests) {
      const meth = this.methods[request.method];
      if (meth == null) {
        console.error(`Method ${request.method} not found`);
        this.respondError(
          -32601,
          "Method not found",
          `Method ${request.method} is not found`
        );
        continue;
      }
      try {
        let result;
        if (hasOwnProperty(request, "params")) {
          if (Array.isArray(request.params)) {
            result = meth.apply(request.params);
          } else {
            result = meth(request.params);
          }
        } else {
          result = meth();
        }
        if (request.id != null) {
          this.outgoingResponses.push({
            jsonrpc: "2.0",
            id: request.id,
            result,
          });
        }
      } catch (error) {
        console.warn(`Request ${request.method} had error:`, error);
        if (request.id != null) {
          this.outgoingResponses.push({
            jsonrpc: "2.0",
            id: request.id,
            error: makeRPCError(error),
          });
        }
      }
    }
  }

  private flushOutgoingRequests(): RPCRequest[] {
    const requests = this.outgoingRequests;
    this.outgoingRequests = [];
    return requests;
  }

  private flushOutgoingResponses(): RPCResponse[] {
    const responses = this.outgoingResponses;
    this.outgoingResponses = [];
    return responses;
  }

  /**
   * Get all string messages that should be sent
   */
  flushOutgoing(): string[] {
    const ret = [];
    const responses = this.flushOutgoingResponses();
    if (responses.length) {
      if (this.batching) {
        ret.push(JSON.stringify(responses));
      } else {
        for (const response of responses) {
          ret.push(JSON.stringify(response));
        }
      }
    }
    const requests = this.flushOutgoingRequests();
    if (requests.length) {
      if (this.batching) {
        ret.push(JSON.stringify(requests));
      } else {
        for (const request of requests) {
          ret.push(JSON.stringify(request));
        }
      }
    }
    return ret;
  }
}

function makeRPCError(err: any): RPCError {
  if (typeof err !== "object" || typeof err.code !== "number") {
    return {
      code: ERR_GENERAL,
      message: err.toString(),
    };
  } else {
    return {
      code: err.code,
      message: err.toString(),
    };
  }
}

function hasOwnProperty<X extends {}, Y extends PropertyKey>(
  obj: X,
  prop: Y
): obj is X & Record<Y, unknown> {
  return obj.hasOwnProperty(prop);
}
