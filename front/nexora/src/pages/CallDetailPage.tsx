import { useState } from "react";
import { Link, useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, Phone, Clock, MapPin, Music, ChevronDown, ChevronRight, Loader2 } from "lucide-react";
import { useCallDetail } from "../hooks/useCallDetail";
import type { CallTreeResponse } from "../api/calls";

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

const DIRECTION_LABELS: Record<string, string> = {
    call_direction_inbound: "Входящий",
    call_direction_outbound_api: "Исходящий (API)",
    call_direction_outbound_dial: "Исходящий (Dial)",
};

const DIRECTION_COLORS: Record<string, string> = {
    call_direction_inbound: "bg-emerald-500/15 text-emerald-700 border-emerald-200",
    call_direction_outbound_api: "bg-blue-500/15 text-blue-700 border-blue-200",
    call_direction_outbound_dial: "bg-violet-500/15 text-violet-700 border-violet-200",
};

export default function CallDetailPage() {
    const { id } = useParams<{ id: string }>();
    const { data: tree, loading, error } = useCallDetail(id);

    if (loading) {
        return (
            <div className="flex-1 flex items-center justify-center min-h-[50vh]">
                <Loader2 className="w-10 h-10 text-slate-400 animate-spin" />
            </div>
        );
    }
    if (error || !tree?.call) {
        return (
            <div className="p-8 max-w-lg mx-auto">
                <div className="rounded-2xl border border-red-200 bg-red-50/50 p-6 text-center">
                    <p className="text-red-700 font-medium">{error ?? "Звонок не найден"}</p>
                    <Link
                        to="/calls"
                        className="inline-flex items-center gap-2 mt-4 text-slate-600 hover:text-slate-900 font-medium"
                    >
                        <ArrowLeft className="w-4 h-4" />
                        К списку звонков
                    </Link>
                </div>
            </div>
        );
    }

    const root = tree.call;
    const created = root.created_at ? new Date(root.created_at) : null;
    const lastEv = root.events?.length ? root.events[root.events.length - 1] : null;
    const dirClass = DIRECTION_COLORS[root.direction ?? ""] ?? "bg-slate-100 text-slate-700 border-slate-200";

    return (
        <div className="flex-1 min-h-0 flex flex-col" style={{ backgroundColor: "var(--theme-bg-page)" }}>
            <div className="border-b border-slate-200 bg-white/80 backdrop-blur-sm sticky top-0 z-10">
                <div className="max-w-4xl mx-auto px-6 py-4">
                    <Link
                        to="/calls"
                        className="inline-flex items-center gap-2 text-sm font-medium text-slate-600 hover:text-slate-900 mb-4 transition-colors"
                    >
                        <ArrowLeft className="w-4 h-4" />
                        К списку звонков
                    </Link>
                    <div className="flex flex-wrap items-center gap-3">
                        <div className="flex items-center gap-2">
                            <div className="w-10 h-10 rounded-xl bg-slate-100 flex items-center justify-center">
                                <Phone className="w-5 h-5 text-slate-600" />
                            </div>
                            <div>
                                <p className="font-semibold text-slate-900 tabular-nums">
                                    {root.from_number ?? "—"} → {root.to_number ?? "—"}
                                </p>
                                <p className="text-sm text-slate-500">
                                    {created
                                        ? `${created.toLocaleDateString("ru-RU")} ${created.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}`
                                        : "—"}
                                </p>
                            </div>
                        </div>
                        <span className={`inline-flex items-center px-3 py-1 rounded-lg text-xs font-medium border ${dirClass}`}>
                            {DIRECTION_LABELS[root.direction ?? ""] ?? root.direction ?? "—"}
                        </span>
                        {lastEv && (
                            <span className="inline-flex items-center px-3 py-1 rounded-lg text-xs font-medium bg-slate-100 text-slate-700 border border-slate-200">
                                {STATUS_LABELS[lastEv.status ?? ""] ?? lastEv.status}
                            </span>
                        )}
                    </div>
                </div>
            </div>

            <div className="max-w-4xl mx-auto w-full px-6 py-8 space-y-8">
                <section className="rounded-2xl border border-slate-200 bg-white shadow-sm overflow-hidden">
                    <h2 className="px-6 py-4 border-b border-slate-100 text-sm font-semibold text-slate-700 uppercase tracking-wider flex items-center gap-2">
                        <Clock className="w-4 h-4 text-slate-400" />
                        Детали
                    </h2>
                    <div className="p-6">
                        {root.details ? (
                            <div className="grid gap-4 sm:grid-cols-2">
                                {root.details.recording_url && (
                                    <div className="flex items-center gap-3 p-3 rounded-xl bg-slate-50">
                                        <Music className="w-5 h-5 text-slate-500" />
                                        <div>
                                            <p className="text-xs font-medium text-slate-500 uppercase">Запись</p>
                                            <a
                                                href={root.details.recording_url}
                                                target="_blank"
                                                rel="noreferrer"
                                                className="text-blue-600 hover:underline font-medium"
                                            >
                                                Прослушать
                                            </a>
                                        </div>
                                    </div>
                                )}
                                {root.details.recording_duration != null && root.details.recording_duration > 0 && (
                                    <div className="flex items-center gap-3 p-3 rounded-xl bg-slate-50">
                                        <Clock className="w-5 h-5 text-slate-500" />
                                        <div>
                                            <p className="text-xs font-medium text-slate-500 uppercase">Длительность</p>
                                            <p className="font-medium text-slate-800">{root.details.recording_duration} сек</p>
                                        </div>
                                    </div>
                                )}
                                {(root.details.from_country || root.details.from_city) && (
                                    <div className="flex items-center gap-3 p-3 rounded-xl bg-slate-50">
                                        <MapPin className="w-5 h-5 text-slate-500" />
                                        <div>
                                            <p className="text-xs font-medium text-slate-500 uppercase">Откуда</p>
                                            <p className="font-medium text-slate-800">
                                                {[root.details.from_country, root.details.from_city].filter(Boolean).join(" ") || "—"}
                                            </p>
                                        </div>
                                    </div>
                                )}
                                {(root.details.to_country || root.details.to_city) && (
                                    <div className="flex items-center gap-3 p-3 rounded-xl bg-slate-50">
                                        <MapPin className="w-5 h-5 text-slate-500" />
                                        <div>
                                            <p className="text-xs font-medium text-slate-500 uppercase">Куда</p>
                                            <p className="font-medium text-slate-800">
                                                {[root.details.to_country, root.details.to_city].filter(Boolean).join(" ") || "—"}
                                            </p>
                                        </div>
                                    </div>
                                )}
                            </div>
                        ) : (
                            <p className="text-sm text-slate-500">Нет дополнительных деталей</p>
                        )}
                    </div>
                </section>

                {root.events && root.events.length > 0 && (
                    <section className="rounded-2xl border border-slate-200 bg-white shadow-sm overflow-hidden">
                        <h2 className="px-6 py-4 border-b border-slate-100 text-sm font-semibold text-slate-700 uppercase tracking-wider">
                            События
                        </h2>
                        <ul className="divide-y divide-slate-100">
                            {[...root.events].reverse().map((ev, idx) => {
                                const t = ev.timestamp ? new Date(ev.timestamp) : null;
                                return (
                                    <li key={ev.id ?? idx} className="flex items-center justify-between px-6 py-3 hover:bg-slate-50/50">
                                        <span className="font-medium text-slate-800">
                                            {STATUS_LABELS[ev.status ?? ""] ?? ev.status}
                                        </span>
                                        <span className="text-sm text-slate-500 tabular-nums">
                                            {t ? t.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit", second: "2-digit" }) : "—"}
                                        </span>
                                    </li>
                                );
                            })}
                        </ul>
                    </section>
                )}

                {tree.children && tree.children.length > 0 && (
                    <section className="rounded-2xl border border-slate-200 bg-white shadow-sm overflow-hidden">
                        <h2 className="px-6 py-4 border-b border-slate-100 text-sm font-semibold text-slate-700 uppercase tracking-wider flex items-center gap-2">
                            <Phone className="w-4 h-4 text-slate-400" />
                            Дочерние звонки ({tree.children.length})
                        </h2>
                        <div className="p-4 space-y-2">
                            {tree.children.map((child, idx) => (
                                <CallTreeBlock key={child.call?.id ?? idx} node={child} depth={0} />
                            ))}
                        </div>
                    </section>
                )}
            </div>
        </div>
    );
}

function CallTreeBlock({ node, depth }: { node: CallTreeResponse; depth: number }) {
    const navigate = useNavigate();
    const [open, setOpen] = useState(depth < 2);
    const r = node.call;
    if (!r?.id) return null;
    const created = r.created_at ? new Date(r.created_at) : null;
    const lastEv = r.events?.length ? r.events[r.events.length - 1] : null;
    const hasChildren = node.children && node.children.length > 0;
    const dirClass = DIRECTION_COLORS[r.direction ?? ""] ?? "bg-slate-100 text-slate-600";

    return (
        <div className={depth > 0 ? "ml-4 pl-4 border-l-2 border-slate-200" : ""}>
            <div
                className={`rounded-xl border transition-colors ${
                    depth === 0 ? "border-slate-200 bg-slate-50/50" : "border-slate-100 bg-white"
                } overflow-hidden`}
            >
                <div
                    role="button"
                    tabIndex={0}
                    className="flex items-center gap-3 p-4 cursor-pointer hover:bg-slate-50/80"
                    onClick={() => navigate(`/calls/${r.id}`)}
                    onKeyDown={(e) => e.key === "Enter" && navigate(`/calls/${r.id}`)}
                >
                    {hasChildren ? (
                        <button
                            type="button"
                            className="p-1 rounded-lg hover:bg-slate-200 text-slate-500"
                            onClick={(e) => {
                                e.stopPropagation();
                                setOpen((o: boolean) => !o);
                            }}
                            aria-label={open ? "Свернуть" : "Развернуть"}
                        >
                            {open ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
                        </button>
                    ) : (
                        <span className="w-6" />
                    )}
                    <div className="min-w-0 flex-1">
                        <p className="font-medium text-slate-900 tabular-nums">
                            {r.from_number ?? "—"} → {r.to_number ?? "—"}
                        </p>
                        <div className="flex items-center gap-2 mt-1 flex-wrap">
                            {created && (
                                <span className="text-xs text-slate-500">
                                    {created.toLocaleDateString("ru-RU")} {created.toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" })}
                                </span>
                            )}
                            <span className={`inline-flex px-2 py-0.5 rounded text-xs font-medium border ${dirClass}`}>
                                {DIRECTION_LABELS[r.direction ?? ""] ?? "—"}
                            </span>
                            {lastEv && (
                                <span className="text-xs text-slate-500">
                                    {STATUS_LABELS[lastEv.status ?? ""] ?? lastEv.status}
                                </span>
                            )}
                        </div>
                    </div>
                </div>
            </div>
            {hasChildren && open && (
                <div className="mt-2 space-y-2">
                    {node.children!.map((ch: CallTreeResponse, idx: number) => (
                        <CallTreeBlock key={ch.call?.id ?? idx} node={ch} depth={depth + 1} />
                    ))}
                </div>
            )}
        </div>
    );
}
