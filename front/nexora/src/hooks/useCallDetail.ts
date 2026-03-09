import { useState, useEffect, useCallback } from "react";
import { getCallById, type CallTreeResponse } from "../api/calls";

export function useCallDetail(id?: string) {
    const [data, setData] = useState<CallTreeResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchDetail = useCallback(async () => {
        if (!id) return;
        setLoading(true);
        setError(null);
        try {
            const res = await getCallById(id);
            setData(res);
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to load call details");
            setData(null);
        } finally {
            setLoading(false);
        }
    }, [id]);

    useEffect(() => {
        fetchDetail();
    }, [fetchDetail]);

    return { data, loading, error, refetch: fetchDetail };
}
