export type RequestOptions = {
  query: { [key: string]: string };
};

export async function get<T>(
  path: string,
  opts?: RequestOptions,
  init?: RequestInit
): Promise<T> {
  const query = opts?.query;
  if (query != null) {
    const pieces = [];
    for (const key in query) {
      pieces.push(`${key}=${encodeURIComponent(query[key].toString())}`);
    }
    path = "?" + pieces.join(",");
  }
  const response = await fetch(path, init);
  return await response.json();
}

export async function post<T>(
  path: string,
  data: any,
  init?: RequestInit
): Promise<T> {
  if (init == null) {
    init = {};
  }
  init.method = "POST";
  init.body = JSON.stringify(data);
  if (init.headers == null) {
    init.headers = {};
  }
  // @ts-ignore
  init.headers.Accept = "application/json";
  // @ts-ignore
  init.headers["Content-Type"] = "application/json";
  const response = await fetch(path, init);
  if (response.ok) {
    return await response.json();
  } else {
    throw await response.text();
  }
}
