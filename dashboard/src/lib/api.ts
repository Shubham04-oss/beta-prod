import { auth } from "./firebase";
import ky, { type BeforeRequestHook, type Options } from "ky";

const injectAuthToken: BeforeRequestHook = async ({ request }) => {
  let token: string | null = null;
  if (auth.currentUser) {
    token = await auth.currentUser.getIdToken();
  } else if (typeof window !== "undefined") {
    token = window.sessionStorage.getItem("synqAuthToken");
  }

  if (token) {
    request.headers.set("Authorization", `Bearer ${token}`);
  }
};

export const apiClient = ky.create({
  prefix: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080",
  throwHttpErrors: false,
  hooks: {
    beforeRequest: [injectAuthToken],
  },
});

export async function fetchAPI<T = unknown>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const kyOptions: Options = {
    method: options.method || "GET",
    headers: options.headers as HeadersInit | undefined,
  };

  if (options.body) {
    if (typeof options.body === "string") {
      try {
        kyOptions.json = JSON.parse(options.body);
      } catch {
        kyOptions.body = options.body;
      }
    } else {
      kyOptions.body = options.body;
    }
  }

  const url = endpoint.startsWith("/") ? endpoint.slice(1) : endpoint;
  const response = await apiClient(url, kyOptions);
  if (!response.ok) {
    throw new Error(await readErrorMessage(response));
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return response.json<T>();
}

async function readErrorMessage(response: Response) {
  const fallback = `API Error: ${response.status} ${response.statusText}`;
  const text = await response.text();
  if (!text) return fallback;
  try {
    const parsed = JSON.parse(text) as { error?: string; message?: string };
    return parsed.error || parsed.message || fallback;
  } catch {
    return text;
  }
}
