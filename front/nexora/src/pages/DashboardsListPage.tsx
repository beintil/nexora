import { Link } from "react-router-dom"
import {
    Phone,
    type LucideIcon,
} from "lucide-react"

export interface DashboardMeta {
    id: string
    title: string
    description: string
    path: string
    icon: LucideIcon
}

/** Каталог всех дашбордов приложения (телефония + аналитика) */
const ALL_DASHBOARDS: DashboardMeta[] = [
    {
        id: "calls",
        title: "Звонки",
        description: "Входящие и исходящие, пропущенные, история по номерам",
        path: "/calls",
        icon: Phone,
    },
]

export default function DashboardsListPage() {
    return (
        <div className="p-8 max-w-5xl">
            <div className="mb-8">
                <h1 className="text-2xl font-semibold text-slate-900 mb-1">Все дашборды</h1>
                <p className="text-slate-600 text-sm">
                    Выберите отчёт для детального просмотра или добавьте виджет на главную.
                </p>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {ALL_DASHBOARDS.map((d) => {
                    const Icon = d.icon
                    return (
                        <Link
                            key={d.id}
                            to={d.path}
                            className="flex gap-4 rounded-2xl border border-slate-200 bg-white p-5 shadow-sm hover:border-slate-300 hover:shadow-md transition group"
                        >
                            <div className="flex items-center justify-center w-12 h-12 rounded-xl bg-slate-100 text-slate-600 group-hover:bg-slate-200 transition shrink-0">
                                <Icon className="h-6 w-6" />
                            </div>
                            <div className="min-w-0">
                                <h2 className="font-semibold text-slate-900 mb-1">{d.title}</h2>
                                <p className="text-sm text-slate-500 line-clamp-2">{d.description}</p>
                            </div>
                        </Link>
                    )
                })}
            </div>
        </div>
    )
}
