import {
    createContext,
    useContext,
    useState,
    useCallback,
    useEffect,
    type ReactNode,
} from "react"

const STORAGE_KEY = "nexora-theme"

export type ThemeChoice = "system" | "light" | "dark" | "gray"

type ResolvedTheme = "light" | "dark" | "gray"

function getStoredTheme(): ThemeChoice {
    try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (raw === "system" || raw === "light" || raw === "dark" || raw === "gray") {
            return raw
        }
    } catch {
        // ignore
    }
    return "system"
}

function resolveTheme(choice: ThemeChoice): ResolvedTheme {
    if (choice === "light" || choice === "dark" || choice === "gray") {
        return choice
    }
    if (typeof window === "undefined" || !window.matchMedia) {
        return "light"
    }
    return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
}

function applyTheme(resolved: ResolvedTheme) {
    document.documentElement.setAttribute("data-theme", resolved)
}

type ThemeContextValue = {
    theme: ThemeChoice
    setTheme: (theme: ThemeChoice) => void
    resolvedTheme: ResolvedTheme
}

const ThemeContext = createContext<ThemeContextValue | null>(null)

export function ThemeProvider({ children }: { children: ReactNode }) {
    const [theme, setThemeState] = useState<ThemeChoice>(getStoredTheme)
    const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>(() =>
        resolveTheme(getStoredTheme())
    )

    const setTheme = useCallback((next: ThemeChoice) => {
        setThemeState(next)
        const resolved = resolveTheme(next)
        setResolvedTheme(resolved)
        applyTheme(resolved)
        try {
            localStorage.setItem(STORAGE_KEY, next)
        } catch {
            // ignore
        }
    }, [])

    useEffect(() => {
        applyTheme(resolvedTheme)
    }, [resolvedTheme])

    useEffect(() => {
        if (theme !== "system") return
        const mq = window.matchMedia("(prefers-color-scheme: dark)")
        const handle = () => {
            const resolved = mq.matches ? "dark" : "light"
            setResolvedTheme(resolved)
            applyTheme(resolved)
        }
        mq.addEventListener("change", handle)
        return () => mq.removeEventListener("change", handle)
    }, [theme])

    return (
        <ThemeContext.Provider value={{ theme, setTheme, resolvedTheme }}>
            {children}
        </ThemeContext.Provider>
    )
}

export function useTheme() {
    const ctx = useContext(ThemeContext)
    if (!ctx) {
        throw new Error("useTheme must be used within ThemeProvider")
    }
    return ctx
}
