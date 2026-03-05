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
            <aside className="w-60 flex flex-col shrink-0 border-r border-slate-200/80 bg-white shadow-[4px_0_24px_-4px_rgba(0,0,0,0.06)]">
                <div className="px-5 py-6 border-b border-slate-100">
                    <span className="text-xl font-bold tracking-tight text-slate-800" style={{ fontFamily: '"Cormorant Garamond", Georgia, serif' }}>
                        Nexora
                    </span>
                </div>
                <nav className="flex-1 py-5 px-3 overflow-auto">
                    {MENU_GROUPS.map((group) => (
                        <div key={group.label} className="mb-5">
                            <p className="px-3 mb-2 text-[10px] font-semibold uppercase tracking-widest text-slate-400">
                                {group.label}
                            </p>
                            <div className="space-y-0.5">
                                {DEFAULT_MENU_ITEMS.filter((i) => group.ids.includes(i.id)).map((item) => {
                                    const Icon = iconByKey[item.icon] ?? Home
                                    return (
                                        <NavLink
                                            key={item.id}
                                            to={item.path}
                                            end={item.path !== "/calls"}
                                            className={({ isActive }) =>
                                                `flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${
                                                    isActive
                                                        ? "bg-indigo-50 text-indigo-700 shadow-sm"
                                                        : "text-slate-600 hover:bg-slate-50 hover:text-slate-900"
                                                }`
                                            }
                                        >
                                            {({ isActive }) => (
                                                <>
                                                    <span className={`flex items-center justify-center w-8 h-8 rounded-lg shrink-0 transition-colors ${isActive ? "bg-indigo-100 text-indigo-600" : "bg-transparent"}`}>
                                                        <Icon className="h-4 w-4" />
                                                    </span>
                                                    <span>{item.label}</span>
                                                </>
                                            )}
                                        </NavLink>
                                    )
                                })}
                            </div>
                        </div>
                    ))}
                </nav>
                <div className="p-3 border-t border-slate-100 bg-slate-50/50">
                    <button
                        onClick={logout}
                        className="flex items-center gap-3 w-full px-3 py-2.5 rounded-xl text-sm font-medium text-slate-600 hover:bg-slate-100 hover:text-slate-900 transition-colors mb-2"
                    >
                        <LogOut className="h-4 w-4 shrink-0" />
                        Выйти
                    </button>
                    <Link
                        to="/profile"
                        className="flex items-center gap-3 px-3 py-3 rounded-xl bg-white border border-slate-200/80 shadow-sm hover:shadow transition-shadow"
                    >
                        {profile?.avatar_url && !avatarImgError ? (
                            <img
                                src={profile.avatar_url}
                                alt=""
                                className="w-10 h-10 rounded-xl object-cover border border-slate-200"
                                onError={() => setAvatarImgError(true)}
                            />
                        ) : (
                            <div className="w-10 h-10 rounded-xl flex items-center justify-center text-sm font-semibold bg-slate-200 text-slate-600 shrink-0" aria-hidden>
                                {initials || <User className="h-4 w-4" />}
                            </div>
                        )}
                        <div className="min-w-0 flex-1">
                            <p className="text-sm font-semibold truncate text-slate-800">{displayName || "Профиль"}</p>
                            {profile?.role_id !== undefined && (
                                <p className="text-xs truncate mt-0.5 text-slate-500">{roleLabel(profile.role_id)}</p>
                            )}
                            <p className="text-xs truncate mt-0.5 text-slate-400">{profile?.email ?? ""}</p>
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
