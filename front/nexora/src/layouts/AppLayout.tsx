import { useState, useEffect } from "react"
import { Outlet, Navigate, NavLink, Link } from "react-router-dom"
import type { LucideIcon } from "lucide-react"
import { useAuth } from "../context/AuthContext"
import { useProfileContext, ProfileProvider } from "../context/ProfileContext"
import { DEFAULT_MENU_ITEMS } from "../config/menu"
import { Home, Phone, Settings, BarChart3, LogOut, Loader2, LayoutDashboard, User } from "lucide-react"

const iconByKey: Record<string, LucideIcon> = {
    home: Home,
    layout: LayoutDashboard,
    phone: Phone,
    settings: Settings,
    chart: BarChart3,
}

const MENU_GROUPS = [
    { label: "Основное", ids: ["dashboard", "dashboards", "calls"] },
    { label: "Аналитика", ids: ["analytics"] },
    { label: "Система", ids: ["settings"] },
]

function AppLayoutContent() {
    const { logout } = useAuth()
    const { profile, loading: profileLoading, roleLabel } = useProfileContext()
    const [avatarImgError, setAvatarImgError] = useState(false)

    useEffect(() => {
        setAvatarImgError(false)
    }, [profile?.avatar_url])

    const displayName =
        profile?.full_name?.trim() ||
        profile?.email?.split("@")[0] ||
        ""
    const initials = displayName
        ? displayName
              .split(/\s+/)
              .map((s) => s[0])
              .join("")
              .toUpperCase()
              .slice(0, 2)
        : ""

    return (
        <div className="min-h-screen flex" style={{ backgroundColor: "var(--theme-bg-page)" }}>
            <aside className="w-64 flex flex-col shrink-0 shadow-sm border-r border-[var(--theme-border)]" style={{ backgroundColor: "var(--theme-bg-sidebar)" }}>
                <div className="p-5 border-b border-[var(--theme-border)]">
                    <span className="text-lg font-semibold tracking-tight" style={{ fontFamily: '"Cormorant Garamond", Georgia, serif', color: "var(--theme-text)" }}>Nexora</span>
                </div>
                <nav className="flex-1 py-4 px-3 overflow-auto">
                    {MENU_GROUPS.map((group) => (
                        <div key={group.label} className="mb-6">
                            <p className="px-3 mb-2 text-[11px] font-semibold uppercase tracking-wider" style={{ color: "var(--theme-text-muted)" }}>{group.label}</p>
                            <div className="space-y-0.5">
                                {DEFAULT_MENU_ITEMS.filter((i) => group.ids.includes(i.id)).map((item) => {
                                    const Icon = iconByKey[item.icon] ?? Home
                                    return (
                                        <NavLink
                                            key={item.id}
                                            to={item.path}
                                            className={({ isActive }) =>
                                                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition ${isActive ? "nav-link-active" : "nav-link-inactive"}`
                                            }
                                        >
                                            <Icon className="h-4 w-4 shrink-0" />
                                            {item.label}
                                        </NavLink>
                                    )
                                })}
                            </div>
                        </div>
                    ))}
                </nav>
                <div className="p-3 border-t border-[var(--theme-border)]">
                    <button
                        onClick={logout}
                        className="flex items-center gap-3 w-full px-3 py-2.5 rounded-lg text-sm transition nav-link-inactive mb-3"
                    >
                        <LogOut className="h-4 w-4 shrink-0" />
                        Выйти
                    </button>
                    <Link
                        to="/profile"
                        className="flex items-center gap-3 px-3 py-3 rounded-xl transition mt-3"
                        style={{ backgroundColor: "var(--theme-hover)" }}
                    >
                        {profile?.avatar_url && !avatarImgError ? (
                            <img
                                src={profile.avatar_url}
                                alt=""
                                className="w-10 h-10 rounded-full object-cover border border-[var(--theme-border)]"
                                onError={() => setAvatarImgError(true)}
                            />
                        ) : (
                            <div className="w-10 h-10 rounded-full flex items-center justify-center text-sm font-medium" style={{ backgroundColor: "var(--theme-active)", color: "var(--theme-text-muted)" }} aria-hidden>
                                {initials || <User className="h-4 w-4" />}
                            </div>
                        )}
                        <div className="min-w-0 flex-1">
                            <p className="text-sm font-medium truncate" style={{ color: "var(--theme-text)" }}>{displayName || "Профиль"}</p>
                            {profile?.role_id !== undefined && (
                                <p className="text-xs truncate mt-0.5" style={{ color: "var(--theme-text-muted)" }}>{roleLabel(profile.role_id)}</p>
                            )}
                            <p className="text-xs truncate mt-0.5" style={{ color: "var(--theme-text-muted)" }}>{profile?.email ?? ""}</p>
                        </div>
                    </Link>
                </div>
            </aside>
            <main className="flex-1 flex flex-col min-h-0">
                <header className="h-14 shrink-0 flex items-center px-6 border-b border-[var(--theme-border)]" style={{ backgroundColor: "var(--theme-bg-header)" }}>
                    <span className="text-sm truncate" style={{ color: "var(--theme-text-muted)" }}>
                        {profileLoading ? (
                            <span className="inline-block w-32 h-4 rounded animate-pulse" style={{ backgroundColor: "var(--theme-active)" }} />
                        ) : (
                            profile?.company_name ?? "—"
                        )}
                    </span>
                </header>
                <div className="flex-1 overflow-auto">
                    <Outlet />
                </div>
            </main>
        </div>
    )
}

export default function AppLayout() {
    const { accessToken, isRestoring } = useAuth()

    if (isRestoring) {
        return (
            <div className="min-h-screen bg-slate-50 flex items-center justify-center">
                <Loader2 className="h-10 w-10 text-slate-400 animate-spin" />
            </div>
        )
    }

    if (!accessToken) {
        return <Navigate to="/" replace />
    }

    return (
        <ProfileProvider accessToken={accessToken}>
            <AppLayoutContent />
        </ProfileProvider>
    )
}
