import { useState, useEffect, useCallback } from "react"
import { getCalls, type CallsListResponse, type CallsListRequest } from "../api/calls"

export function useCallsList(initialParams: Partial<CallsListRequest> = {}) {
    const [data, setData] = useState<CallsListResponse | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [params, setParams] = useState<Partial<CallsListRequest>>({
        page: { limit: 20, offset: 0, page: 1 },
        ...initialParams,
    })

    const fetchCalls = useCallback(async () => {
        setLoading(true)
        setError(null)
        try {
            const res = await getCalls(params)
            setData(res)
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to load calls")
        } finally {
            setLoading(false)
        }
    }, [params])

    useEffect(() => {
        fetchCalls()
    }, [fetchCalls])

    const updateParams = useCallback((newParams: Partial<CallsListRequest>) => {
        setParams((prev) => ({
            ...prev,
            ...newParams,
            page: {
                limit: newParams.page?.limit ?? prev.page?.limit ?? 20,
                offset: newParams.page?.offset ?? (newParams.page ? prev.page?.offset ?? 0 : 0),
                page: newParams.page?.page ?? (newParams.page ? prev.page?.page ?? 1 : 1),
            },
        }))
    }, [])

    return { data, loading, error, params, updateParams, refetch: fetchCalls }
}
