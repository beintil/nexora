import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useRef,
    useState,
    type ReactNode,
} from "react";
import { refreshRequest, logoutRequest } from "../api/auth";
import { setAuthBridge } from "../api/authBridge";

type AuthContextValue = {
    accessToken: string | null;
    setAccessToken: (token: string | null) => void;
    logout: () => void;
    refreshTokens: () => Promise<string | null>;
    isRestoring: boolean;
};

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [accessToken, setAccessTokenState] = useState<string | null>(null);
    const [isRestoring, setIsRestoring] = useState(true);
    const tokenRef = useRef<string | null>(null);

    useEffect(() => {
        tokenRef.current = accessToken;
    }, [accessToken]);

    const setAccessToken = useCallback((token: string | null) => {
        setAccessTokenState(token);
    }, []);

    const logout = useCallback(() => {
        const storedRefresh = sessionStorage.getItem("refresh_token");
        sessionStorage.removeItem("refresh_token");
        setAccessTokenState(null);
        tokenRef.current = null;
        if (storedRefresh) {
            logoutRequest(storedRefresh).catch(() => {});
        }
    }, []);

    const refreshTokens = useCallback(async (): Promise<string | null> => {
        try {
            const storedRefresh = sessionStorage.getItem("refresh_token");
            const data = await refreshRequest(storedRefresh ?? undefined);
            setAccessTokenState(data.accessToken);
            tokenRef.current = data.accessToken;
            if (data.refreshToken) {
                sessionStorage.setItem("refresh_token", data.refreshToken);
            }
            return data.accessToken;
        } catch {
            setAccessTokenState(null);
            tokenRef.current = null;
            sessionStorage.removeItem("refresh_token");
            return null;
        }
    }, []);

    useEffect(() => {
        setAuthBridge({
            getToken: () => tokenRef.current,
            refreshTokens,
            logout,
        });
        return () => setAuthBridge(null);
    }, [refreshTokens, logout]);

    useEffect(() => {
        const hash = window.location.hash;
        if (hash) {
            const params = new URLSearchParams(hash.replace(/^#/, ""));
            const access = params.get("access_token");
            const refresh = params.get("refresh_token");
            if (access) {
                setAccessTokenState(access);
                if (refresh) sessionStorage.setItem("refresh_token", refresh);
                window.history.replaceState(null, "", window.location.pathname + window.location.search);
            }
        }
    }, []);

    const restoringRef = useRef(false);
    useEffect(() => {
        const storedRefresh = sessionStorage.getItem("refresh_token");
        if (!storedRefresh) {
            setIsRestoring(false);
            return;
        }
        if (restoringRef.current) return;
        restoringRef.current = true;
        refreshTokens().finally(() => {
            setIsRestoring(false);
            restoringRef.current = false;
        });
    }, [refreshTokens]);

    const value: AuthContextValue = {
        accessToken,
        setAccessToken,
        logout,
        refreshTokens,
        isRestoring,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth(): AuthContextValue {
    const ctx = useContext(AuthContext);
    if (!ctx) {
        throw new Error("useAuth must be used within AuthProvider");
    }
    return ctx;
}
