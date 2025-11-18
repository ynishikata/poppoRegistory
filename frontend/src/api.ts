import { supabase } from "./supabase";

export type User = {
  id: string; // Changed to string (UUID) for Supabase
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

// Get Supabase JWT token for API requests
async function getAuthToken(): Promise<string | null> {
  const { data: { session }, error } = await supabase.auth.getSession();
  if (error) {
    console.error("Failed to get session:", error);
    return null;
  }
  if (!session) {
    console.warn("No active session. Please log in.");
    return null;
  }
  return session.access_token || null;
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let message = `エラーが発生しました (${res.status})`;
    try {
      const data = (await res.json()) as { error?: string };
      if (data.error) {
        message = data.error;
        // Translate common error messages to Japanese
        if (message.includes("unauthorized")) {
          message = "認証が必要です。ログインしてください。";
        } else if (message.includes("not found")) {
          message = "見つかりませんでした。";
        } else if (message.includes("failed to")) {
          message = "処理に失敗しました。";
        }
      }
    } catch {
      // ignore JSON parse errors
    }
    throw new Error(message);
  }
  if (res.status === 204) {
    // no content
    return undefined as unknown as T;
  }
  return (await res.json()) as T;
}

// Supabase Auth functions
export async function apiRegister(email: string, password: string): Promise<User> {
  const { data, error } = await supabase.auth.signUp({
    email,
    password,
  });
  if (error) throw new Error(error.message);
  if (!data.user) throw new Error("Failed to create user");
  return {
    id: data.user.id,
    email: data.user.email || "",
  };
}

export async function apiLogin(email: string, password: string): Promise<User> {
  const { data, error } = await supabase.auth.signInWithPassword({
    email,
    password,
  });
  if (error) throw new Error(error.message);
  if (!data.user) throw new Error("Failed to login");
  return {
    id: data.user.id,
    email: data.user.email || "",
  };
}

export async function apiLogout(): Promise<void> {
  const { error } = await supabase.auth.signOut();
  if (error) throw new Error(error.message);
}

export async function apiMe(): Promise<User> {
  const { data: { user }, error } = await supabase.auth.getUser();
  if (error) throw new Error(error.message);
  if (!user) throw new Error("Not authenticated");
  return {
    id: user.id,
    email: user.email || "",
  };
}

export async function apiListPlushies(): Promise<Plushie[]> {
  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies`, {
    method: "GET",
    headers,
  });
  const list = await handleResponse<Plushie[]>(res);
  if (!list || !Array.isArray(list)) {
    return [];
  }
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

  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies`, {
    method: "POST",
    headers,
    body: form,
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

  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies/${id}`, {
    method: "PUT",
    headers,
    body: form,
  });
  await handleResponse<unknown>(res);
}

export async function apiGetPlushie(id: number): Promise<Plushie> {
  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies/${id}`, {
    method: "GET",
    headers,
  });
  const plushie = await handleResponse<Plushie>(res);
  if (plushie.image_url) {
    plushie.image_url = `${API_ORIGIN}${plushie.image_url}`;
  }
  return plushie;
}

export async function apiUpdateConversation(id: number, conversationHistory: string): Promise<void> {
  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies/${id}/conversation`, {
    method: "PUT",
    headers,
    body: JSON.stringify({ conversation_history: conversationHistory })
  });
  await handleResponse<unknown>(res);
}

export async function apiChat(id: number): Promise<{ message: string }> {
  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies/${id}/chat`, {
    method: "POST",
    headers,
  });
  return handleResponse<{ message: string }>(res);
}

export async function apiDeletePlushie(id: number): Promise<void> {
  const token = await getAuthToken();
  if (!token) {
    throw new Error("Not authenticated. Please log in.");
  }
  const headers: HeadersInit = {
    "Authorization": `Bearer ${token}`,
  };
  const res = await fetch(`${API_BASE}/plushies/${id}`, {
    method: "DELETE",
    headers,
  });
  await handleResponse<unknown>(res);
}


