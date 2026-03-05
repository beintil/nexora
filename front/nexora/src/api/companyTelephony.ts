import type { components } from "./generated/swagger-types";
import { BACKEND_URL } from "../config";
import { getAuthToken, refreshAuthTokens, logoutSession } from "./authBridge";

export type CompanyTelephonyItem = components["schemas"]["CompanyTelephonyItem"];
export type CompanyTelephonyListResponse = components["schemas"]["CompanyTelephonyListResponse"];
export type CompanyTelephonyCreateRequest = components["schemas"]["CompanyTelephonyCreateRequest"];
export type TelephonyDictionaryItem = components["schemas"]["TelephonyDictionaryItem"];
export type TelephonyDictionaryResponse = components["schemas"]["TelephonyDictionaryResponse"];

async function fetchWithAuthRetry(
    path: string,
    init: RequestInit & { headers?: Record<string, string> } = {}
): Promise<Response> {
    let token = getAuthToken();
    if (!token) {
        const err = new Error("Unauthorized");
        (err as { code?: number }).code = 401;
        throw err;
    }
    const headers = { ...init.headers, Authorization: `Bearer ${token}` };
    let res = await fetch(`${BACKEND_URL}${path}`, { ...init, credentials: "include", headers });
    if (res.status === 401) {
        token = await refreshAuthTokens();
        if (!token) {
            logoutSession();
            const err = new Error("Unauthorized");
            (err as { code?: number }).code = 401;
            throw err;
        }
        res = await fetch(`${BACKEND_URL}${path}`, {
            ...init,
            credentials: "include",
            headers: { ...init.headers, Authorization: `Bearer ${token}` },
        });
    }
    return res;
}

export async function getCompanyTelephony(): Promise<CompanyTelephonyItem[]> {
    const res = await fetchWithAuthRetry("/v1/company/telephony", { method: "GET" });
    if (!res.ok) {
        const err = new Error("Failed to load telephony list");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    const data = (await res.json()) as CompanyTelephonyListResponse;
    return data.items ?? [];
}

export async function createCompanyTelephony(payload: CompanyTelephonyCreateRequest): Promise<CompanyTelephonyItem> {
    const res = await fetchWithAuthRetry("/v1/company/telephony", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
    });
    if (!res.ok) {
        const err = new Error("Failed to connect telephony");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<CompanyTelephonyItem>;
}

export async function deleteCompanyTelephony(id: string): Promise<void> {
    const res = await fetchWithAuthRetry(`/v1/company/telephony/${encodeURIComponent(id)}`, {
        method: "DELETE",
    });
    if (!res.ok) {
        const err = new Error("Failed to disconnect telephony");
        (err as { code?: number }).code = res.status;
        throw err;
    }
}

export async function getTelephonyDictionary(): Promise<TelephonyDictionaryItem[]> {
    const res = await fetchWithAuthRetry("/v1/company/telephony/dictionary", { method: "GET" });
    if (!res.ok) {
        const err = new Error("Failed to load telephony dictionary");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    const data = (await res.json()) as TelephonyDictionaryResponse;
    return data.items ?? [];
}
