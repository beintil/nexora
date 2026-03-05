import type { components } from "./generated/swagger-types";
import { BACKEND_URL } from "../config";
import { getAuthToken, refreshAuthTokens, logoutSession } from "./authBridge";

export type CallsListResponse = components["schemas"]["CallsListResponse"];
export type CallTreeResponse = components["schemas"]["CallTreeResponse"];
export type CallsListItem = components["schemas"]["CallsListItem"];
export type PaginationMeta = components["schemas"]["PaginationMeta"];
/** Direction enum from schema */
export type CallDirection = NonNullable<components["schemas"]["CallsListItem"]["direction"]>;

/** Body for POST /v1/calls (from schema) */
export type CallsListRequest = components["schemas"]["CallsListRequest"];

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

export async function getCalls(body: Partial<CallsListRequest> = {}): Promise<CallsListResponse> {
    const payload: CallsListRequest = {
        page: {
            limit: body.page?.limit ?? 50,
            offset: body.page?.offset ?? 0,
            page: body.page?.page ?? 1,
        },
        date_from: body.date_from,
        date_to: body.date_to,
        direction: body.direction,
        status: body.status,
        company_telephony_id: body.company_telephony_id,
    };
    const res = await fetchWithAuthRetry("/v1/calls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
    });
    if (!res.ok) {
        const err = new Error("Failed to load calls");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<CallsListResponse>;
}

export async function getCallById(id: string): Promise<CallTreeResponse> {
    const res = await fetchWithAuthRetry(`/v1/calls/${encodeURIComponent(id)}`, {
        method: "GET",
    });
    if (!res.ok) {
        const err = new Error("Failed to load call");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<CallTreeResponse>;
}
