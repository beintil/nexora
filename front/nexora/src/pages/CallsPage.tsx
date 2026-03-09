import { useEffect, useMemo, useRef, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { ChevronDown, ChevronRight, Calendar, SlidersHorizontal, Search } from "lucide-react";
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

const PAGE_LIMIT = 20;

const STATUS_LABELS: Record<string, string> = {
  "": "—",
  call_event_status_queued: "В очереди",
  call_event_status_initiated: "Инициирован",
  call_event_status_ringing: "Звонит",
  call_event_status_in_progress: "В разговоре",
  call_event_status_completed: "Завершён",
  call_event_status_busy: "Занято",
  call_event_status_failed: "Ошибка",
  call_event_status_no_answer: "Нет ответа",
  call_event_status_canceled: "Отменён",
  call_event_status_timeout: "Таймаут",
};

const DIRECTION_LABELS: Record<CallDirection, string> = {
  call_direction_inbound: "Входящий",
  call_direction_outbound_api: "Исходящий",
  call_direction_outbound_dial: "Исходящий",
};

const DIRECTION_ROW_ACCENT: Record<string, string> = {
  call_direction_inbound: "border-l-emerald-500",
  call_direction_outbound_api: "border-l-blue-500",
  call_direction_outbound_dial: "border-l-violet-500",
};

const DIRECTION_BADGE_CLASS: Record<string, string> = {
  call_direction_inbound: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20",
  call_direction_outbound_api: "bg-blue-500/15 text-blue-700 dark:text-blue-400 border border-blue-500/20",
  call_direction_outbound_dial: "bg-violet-500/15 text-violet-700 dark:text-violet-400 border border-violet-500/20",
};

const STATUS_BADGE_CLASS: Record<string, string> = {
  call_event_status_completed: "bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/20",
  call_event_status_in_progress: "bg-blue-500/15 text-blue-700 dark:text-blue-400 border-blue-500/20",
  call_event_status_ringing: "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/20",
  call_event_status_initiated: "bg-slate-500/15 text-slate-700 dark:text-slate-300 border-slate-500/20",
  call_event_status_no_answer: "bg-orange-500/15 text-orange-700 dark:text-orange-400 border-orange-500/20",
  call_event_status_busy: "bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/20",
  call_event_status_failed: "bg-red-500/15 text-red-700 dark:text-red-400 border-red-500/20",
  call_event_status_canceled: "bg-slate-500/15 text-slate-600 dark:text-slate-400 border-slate-500/20",
  call_event_status_timeout: "bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/20",
  call_event_status_queued: "bg-slate-500/15 text-slate-600 dark:text-slate-400 border-slate-500/20",
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
                className="h-8 bg-secondary/30 text-xs font-medium border-border"
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
                className="h-8 bg-secondary/30 text-xs font-medium border-border"
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

      <div className="flex-1 flex flex-col border border-border rounded-md bg-card overflow-hidden min-h-0">
        {loading && (
          <div className="flex-1 flex flex-col items-center justify-center text-muted-foreground">
            <div className="size-6 rounded-full border-2 border-primary border-r-transparent animate-spin mb-3" />
            <p className="text-xs">Loading calls...</p>
          </div>
        )}
        
        {error && (
          <div className="p-4 flex flex-col sm:flex-row items-center justify-between gap-4 bg-destructive/10 text-destructive text-sm">
            <span className="font-medium">{error}</span>
            <Button variant="destructive" size="sm" className="h-8" onClick={() => refetch()}>Retry</Button>
          </div>
        )}
        
        {!loading && data && data.items.length === 0 && (
          <div className="flex-1 flex flex-col items-center justify-center text-center p-12">
            <div className="flex size-12 items-center justify-center rounded-full bg-secondary text-muted-foreground mb-3">
               <Search className="size-5" />
            </div>
            <h3 className="text-sm font-semibold text-foreground mb-1">No calls found</h3>
            <p className="text-[11px] text-muted-foreground">Try adjusting your filters or date range.</p>
          </div>
        )}
        
        {!loading && data && data.items.length > 0 && (
          <div className="flex-1 overflow-auto">
            <table className="w-full min-w-[800px] text-[11px] text-left align-middle border-collapse">
              <thead className="sticky top-0 z-10 bg-secondary/80 backdrop-blur-sm shadow-sm ring-1 ring-border/50">
                <tr className="text-xs text-muted-foreground font-semibold uppercase tracking-wider">
                  <th className="py-2 pl-4 pr-3 w-[260px] font-medium">Call ID</th>
                  <th className="py-2 px-3 font-medium">Date & Time</th>
                  <th className="py-2 px-3 font-medium">Direction</th>
                  <th className="py-2 px-3 font-medium">Telephony</th>
                  <th className="py-2 px-3 font-medium">From</th>
                  <th className="py-2 px-3 font-medium">To</th>
                  <th className="py-2 px-4 font-medium text-right">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40 text-foreground">
                {callRows.map((row) => {
                  const c = row.listItem ?? row.treeNode!.call;
                  if (!c) return null;
                  const created = c.created_at ? new Date(c.created_at) : null;
                  const lastStatus = row.listItem
                    ? row.listItem.last_status
                    : row.treeNode!.call?.events?.length
                      ? row.treeNode!.call.events[row.treeNode!.call.events.length - 1].status
                      : "";
                  const dirAccent = DIRECTION_ROW_ACCENT[c.direction ?? ""] ?? "border-l-transparent";
                  const statusClass = STATUS_BADGE_CLASS[lastStatus ?? ""] ?? "bg-slate-500/10 text-slate-600 border border-slate-500/20";
                  const hasChildren =
                    row.depth === 0
                      ? Boolean(row.listItem?.has_children)
                      : (row.treeNode?.children?.length ?? 0) > 0;
                  const isExpanded = expandedIds.has(row.id);
                  const isTreeLoading = loadingTreeId === row.id;
                  const indentPx = Math.max(0, row.depth * 28);
                  const isRoot = row.depth === 0;

                  return (
                    <tr
                      key={row.id}
                      className={`
                        group cursor-pointer transition-colors
                        ${isRoot ? "bg-background hover:bg-secondary/40" : "bg-secondary/20 hover:bg-secondary/60"}
                      `}
                      onClick={() => navigate(`/calls/${row.id}`)}
                    >
                      <td className="py-1.5 pl-4 pr-3">
                        <div className="flex items-center" style={{ paddingLeft: indentPx }}>
                          <div className={`w-0.5 h-6 rounded-full ${dirAccent.replace('border-l-', 'bg-')} mr-2 shrink-0 opacity-70 group-hover:opacity-100 transition-opacity`} />
                           
                          <button
                            type="button"
                            className={`flex size-5 items-center justify-center rounded shrink-0 transition-colors mr-1.5 ${
                              hasChildren 
                                ? "bg-primary/10 text-primary hover:bg-primary/20" 
                                : "text-muted-foreground opacity-30"
                            }`}
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
                              <div className="size-2 animate-pulse bg-current rounded-full" />
                            ) : isExpanded ? (
                              <ChevronDown className="size-3" />
                            ) : (
                              <ChevronRight className="size-3" />
                            )}
                          </button>

                          <div className="flex flex-col min-w-0">
                            <span className="font-mono text-[11px] font-semibold text-foreground truncate max-w-[100px]" title={row.id}>
                                {row.id.split('-')[0]}
                            </span>
                            {row.depth > 0 && (
                                <span className="text-[9px] uppercase font-bold text-muted-foreground -mt-0.5">
                                  subcall {row.depth}
                                </span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td className="py-1.5 px-3 whitespace-nowrap text-muted-foreground">
                        {created ? (
                            <div className="flex flex-col">
                                <span className="text-foreground font-medium">{created.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}</span>
                                <span className="text-[10px] leading-tight">{created.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}</span>
                            </div>
                        ) : "—"}
                      </td>
                      <td className="py-1.5 px-3">
                        <span className={`inline-flex px-1.5 py-0.5 rounded text-[10px] font-semibold ${DIRECTION_BADGE_CLASS[c.direction ?? ""] ?? "bg-secondary text-secondary-foreground"}`}>
                          {DIRECTION_LABELS[c.direction as CallDirection] ?? c.direction ?? "Unknown"}
                        </span>
                      </td>
                      <td className="py-1.5 px-3 text-muted-foreground truncate max-w-[120px]">
                        {isRoot && c.company_telephony_id
                          ? telephonyById[c.company_telephony_id as string] ?? "—"
                          : "—"}
                      </td>
                      <td className="py-1.5 px-3 font-mono text-[10px] font-medium tracking-tight bg-secondary/20">{c.from_number ?? "—"}</td>
                      <td className="py-1.5 px-3 font-mono text-[10px] font-medium tracking-tight bg-secondary/20">{c.to_number ?? "—"}</td>
                      <td className="py-1.5 px-4 text-right">
                        <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold ${statusClass}`}>
                          {(lastStatus && STATUS_LABELS[lastStatus] ? STATUS_LABELS[lastStatus] : lastStatus) || "—"}
                        </span>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Pagination - Slim footer */}
      {data && data.meta.total > PAGE_LIMIT && (
        <div className="flex items-center justify-between text-xs text-muted-foreground px-1 shrink-0">
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

