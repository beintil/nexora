import { useState, useEffect, memo, Suspense } from "react"
import { Navigate, NavLink, useLocation, useNavigate, useOutlet } from "react-router-dom"
import type { LucideIcon } from "lucide-react"
import { useAuth } from "../context/AuthContext"
import { useProfileContext, ProfileProvider } from "../context/ProfileContext"
import { DEFAULT_MENU_ITEMS } from "../config/menu"
import { CommandMenu } from "../components/CommandMenu"
import { Home, Phone, Settings, BarChart3, LogOut, LayoutDashboard, Bell, Users } from "lucide-react"
import { motion, AnimatePresence } from "framer-motion"
import { cn } from "@/lib/utils"
import { NotificationPanel } from "../components/NotificationPanel"
import { useNotifications } from "../hooks/useNotifications"

const iconByKey: Record<string, LucideIcon> = {
  home: Home,
  layout: LayoutDashboard,
  phone: Phone,
  settings: Settings,
  chart: BarChart3,
  users: Users,
}

const MENU_GROUPS = [
  { label: "Overview", ids: ["dashboard", "dashboards", "calls"] },
  { label: "Analytics", ids: ["analytics"] },
  { label: "System", ids: ["team", "settings"] },
]

const Sidebar = memo(({ profile, profileLoading, location, navigate, logout }: any) => {
  const [avatarImgError, setAvatarImgError] = useState(false)
  
  useEffect(() => {
    setAvatarImgError(false)
  }, [profile?.avatar_url])

  const displayName = profile?.full_name?.trim() || profile?.email?.split("@")[0] || "User"
  const initials = displayName.slice(0, 2).toUpperCase()

  return (
    <aside className="fixed inset-y-0 left-0 z-20 flex w-64 flex-col border-r border-border/40 bg-sidebar-background transition-colors duration-200">
      {/* Workspace Selector */}
      <div className="flex h-14 w-full items-center px-4 mt-2 mb-2">
        <div className="flex w-full items-center gap-3 rounded-xl p-2 hover:bg-black/5 dark:hover:bg-white/5 cursor-pointer transition-all active:scale-[0.98]">
            <div className="flex size-8 shadow-sm items-center justify-center rounded-lg bg-primary text-primary-foreground shrink-0 border border-border/10">
            <span className="font-mono text-sm font-bold">N</span>
            </div>
            <div className="flex flex-col flex-1 overflow-hidden">
            <span className="text-sm font-bold text-foreground truncate tracking-tight">
                {profileLoading ? "Loading..." : profile?.company_name || "Personal Workspace"}
            </span>
            {/* Subscription Plan Badge */}
            <div className="mt-1">
              {(profile as any)?.plan?.name ? (
                 <div className="flex flex-col gap-0.5">
                    <span className="inline-flex w-fit items-center px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-widest bg-emerald-500/10 text-emerald-600 border border-emerald-500/20">
                      {(profile as any).plan.name}
                    </span>
                    <span className="text-[9px] text-muted-foreground font-medium">
                      Осталось: {(profile as any).plan.daysLeft ?? 0} дн.
                    </span>
                 </div>
              ) : (
                <span className="inline-flex w-fit items-center px-1.5 py-0.5 rounded text-[10px] font-bold uppercase tracking-widest bg-slate-500/10 text-slate-500 border border-slate-500/20">
                  Плана нет
                </span>
              )}
            </div>
            </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex flex-1 flex-col overflow-y-auto px-4 py-2 gap-6 scrollbar-none">
        {MENU_GROUPS.map((group, idx) => (
          <div key={idx} className="flex flex-col gap-1 w-full">
            <span className="px-2 text-[10px] font-bold uppercase tracking-widest text-muted-foreground/70 mb-2">
              {group.label}
            </span>
            {DEFAULT_MENU_ITEMS.filter((i) => group.ids.includes(i.id)).map((item) => {
              const Icon = iconByKey[item.icon] ?? Home
              const isActive = location.pathname.startsWith(item.path) && (item.path !== "/" || location.pathname === "/")
              
              return (
                <NavLink
                  key={item.id}
                  to={item.path}
                  className={cn(
                      "relative flex items-center h-9 rounded-lg px-2.5 text-sm transition-all duration-200 group",
                      isActive 
                          ? "bg-black/5 dark:bg-white/10 text-foreground font-semibold shadow-[inset_0_1px_rgba(255,255,255,0.1)]" 
                          : "text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground font-medium"
                  )}
                >
                  <Icon className={cn("size-4 mr-3 transition-colors", isActive ? "text-foreground" : "text-muted-foreground group-hover:text-foreground")} />
                  {item.label}
                </NavLink>
              )
            })}
          </div>
        ))}
      </nav>

      {/* User Profile */}
      <div className="p-4">
          <div 
              onClick={() => navigate("/profile")}
              className="flex items-center p-2 rounded-xl hover:bg-black/5 dark:hover:bg-white/5 transition-all cursor-pointer group border border-transparent hover:border-border/30 active:scale-[0.98]"
          >
              <div className="size-9 rounded-full overflow-hidden bg-gradient-to-br from-secondary to-muted flex items-center justify-center border border-border/50 shrink-0 shadow-sm">
                  {profile?.avatar_url && !avatarImgError ? (
                      <img src={profile.avatar_url} alt="" className="size-full object-cover" onError={() => setAvatarImgError(true)} />
                  ) : (
                      <span className="text-xs font-bold text-foreground">{initials}</span>
                  )}
              </div>
              <div className="flex flex-col ml-3 flex-1 overflow-hidden">
                  <span className="text-sm font-semibold text-foreground truncate">{displayName}</span>
                  <span className="text-[11px] font-medium text-muted-foreground truncate">{profile?.email || "No email"}</span>
              </div>
              <button
                  onClick={(e) => { 
                      e.preventDefault(); 
                      e.stopPropagation();
                      logout(); 
                  }}
                  className="p-1.5 rounded-md text-muted-foreground hover:text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-all"
                  title="Log out"
              >
                  <LogOut className="size-4" />
              </button>
          </div>
      </div>
    </aside>
  )
})

const Header = memo(() => {
  const [isNotifOpen, setIsNotifOpen] = useState(false)
  const { unreadCount } = useNotifications()

  return (
    <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center justify-between border-b border-border/40 bg-background/80 backdrop-blur-md px-6 transition-all">
      <div className="flex-1 max-w-md">
         <CommandMenu />
      </div>

      <div className="flex items-center gap-4 relative">
         <button 
            onClick={() => setIsNotifOpen(!isNotifOpen)}
            className={cn(
                "relative size-9 flex items-center justify-center rounded-full text-muted-foreground hover:bg-black/5 dark:hover:bg-white/5 hover:text-foreground transition-all active:scale-[0.95]",
                isNotifOpen && "bg-black/5 dark:bg-white/5 text-foreground shadow-inner"
            )}
         >
            <Bell className="size-4" />
            {unreadCount > 0 && (
                <span className="absolute top-2 right-2 size-2 rounded-full bg-primary border-2 border-background" />
            )}
         </button>
         <NotificationPanel isOpen={isNotifOpen} onClose={() => setIsNotifOpen(false)} />
      </div>
    </header>
  )
})

function AppLayoutContent() {
  const { logout } = useAuth()
  const { profile, loading: profileLoading } = useProfileContext()
  const location = useLocation()
  const navigate = useNavigate()
  const outlet = useOutlet()

  return (
    <div className="flex min-h-screen w-full bg-background text-foreground font-sans selection:bg-primary/20">
      
      <Sidebar 
        profile={profile} 
        profileLoading={profileLoading} 
        location={location} 
        navigate={navigate} 
        logout={logout} 
      />

      <main className="flex flex-1 flex-col pl-64 transition-all duration-200">
        <Header />

        {/* Page Content */}
        <div className="flex-1 overflow-auto bg-transparent">
          <div className="w-full h-full p-6 md:p-8">
            <AnimatePresence mode="wait">
                <motion.div
                key={location.pathname}
                initial={{ opacity: 0, y: 15, filter: "blur(4px)" }}
                animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                exit={{ opacity: 0, y: -15, filter: "blur(4px)" }}
                transition={{ duration: 0.4, ease: [0.22, 1, 0.36, 1] }}
                className="mx-auto w-full max-w-7xl h-full"
                >
                  <Suspense fallback={<div className="h-full w-full" />}>
                    {outlet}
                  </Suspense>
                </motion.div>
            </AnimatePresence>
          </div>
        </div>
      </main>
    </div>
  )
}

function AppLayoutSkeleton() {
  return (
    <div className="flex min-h-screen w-full bg-background text-foreground font-sans">
      <aside className="fixed inset-y-0 left-0 z-20 flex w-64 flex-col border-r border-border/40 bg-sidebar-background">
        <div className="flex h-14 w-full items-center px-4 mt-2">
          <div className="flex size-8 items-center justify-center rounded-lg bg-muted text-transparent shadow-sm mr-3 shrink-0">N</div>
          <div className="h-4 w-28 bg-muted rounded animate-pulse"></div>
        </div>
        <nav className="flex flex-1 flex-col px-4 py-4 gap-6 scrollbar-none"></nav>
      </aside>
      <main className="flex flex-1 flex-col pl-64">
        <header className="sticky top-0 z-10 flex h-14 shrink-0 items-center justify-between border-b border-border/40 bg-background/80 px-6">
           <div className="h-8 w-64 bg-muted/50 rounded-lg animate-pulse" />
        </header>
        <div className="flex-1 overflow-auto bg-transparent pt-8 px-8">
           <div className="mx-auto w-full max-w-7xl h-full" />
        </div>
      </main>
    </div>
  )
}

export default function AppLayout() {
  const { accessToken, isRestoring } = useAuth()

  if (isRestoring) {
    return <AppLayoutSkeleton />
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
