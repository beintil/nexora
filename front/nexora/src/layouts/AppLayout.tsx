import { useState, useEffect, memo } from "react"
import { Navigate, NavLink, useLocation, useNavigate, useOutlet } from "react-router-dom"
import type { LucideIcon } from "lucide-react"
import { useAuth } from "../context/AuthContext"
import { useProfileContext, ProfileProvider } from "../context/ProfileContext"
import { DEFAULT_MENU_ITEMS } from "../config/menu"
import { CommandMenu } from "../components/CommandMenu"
import { Home, Phone, Settings, BarChart3, LogOut, Loader2, LayoutDashboard, Bell, Users } from "lucide-react"
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
    <aside className="fixed inset-y-0 left-0 z-20 flex w-60 flex-col border-r border-border bg-sidebar-background">
      {/* Workspace Selector */}
      <div className="flex h-12 w-full items-center px-4 border-b border-border hover:bg-sidebar-accent cursor-pointer transition-colors">
        <div className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground shadow-sm mr-2 shrink-0">
          <span className="font-mono text-xs font-bold">N</span>
        </div>
        <div className="flex flex-col flex-1 overflow-hidden">
           <span className="text-sm font-medium text-sidebar-foreground truncate">
              {profileLoading ? "Loading..." : profile?.company_name || "Personal Workspace"}
           </span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex flex-1 flex-col overflow-y-auto px-3 py-4 gap-6 scrollbar-none">
        {MENU_GROUPS.map((group, idx) => (
          <div key={idx} className="flex flex-col gap-1 w-full">
            <span className="px-2 text-[10px] font-medium uppercase tracking-wider text-muted-foreground mb-1">
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
                      "relative flex items-center h-8 rounded-md px-2 text-sm transition-colors",
                      isActive 
                          ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium" 
                          : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                  )}
                >
                  <Icon className={cn("size-4 mr-2", isActive ? "text-primary dark:text-primary" : "text-muted-foreground")} />
                  {item.label}
                </NavLink>
              )
            })}
          </div>
        ))}
      </nav>

      {/* User Profile */}
      <div className="p-3 border-t border-border">
          <div 
              onClick={() => navigate("/profile")}
              className="flex items-center p-2 rounded-md hover:bg-sidebar-accent transition-colors cursor-pointer group"
          >
              <div className="size-8 rounded-full overflow-hidden bg-muted flex items-center justify-center border border-border shrink-0">
                  {profile?.avatar_url && !avatarImgError ? (
                      <img src={profile.avatar_url} alt="" className="size-full object-cover" onError={() => setAvatarImgError(true)} />
                  ) : (
                      <span className="text-xs font-mono text-muted-foreground">{initials}</span>
                  )}
              </div>
              <div className="flex flex-col ml-3 flex-1 overflow-hidden">
                  <span className="text-xs font-medium text-sidebar-foreground truncate">{displayName}</span>
                  <span className="text-[10px] text-muted-foreground truncate">{profile?.email || "No email"}</span>
              </div>
              <button
                  onClick={(e) => { 
                      e.preventDefault(); 
                      e.stopPropagation();
                      logout(); 
                  }}
                  className="p-1 rounded text-muted-foreground hover:text-destructive opacity-0 group-hover:opacity-100 transition-opacity"
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
    <header className="sticky top-0 z-10 flex h-12 shrink-0 items-center justify-between border-b border-border bg-background px-6">
      <div className="flex-1 max-w-md">
         <CommandMenu />
      </div>

      <div className="flex items-center gap-4 relative">
         <button 
            onClick={() => setIsNotifOpen(!isNotifOpen)}
            className={cn(
                "relative size-8 flex items-center justify-center rounded-md text-muted-foreground hover:bg-secondary transition-colors",
                isNotifOpen && "bg-secondary text-foreground"
            )}
         >
            <Bell className="size-4" />
            {unreadCount > 0 && (
                <span className="absolute top-1.5 right-1.5 size-2 rounded-full bg-primary border-2 border-background" />
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
    <div className="flex min-h-screen w-full bg-background text-foreground font-sans">
      
      <Sidebar 
        profile={profile} 
        profileLoading={profileLoading} 
        location={location} 
        navigate={navigate} 
        logout={logout} 
      />

      <main className="flex flex-1 flex-col pl-60">
        <Header />

        {/* Page Content */}
        <div className="flex-1 overflow-auto bg-muted/20 dark:bg-background">
          <div className="w-full h-full p-6">
            <AnimatePresence mode="wait">
                <motion.div
                key={location.pathname}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.3, ease: "easeOut" }}
                className="mx-auto w-full max-w-7xl h-full"
                >
                {outlet}
                </motion.div>
            </AnimatePresence>
          </div>
        </div>
      </main>
    </div>
  )
}

export default function AppLayout() {
  const { accessToken, isRestoring } = useAuth()

  if (isRestoring) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <Loader2 className="size-8 animate-spin text-muted-foreground/50" />
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
