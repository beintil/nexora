import { useEffect, useState } from "react"
import { Command } from "cmdk"
import { useNavigate } from "react-router-dom"
import { 
  Home, 
  BarChart3, 
  Settings, 
  Phone,
  LayoutDashboard,
  Search
} from "lucide-react"

export function CommandMenu() {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((open) => !open)
      }
    }
    document.addEventListener("keydown", down)
    return () => document.removeEventListener("keydown", down)
  }, [])

  const runCommand = (command: () => void) => {
    setOpen(false)
    command()
  }

  return (
    <>
      <button 
        onClick={() => setOpen(true)}
        className="flex items-center w-full bg-secondary/50 hover:bg-secondary border border-border rounded-md px-3 h-8 transition-colors text-muted-foreground group"
      >
        <Search className="size-4 mr-2 opacity-50 group-hover:opacity-100" />
        <span className="text-xs flex-1 text-left">Search anything...</span>
        <kbd className="hidden sm:inline-flex h-5 items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100">
            <span className="text-xs">⌘</span>K
        </kbd>
      </button>

      <Command.Dialog
        open={open}
        onOpenChange={setOpen}
        label="Global Search"
        className="fixed inset-0 z-50 flex items-start justify-center pt-[15vh] sm:pt-[20vh] bg-background/80 backdrop-blur-sm"
      >
        <div className="w-full max-w-xl overflow-hidden rounded-xl border border-border bg-popover text-popover-foreground shadow-2xl">
          <Command 
             className="[&_[cmdk-root]]:w-full [&_[cmdk-input]]:h-12 [&_[cmdk-input]]:px-4 [&_[cmdk-input]]:text-sm [&_[cmdk-item]]:px-4 [&_[cmdk-item]]:py-3 [&_[cmdk-item]]:text-sm"
             shouldFilter={true}
          >
            <div className="flex items-center border-b border-border px-3">
              <Search className="size-4 shrink-0 text-muted-foreground" />
              <Command.Input 
                 placeholder="Type a command or search..." 
                 className="flex h-12 w-full rounded-md bg-transparent px-2 py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>
            
            <Command.List className="max-h-[300px] overflow-y-auto overflow-x-hidden p-2">
              <Command.Empty className="py-6 text-center text-sm text-muted-foreground">
                No results found.
              </Command.Empty>
              
              <Command.Group heading="Navigation" className="px-2 py-1 text-xs font-semibold text-muted-foreground">
                <Command.Item 
                  onSelect={() => runCommand(() => navigate("/dashboard"))}
                  className="relative flex cursor-default select-none items-center rounded-sm px-2 py-2 text-sm outline-none aria-selected:bg-accent aria-selected:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 mt-1"
                >
                  <Home className="mr-2 size-4" />
                  <span>Dashboard</span>
                </Command.Item>
                <Command.Item 
                  onSelect={() => runCommand(() => navigate("/calls"))}
                  className="relative flex cursor-default select-none items-center rounded-sm px-2 py-2 text-sm outline-none aria-selected:bg-accent aria-selected:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 mt-1"
                >
                  <Phone className="mr-2 size-4" />
                  <span>Calls Log</span>
                </Command.Item>
                <Command.Item 
                  onSelect={() => runCommand(() => navigate("/analytics"))}
                  className="relative flex cursor-default select-none items-center rounded-sm px-2 py-2 text-sm outline-none aria-selected:bg-accent aria-selected:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 mt-1"
                >
                  <BarChart3 className="mr-2 size-4" />
                  <span>Analytics</span>
                </Command.Item>
              </Command.Group>
              
              <Command.Separator className="-mx-2 h-px bg-border my-2" />
              
              <Command.Group heading="Settings" className="px-2 py-1 text-xs font-semibold text-muted-foreground">
                <Command.Item 
                  onSelect={() => runCommand(() => navigate("/profile"))}
                  className="relative flex cursor-default select-none items-center rounded-sm px-2 py-2 text-sm outline-none aria-selected:bg-accent aria-selected:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 mt-1"
                >
                  <Settings className="mr-2 size-4" />
                  <span>Profile Settings</span>
                </Command.Item>
                <Command.Item 
                  onSelect={() => runCommand(() => navigate("/dashboards"))}
                  className="relative flex cursor-default select-none items-center rounded-sm px-2 py-2 text-sm outline-none aria-selected:bg-accent aria-selected:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 mt-1"
                >
                  <LayoutDashboard className="mr-2 size-4" />
                  <span>Manage Dashboards</span>
                </Command.Item>
              </Command.Group>
            </Command.List>
          </Command>
        </div>
      </Command.Dialog>
    </>
  )
}
