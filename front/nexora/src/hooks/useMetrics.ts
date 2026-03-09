import { useState, useEffect, useCallback } from "react"
import { getCallMetrics, type CallMetricsResponse, type CallMetricsRequest } from "../api/metrics"

export function useMetrics(initialParams: Partial<CallMetricsRequest> = {}) {
    const [data, setData] = useState<CallMetricsResponse | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [params, setParams] = useState<Partial<CallMetricsRequest>>(initialParams)

    const fetchMetrics = useCallback(async () => {
        setLoading(true)
        setError(null)
        try {
            const res = await getCallMetrics(params)
            setData(res)
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to load metrics")
        } finally {
            setLoading(false)
        }
    }, [params])

    useEffect(() => {
        fetchMetrics()
    }, [fetchMetrics])

    const updateRange = useCallback((date_from: string, date_to: string) => {
        setParams({ date_from, date_to })
    }, [])

    return { data, loading, error, params, updateRange, refetch: fetchMetrics }
}
