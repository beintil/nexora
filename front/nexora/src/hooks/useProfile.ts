import { useState, useEffect } from "react";
import { getProfile, roleLabel, type ProfileResponse } from "../api/profile";

export function useProfile(accessToken: string | null) {
    const [profile, setProfile] = useState<ProfileResponse | null>(null);
    const [loading, setLoading] = useState(!!accessToken);
    const [error, setError] = useState(false);

    useEffect(() => {
        if (!accessToken) {
            setProfile(null);
            setLoading(false);
            setError(false);
            return;
        }
        setLoading(true);
        setError(false);
        getProfile()
            .then(setProfile)
            .catch(() => setError(true))
            .finally(() => setLoading(false));
    }, [accessToken]);

    const refetch = () => {
        if (!accessToken) return;
        setLoading(true);
        getProfile()
            .then(setProfile)
            .catch(() => setError(true))
            .finally(() => setLoading(false));
    };

    return { profile, loading, error, roleLabel, refetch };
}
