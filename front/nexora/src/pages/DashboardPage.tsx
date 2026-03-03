import { Phone, PhoneCall, PhoneMissed } from "lucide-react"
import { AreaChart, Area, ResponsiveContainer } from "recharts"

const sparkData = [
    { v: 4 }, { v: 6 }, { v: 5 }, { v: 8 }, { v: 7 }, { v: 9 }, { v: 11 }, { v: 10 }, { v: 12 }, { v: 14 },
]

const METRICS = [
    { icon: Phone, title: "Всего звонков", value: "—", chartColor: "#94a3b8" },
    { icon: PhoneCall, title: "Отвеченных звонков", value: "—", chartColor: "#10b981" },
    { icon: PhoneMissed, title: "Пропущенных", value: "—", chartColor: "#f59e0b" },
]

export default function DashboardPage() {
    return (
        <div className="flex-1 min-h-0 overflow-auto" style={{ backgroundColor: "var(--theme-bg-page)" }}>
            <div className="max-w-2xl mx-auto px-8 py-8">
                <div className="grid gap-4 sm:grid-cols-3">
                    {METRICS.map(({ icon: Icon, title, value, chartColor }) => (
                        <div
                            key={title}
                            className="rounded-xl shadow-sm p-4 border"
                            style={{
                                backgroundColor: "var(--theme-bg-card)",
                                borderColor: "var(--theme-border)",
                            }}
                        >
                            <div className="flex items-center gap-2 mb-3">
                                <span
                                    className="flex items-center justify-center w-9 h-9 rounded-lg shrink-0"
                                    style={{ backgroundColor: `${chartColor}20`, color: chartColor }}
                                >
                                    <Icon className="h-4 w-4" />
                                </span>
                                <span className="font-medium text-sm truncate" style={{ color: "var(--theme-text)" }}>{title}</span>
                            </div>
                            <p className="text-2xl font-light tabular-nums mb-3" style={{ fontFamily: '"Cormorant Garamond", Georgia, serif', color: "var(--theme-text)" }}>
                                {value}
                            </p>
                            <div className="h-12 -mx-1">
                                <ResponsiveContainer width="100%" height="100%">
                                    <AreaChart data={sparkData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                                        <defs>
                                            <linearGradient id={`fill-${title.replace(/\s/g, "")}`} x1="0" y1="0" x2="0" y2="1">
                                                <stop offset="0%" stopColor={chartColor} stopOpacity={0.35} />
                                                <stop offset="100%" stopColor={chartColor} stopOpacity={0} />
                                            </linearGradient>
                                        </defs>
                                        <Area
                                            type="monotone"
                                            dataKey="v"
                                            stroke={chartColor}
                                            strokeWidth={1.5}
                                            fill={`url(#fill-${title.replace(/\s/g, "")})`}
                                            isAnimationActive={false}
                                        />
                                    </AreaChart>
                                </ResponsiveContainer>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    )
}
