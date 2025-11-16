export type User = {
  id: number;
  email: string;
};

export type Plushie = {
  id: number;
  name: string;
  kind: string;
  adopted_at?: string;
  image_url?: string;
  conversation_history?: string;
  created_at?: string;
};

const API_BASE = "http://localhost:8080/api";
let API_ORIGIN = "";
try {
  const url = new URL(API_BASE);
  API_ORIGIN = url.origin;
} catch {
  API_ORIGIN = "http://localhost:8080";
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let message = `Error ${res.status}`;
    try {
      const data = (await res.json()) as { error?: string };
      if (data.error) message = data.error;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
  if (res.status === 204) {
    // no content
    return undefined as unknown as T;
  }
  return (await res.json()) as T;
}

export async function apiRegister(email: string, password: string): Promise<User> {
  const res = await fetch(`${API_BASE}/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password })
  });
  return handleResponse<User>(res);
}

export async function apiLogin(email: string, password: string): Promise<User> {
  const res = await fetch(`${API_BASE}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, password })
  });
  return handleResponse<User>(res);
}

export async function apiLogout(): Promise<void> {
  const res = await fetch(`${API_BASE}/logout`, {
    method: "POST",
    credentials: "include"
  });
  await handleResponse<unknown>(res);
}

export async function apiMe(): Promise<User> {
  const res = await fetch(`${API_BASE}/me`, {
    method: "GET",
    credentials: "include"
  });
  return handleResponse<User>(res);
}

export async function apiListPlushies(): Promise<Plushie[]> {
  const res = await fetch(`${API_BASE}/plushies`, {
    method: "GET",
    credentials: "include"
  });
  const list = await handleResponse<Plushie[]>(res);
  return list.map((p) =>
    p.image_url
      ? {
          ...p,
          image_url: `${API_ORIGIN}${p.image_url}`
        }
      : p
  );
}

export async function apiCreatePlushie(params: {
  name: string;
  kind: string;
  adoptedAt?: string;
  imageFile?: File | null;
}): Promise<{ id: number }> {
  const form = new FormData();
  form.set("name", params.name);
  form.set("kind", params.kind);
  if (params.adoptedAt) form.set("adopted_at", params.adoptedAt);
  if (params.imageFile) form.set("image", params.imageFile);

  const res = await fetch(`${API_BASE}/plushies`, {
    method: "POST",
    body: form,
    credentials: "include"
  });
  return handleResponse<{ id: number }>(res);
}

export async function apiUpdatePlushie(
  id: number,
  params: {
    name: string;
    kind: string;
    adoptedAt?: string;
    imageFile?: File | null;
  }
): Promise<void> {
  const form = new FormData();
  form.set("name", params.name);
  form.set("kind", params.kind);
  if (params.adoptedAt) form.set("adopted_at", params.adoptedAt);
  if (params.imageFile) form.set("image", params.imageFile);

  const res = await fetch(`${API_BASE}/plushies/${id}`, {
    method: "PUT",
    body: form,
    credentials: "include"
  });
  await handleResponse<unknown>(res);
}

export async function apiGetPlushie(id: number): Promise<Plushie> {
  const res = await fetch(`${API_BASE}/plushies/${id}`, {
    method: "GET",
    credentials: "include"
  });
  const plushie = await handleResponse<Plushie>(res);
  if (plushie.image_url) {
    plushie.image_url = `${API_ORIGIN}${plushie.image_url}`;
  }
  return plushie;
}

export async function apiUpdateConversation(id: number, conversationHistory: string): Promise<void> {
  const res = await fetch(`${API_BASE}/plushies/${id}/conversation`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ conversation_history: conversationHistory })
  });
  await handleResponse<unknown>(res);
}

export async function apiChat(id: number): Promise<{ message: string }> {
  const res = await fetch(`${API_BASE}/plushies/${id}/chat`, {
    method: "POST",
    credentials: "include"
  });
  return handleResponse<{ message: string }>(res);
}

export async function apiDeletePlushie(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/plushies/${id}`, {
    method: "DELETE",
    credentials: "include"
  });
  await handleResponse<unknown>(res);
}


