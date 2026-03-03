import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useRef,
    useState,
    type ReactNode,
} from "react";
import { getProfile, roleLabel, type ProfileResponse } from "../api/profile";

type ProfileContextValue = {
    profile: ProfileResponse | null;
    loading: boolean;
    error: boolean;
    roleLabel: (roleId: number) => string;
    refetch: () => void;
};

const ProfileContext = createContext<ProfileContextValue | null>(null);

export function ProfileProvider({
    accessToken,
    children,
}: {
    accessToken: string;
    children: ReactNode;
}) {
    const [profile, setProfile] = useState<ProfileResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(false);
    const fetchedForTokenRef = useRef<string | null>(null);

    const fetchProfile = useCallback(() => {
        setLoading(true);
        getProfile()
            .then(setProfile)
            .catch(() => setError(true))
            .finally(() => setLoading(false));
    }, []);

    useEffect(() => {
        if (fetchedForTokenRef.current === accessToken) return;
        fetchedForTokenRef.current = accessToken;
        fetchProfile();
    }, [accessToken, fetchProfile]);

    const value: ProfileContextValue = {
        profile,
        loading,
        error,
        roleLabel,
        refetch: fetchProfile,
    };

    return (
        <ProfileContext.Provider value={value}>
            {children}
        </ProfileContext.Provider>
    );
}

export function useProfileContext(): ProfileContextValue {
    const ctx = useContext(ProfileContext);
    if (!ctx) {
        throw new Error("useProfileContext must be used within ProfileProvider");
    }
    return ctx;
}
