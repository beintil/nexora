import { useState, useEffect, useCallback } from "react"
import { 
    getCompanyTelephony, 
    getTelephonyDictionary, 
    createCompanyTelephony, 
    deleteCompanyTelephony,
    type CompanyTelephonyItem, 
    type TelephonyDictionaryItem,
    type CompanyTelephonyCreateRequest
} from "../api/companyTelephony"

export function useTelephony() {
    const [data, setData] = useState<CompanyTelephonyItem[]>([])
    const [dictionary, setDictionary] = useState<TelephonyDictionaryItem[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const fetchTelephony = useCallback(async () => {
        setLoading(true)
        setError(null)
        try {
            const [list, dict] = await Promise.all([
                getCompanyTelephony(),
                getTelephonyDictionary()
            ])
            setData(list)
            setDictionary(dict)
        } catch (e) {
            setError(e instanceof Error ? e.message : "Failed to load telephony data")
            setData([])
            setDictionary([])
        } finally {
            setLoading(false)
        }
    }, [])

    const connect = useCallback(async (payload: CompanyTelephonyCreateRequest) => {
        try {
            await createCompanyTelephony(payload)
            await fetchTelephony()
            return { success: true }
        } catch (e) {
            return { success: false, error: e instanceof Error ? e.message : "Failed to connect telephony" }
        }
    }, [fetchTelephony])

    const detach = useCallback(async (id: string) => {
        try {
            await deleteCompanyTelephony(id)
            await fetchTelephony()
            return { success: true }
        } catch (e) {
            return { success: false, error: e instanceof Error ? e.message : "Failed to detach telephony" }
        }
    }, [fetchTelephony])

    useEffect(() => {
        fetchTelephony()
    }, [fetchTelephony])

    return { 
        data, 
        dictionary, 
        loading, 
        error, 
        refetch: fetchTelephony,
        connect,
        detach
    }
}
