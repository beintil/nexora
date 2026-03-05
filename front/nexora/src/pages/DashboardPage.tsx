import { useEffect, useState } from "react";
import { Phone, PhoneCall, PhoneMissed } from "lucide-react";
import { AreaChart, Area, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer } from "recharts";
import { getCallMetrics } from "../api/metrics";

const CHART_COLORS = {
    total: "#94a3b8",
    answered: "#10b981",
    missed: "#f59e0b",
};

function todayISO(): string {
    return new Date().toISOString().slice(0, 10);
}

function yesterdayISO(): string {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    return d.toISOString().slice(0, 10);
}

export default function DashboardPage() {
    const [metrics, setMetrics] = useState<Awaited<ReturnType<typeof getCallMetrics>> | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    function loadMetrics() {
        setLoading(true);
        setError(null);
        const today = todayISO();
        const from = new Date();
        from.setDate(from.getDate() - 6);
        getCallMetrics({
            date_from: from.toISOString().slice(0, 10),
            date_to: today,
        })
            .then((data) => {
                setMetrics(data);
                setLoading(false);
            })
            .catch((e) => {
                setError((e as Error).message || "Не удалось загрузить метрики");
                setLoading(false);
            });
    }

    useEffect(() => {
        loadMetrics();
    }, []);

    const sparkDataTotal =
        metrics?.timeseries && metrics.timeseries.length > 0
            ? metrics.timeseries.map((p) => ({ v: p.total }))
            : [{ v: 0 }];
    const sparkDataAnswered =
        metrics?.timeseries && metrics.timeseries.length > 0
            ? metrics.timeseries.map((p) => ({ v: p.answered }))
            : [{ v: 0 }];
    const sparkDataMissed =
        metrics?.timeseries && metrics.timeseries.length > 0
            ? metrics.timeseries.map((p) => ({ v: p.missed }))
            : [{ v: 0 }];

    const cards = [
        {
            icon: Phone,
            title: "Всего звонков",
            value: loading ? "—" : String(metrics?.summary?.total ?? 0),
            chartColor: CHART_COLORS.total,
            sparkData: sparkDataTotal,
        },
        {
            icon: PhoneCall,
            title: "Отвеченных звонков",
            value: loading ? "—" : String(metrics?.summary?.answered ?? 0),
            chartColor: CHART_COLORS.answered,
            sparkData: sparkDataAnswered,
        },
        {
            icon: PhoneMissed,
            title: "Пропущенных",
            value: loading ? "—" : String(metrics?.summary?.missed ?? 0),
            chartColor: CHART_COLORS.missed,
            sparkData: sparkDataMissed,
        },
    ];

    const chartData =
        metrics?.timeseries && metrics.timeseries.length > 0
            ? metrics.timeseries.map((p) => ({
                  date: p.date ? new Date(p.date).toLocaleDateString("ru-RU", { day: "numeric", month: "short" }) : "",
                  dateISO: p.date ?? "",
                  total: p.total ?? 0,
                  answered: p.answered ?? 0,
                  missed: p.missed ?? 0,
              }))
            : [];

    const todayStr = todayISO();
    const yesterdayStr = yesterdayISO();
    const todayPoint = chartData.find((p) => p.dateISO === todayStr);
    const yesterdayPoint = chartData.find((p) => p.dateISO === yesterdayStr);

    return (
        <div className="flex-1 min-h-0 overflow-auto" style={{ backgroundColor: "var(--theme-bg-page)" }}>
            <div className="max-w-4xl mx-auto px-8 py-8">
                <h2 className="text-lg font-medium mb-6" style={{ color: "var(--theme-text)" }}>
                    Данные за неделю
                </h2>
                {error && (
                    <div className="mb-4 p-4 rounded-xl border flex items-center justify-between" style={{ backgroundColor: "var(--theme-bg-card)", borderColor: "var(--theme-border)" }}>
                        <span style={{ color: "var(--theme-text)" }}>{error}</span>
                        <button
                            type="button"
                            className="px-3 py-1.5 text-sm rounded-lg border"
                            style={{ borderColor: "var(--theme-border)", color: "var(--theme-text)" }}
                            onClick={loadMetrics}
                        >
                            Повторить
                        </button>
                    </div>
                )}
                <div className="grid gap-4 sm:grid-cols-3">
                    {cards.map(({ icon: Icon, title, value, chartColor, sparkData }, idx) => (
                        <div
                            key={title}
                            className="rounded-xl shadow-sm p-4 border"
                            style={{
                                backgroundColor: "var(--theme-bg-card)",
                                borderColor: "var(--theme-border)",
                            }}
                        >
                            <div className="flex items-center gap-2 mb-2">
                                <span
                                    className="flex items-center justify-center w-8 h-8 rounded-lg shrink-0"
                                    style={{ backgroundColor: `${chartColor}20`, color: chartColor }}
                                >
                                    <Icon className="h-3.5 w-3.5" />
                                </span>
                                <span className="font-medium text-sm truncate" style={{ color: "var(--theme-text)" }}>
                                    {title}
                                </span>
                            </div>
                            <p
                                className="text-xl font-light tabular-nums mb-2"
                                style={{ fontFamily: '"Cormorant Garamond", Georgia, serif', color: "var(--theme-text)" }}
                            >
                                {value}
                            </p>
                            {idx === 0 && !loading && chartData.length > 0 ? (
                                <div className="h-16 -mx-1 mt-1">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <BarChart
                                            data={chartData}
                                            margin={{ top: 2, right: 2, left: 2, bottom: 2 }}
                                            barCategoryGap="15%"
                                        >
                                            <XAxis dataKey="date" tick={{ fill: "var(--theme-text-muted)", fontSize: 9 }} hide />
                                            <YAxis tick={{ fill: "var(--theme-text-muted)", fontSize: 9 }} width={18} />
                                            <Tooltip
                                                contentStyle={{
                                                    backgroundColor: "var(--theme-bg-card)",
                                                    border: "1px solid var(--theme-border)",
                                                    borderRadius: "6px",
                                                    fontSize: "11px",
                                                }}
                                                formatter={(value: number) => [value, "звонков"]}
                                            />
                                            <Bar dataKey="total" fill={CHART_COLORS.total} radius={[2, 2, 0, 0]} />
                                        </BarChart>
                                    </ResponsiveContainer>
                                </div>
                            ) : (
                                <div className="h-8 -mx-1">
                                    <ResponsiveContainer width="100%" height="100%">
                                        <AreaChart data={sparkData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                                            <defs>
                                                <linearGradient
                                                    id={`dashboard-spark-${idx}`}
                                                    x1="0"
                                                    y1="0"
                                                    x2="0"
                                                    y2="1"
                                                >
                                                    <stop offset="0%" stopColor={chartColor} stopOpacity={0.35} />
                                                    <stop offset="100%" stopColor={chartColor} stopOpacity={0} />
                                                </linearGradient>
                                            </defs>
                                            <Area
                                                type="monotone"
                                                dataKey="v"
                                                stroke={chartColor}
                                                strokeWidth={1}
                                                fill={`url(#dashboard-spark-${idx})`}
                                                isAnimationActive={false}
                                            />
                                        </AreaChart>
                                    </ResponsiveContainer>
                                </div>
                            )}
                        </div>
                    ))}
                </div>

                {(yesterdayPoint || todayPoint) && !loading && (
                    <div
                        className="mt-6 rounded-xl border p-4"
                        style={{
                            backgroundColor: "var(--theme-bg-card)",
                            borderColor: "var(--theme-border)",
                        }}
                    >
                        <p className="text-sm font-medium mb-3" style={{ color: "var(--theme-text-muted)" }}>
                            Сравнение с вчера
                        </p>
                        <div className="grid grid-cols-2 gap-4 sm:gap-6">
                            <div>
                                <p className="text-xs font-medium mb-2 uppercase tracking-wider" style={{ color: "var(--theme-text-muted)" }}>
                                    Вчера
                                </p>
                                <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
                                    <span style={{ color: "var(--theme-text)" }}>
                                        Всего: <strong>{yesterdayPoint?.total ?? 0}</strong>
                                    </span>
                                    <span style={{ color: "var(--theme-text)" }}>
                                        Отвечено: <strong style={{ color: CHART_COLORS.answered }}>{yesterdayPoint?.answered ?? 0}</strong>
                                    </span>
                                    <span style={{ color: "var(--theme-text)" }}>
                                        Пропущено: <strong style={{ color: CHART_COLORS.missed }}>{yesterdayPoint?.missed ?? 0}</strong>
                                    </span>
                                </div>
                            </div>
                            <div>
                                <p className="text-xs font-medium mb-2 uppercase tracking-wider" style={{ color: "var(--theme-text-muted)" }}>
                                    Сегодня
                                </p>
                                <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
                                    <span style={{ color: "var(--theme-text)" }}>
                                        Всего: <strong>{todayPoint?.total ?? 0}</strong>
                                    </span>
                                    <span style={{ color: "var(--theme-text)" }}>
                                        Отвечено: <strong style={{ color: CHART_COLORS.answered }}>{todayPoint?.answered ?? 0}</strong>
                                    </span>
                                    <span style={{ color: "var(--theme-text)" }}>
                                        Пропущено: <strong style={{ color: CHART_COLORS.missed }}>{todayPoint?.missed ?? 0}</strong>
                                    </span>
                                </div>
                                {yesterdayPoint && todayPoint && (
                                    <p className="text-xs mt-2" style={{ color: "var(--theme-text-muted)" }}>
                                        {todayPoint.total > yesterdayPoint.total && (
                                            <>+{todayPoint.total - yesterdayPoint.total} звонков к вчера</>
                                        )}
                                        {todayPoint.total < yesterdayPoint.total && (
                                            <>{todayPoint.total - yesterdayPoint.total} звонков к вчера</>
                                        )}
                                        {todayPoint.total === yesterdayPoint.total && <>Без изменений к вчера</>}
                                    </p>
                                )}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}
