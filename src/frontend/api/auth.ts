import { api } from "./client";
import type { MortarUser } from "../types/plugin";

export interface AuthSessionResponse {
  user: MortarUser;
}

export function fetchSession(
  signal?: AbortSignal,
): Promise<AuthSessionResponse> {
  return api.get<AuthSessionResponse>("/api/v1/auth/session", { signal });
}

export function login(
  username: string,
  password: string,
  signal?: AbortSignal,
): Promise<AuthSessionResponse> {
  return api.post<AuthSessionResponse>(
    "/api/v1/auth/login",
    { username, password },
    { signal },
  );
}

export function logout(signal?: AbortSignal): Promise<void> {
  return api.post<void>("/api/v1/auth/logout", undefined, { signal });
}
