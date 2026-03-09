import { useMemo, useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Phone, PhoneCall, PhoneMissed, ArrowUpRight, ArrowDownRight, Minus } from "lucide-react";
import { AreaChart, Area, BarChart, Bar, Tooltip, ResponsiveContainer } from "recharts";
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

  const cards = [
    {
      id: "all",
      icon: Phone,
      title: "Total Calls",
      value: loading ? "—" : String(metrics?.summary?.total ?? 0),
      chartColor: CHART_COLORS.total,
      sparkData: sparkDataTotal,
      colSpan: "col-span-1 md:col-span-2 lg:col-span-1", // Bento size
      onClick: () => navigate("/calls"),
    },
    {
      id: "answered",
      icon: PhoneCall,
      title: "Answered",
      value: loading ? "—" : String(metrics?.summary?.answered ?? 0),
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
      value: loading ? "—" : String(metrics?.summary?.missed ?? 0),
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

  const getTrend = (current: number = 0, previous: number = 0) => {
    if (current > previous) return { icon: ArrowUpRight, color: "text-emerald-500", bg: "bg-emerald-500/10", text: `+${current - previous}` };
    if (current < previous) return { icon: ArrowDownRight, color: "text-red-500", bg: "bg-red-500/10", text: `${current - previous}` };
    return { icon: Minus, color: "text-muted-foreground", bg: "bg-secondary", text: "0" };
  };

  return (
    <div className="flex flex-col h-full font-sans">
      
      {/* Page Header */}
      <div className="flex items-center justify-between pb-4 mb-4 border-b border-border shrink-0">
        <div className="flex flex-col">
           <h1 className="text-xl font-bold tracking-tight text-foreground">Analytics Overview</h1>
           <span className="text-xs text-muted-foreground">Performance metrics & activity for the last 7 days.</span>
        </div>
        <div className="flex gap-2">
            <button className="h-8 px-3 text-xs font-medium bg-secondary text-secondary-foreground hover:bg-secondary/80 rounded-md border border-border transition-colors">Last 7 Days</button>
            <button className="h-8 px-3 text-xs font-medium bg-primary text-primary-foreground hover:bg-primary/90 rounded-md shadow-sm transition-colors">Export CSV</button>
        </div>
      </div>

      {error && (
        <div className="mb-4 flex items-center justify-between rounded-md border border-destructive/50 bg-destructive/10 px-3 py-2 text-destructive shrink-0">
          <p className="text-xs font-medium">{error}</p>
          <button onClick={() => refetch()} className="rounded-md bg-destructive/20 px-2 py-1 text-[10px] font-semibold hover:bg-destructive/30 transition-colors">Retry</button>
        </div>
      )}

      {/* Top Metrics Row */}
      <motion.div variants={containerVars} initial="hidden" animate="show" className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4 shrink-0">
        {cards.map(({ id, icon: Icon, title, value, chartColor, sparkData, onClick }, idx) => {
          const trend = idx === 0 ? getTrend(todayPoint?.total, yesterdayPoint?.total) : 
                        idx === 1 ? getTrend(todayPoint?.answered, yesterdayPoint?.answered) : 
                        getTrend(todayPoint?.missed, yesterdayPoint?.missed);
          
          return (
          <motion.div key={id} variants={itemVars}>
            <div 
              onClick={onClick}
              className="group relative flex flex-col overflow-hidden rounded-md border border-border bg-card p-4 transition-all hover:bg-accent/50 cursor-pointer"
            >
              <div className="flex items-center justify-between mb-2">
                <span className="text-xs font-medium text-muted-foreground uppercase tracking-wider">{title}</span>
                <Icon className="size-4 text-muted-foreground opacity-50" />
              </div>
              <div className="flex items-end justify-between">
                <div className="flex flex-col">
                  <span className="text-2xl font-mono font-bold tracking-tight text-foreground">{value}</span>
                  {!loading && (todayPoint || yesterdayPoint) && (
                    <div className="mt-1 flex items-center gap-1.5">
                      <div className={`flex items-center gap-0.5 rounded px-1 py-0.5 text-[10px] font-semibold ${trend.bg} ${trend.color}`}>
                        <trend.icon className="size-3" />
                        <span>{trend.text}</span>
                      </div>
                      <span className="text-[10px] text-muted-foreground">vs yesterday</span>
                    </div>
                  )}
                </div>
                
                {/* Mini Sparkline embedded in card */}
                {!loading && (
                  <div className="h-8 w-24">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart data={sparkData}>
                        <defs>
                          <linearGradient id={`spark-${id}`} x1="0" y1="0" x2="0" y2="1">
                            <stop offset="0%" stopColor={chartColor} stopOpacity={0.3} />
                            <stop offset="100%" stopColor={chartColor} stopOpacity={0} />
                          </linearGradient>
                        </defs>
                        <Area type="monotone" dataKey="v" stroke={chartColor} strokeWidth={1.5} fill={`url(#spark-${id})`} isAnimationActive={false} />
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
      <motion.div variants={containerVars} initial="hidden" animate="show" className="grid grid-cols-1 lg:grid-cols-3 gap-4 flex-1 min-h-[400px]">
        
        {/* Main Chart Pane */}
        <motion.div variants={itemVars} className="lg:col-span-2 flex flex-col rounded-md border border-border bg-card overflow-hidden">
          <div className="flex items-center justify-between border-b border-border px-4 py-3 bg-secondary/20">
            <h3 className="text-sm font-semibold text-foreground">Call Volume Analysis</h3>
            <span className="text-[10px] font-mono text-muted-foreground bg-secondary px-2 py-0.5 rounded border border-border">Total / Answered</span>
          </div>
          <div className="flex-1 p-4">
            {!loading && chartData.length > 0 ? (
                <ResponsiveContainer width="100%" height="100%">
                    <BarChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }} barCategoryGap="25%">
                    <Tooltip
                        cursor={{ fill: "var(--color-muted)", opacity: 0.3 }}
                        contentStyle={{ backgroundColor: "var(--color-card)", borderColor: "var(--color-border)", borderRadius: "6px", fontSize: "11px", color: "var(--color-foreground)", padding: "8px" }}
                        itemStyle={{ color: "var(--color-foreground)", fontSize: "12px", fontWeight: 600 }}
                        labelStyle={{ color: "var(--color-muted-foreground)", marginBottom: "4px" }}
                    />
                    <Bar dataKey="total" fill={CHART_COLORS.total} opacity={0.3} radius={[2, 2, 0, 0]} />
                    <Bar dataKey="answered" fill={CHART_COLORS.answered} radius={[2, 2, 0, 0]} />
                    </BarChart>
                </ResponsiveContainer>
            ) : (
                <div className="flex h-full items-center justify-center text-xs text-muted-foreground">Loading chart data...</div>
            )}
          </div>
        </motion.div>

        {/* Recent Activity Log */}
        <motion.div variants={itemVars} className="flex flex-col rounded-md border border-border bg-card overflow-hidden">
            <div className="flex items-center justify-between border-b border-border px-4 py-3 bg-secondary/20">
                <h3 className="text-sm font-semibold text-foreground">Recent Activity</h3>
                <span className="relative flex size-2">
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
                    <span className="relative inline-flex rounded-full size-2 bg-emerald-500"></span>
                </span>
            </div>
            <div className="flex-1 overflow-y-auto p-0">
                <div className="divide-y divide-border">
                    {callsLoading ? (
                        <div className="p-8 text-center text-xs text-muted-foreground">Loading activity...</div>
                    ) : recentCalls.length === 0 ? (
                        <div className="p-8 text-center text-xs text-muted-foreground">No recent calls</div>
                    ) : (
                        recentCalls.map((call) => (
                            <div 
                                key={call.id} 
                                onClick={() => navigate(`/calls/${call.id}`)}
                                className="flex flex-col gap-1 p-3 hover:bg-secondary/30 transition-colors cursor-pointer"
                            >
                                <div className="flex items-center justify-between">
                                    <span className="text-[11px] font-mono font-medium text-foreground">
                                        {call.from_number === profile?.email ? call.to_number : call.from_number}
                                    </span>
                                    <span className="text-[10px] text-muted-foreground">
                                        {call.created_at ? timeAgo(call.created_at) : "—"}
                                    </span>
                                </div>
                                <div className="flex items-center justify-between">
                                    <span className="text-[11px] text-muted-foreground truncate max-w-[150px]">
                                        {call.direction === 'call_direction_inbound' ? "Inbound from Customer" : "Outbound to Lead"}
                                    </span>
                                    <span className={cn(
                                        "text-[10px] font-semibold px-1.5 py-0.5 rounded", 
                                        call.last_status === 'call_event_status_completed' ? "bg-emerald-500/10 text-emerald-500" : "bg-red-500/10 text-red-500"
                                    )}>
                                        {call.last_status === 'call_event_status_completed' ? "Completed" : "Missed"}
                                    </span>
                                </div>
                            </div>
                        ))
                    )}
                </div>
            </div>
            {recentCalls.length > 0 && (
                <div className="border-t border-border p-2">
                    <button 
                        onClick={() => navigate("/calls")}
                        className="w-full rounded-md py-1.5 text-center text-[11px] font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
                    >
                        View All Activity
                    </button>
                </div>
            )}
        </motion.div>

      </motion.div>
    </div>
  );
}
