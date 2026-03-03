import { useTheme, type ThemeChoice } from "../context/ThemeContext"
import { Monitor, Moon, Sun, Palette } from "lucide-react"

const THEMES: { value: ThemeChoice; label: string; icon: typeof Monitor }[] = [
    { value: "system", label: "Системная", icon: Monitor },
    { value: "light", label: "Светлая", icon: Sun },
    { value: "dark", label: "Тёмная", icon: Moon },
    { value: "gray", label: "Серая", icon: Palette },
]

export default function SettingsPage() {
    const { theme, setTheme } = useTheme()

    return (
        <div className="p-8 max-w-2xl">
            <h1 className="text-2xl font-semibold mb-2" style={{ color: "var(--theme-text)" }}>
                Настройки
            </h1>
            <p className="text-sm mb-8" style={{ color: "var(--theme-text-muted)" }}>
                Внешний вид и параметры приложения. Настройки хранятся только в браузере, сервер о них не знает.
            </p>

            <section className="mb-8">
                <h2 className="text-sm font-semibold uppercase tracking-wider mb-4" style={{ color: "var(--theme-text-muted)" }}>
                    Тема
                </h2>
                <div className="flex flex-wrap gap-3">
                    {THEMES.map(({ value, label, icon: Icon }) => (
                        <button
                            key={value}
                            type="button"
                            onClick={() => setTheme(value)}
                            className="flex items-center gap-2 px-4 py-3 rounded-xl border text-sm font-medium transition"
                            style={{
                                borderColor: theme === value ? "var(--theme-text)" : "var(--theme-border)",
                                backgroundColor: theme === value ? "var(--theme-active)" : "transparent",
                                color: theme === value ? "var(--theme-text)" : "var(--theme-text-muted)",
                            }}
                        >
                            <Icon className="h-4 w-4 shrink-0" />
                            {label}
                        </button>
                    ))}
                </div>
            </section>
        </div>
    )
}
