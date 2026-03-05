import type { components } from "./generated/swagger-types";
import { BACKEND_URL } from "../config";
import { getAuthToken, refreshAuthTokens, logoutSession } from "./authBridge";

export type CallMetricsResponse = components["schemas"]["CallMetricsResponse"];
export type CallMetricsRequest = components["schemas"]["CallMetricsRequest"];

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

export async function getCallMetrics(body: Partial<CallMetricsRequest> = {}): Promise<CallMetricsResponse> {
    const res = await fetchWithAuthRetry("/v1/metrics/calls", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
            date_from: body.date_from,
            date_to: body.date_to,
        }),
    });
    if (!res.ok) {
        const err = new Error("Failed to load metrics");
        (err as { code?: number }).code = res.status;
        throw err;
    }
    return res.json() as Promise<CallMetricsResponse>;
}
