import { BACKEND_URL } from "../config";
import {
    getAuthToken,
    refreshAuthTokens,
    logoutSession,
} from "./authBridge";

export type ProfileResponse = {
    id: string;
    company_id: string;
    company_name: string;
    role_id: number;
    email?: string;
    full_name?: string;
    avatar_url?: string;
    avatar_id?: string;
    created_at: string;
    updated_at: string;
};

const roleLabels: Record<number, string> = {
    0: "Admin",
    1: "Support",
    2: "Owner",
    3: "Manager",
};

export function roleLabel(roleId: number): string {
    return roleLabels[roleId] ?? "Manager";
}

async function fetchWithAuthRetry(
    url: string,
    init: RequestInit & { headers?: Record<string, string> }
): Promise<Response> {
    let token = getAuthToken();
    if (!token) {
        const err = new Error("Unauthorized");
        (err as { code?: number }).code = 401;
        throw err;
    }
    const headers = { ...init.headers, Authorization: `Bearer ${token}` };
    let res = await fetch(url, { ...init, credentials: "include", headers });
    if (res.status === 401) {
        token = await refreshAuthTokens();
        if (!token) {
            logoutSession();
            const err = new Error("Unauthorized");
            (err as { code?: number }).code = 401;
            throw err;
        }
        res = await fetch(url, { ...init, credentials: "include", headers: { ...init.headers, Authorization: `Bearer ${token}` } });
    }
    return res;
}

export async function getProfile(): Promise<ProfileResponse> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/profile`, { method: "GET" });
    if (!res.ok) {
        const err = new Error(res.status === 401 ? "Unauthorized" : "Failed to load profile");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<ProfileResponse>;
}

export async function updateProfile(data: {
    full_name?: string;
}): Promise<ProfileResponse> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/profile`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    });
    if (!res.ok) {
        const err = new Error("Failed to update profile");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<ProfileResponse>;
}

export async function uploadAvatar(file: File): Promise<ProfileResponse> {
    const form = new FormData();
    form.append("avatar", file);
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/profile/avatar`, {
        method: "POST",
        body: form,
    });
    if (!res.ok) {
        const err = new Error("Failed to upload avatar");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<ProfileResponse>;
}
