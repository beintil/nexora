import { useMemo } from "react";
import { useNavigate, Link } from "react-router-dom";
import { Phone, PhoneCall, PhoneMissed, ArrowUpRight, ArrowDownLeft, Minus } from "lucide-react";
import { AreaChart, Area, Tooltip, ResponsiveContainer, XAxis, YAxis, CartesianGrid } from "recharts";
import { cn } from "@/lib/utils";
import { motion } from "framer-motion";
import type { Variants } from "framer-motion";
import { useMetrics } from "../hooks/useMetrics";
import { useCallsList } from "../hooks/useCallsList";
import { timeAgo } from "@/lib/time";
import { useProfileContext } from "../context/ProfileContext";

const CHART_COLORS = {
  total: "#a1a1aa", // zinc-400
  answered: "#10b981", // emerald-500
  missed: "#ef4444", // red-500
};

function todayISO(): string {
  return new Date().toISOString().slice(0, 10);
}

function yesterdayISO(): string {
  const d = new Date();
  d.setDate(d.getDate() - 1);
  return d.toISOString().slice(0, 10);
}

const containerVars: Variants = {
  hidden: { opacity: 0 },
  show: { opacity: 1, transition: { staggerChildren: 0.05 } },
};

const itemVars: Variants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { type: "spring", stiffness: 300, damping: 24 } },
};

export default function DashboardPage() {
  const navigate = useNavigate();
  const { profile } = useProfileContext();

  const initialDateFrom = useMemo(() => {
    const from = new Date();
    from.setDate(from.getDate() - 6);
    return from.toISOString().slice(0, 10);
  }, []);
  const initialDateTo = useMemo(() => todayISO(), []);

  const { data: metrics, loading, error, refetch } = useMetrics({
    date_from: initialDateFrom,
    date_to: initialDateTo,
  });

  const { data: callsData, loading: callsLoading } = useCallsList({
    page: { limit: 6, offset: 0, page: 1 }
  });

  const recentCalls = callsData?.items || [];

  const sparkDataTotal = useMemo(() => 
    metrics?.timeseries && metrics.timeseries.length > 0
      ? metrics.timeseries.map((p) => ({ v: p.total }))
      : [{ v: 0 }], [metrics]);

  const sparkDataAnswered = useMemo(() => 
    metrics?.timeseries && metrics.timeseries.length > 0
      ? metrics.timeseries.map((p) => ({ v: p.answered }))
      : [{ v: 0 }], [metrics]);

  const sparkDataMissed = useMemo(() => 
    metrics?.timeseries && metrics.timeseries.length > 0
      ? metrics.timeseries.map((p) => ({ v: p.missed }))
      : [{ v: 0 }], [metrics]);

  const chartData = useMemo(() => 
    metrics?.timeseries && metrics.timeseries.length > 0
      ? metrics.timeseries.map((p) => ({
        date: p.date ? new Date(p.date).toLocaleDateString("en-US", { day: "numeric", month: "short" }) : "",
        dateISO: p.date ?? "",
        total: p.total ?? 0,
        answered: p.answered ?? 0,
        missed: p.missed ?? 0,
      }))
      : [], [metrics]);

  const todayStr = todayISO();
  const yesterdayStr = yesterdayISO();
  const todayPoint = chartData.find((p) => p.dateISO === todayStr);
  const yesterdayPoint = chartData.find((p) => p.dateISO === yesterdayStr);

  const cards = [
    {
      id: "all",
      icon: Phone,
      title: "Total Calls",
      value: loading ? "—" : String(todayPoint?.total ?? 0),
      chartColor: CHART_COLORS.total,
      sparkData: sparkDataTotal,
      colSpan: "col-span-1 md:col-span-2 lg:col-span-1", // Bento size
      onClick: () => navigate("/calls"),
    },
    {
      id: "answered",
      icon: PhoneCall,
      title: "Answered",
      value: loading ? "—" : String(todayPoint?.answered ?? 0),
      chartColor: CHART_COLORS.answered,
      sparkData: sparkDataAnswered,
      colSpan: "col-span-1",
      onClick: () =>
        navigate("/calls", {
          state: {
            initialStatus: "call_event_status_completed",
          },
        }),
    },
    {
      id: "missed",
      icon: PhoneMissed,
      title: "Missed",
      value: loading ? "—" : String(todayPoint?.missed ?? 0),
      chartColor: CHART_COLORS.missed,
      sparkData: sparkDataMissed,
      colSpan: "col-span-1",
      onClick: () =>
        navigate("/calls", {
          state: {
            initialStatus: "call_event_status_no_answer",
            initialDirection: "call_direction_inbound",
          },
        }),
    },
  ];



  const getTrend = (current: number = 0, previous: number = 0) => {
    if (current > previous) return { icon: ArrowUpRight, color: "text-emerald-500", bg: "bg-emerald-500/10", text: `+${current - previous}` };
    if (current < previous) return { icon: ArrowDownLeft, color: "text-red-500", bg: "bg-red-500/10", text: `${current - previous}` };
    return { icon: Minus, color: "text-muted-foreground", bg: "bg-secondary", text: "0" };
  };

  const getStatusConfig = (status: string | undefined) => {
    switch (status) {
      case "call_event_status_completed":
        return { label: "Completed", colorClass: "text-emerald-600 dark:text-emerald-500", dotClass: "bg-emerald-500" };
      case "call_event_status_no_answer":
        return { label: "No Answer", colorClass: "text-red-600 dark:text-red-500", dotClass: "bg-red-500" };
      case "call_event_status_canceled":
        return { label: "Canceled", colorClass: "text-muted-foreground", dotClass: "bg-muted-foreground" };
      case "call_event_status_busy":
        return { label: "Busy", colorClass: "text-amber-600 dark:text-amber-500", dotClass: "bg-amber-500" };
      case "call_event_status_failed":
        return { label: "Failed", colorClass: "text-red-600 dark:text-red-500", dotClass: "bg-red-700" };
      case "call_event_status_in_progress":
        return { label: "In Progress", colorClass: "text-blue-600 dark:text-blue-500", dotClass: "bg-blue-500" };
      case "call_event_status_ringing":
        return { label: "Ringing", colorClass: "text-blue-400", dotClass: "bg-blue-400 animate-pulse" };
      case "call_event_status_queued":
        return { label: "Queued", colorClass: "text-amber-600 dark:text-amber-500", dotClass: "bg-amber-300" };
      case "call_event_status_timeout":
        return { label: "Timeout", colorClass: "text-red-600 dark:text-red-500", dotClass: "bg-red-500" };
      default:
        return { label: status ? status.replace("call_event_status_", "").replace("_", " ") : "Unknown", colorClass: "text-muted-foreground", dotClass: "bg-muted" };
    }
  };

  return (
    <div className="flex flex-col gap-6 font-sans">
      
      {/* Premium Minimalist Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 pb-4 border-b border-border/40">
        <div className="flex flex-col gap-1">
            <h1 className="text-2xl font-bold tracking-tight text-foreground">Overview</h1>
            <p className="text-sm font-medium text-muted-foreground/80">
                Your communication metrics for the last 7 days.
            </p>
        </div>
        <div className="flex items-center gap-2">
            <button className="flex items-center gap-2 h-9 rounded-lg bg-secondary/50 px-3.5 text-xs font-semibold text-foreground ring-1 ring-inset ring-border/50 hover:bg-secondary transition-all active:scale-[0.98]">
                Last 7 Days
            </button>
            <button className="flex items-center gap-2 h-9 rounded-lg bg-foreground px-3.5 text-xs font-semibold text-background hover:bg-foreground/90 transition-all shadow-sm active:scale-[0.98]">
                Export CSV
            </button>
        </div>
      </div>

      {error && (
        <div className="flex items-center justify-between rounded-lg border border-destructive/20 bg-destructive/5 px-4 py-3 text-destructive shrink-0">
          <p className="text-sm font-medium">{error}</p>
          <button onClick={() => refetch()} className="rounded-md bg-destructive/10 px-3 py-1.5 text-xs font-semibold hover:bg-destructive/20 transition-colors">Retry</button>
        </div>
      )}

      {/* Top Metrics Row - Physical Cards */}
      <motion.div variants={containerVars} initial="hidden" animate="show" className="grid grid-cols-1 md:grid-cols-3 gap-6 shrink-0">
        {cards.map(({ id, icon: Icon, title, value, chartColor, sparkData, onClick }, idx) => {
          const trend = idx === 0 ? getTrend(todayPoint?.total, yesterdayPoint?.total) : 
                        idx === 1 ? getTrend(todayPoint?.answered, yesterdayPoint?.answered) : 
                        getTrend(todayPoint?.missed, yesterdayPoint?.missed);
          
          return (
          <motion.div key={id} variants={itemVars} className={idx === 0 ? "md:col-span-1" : ""}>
            <div 
              onClick={onClick}
              className="group flex flex-col justify-between overflow-hidden bg-card p-5 cursor-pointer card-solid transition-all min-h-[140px] hover:border-foreground/20 hover:shadow-lg"
            >
              <div className="flex items-center justify-between mb-4">
                <span className="text-xs font-semibold text-muted-foreground tracking-wide text-foreground/70">{title}</span>
                <Icon className="size-4 text-muted-foreground/50 transition-colors group-hover:text-foreground/80" />
              </div>
              
              <div className="flex items-end justify-between z-10">
                <div className="flex flex-col gap-1.5">
                  <span className="text-3xl font-bold tracking-tight text-foreground">{value}</span>
                  {!loading && (todayPoint || yesterdayPoint) && (
                     <div className="flex items-center gap-1.5 h-5">
                       <span className={cn("text-xs font-semibold flex items-center gap-0.5", trend.color)}>
                         <trend.icon className="size-3" />
                         {trend.text}
                       </span>
                       <span className="text-[11px] font-medium text-muted-foreground/70">from yesterday</span>
                     </div>
                  )}
                </div>
                
                {/* Embedded Mini Sparkline */}
                {!loading && (
                  <div className="h-10 w-24 opacity-60 group-hover:opacity-100 transition-opacity">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={sparkData}>
                        <defs>
                          <linearGradient id={`spark-${id}`} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" stopColor={chartColor} stopOpacity={0.2} />
                            <stop offset="100%" stopColor={chartColor} stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <Area type="monotone" dataKey="v" stroke={chartColor} strokeWidth={2} fill={`url(#spark-${id})`} isAnimationActive={false} />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                )}
              </div>
            </div>
          </motion.div>
        )})}
      </motion.div>

      {/* Main Analysis Area - Split Pane */}
      <motion.div variants={containerVars} initial="hidden" animate="show" className="grid grid-cols-1 xl:grid-cols-3 gap-6 flex-1 min-h-[450px]">
        
        {/* Main Chart Pane */}
        <motion.div variants={itemVars} className="xl:col-span-2 flex flex-col card-solid p-6">
          <div className="flex items-center justify-between mb-6">
            <h3 className="text-sm font-bold text-foreground tracking-tight">Call Volume Analysis</h3>
            <div className="flex items-center gap-4">
                <div className="flex items-center gap-2 text-xs font-semibold text-muted-foreground">
                    <div className="size-2 rounded-full" style={{ backgroundColor: CHART_COLORS.total }}></div>
                    Total
                </div>
                <div className="flex items-center gap-2 text-xs font-semibold text-foreground">
                    <div className="size-2 rounded-full" style={{ backgroundColor: CHART_COLORS.answered }}></div>
                    Answered
                </div>
            </div>
          </div>
          <div className="flex-1 w-full h-[300px] min-h-[300px] mt-4">
             {!loading && chartData.length > 0 ? (
                 <ResponsiveContainer width="100%" height="100%">
                     <AreaChart data={chartData} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
                     <defs>
                        <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor={CHART_COLORS.total} stopOpacity={0.2} />
                          <stop offset="95%" stopColor={CHART_COLORS.total} stopOpacity={0} />
                        </linearGradient>
                        <linearGradient id="colorAnswered" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="5%" stopColor={CHART_COLORS.answered} stopOpacity={0.25} />
                          <stop offset="95%" stopColor={CHART_COLORS.answered} stopOpacity={0} />
                        </linearGradient>
                     </defs>
                     <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="hsl(var(--border))" opacity={0.4} />
                     <XAxis 
                        dataKey="date" 
                        axisLine={false} 
                        tickLine={false} 
                        tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 11, fontWeight: 500 }}
                        dy={10}
                     />
                     <YAxis 
                        axisLine={false} 
                        tickLine={false} 
                        tick={{ fill: "hsl(var(--muted-foreground))", fontSize: 11, fontWeight: 500 }}
                        dx={-10}
                     />
                     <Tooltip
                         cursor={{ stroke: "hsl(var(--border))", strokeWidth: 1, strokeDasharray: "4 4", fill: "transparent" }}
                         contentStyle={{ backgroundColor: "hsl(var(--card))", borderColor: "hsl(var(--border))", borderRadius: "10px", fontSize: "12px", color: "hsl(var(--foreground))", padding: "12px", boxShadow: "0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)" }}
                         itemStyle={{ color: "hsl(var(--foreground))", fontSize: "13px", fontWeight: 600 }}
                         labelStyle={{ color: "hsl(var(--muted-foreground))", marginBottom: "6px", fontWeight: 500, fontSize: "11px", letterSpacing: "0.02em" }}
                     />
                     <Area type="monotone" dataKey="total" stroke={CHART_COLORS.total} strokeWidth={2.5} fillOpacity={1} fill="url(#colorTotal)" activeDot={{ r: 4, strokeWidth: 0, fill: CHART_COLORS.total }} />
                     <Area type="monotone" dataKey="answered" stroke={CHART_COLORS.answered} strokeWidth={2.5} fillOpacity={1} fill="url(#colorAnswered)" activeDot={{ r: 4, strokeWidth: 0, fill: CHART_COLORS.answered }} />
                     </AreaChart>
                 </ResponsiveContainer>
             ) : (
                 <div className="flex h-full items-center justify-center text-sm font-medium text-muted-foreground bg-transparent"></div>
             )}
          </div>
        </motion.div>

        {/* Recent Activity Log - Timeline Style */}
        <motion.div variants={itemVars} className="flex flex-col card-solid">
            <div className="flex items-center justify-between p-5 border-b border-border/40">
                <h3 className="text-sm font-bold text-foreground tracking-tight">Recent Activity</h3>
                <Link to="/calls" className="text-xs font-semibold text-muted-foreground hover:text-foreground transition-colors">
                    View All
                </Link>
            </div>
            
            <div className="flex-1 overflow-y-auto p-0 scrollbar-none">
                <div className="flex flex-col">
                    {callsLoading ? (
                        <div className="p-8 bg-transparent"></div>
                    ) : recentCalls.length === 0 ? (
                        <div className="p-12 text-center text-sm font-medium text-muted-foreground flex flex-col items-center gap-3">
                            <Phone className="size-4 text-muted-foreground/30" />
                            No recent calls
                        </div>
                    ) : (
                        recentCalls.map((call, idx) => {
                            const isIncoming = call.direction === 'call_direction_inbound';
                            const activeNumber = call.from_number === profile?.email ? call.to_number : call.from_number;
                            const statusConf = getStatusConfig(call.last_status);
                            
                            return (
                            <div 
                                key={call.id} 
                                onClick={() => navigate(`/calls/${call.id}`)}
                                className={cn(
                                    "group flex items-center gap-4 p-4 hover:bg-black/5 dark:hover:bg-white/5 transition-colors cursor-pointer",
                                    idx !== recentCalls.length - 1 && "border-b border-border/40"
                                )}
                            >
                                {/* Status Dot */}
                                <div className="flex flex-col items-center justify-center shrink-0 w-6">
                                    <div className={cn("size-2 rounded-full", statusConf.dotClass)} />
                                </div>
                                
                                <div className="flex flex-col flex-1 min-w-0">
                                    <div className="flex items-center justify-between gap-2 mb-1">
                                        <span className="text-sm font-semibold font-mono text-foreground truncate tracking-tight">{activeNumber}</span>
                                        <span className="text-[11px] font-medium text-muted-foreground shrink-0 tabular-nums">
                                            {call.created_at ? timeAgo(call.created_at) : "—"}
                                        </span>
                                    </div>
                                    <div className="flex items-center justify-between gap-2">
                                        <span className="text-[11px] font-medium text-muted-foreground flex items-center gap-1.5 truncate">
                                            {isIncoming ? (
                                                <ArrowDownLeft className="size-3 text-muted-foreground" />
                                            ) : (
                                                <ArrowUpRight className="size-3 text-muted-foreground" />
                                            )}
                                            {isIncoming ? "Inbound" : "Outbound"}
                                        </span>
                                        <span className={cn(
                                            "text-[11px] font-medium capitalize", 
                                            statusConf.colorClass
                                        )}>
                                            {statusConf.label}
                                        </span>
                                    </div>
                                </div>
                            </div>
                        )})
                    )}
                </div>
            </div>
        </motion.div>

      </motion.div>
    </div>
  );
}
