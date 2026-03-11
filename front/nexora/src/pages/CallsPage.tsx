import { useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ChevronDown, ChevronRight, Calendar, SlidersHorizontal, Search, ArrowUpRight, ArrowDownLeft } from "lucide-react";
import type {
  CallsListItem,
  CallTreeResponse,
  CallDirection,
} from "../api/calls";
import { getCallById } from "../api/calls";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { motion, AnimatePresence } from "framer-motion";
import { useCallsList } from "../hooks/useCallsList";
import { useTelephony } from "../hooks/useTelephony";
import { cn } from "@/lib/utils";

const PAGE_LIMIT = 20;

const DIRECTION_LABELS: Record<CallDirection, string> = {
  call_direction_inbound: "Входящий",
  call_direction_outbound_api: "Исходящий",
  call_direction_outbound_dial: "Исходящий",
};

const getStatusConfig = (status: string | undefined) => {
    switch (status) {
      case "call_event_status_completed":
        return { label: "Завершён", colorClass: "text-emerald-700 dark:text-emerald-400 bg-emerald-500/15 border-emerald-500/20", dotClass: "bg-emerald-500" };
      case "call_event_status_no_answer":
        return { label: "Нет ответа", colorClass: "text-red-700 dark:text-red-400 bg-red-500/15 border-red-500/20", dotClass: "bg-red-500" };
      case "call_event_status_canceled":
        return { label: "Отменён", colorClass: "text-slate-600 dark:text-slate-400 bg-slate-500/15 border-slate-500/20", dotClass: "bg-slate-500" };
      case "call_event_status_busy":
        return { label: "Занято", colorClass: "text-amber-700 dark:text-amber-400 bg-amber-500/15 border-amber-500/20", dotClass: "bg-amber-500" };
      case "call_event_status_failed":
        return { label: "Ошибка", colorClass: "text-red-700 dark:text-red-400 bg-red-500/15 border-red-500/20", dotClass: "bg-red-700" };
      case "call_event_status_in_progress":
        return { label: "В разговоре", colorClass: "text-blue-700 dark:text-blue-400 bg-blue-500/15 border-blue-500/20", dotClass: "bg-blue-500" };
      case "call_event_status_ringing":
        return { label: "Звонит", colorClass: "text-blue-600 dark:text-blue-400 bg-blue-500/10 border-blue-500/20", dotClass: "bg-blue-400 animate-pulse" };
      case "call_event_status_queued":
        return { label: "В очереди", colorClass: "text-amber-700 dark:text-amber-400 bg-amber-500/15 border-amber-500/20", dotClass: "bg-amber-400" };
      case "call_event_status_timeout":
        return { label: "Таймаут", colorClass: "text-red-700 dark:text-red-400 bg-red-500/15 border-red-500/20", dotClass: "bg-red-500" };
      case "call_event_status_initiated":
        return { label: "Инициирован", colorClass: "text-slate-700 dark:text-slate-300 bg-slate-500/15 border-slate-500/20", dotClass: "bg-slate-500" };
      default:
        const clean = status ? status.replace("call_event_status_", "").replace("_", " ") : "Неизвестно";
        return { label: clean, colorClass: "text-slate-600 dark:text-slate-400 bg-slate-500/10 border-slate-500/20", dotClass: "bg-slate-500" };
    }
};

interface CallsPageLocationState {
  initialStatus?: string;
  initialDirection?: CallDirection | "all";
}

export default function CallsPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const locationState = (location.state as CallsPageLocationState | null) ?? null;

  const { data: telephonyData } = useTelephony();
  const telephonyList = telephonyData ?? [];

  const initialDateFrom = useMemo(() => {
    const d = new Date();
    d.setDate(d.getDate() - 6);
    return d.toISOString().slice(0, 10);
  }, []);
  const initialDateTo = useMemo(() => new Date().toISOString().slice(0, 10), []);

  const { data, loading, error, params, updateParams, refetch } = useCallsList({
    page: { limit: PAGE_LIMIT, offset: 0, page: 1 },
    direction: locationState?.initialDirection === "all" ? undefined : locationState?.initialDirection,
    status: locationState?.initialStatus === "all" ? undefined : locationState?.initialStatus,
    date_from: initialDateFrom,
    date_to: initialDateTo,
  });

  const [periodOpen, setPeriodOpen] = useState(false);
  const [filtersOpen, setFiltersOpen] = useState(false);
  const periodRef = useRef<HTMLDivElement>(null);
  const filtersRef = useRef<HTMLDivElement>(null);

  const periodLabel = useMemo(() => {
    const today = new Date().toISOString().slice(0, 10);
    const weekStart = new Date();
    weekStart.setDate(weekStart.getDate() - 6);
    const weekStartStr = weekStart.toISOString().slice(0, 10);
    const monthStart = new Date();
    monthStart.setDate(monthStart.getDate() - 29);
    const monthStartStr = monthStart.toISOString().slice(0, 10);
    
    const { date_from, date_to } = params;
    
    if (date_from === today && date_to === today) return "Today";
    if (date_from === weekStartStr && date_to === today) return "Last 7 days";
    if (date_from === monthStartStr && date_to === today) return "Last 30 days";
    const from = date_from ? new Date(date_from).toLocaleDateString("en-US", { day: "2-digit", month: "short" }) : "";
    const to = date_to ? new Date(date_to).toLocaleDateString("en-US", { day: "2-digit", month: "short" }) : "";
    return from && to ? `${from} — ${to}` : "Period";
  }, [params.date_from, params.date_to]);

  useEffect(() => {
    const close = (e: MouseEvent) => {
      if (periodRef.current && !periodRef.current.contains(e.target as Node)) setPeriodOpen(false);
      if (filtersRef.current && !filtersRef.current.contains(e.target as Node)) setFiltersOpen(false);
    };
    document.addEventListener("click", close);
    return () => document.removeEventListener("click", close);
  }, []);
  
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [treeByCallId, setTreeByCallId] = useState<Record<string, CallTreeResponse>>({});
  const [loadingTreeId, setLoadingTreeId] = useState<string | null>(null);

  const toggleExpand = (id: string) => {
    if (treeByCallId[id]) {
      setExpandedIds((prev) => {
        const next = new Set(prev);
        if (next.has(id)) next.delete(id);
        else next.add(id);
        return next;
      });
      return;
    }
    setLoadingTreeId(id);
    getCallById(id)
      .then((tree) => {
        setTreeByCallId((prev) => {
          const next = { ...prev };
          const store = (t: CallTreeResponse) => {
            if (!t.call?.id) return;
            if (next[t.call.id] != null) return;
            next[t.call.id] = t;
            t.children?.forEach(store);
          };
          store(tree);
          return next;
        });
        setExpandedIds((prev) => new Set(prev).add(id));
      })
      .finally(() => setLoadingTreeId(null));
  };

  const callRows = useMemo(() => {
    if (!data?.items.length) return [];
    const rows: { id: string; depth: number; listItem?: CallsListItem; treeNode?: CallTreeResponse }[] = [];
    const pushChildren = (children: CallTreeResponse[], depth: number, ancestorIds: Set<string>, seenInBranch: Set<string>) => {
      for (const ch of children) {
        const id = ch.call?.id as string;
        if (!id) continue;
        if (ancestorIds.has(id) || seenInBranch.has(id)) continue;
        seenInBranch.add(id);
        rows.push({ id, depth, treeNode: ch });
        if (expandedIds.has(id) && treeByCallId[id]?.children?.length) {
          pushChildren(treeByCallId[id].children!, depth + 1, new Set([...ancestorIds, id]), seenInBranch);
        }
      }
    };
    for (const item of data.items) {
      const itemId = item.id as string;
      if (!itemId) continue;
      const seenInBranch = new Set<string>();
      rows.push({ id: itemId, depth: 0, listItem: item });
      if (expandedIds.has(itemId) && treeByCallId[itemId]?.children?.length) {
        pushChildren(treeByCallId[itemId].children!, 1, new Set([itemId]), seenInBranch);
      }
    }
    return rows;
  }, [data?.items, expandedIds, treeByCallId]);

  const telephonyById = useMemo(() => {
    const map: Record<string, string> = {};
    for (const t of telephonyList) {
      if (t.id) {
        map[t.id] = t.telephony_name ?? t.id;
      }
    }
    return map;
  }, [telephonyList]);

  const totalPages = useMemo(() => {
    if (!data) return 0;
    return Math.max(1, Math.ceil(data.meta.total / PAGE_LIMIT));
  }, [data]);

  const currentPage = params.page?.page ?? 1;

  return (
    <div className="flex-1 flex flex-col h-full font-sans">
      
      {/* Page Header & Actions Row */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-3 mb-3 border-b border-border shrink-0">
        <div className="flex flex-col mb-3 sm:mb-0">
          <h1 className="text-xl font-bold tracking-tight text-foreground">Calls Log</h1>
          <span className="text-xs text-muted-foreground">Monitor and analyze communication history.</span>
        </div>

        <div className="flex items-center gap-2">
        <div className="flex flex-wrap items-center gap-3">
          <div className="relative" ref={periodRef}>
            <Button
                variant="outline"
                size="sm"
                className="h-9 bg-secondary/30 text-xs font-semibold border-border/60 hover:bg-secondary/60 hover:text-foreground transition-all shadow-sm"
                onClick={(e: React.MouseEvent<HTMLButtonElement>) => {
                    e.stopPropagation();
                    setPeriodOpen((v) => !v);
                    setFiltersOpen(false);
                }}
            >
                <Calendar className="mr-2 size-3 text-muted-foreground" />
                {periodLabel}
                <ChevronDown className={`ml-2 size-3 text-muted-foreground transition ${periodOpen ? "rotate-180" : ""}`} />
            </Button>
            
            <AnimatePresence>
              {periodOpen && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ duration: 0.15 }}
                  className="absolute left-0 top-full z-50 mt-2 min-w-[300px] rounded-xl border border-border bg-card p-4 shadow-premium"
                >
                  <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Quick Select</p>
                  <div className="flex gap-2 mb-4">
                    {[
                      { label: "Today", days: 0 },
                      { label: "7 days", days: 6 },
                      { label: "30 days", days: 29 },
                    ].map(({ label, days }) => {
                      const to = new Date();
                      const from = new Date();
                      from.setDate(from.getDate() - days);
                      const fromStr = from.toISOString().slice(0, 10);
                      const toStr = to.toISOString().slice(0, 10);
                      const isActive = params.date_from === fromStr && params.date_to === toStr;
                      return (
                        <Button
                          key={label}
                          variant={isActive ? "default" : "secondary"}
                          size="sm"
                          className="flex-1"
                          onClick={() => {
                            updateParams({ date_from: fromStr, date_to: toStr });
                          }}
                        >
                          {label}
                        </Button>
                      );
                    })}
                  </div>
                  <p className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted-foreground">Custom Range</p>
                  <div className="flex items-center gap-2">
                    <Input
                      type="date"
                      value={params.date_from ?? ""}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateParams({ date_from: e.target.value })}
                      className="flex-1 h-9"
                    />
                    <span className="text-muted-foreground">—</span>
                    <Input
                      type="date"
                      value={params.date_to ?? ""}
                      onChange={(e: React.ChangeEvent<HTMLInputElement>) => updateParams({ date_to: e.target.value })}
                      className="flex-1 h-9"
                    />
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          <div className="relative" ref={filtersRef}>
            <Button
                variant="outline"
                size="sm"
                className="h-9 bg-secondary/30 text-xs font-semibold border-border/60 hover:bg-secondary/60 hover:text-foreground transition-all shadow-sm"
                onClick={(e: React.MouseEvent<HTMLButtonElement>) => {
                    e.stopPropagation();
                    setFiltersOpen((v) => !v);
                    setPeriodOpen(false);
                }}
            >
                <SlidersHorizontal className="mr-2 size-3 text-muted-foreground" />
                Filters
                {(params.direction || params.status || params.company_telephony_id) && (
                    <span className="ml-2 flex h-4 min-w-[16px] items-center justify-center rounded-full bg-primary px-1 text-[9px] font-bold text-primary-foreground">
                        {[params.direction, params.status, params.company_telephony_id].filter(Boolean).length}
                    </span>
                )}
                <ChevronDown className={`ml-2 size-3 text-muted-foreground transition ${filtersOpen ? "rotate-180" : ""}`} />
            </Button>
            
            <AnimatePresence>
              {filtersOpen && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, scale: 0.95 }}
                  transition={{ duration: 0.15 }}
                  className="absolute left-0 top-full z-50 mt-2 min-w-[320px] rounded-xl border border-border bg-card p-4 shadow-premium"
                >
                  <div className="space-y-4">
                    <div>
                      <label className="mb-2 block text-xs font-semibold uppercase tracking-wider text-muted-foreground">Telephony</label>
                      <select
                        value={params.company_telephony_id ?? "all"}
                        onChange={(e: React.ChangeEvent<HTMLSelectElement>) => updateParams({ company_telephony_id: e.target.value === "all" ? undefined : e.target.value })}
                        className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      >
                        <option value="all">All</option>
                        {telephonyList.map((t) => (
                          <option key={t.id} value={t.id ?? ""}>
                            {t.telephony_name ?? t.id ?? "—"}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label className="mb-2 block text-xs font-semibold uppercase tracking-wider text-muted-foreground">Direction</label>
                      <select
                        value={params.direction ?? "all"}
                        onChange={(e: React.ChangeEvent<HTMLSelectElement>) => updateParams({ direction: e.target.value === "all" ? undefined : e.target.value as CallDirection })}
                        className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      >
                        <option value="all">All</option>
                        <option value="call_direction_inbound">Inbound</option>
                        <option value="call_direction_outbound_api">Outbound (API)</option>
                        <option value="call_direction_outbound_dial">Outbound (Dial)</option>
                      </select>
                    </div>
                    <div>
                      <label className="mb-2 block text-xs font-semibold uppercase tracking-wider text-muted-foreground">Status</label>
                      <select
                        value={params.status ?? "all"}
                        onChange={(e: React.ChangeEvent<HTMLSelectElement>) => updateParams({ status: e.target.value === "all" ? undefined : e.target.value })}
                        className="w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                      >
                        <option value="all">All</option>
                        <option value="call_event_status_completed">Completed</option>
                        <option value="call_event_status_no_answer">No Answer</option>
                        <option value="call_event_status_busy">Busy</option>
                        <option value="call_event_status_canceled">Canceled</option>
                        <option value="call_event_status_failed">Failed</option>
                      </select>
                    </div>
                  </div>
                </motion.div>
              )}
            </AnimatePresence>
          </div>
        </div>

        {(params.direction || params.status || params.company_telephony_id) && (
          <Button
            variant="ghost"
            className="text-muted-foreground hover:text-foreground hover:bg-muted"
            onClick={() => {
                updateParams({
                    direction: undefined,
                    status: undefined,
                    company_telephony_id: undefined,
                    date_from: initialDateFrom,
                    date_to: initialDateTo,
                    page: { limit: PAGE_LIMIT, offset: 0, page: 1 }
                });
            }}
          >
            Clear Filters
          </Button>
        )}
      </div>
      </div>

      <div className="flex-1 flex flex-col min-h-0 bg-transparent">
        {loading && (
          <div className="flex-1 bg-transparent">
          </div>
        )}
        
        {error && (
          <div className="p-4 rounded-xl mb-4 flex flex-col sm:flex-row items-center justify-between gap-4 bg-destructive/10 text-destructive text-sm border border-destructive/20 shadow-sm">
            <span className="font-medium">{error}</span>
            <Button variant="destructive" size="sm" className="h-8" onClick={() => refetch()}>Retry</Button>
          </div>
        )}
        
        {!loading && data && data.items.length === 0 && (
          <div className="flex-1 flex flex-col items-center justify-center text-center p-12 bg-card rounded-2xl border border-border/40 shadow-sm">
            <div className="flex size-14 items-center justify-center rounded-full bg-secondary text-muted-foreground/60 mb-4 shadow-sm border border-border/40">
               <Search className="size-6" />
            </div>
            <h3 className="text-base font-bold text-foreground mb-1">No calls found</h3>
            <p className="text-sm font-medium text-muted-foreground">Try adjusting your filters or date range.</p>
          </div>
        )}
        
        {!loading && data && data.items.length > 0 && (
          <div className="flex-1 overflow-auto rounded-xl">
             <div className="flex flex-col gap-2 min-w-[800px] pb-4">
                
                {/* List Header */}
                <div className="sticky top-0 z-10 flex bg-background/95 backdrop-blur-sm px-6 py-3 border-b border-border/40 text-xs font-bold uppercase tracking-widest text-muted-foreground">
                   <div className="w-[140px] shrink-0">Статус</div>
                   <div className="w-[180px] shrink-0">Контакт</div>
                   <div className="w-[140px] shrink-0">Направление</div>
                   <div className="flex-1">Дата и Время</div>
                   <div className="w-[120px] shrink-0 text-right">Провайдер</div>
                   <div className="w-[60px] shrink-0"></div>
                </div>

                {/* Rows Space */}
                <div className="flex flex-col gap-2 px-1">
                    {callRows.map((row) => {
                    const c = row.listItem ?? row.treeNode!.call;
                    if (!c) return null;
                    const created = c.created_at ? new Date(c.created_at) : null;
                    const lastStatus = row.listItem
                        ? row.listItem.last_status
                        : row.treeNode!.call?.events?.length
                        ? row.treeNode!.call.events[row.treeNode!.call.events.length - 1].status
                        : undefined;
                    
                    const isIncoming = c.direction === 'call_direction_inbound';
                    const activeNumber = isIncoming ? c.from_number : c.to_number;
                    const secondaryNumber = isIncoming ? c.to_number : c.from_number;
                    const statusConf = getStatusConfig(lastStatus);

                    const hasChildren =
                        row.depth === 0
                        ? Boolean(row.listItem?.has_children)
                        : (row.treeNode?.children?.length ?? 0) > 0;
                    const isExpanded = expandedIds.has(row.id);
                    const isTreeLoading = loadingTreeId === row.id;
                    const indentPx = row.depth > 0 ? 32 + (row.depth - 1) * 24 : 0;
                    const isRoot = row.depth === 0;

                    return (
                        <div
                           key={row.id}
                           onClick={() => navigate(`/calls/${row.id}`)}
                           className={cn(
                               "group relative flex items-center bg-card rounded-2xl px-6 py-4 cursor-pointer transition-all border shadow-sm",
                               isRoot ? "border-border/60 hover:border-foreground/20 hover:shadow-md" : "border-border/30 bg-secondary/10 hover:bg-secondary/30",
                               row.depth > 0 && "mt-0"
                           )}
                           style={{ marginLeft: indentPx }}
                        >
                            {/* Connector Line for Subcalls */}
                            {row.depth > 0 && (
                                <div className="absolute -left-6 top-1/2 w-4 border-t-2 border-border/40 rounded-tl-lg" />
                            )}
                            {row.depth > 0 && row.id !== callRows[callRows.length-1]?.id && (
                                <div className="absolute -left-6 -top-6 bottom-1/2 border-l-2 border-border/40" />
                            )}

                            {/* Status Indicator */}
                            <div className="w-[140px] shrink-0 flex items-center pr-4">
                                <div className={cn(
                                   "flex items-center justify-center rounded-xl font-bold uppercase tracking-wider text-[10px] px-2.5 py-1 border whitespace-nowrap",
                                   statusConf.colorClass
                                )}>
                                    <div className={cn("size-1.5 rounded-full mr-1.5 shrink-0", statusConf.dotClass)} />
                                    {statusConf.label}
                                </div>
                            </div>

                            {/* Contact Info */}
                            <div className="w-[180px] shrink-0 flex flex-col justify-center">
                                <span className={cn("text-[15px] font-bold font-mono tracking-tight", isRoot ? "text-foreground" : "text-muted-foreground")}>
                                    {activeNumber || "—"}
                                </span>
                                {isRoot && secondaryNumber && (
                                   <span className="text-xs font-medium text-muted-foreground mt-0.5 max-w-[150px] truncate">
                                      {isIncoming ? "to" : "from"} <span className="font-mono text-[11px]">{secondaryNumber}</span>
                                   </span>
                                )}
                            </div>

                            {/* Direction */}
                            <div className="w-[140px] shrink-0 flex items-center pr-4">
                                <span className={cn(
                                    "flex items-center gap-1.5 text-xs font-semibold px-2.5 py-1 rounded-lg border whitespace-nowrap",
                                    isIncoming ? "bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/20" : "bg-violet-500/10 text-violet-700 border-violet-500/20"
                                )}>
                                    {isIncoming ? <ArrowDownLeft className="size-3.5" /> : <ArrowUpRight className="size-3.5" />}
                                    {DIRECTION_LABELS[c.direction as CallDirection] || "Неизвестно"}
                                </span>
                            </div>

                            {/* Date & Time */}
                            <div className="flex-1 flex flex-col justify-center">
                                {created ? (
                                    <>
                                       <span className="text-sm font-semibold text-foreground/90">
                                            {created.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
                                       </span>
                                       <span className="text-xs font-medium text-muted-foreground mt-0.5">
                                            {created.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" })}
                                       </span>
                                    </>
                                ) : <span className="text-muted-foreground">—</span>}
                            </div>

                            {/* Telephony */}
                            <div className="w-[120px] shrink-0 flex flex-col justify-center items-end text-right px-4">
                                {isRoot && c.company_telephony_id ? (
                                    <>
                                        <span className="text-sm font-semibold text-foreground truncate w-full">
                                            {telephonyById[c.company_telephony_id as string] ?? "SIP Trunk"}
                                        </span>
                                        <span className="text-[10px] font-bold uppercase tracking-wider text-muted-foreground mt-0.5">Provider</span>
                                    </>
                                ) : <span className="text-muted-foreground">—</span>}
                            </div>

                            {/* Expand / Actions */}
                            <div className="w-[60px] shrink-0 flex items-center justify-end">
                               {hasChildren ? (
                                    <button
                                        type="button"
                                        className={cn(
                                            "flex size-8 items-center justify-center rounded-full transition-all border shadow-sm",
                                            isExpanded ? "bg-primary text-primary-foreground border-transparent" : "bg-background text-foreground border-border/80 hover:bg-secondary"
                                        )}
                                        disabled={isTreeLoading}
                                        aria-label={isExpanded ? "Collapse" : "Expand"}
                                        onClick={(e: React.MouseEvent<HTMLButtonElement>) => {
                                            e.stopPropagation();
                                            if (row.depth === 0 || treeByCallId[row.id] != null) {
                                                toggleExpand(row.id);
                                            } else if (row.treeNode?.children?.length) {
                                                toggleExpand(row.id);
                                            }
                                        }}
                                    >
                                        {isTreeLoading ? (
                                            <div className="size-3 animate-pulse bg-current rounded-full" />
                                        ) : (
                                            <ChevronRight className={cn("size-4 transition-transform", isExpanded && "rotate-90")} />
                                        )}
                                    </button>
                               ) : (
                                   <div className="size-8 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                                       <ChevronRight className="size-4 text-muted-foreground" />
                                   </div>
                               )}
                            </div>
                        </div>
                    );
                    })}
                </div>
             </div>
          </div>
        )}
      </div>

      {/* Pagination - Slim footer */}
      {data && data.meta.total > PAGE_LIMIT && (
        <div className="flex items-center justify-between text-xs font-medium text-muted-foreground px-2 shrink-0 border-t border-border/40 pt-4 mt-2">
          <span>
            Page <strong className="text-foreground">{currentPage}</strong> of <strong className="text-foreground">{totalPages}</strong>
          </span>
          <div className="flex gap-1.5">
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-3 text-[10px]"
              disabled={currentPage === 1}
              onClick={() => updateParams({ page: { limit: PAGE_LIMIT, offset: (currentPage - 2) * PAGE_LIMIT, page: currentPage - 1 } })}
            >
              Prev
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-3 text-[10px]"
              disabled={currentPage >= totalPages}
              onClick={() => updateParams({ page: { limit: PAGE_LIMIT, offset: currentPage * PAGE_LIMIT, page: currentPage + 1 } })}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}

