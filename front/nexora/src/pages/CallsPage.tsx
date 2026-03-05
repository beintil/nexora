import { useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ChevronDown, ChevronRight, Calendar, SlidersHorizontal } from "lucide-react";
import type {
    CallsListItem,
    CallsListResponse,
    CallTreeResponse,
    CallDirection,
} from "../api/calls";
import { getCalls, getCallById } from "../api/calls";
import { getCompanyTelephony } from "../api/companyTelephony";
import type { CompanyTelephonyItem } from "../api/companyTelephony";

type LoadingState = "idle" | "loading" | "error";

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
    call_direction_inbound: "bg-emerald-500/15 text-emerald-700 border-emerald-200",
    call_direction_outbound_api: "bg-blue-500/15 text-blue-700 border-blue-200",
    call_direction_outbound_dial: "bg-violet-500/15 text-violet-700 border-violet-200",
};

const STATUS_BADGE_CLASS: Record<string, string> = {
    call_event_status_completed: "bg-emerald-100 text-emerald-800 border-emerald-200",
    call_event_status_in_progress: "bg-blue-100 text-blue-800 border-blue-200",
    call_event_status_ringing: "bg-amber-100 text-amber-800 border-amber-200",
    call_event_status_initiated: "bg-slate-100 text-slate-700 border-slate-200",
    call_event_status_no_answer: "bg-orange-100 text-orange-800 border-orange-200",
    call_event_status_busy: "bg-rose-100 text-rose-800 border-rose-200",
    call_event_status_failed: "bg-red-100 text-red-800 border-red-200",
    call_event_status_canceled: "bg-slate-100 text-slate-600 border-slate-200",
    call_event_status_timeout: "bg-amber-100 text-amber-700 border-amber-200",
    call_event_status_queued: "bg-slate-100 text-slate-600 border-slate-200",
};

export default function CallsPage() {
    const [data, setData] = useState<CallsListResponse | null>(null);
    const [loading, setLoading] = useState<LoadingState>("idle");
    const [error, setError] = useState<string | null>(null);

    const [page, setPage] = useState(0);
    const [direction, setDirection] = useState<CallDirection | "all">("all");
    const [status, setStatus] = useState<string | "all">("all");
    const [dateFrom, setDateFrom] = useState<string>(() => {
        const d = new Date();
        d.setDate(d.getDate() - 6);
        return d.toISOString().slice(0, 10);
    });
    const [dateTo, setDateTo] = useState<string>(() => new Date().toISOString().slice(0, 10));
    const [telephonyId, setTelephonyId] = useState<string>("all");
    const [telephonyList, setTelephonyList] = useState<CompanyTelephonyItem[]>([]);
    const [periodOpen, setPeriodOpen] = useState(false);
    const [filtersOpen, setFiltersOpen] = useState(false);
    const periodRef = useRef<HTMLDivElement>(null);
    const filtersRef = useRef<HTMLDivElement>(null);

    const navigate = useNavigate();

    const periodLabel = useMemo(() => {
        const today = new Date().toISOString().slice(0, 10);
        const weekStart = new Date();
        weekStart.setDate(weekStart.getDate() - 6);
        const weekStartStr = weekStart.toISOString().slice(0, 10);
        const monthStart = new Date();
        monthStart.setDate(monthStart.getDate() - 29);
        const monthStartStr = monthStart.toISOString().slice(0, 10);
        if (dateFrom === today && dateTo === today) return "Сегодня";
        if (dateFrom === weekStartStr && dateTo === today) return "7 дней";
        if (dateFrom === monthStartStr && dateTo === today) return "30 дней";
        const from = dateFrom ? new Date(dateFrom).toLocaleDateString("ru-RU", { day: "2-digit", month: "2-digit", year: "2-digit" }) : "";
        const to = dateTo ? new Date(dateTo).toLocaleDateString("ru-RU", { day: "2-digit", month: "2-digit", year: "2-digit" }) : "";
        return from && to ? `${from} — ${to}` : "Период";
    }, [dateFrom, dateTo]);

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
                const id = ch.call?.id;
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
            const seenInBranch = new Set<string>();
            rows.push({ id: item.id, depth: 0, listItem: item });
            if (expandedIds.has(item.id) && treeByCallId[item.id]?.children?.length) {
                pushChildren(treeByCallId[item.id].children!, 1, new Set([item.id]), seenInBranch);
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

    useEffect(() => {
        getCompanyTelephony()
            .then(setTelephonyList)
            .catch(() => setTelephonyList([]));
    }, []);

    useEffect(() => {
        let cancelled = false;
        async function load() {
            setLoading("loading");
            setError(null);
            try {
                const res = await getCalls({
                    page: { limit: PAGE_LIMIT, offset: page * PAGE_LIMIT, page: page + 1 },
                    direction: direction === "all" ? undefined : direction,
                    status: status === "all" ? undefined : status,
                    date_from: dateFrom || undefined,
                    date_to: dateTo || undefined,
                    company_telephony_id: telephonyId === "all" ? undefined : telephonyId,
                });
                if (!cancelled) {
                    setData(res);
                    setLoading("idle");
                }
            } catch (e) {
                if (!cancelled) {
                    setLoading("error");
                    setError((e as Error).message || "Не удалось загрузить звонки");
                }
            }
        }
        load();
        return () => {
            cancelled = true;
        };
    }, [page, direction, status, dateFrom, dateTo, telephonyId]);

    const totalPages = useMemo(() => {
        if (!data) return 0;
        return Math.max(1, Math.ceil(data.meta.total / PAGE_LIMIT));
    }, [data]);

    return (
        <div className="flex-1 min-h-0 flex flex-col" style={{ color: "var(--theme-text)" }}>
            <div className="flex-1 p-6 md:p-8 min-w-0 max-w-5xl mx-auto w-full">
                <div className="flex items-center justify-between mb-6">
                    <h1 className="text-2xl font-bold tracking-tight" style={{ color: "var(--theme-text)" }}>
                        Звонки
                    </h1>
                </div>

                {/* Тулбар: период и фильтры в выпадающих блоках */}
                <div className="mb-5 flex flex-col gap-3">
                    <div className="flex flex-wrap items-center gap-2">
                        <div className="relative" ref={periodRef}>
                            <button
                                type="button"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    setPeriodOpen((v) => !v);
                                    setFiltersOpen(false);
                                }}
                                className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"
                            >
                                <Calendar className="h-4 w-4 text-slate-500" />
                                {periodLabel}
                                <ChevronDown className={`h-4 w-4 text-slate-400 transition ${periodOpen ? "rotate-180" : ""}`} />
                            </button>
                            {periodOpen && (
                                <div className="absolute left-0 top-full z-20 mt-1.5 min-w-[280px] rounded-xl border border-slate-200 bg-white p-4 shadow-lg">
                                    <p className="mb-3 text-xs font-medium uppercase tracking-wider text-slate-400">Быстрый выбор</p>
                                    <div className="flex gap-2">
                                        {[
                                            { label: "Сегодня", days: 0 },
                                            { label: "7 дней", days: 6 },
                                            { label: "30 дней", days: 29 },
                                        ].map(({ label, days }) => {
                                            const to = new Date();
                                            const from = new Date();
                                            from.setDate(from.getDate() - days);
                                            const fromStr = from.toISOString().slice(0, 10);
                                            const toStr = to.toISOString().slice(0, 10);
                                            const isActive = dateFrom === fromStr && dateTo === toStr;
                                            return (
                                                <button
                                                    key={label}
                                                    type="button"
                                                    onClick={() => {
                                                        setDateFrom(fromStr);
                                                        setDateTo(toStr);
                                                    }}
                                                    className={`flex-1 rounded-lg px-3 py-2.5 text-sm font-medium transition ${
                                                        isActive ? "bg-slate-900 text-white shadow-sm" : "bg-slate-100 text-slate-700 hover:bg-slate-200"
                                                    }`}
                                                >
                                                    {label}
                                                </button>
                                            );
                                        })}
                                    </div>
                                    <p className="mb-2 mt-4 text-xs font-medium uppercase tracking-wider text-slate-400">Свой период</p>
                                    <div className="flex items-center gap-2">
                                        <input
                                            type="date"
                                            value={dateFrom}
                                            onChange={(e) => setDateFrom(e.target.value)}
                                            className="flex-1 rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2 text-sm text-slate-800 focus:border-slate-400 focus:bg-white focus:outline-none focus:ring-1 focus:ring-slate-400"
                                        />
                                        <span className="text-slate-300">—</span>
                                        <input
                                            type="date"
                                            value={dateTo}
                                            onChange={(e) => setDateTo(e.target.value)}
                                            className="flex-1 rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2 text-sm text-slate-800 focus:border-slate-400 focus:bg-white focus:outline-none focus:ring-1 focus:ring-slate-400"
                                        />
                                    </div>
                                </div>
                            )}
                        </div>
                        <div className="relative" ref={filtersRef}>
                            <button
                                type="button"
                                onClick={(e) => {
                                    e.stopPropagation();
                                    setFiltersOpen((v) => !v);
                                    setPeriodOpen(false);
                                }}
                                className="flex items-center gap-2 rounded-lg border border-slate-200 bg-white px-3 py-2 text-sm font-medium text-slate-700 shadow-sm transition hover:border-slate-300 hover:bg-slate-50"
                            >
                                <SlidersHorizontal className="h-4 w-4 text-slate-500" />
                                Фильтры
                                {(direction !== "all" || status !== "all" || telephonyId !== "all") && (
                                    <span className="flex h-5 min-w-[20px] items-center justify-center rounded-full bg-slate-200 px-1.5 text-xs font-semibold text-slate-700">
                                        {[direction, status, telephonyId].filter((v) => v !== "all").length}
                                    </span>
                                )}
                                <ChevronDown className={`h-4 w-4 text-slate-400 transition ${filtersOpen ? "rotate-180" : ""}`} />
                            </button>
                            {filtersOpen && (
                                <div className="absolute left-0 top-full z-20 mt-1.5 min-w-[320px] rounded-xl border border-slate-200 bg-white p-4 shadow-lg">
                                    <p className="mb-3 text-xs font-medium uppercase tracking-wider text-slate-400">Телефония</p>
                                    <select
                                        value={telephonyId}
                                        onChange={(e) => setTelephonyId(e.target.value)}
                                        className="mb-4 w-full rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2.5 text-sm text-slate-800 focus:border-slate-400 focus:bg-white focus:outline-none focus:ring-1 focus:ring-slate-400"
                                    >
                                        <option value="all">Все</option>
                                        {telephonyList.map((t) => (
                                            <option key={t.id} value={t.id ?? ""}>
                                                {t.telephony_name ?? t.id ?? "—"}
                                            </option>
                                        ))}
                                    </select>
                                    <p className="mb-3 text-xs font-medium uppercase tracking-wider text-slate-400">Направление</p>
                                    <select
                                        value={direction}
                                        onChange={(e) => setDirection(e.target.value as CallDirection | "all")}
                                        className="mb-4 w-full rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2.5 text-sm text-slate-800 focus:border-slate-400 focus:bg-white focus:outline-none focus:ring-1 focus:ring-slate-400"
                                    >
                                        <option value="all">Все</option>
                                        <option value="call_direction_inbound">Входящий</option>
                                        <option value="call_direction_outbound_api">Исходящий</option>
                                        <option value="call_direction_outbound_dial">Исходящий</option>
                                    </select>
                                    <p className="mb-3 text-xs font-medium uppercase tracking-wider text-slate-400">Статус</p>
                                    <select
                                        value={status}
                                        onChange={(e) => setStatus(e.target.value as string | "all")}
                                        className="w-full rounded-lg border border-slate-200 bg-slate-50/80 px-3 py-2.5 text-sm text-slate-800 focus:border-slate-400 focus:bg-white focus:outline-none focus:ring-1 focus:ring-slate-400"
                                    >
                                        <option value="all">Все</option>
                                        <option value="call_event_status_completed">Отвеченные</option>
                                        <option value="call_event_status_no_answer">Нет ответа</option>
                                        <option value="call_event_status_busy">Занято</option>
                                        <option value="call_event_status_canceled">Отменённые</option>
                                        <option value="call_event_status_failed">С ошибкой</option>
                                    </select>
                                </div>
                            )}
                        </div>
                    </div>
                    <div className="flex items-center justify-between gap-2">
                        <p className="text-[13px] text-slate-500">
                            Нажмите на звонок, чтобы открыть детали и дочерние звонки.
                        </p>
                        {(direction !== "all" || status !== "all" || telephonyId !== "all") && (
                            <button
                                type="button"
                                onClick={() => {
                                    setDirection("all");
                                    setStatus("all");
                                    setTelephonyId("all");
                                    const to = new Date();
                                    const from = new Date();
                                    from.setDate(from.getDate() - 6);
                                    setDateFrom(from.toISOString().slice(0, 10));
                                    setDateTo(to.toISOString().slice(0, 10));
                                }}
                                className="text-[13px] font-medium text-slate-500 hover:text-slate-700"
                            >
                                Сбросить фильтры
                            </button>
                        )}
                    </div>
                </div>

                <div className="rounded-2xl border border-slate-200 bg-white shadow-md overflow-hidden">
                    {loading === "loading" && (
                        <div className="p-8 text-sm text-slate-500 text-center">Загрузка звонков…</div>
                    )}
                    {loading === "error" && (
                        <div className="p-6 flex items-center justify-between gap-4 border-b border-slate-100 bg-red-50/50">
                            <span className="text-sm text-red-700">{error}</span>
                            <button
                                type="button"
                                className="px-4 py-2 rounded-lg text-sm font-medium border border-red-200 text-red-700 hover:bg-red-100"
                                onClick={() => setPage((p) => p)}
                            >
                                Повторить
                            </button>
                        </div>
                    )}
                    {loading === "idle" && data && data.items.length === 0 && (
                        <div className="p-12 text-sm text-slate-500 text-center">Звонков пока нет.</div>
                    )}
                    {loading === "idle" && data && data.items.length > 0 && (
                        <div className="overflow-x-auto">
                            <table className="min-w-full text-sm border-collapse">
                                <thead>
                                    <tr className="bg-slate-100/80 border-b-2 border-slate-200">
                                        <th className="text-left py-4 pl-5 pr-3 text-xs font-semibold uppercase tracking-wider text-slate-500 w-[240px] min-w-[240px]">
                                            Звонок
                                        </th>
                                        <th className="text-left py-4 px-4 text-xs font-semibold uppercase tracking-wider text-slate-500">Когда</th>
                                        <th className="text-left py-4 px-4 text-xs font-semibold uppercase tracking-wider text-slate-500">Направление</th>
                                        <th className="text-left py-4 px-4 text-xs font-semibold uppercase tracking-wider text-slate-500">Телефония</th>
                                        <th className="text-left py-4 px-4 text-xs font-semibold uppercase tracking-wider text-slate-500">От</th>
                                        <th className="text-left py-4 px-4 text-xs font-semibold uppercase tracking-wider text-slate-500">Кому</th>
                                        <th className="text-left py-4 px-4 text-xs font-semibold uppercase tracking-wider text-slate-500">Статус</th>
                                    </tr>
                                </thead>
                                <tbody>
                                    {callRows.map((row) => {
                                        const c = row.listItem ?? row.treeNode!.call;
                                        if (!c) return null;
                                        const created = c.created_at ? new Date(c.created_at) : null;
                                        const lastStatus = row.listItem
                                            ? row.listItem.last_status
                                            : row.treeNode!.call?.events?.length
                                              ? row.treeNode!.call.events[row.treeNode!.call.events.length - 1].status
                                              : "";
                                        const dirAccent = DIRECTION_ROW_ACCENT[c.direction ?? ""] ?? "border-l-slate-300";
                                        const statusClass = STATUS_BADGE_CLASS[lastStatus ?? ""] ?? "bg-slate-100 text-slate-600 border-slate-200";
                                        const hasChildren =
                                            row.depth === 0
                                                ? Boolean(row.listItem?.has_children)
                                                : (row.treeNode?.children?.length ?? 0) > 0;
                                        const isExpanded = expandedIds.has(row.id);
                                        const isLoading = loadingTreeId === row.id;
                                        const indentPx = 16 + row.depth * 32;
                                        const isRoot = row.depth === 0;
                                        return (
                                            <tr
                                                key={row.id}
                                                className={`
                                                    cursor-pointer transition-colors border-b border-slate-100
                                                    ${isRoot ? "bg-white hover:bg-slate-50/80" : "bg-slate-50/70 hover:bg-slate-100/80"}
                                                    border-l-4 ${dirAccent}
                                                `}
                                                onClick={() => navigate(`/calls/${row.id}`)}
                                            >
                                                <td
                                                    className="py-3 align-middle"
                                                    style={{ paddingLeft: indentPx }}
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        if (row.depth === 0 || treeByCallId[row.id] != null) {
                                                            toggleExpand(row.id);
                                                        } else if (row.treeNode?.children?.length) {
                                                            toggleExpand(row.id);
                                                        }
                                                    }}
                                                >
                                                    <div className="flex items-center gap-2">
                                                        <button
                                                            type="button"
                                                            className={`w-8 h-8 flex items-center justify-center rounded-lg shrink-0 transition-colors hover:opacity-80 ${
                                                                hasChildren ? "text-red-600" : "text-slate-400"
                                                            }`}
                                                            disabled={isLoading}
                                                            aria-label={isExpanded ? "Свернуть" : "Развернуть"}
                                                        >
                                                            {isLoading ? (
                                                                <span className="text-xs">…</span>
                                                            ) : isExpanded ? (
                                                                <ChevronDown className="w-4 h-4" />
                                                            ) : (
                                                                <ChevronRight className="w-4 h-4" />
                                                            )}
                                                        </button>
                                                        {row.depth > 0 && (
                                                            <span className="text-xs font-medium text-slate-400 uppercase tracking-wider shrink-0">
                                                                вложенный {row.depth}
                                                            </span>
                                                        )}
                                                    </div>
                                                </td>
                                                <td className="py-3 px-4 whitespace-nowrap text-slate-600">
                                                    {created
                                                        ? `${created.toLocaleDateString()} ${created.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}`
                                                        : "—"}
                                                </td>
                                                <td className="py-3 px-4">
                                                    <span className={`inline-flex px-2.5 py-0.5 rounded-lg text-xs font-medium border ${DIRECTION_BADGE_CLASS[c.direction ?? ""] ?? "bg-slate-100 text-slate-600 border-slate-200"}`}>
                                                        {DIRECTION_LABELS[c.direction as CallDirection] ?? c.direction ?? "—"}
                                                    </span>
                                                </td>
                                                <td className="py-3 px-4 text-slate-600">
                                                    {isRoot && c.company_telephony_id
                                                        ? telephonyById[c.company_telephony_id] ?? "—"
                                                        : "—"}
                                                </td>
                                                <td className="py-3 px-4 font-medium text-slate-800 tabular-nums">{c.from_number ?? "—"}</td>
                                                <td className="py-3 px-4 font-medium text-slate-800 tabular-nums">{c.to_number ?? "—"}</td>
                                                <td className="py-3 px-4">
                                                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-lg text-xs font-medium border ${statusClass}`}>
                                                        {(STATUS_LABELS[lastStatus] ?? lastStatus) || "—"}
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

                {data && data.meta.total > PAGE_LIMIT && (
                    <div className="flex items-center justify-between mt-4 text-sm text-slate-600">
                        <span>
                            Страница {page + 1} из {totalPages}
                        </span>
                        <div className="flex gap-2">
                            <button
                                type="button"
                                disabled={page === 0}
                                className="px-3 py-1 border rounded-lg disabled:opacity-40"
                                onClick={() => setPage((p) => Math.max(0, p - 1))}
                            >
                                Назад
                            </button>
                            <button
                                type="button"
                                disabled={page + 1 >= totalPages}
                                className="px-3 py-1 border rounded-lg disabled:opacity-40"
                                onClick={() => setPage((p) => p + 1)}
                            >
                                Вперёд
                            </button>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

