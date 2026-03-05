import { useEffect, useState } from "react";
import { useTheme, type ThemeChoice } from "../context/ThemeContext";
import { Monitor, Moon, Sun, Palette } from "lucide-react";
import {
    getCompanyTelephony,
    createCompanyTelephony,
    deleteCompanyTelephony,
    getTelephonyDictionary,
    type CompanyTelephonyItem,
    type TelephonyDictionaryItem,
} from "../api/companyTelephony";

const THEMES: { value: ThemeChoice; label: string; icon: typeof Monitor }[] = [
    { value: "system", label: "Системная", icon: Monitor },
    { value: "light", label: "Светлая", icon: Sun },
    { value: "dark", label: "Тёмная", icon: Moon },
    { value: "gray", label: "Серая", icon: Palette },
];

export default function SettingsPage() {
    const { theme, setTheme } = useTheme();

    const [telephonyList, setTelephonyList] = useState<CompanyTelephonyItem[]>([]);
    const [dictionary, setDictionary] = useState<TelephonyDictionaryItem[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const [formTelephonyName, setFormTelephonyName] = useState("");
    const [formExternalId, setFormExternalId] = useState("");
    const [submitLoading, setSubmitLoading] = useState(false);
    const [submitError, setSubmitError] = useState<string | null>(null);

    const [detachId, setDetachId] = useState<string | null>(null);

    function loadData() {
        setLoading(true);
        setError(null);
        Promise.all([getCompanyTelephony(), getTelephonyDictionary()])
            .then(([list, dict]) => {
                setTelephonyList(list);
                setDictionary(dict);
                if (dict.length > 0 && !formTelephonyName) setFormTelephonyName(dict[0].name ?? "");
                setLoading(false);
            })
            .catch((e) => {
                setError((e as Error).message || "Не удалось загрузить данные");
                setLoading(false);
            });
    }

    useEffect(() => {
        loadData();
    }, []);

    function handleConnect() {
        if (!formTelephonyName.trim() || !formExternalId.trim()) {
            setSubmitError("Укажите телефонию и идентификатор аккаунта");
            return;
        }
        setSubmitLoading(true);
        setSubmitError(null);
        createCompanyTelephony({
            telephony_name: formTelephonyName.trim(),
            external_account_id: formExternalId.trim(),
        })
            .then(() => {
                setFormExternalId("");
                loadData();
                setSubmitLoading(false);
            })
            .catch((e) => {
                setSubmitError((e as Error).message || "Не удалось подключить");
                setSubmitLoading(false);
            });
    }

    function handleDetach(id: string) {
        if (!window.confirm("Отключить эту телефонию от компании?")) return;
        setDetachId(id);
        deleteCompanyTelephony(id)
            .then(() => {
                loadData();
                setDetachId(null);
            })
            .catch(() => {
                setDetachId(null);
            });
    }

    return (
        <div className="p-8 max-w-2xl">
            <h1 className="text-2xl font-semibold mb-2" style={{ color: "var(--theme-text)" }}>
                Настройки
            </h1>
            <p className="text-sm mb-8" style={{ color: "var(--theme-text-muted)" }}>
                Внешний вид и параметры приложения. Подключение телефоний для приёма звонков.
            </p>

            <section className="mb-8">
                <h2 className="text-sm font-semibold uppercase tracking-wider mb-4" style={{ color: "var(--theme-text-muted)" }}>
                    Тема
                </h2>
                <div className="flex flex-wrap gap-3">
                    {THEMES.map(({ value, label, icon: Icon }) => (
                        <button
                            key={value}
                            type="button"
                            onClick={() => setTheme(value)}
                            className="flex items-center gap-2 px-4 py-3 rounded-xl border text-sm font-medium transition"
                            style={{
                                borderColor: theme === value ? "var(--theme-text)" : "var(--theme-border)",
                                backgroundColor: theme === value ? "var(--theme-active)" : "transparent",
                                color: theme === value ? "var(--theme-text)" : "var(--theme-text-muted)",
                            }}
                        >
                            <Icon className="h-4 w-4 shrink-0" />
                            {label}
                        </button>
                    ))}
                </div>
            </section>

            <section className="mb-8">
                <h2 className="text-sm font-semibold uppercase tracking-wider mb-4" style={{ color: "var(--theme-text-muted)" }}>
                    Телефонии
                </h2>
                {loading && (
                    <p className="text-sm mb-4" style={{ color: "var(--theme-text-muted)" }}>
                        Загрузка…
                    </p>
                )}
                {error && (
                    <p className="text-sm mb-4 text-red-500">
                        {error}
                        <button
                            type="button"
                            className="ml-2 underline"
                            onClick={loadData}
                        >
                            Повторить
                        </button>
                    </p>
                )}
                {!loading && !error && (
                    <>
                        <div className="rounded-xl border overflow-hidden mb-6" style={{ borderColor: "var(--theme-border)" }}>
                            <table className="min-w-full text-sm">
                                <thead style={{ backgroundColor: "var(--theme-bg-card)" }}>
                                    <tr>
                                        <th className="text-left px-4 py-2" style={{ color: "var(--theme-text-muted)" }}>
                                            Телефония
                                        </th>
                                        <th className="text-left px-4 py-2" style={{ color: "var(--theme-text-muted)" }}>
                                            ID аккаунта
                                        </th>
                                        <th className="text-left px-4 py-2" style={{ color: "var(--theme-text-muted)" }}>
                                            Подключено
                                        </th>
                                        <th className="w-24" />
                                    </tr>
                                </thead>
                                <tbody style={{ color: "var(--theme-text)" }}>
                                    {telephonyList.length === 0 && (
                                        <tr>
                                            <td colSpan={4} className="px-4 py-3 text-center" style={{ color: "var(--theme-text-muted)" }}>
                                                Нет подключённых телефоний
                                            </td>
                                        </tr>
                                    )}
                                    {telephonyList.map((item) => (
                                        <tr key={item.id} className="border-t" style={{ borderColor: "var(--theme-border)" }}>
                                            <td className="px-4 py-2">{item.telephony_name}</td>
                                            <td className="px-4 py-2 font-mono text-xs">{item.external_account_id}</td>
                                            <td className="px-4 py-2">
                                                {item.created_at ? new Date(item.created_at).toLocaleDateString() : "—"}
                                            </td>
                                            <td className="px-4 py-2">
                                                <button
                                                    type="button"
                                                    disabled={detachId === item.id}
                                                    className="text-xs text-red-600 hover:underline disabled:opacity-50"
                                                    onClick={() => handleDetach(item.id)}
                                                >
                                                    {detachId === item.id ? "…" : "Отключить"}
                                                </button>
                                            </td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>

                        <div className="rounded-xl border p-4" style={{ borderColor: "var(--theme-border)", backgroundColor: "var(--theme-bg-card)" }}>
                            <h3 className="text-sm font-medium mb-3" style={{ color: "var(--theme-text)" }}>
                                Подключить телефонию
                            </h3>
                            {submitError && (
                                <p className="text-sm text-red-500 mb-3">{submitError}</p>
                            )}
                            <div className="flex flex-wrap gap-3 items-end">
                                <label className="flex flex-col gap-1">
                                    <span className="text-xs" style={{ color: "var(--theme-text-muted)" }}>
                                        Телефония
                                    </span>
                                    <select
                                        value={formTelephonyName}
                                        onChange={(e) => setFormTelephonyName(e.target.value)}
                                        className="border rounded-lg px-3 py-2 text-sm min-w-[140px]"
                                        style={{ borderColor: "var(--theme-border)", color: "var(--theme-text)" }}
                                    >
                                        {dictionary.map((d) => (
                                            <option key={d.id} value={d.name}>
                                                {d.name}
                                            </option>
                                        ))}
                                    </select>
                                </label>
                                <label className="flex flex-col gap-1">
                                    <span className="text-xs" style={{ color: "var(--theme-text-muted)" }}>
                                        ID аккаунта (external_account_id)
                                    </span>
                                    <input
                                        type="text"
                                        value={formExternalId}
                                        onChange={(e) => setFormExternalId(e.target.value)}
                                        placeholder="Например Account SID"
                                        className="border rounded-lg px-3 py-2 text-sm min-w-[200px]"
                                        style={{ borderColor: "var(--theme-border)", color: "var(--theme-text)" }}
                                    />
                                </label>
                                <button
                                    type="button"
                                    disabled={submitLoading || dictionary.length === 0}
                                    onClick={handleConnect}
                                    className="px-4 py-2 rounded-lg text-sm font-medium border disabled:opacity-50"
                                    style={{ borderColor: "var(--theme-border)", color: "var(--theme-text)" }}
                                >
                                    {submitLoading ? "Подключение…" : "Подключить"}
                                </button>
                            </div>
                        </div>
                    </>
                )}
            </section>
        </div>
    );
}
