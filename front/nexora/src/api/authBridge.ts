/**
 * Мост для доступа к токену и обновлению сессии из слоя API (без React).
 * Устанавливается AuthProvider при монтировании.
 */

let getToken: () => string | null = () => null;
let refreshTokens: () => Promise<string | null> = async () => null;
let logout: () => void = () => {};

export function setAuthBridge(b: {
    getToken: () => string | null;
    refreshTokens: () => Promise<string | null>;
    logout: () => void;
} | null) {
    if (!b) {
        getToken = () => null;
        refreshTokens = async () => null;
        logout = () => {};
        return;
    }
    getToken = b.getToken;
    refreshTokens = b.refreshTokens;
    logout = b.logout;
}

export function getAuthToken(): string | null {
    return getToken();
}

export async function refreshAuthTokens(): Promise<string | null> {
    return refreshTokens();
}

export function logoutSession(): void {
    logout();
}
