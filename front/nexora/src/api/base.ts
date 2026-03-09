import {
    getAuthToken,
    refreshAuthTokens,
    logoutSession,
} from "./authBridge";

export async function fetchWithAuthRetry(
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
