import { useState, useEffect } from "react";
import { useLocation } from "react-router-dom";
import { PhoneCall, Zap, ShieldCheck } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import Login from "./Login"
import Register from "./Register"

export default function TelephonyApp() {
    const location = useLocation();
    const [page, setPage] = useState<"landing" | "login" | "register">("landing");
    const authOpen = page !== "landing";

    useEffect(() => {
        const state = location.state as { openAuth?: "login" | "register" } | null;
        if (state?.openAuth === "login") setPage("login");
        if (state?.openAuth === "register") setPage("register");
    }, [location.state]);

    return (
        <div className="relative overflow-hidden">
            <motion.div
                animate={{
                    opacity: authOpen ? 0 : 1,
                }}
                transition={{ duration: 0.28, ease: "easeInOut" }}
            >
                <LandingPage
                    onLogin={() => setPage("login")}
                    onRegister={() => setPage("register")}
                />
            </motion.div>

            <AnimatePresence mode="wait">
                {page !== "landing" && (
                    <motion.div
                        key={page}
                        className="absolute inset-0 z-20 bg-slate-100"
                    >
                        <motion.div
                            initial={{ opacity: 0 }}
                            animate={{ opacity: 1 }}
                            exit={{ opacity: 0 }}
                            transition={{ duration: 0.5, ease: "easeInOut" }}
                            className="h-full"
                        >
                            {page === "login" ? (
                                <Login
                                    onBack={() => setPage("landing")}
                                    onSwitch={() => setPage("register")}
                                />
                            ) : (
                                <Register
                                    onBack={() => setPage("landing")}
                                    onSwitch={() => setPage("login")}
                                />
                            )}
                        </motion.div>
                    </motion.div>
                )}
            </AnimatePresence>
        </div>
    );
}

function LandingPage({ onLogin, onRegister }: any) {
    return (
        <div className="relative min-h-screen bg-gradient-to-br from-slate-100 to-slate-200 text-slate-900 overflow-hidden">
            {/* subtle noise texture */}
            <div className="pointer-events-none absolute inset-0 opacity-[0.02] mix-blend-multiply" style={{backgroundImage:'url("https://www.transparenttextures.com/patterns/noise.png")'}} />
            <header className="flex items-center justify-between px-8 py-8 max-w-7xl mx-auto">
                <div className="text-2xl font-semibold tracking-tight">Nexora</div>
                <div className="flex gap-4">
                    <button
                        onClick={onLogin}
                        className="px-6 py-2.5 rounded-2xl border border-slate-300 bg-white hover:bg-slate-50 transition"
                    >
                        Login
                    </button>
                    <button
                        onClick={onRegister}
                        className="px-6 py-2.5 rounded-2xl bg-slate-900 text-white hover:opacity-90 transition shadow-lg"
                    >
                        Get Started
                    </button>
                </div>
            </header>

            <section className="px-8 pt-20 pb-7 max-w-5xl mx-auto text-center relative">
                <h1 className="text-5xl md:text-6xl font-semibold leading-tight mb-8 tracking-tight">
                    Ни одного пропущенного
                    <br />
                    делового звонка
                </h1>
                <p className="text-lg text-slate-600 mb-10 leading-relaxed">
                    Nexora принимает входящие звонки, восстанавливает пропущенные лиды
                    мгновенной SMS-рассылкой и превращает каждый запрос клиента в выручку.
                </p>
            </section>

            <section className="px-8 py-24">
                <div className="max-w-6xl mx-auto bg-white rounded-[32px] shadow-2xl p-12">
                    <h2 className="text-3xl font-semibold text-center mb-16">Как это работает</h2>
                    <div className="grid md:grid-cols-3 gap-12 text-center">
                        <Step
                            number="01"
                            title="Подключите номер"
                            description="Подключите корпоративный номер к Nexora за несколько минут."
                        />
                        <Step
                            number="02"
                            title="SMS при пропущенном"
                            description="Если вы не ответили — клиент сразу получает SMS от Nexora."
                        />
                        <Step
                            number="03"
                            title="Все лиды в одном месте"
                            description="История звонков и контакты — ни один клиент не теряется."
                        />
                    </div>
                </div>
            </section>

            <motion.section
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, margin: "-80px" }}
                transition={{ duration: 0.6 }}
                className="px-8 pb-24"
            >
                <div className="max-w-7xl mx-auto">
                    <h2 className="text-3xl font-semibold text-center mb-16">
                        Для растущего бизнеса
                    </h2>
                    <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
                        <Feature
                            icon={<PhoneCall size={26} />}
                            title="Учёт звонков"
                            description="Каждый входящий звонок с логами и временными метками."
                        />
                        <Feature
                            icon={<Zap size={26} />}
                            title="Мгновенное SMS"
                            description="Автоматическая отправка SMS пропустившим звонок за секунды."
                        />
                        <Feature
                            icon={<ShieldCheck size={26} />}
                            title="Надёжность"
                            description="Облачная инфраструктура для стабильной работы бизнеса."
                        />
                    </div>
                </div>
            </motion.section>

            <motion.section
                initial={{ opacity: 0, y: 50 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, margin: "-80px" }}
                transition={{ duration: 0.6 }}
                className="px-8 pb-28"
            >
                <div className="max-w-6xl mx-auto bg-white rounded-[32px] shadow-2xl p-12 text-center">
                    <h2 className="text-3xl font-semibold mb-16">
                        Почему выбирают Nexora
                    </h2>
                    <div className="grid md:grid-cols-3 gap-12">
                        <Benefit
                            title="Восстановление выручки"
                            description="Пропущенные звонки превращаются в платящих клиентов."
                        />
                        <Benefit
                            title="Скорость ответа"
                            description="Реакция на каждый пропущенный звонок за секунды — даже после работы."
                        />
                        <Benefit
                            title="Простое подключение"
                            description="Без сложных интеграций: перенаправили номер — и в работе."
                        />
                    </div>
                </div>
            </motion.section>

            <footer className="px-8 py-12 text-center text-slate-500 text-sm">
                © {new Date().getFullYear()} Nexora. Все права защищены.
            </footer>
        </div>
    );
}

function Feature({ icon, title, description }: any) {
    return (
        <div className="bg-white p-8 rounded-2xl shadow-sm hover:shadow-md transition">
            <div className="mb-4 text-blue-600">{icon}</div>
            <h3 className="font-semibold text-lg mb-2">{title}</h3>
            <p className="text-gray-600 text-sm">{description}</p>
        </div>
    );
}

function Step({ number, title, description }: any) {
    return (
        <div className="p-6">
            <div className="text-blue-600 font-bold text-xl mb-4">{number}</div>
            <h3 className="font-semibold text-lg mb-2">{title}</h3>
            <p className="text-gray-600 text-sm">{description}</p>
        </div>
    );
}

function Benefit({ title, description }: any) {
    return (
        <div className="p-8 bg-gray-50 rounded-2xl">
            <h3 className="font-semibold text-lg mb-3">{title}</h3>
            <p className="text-gray-600 text-sm">{description}</p>
        </div>
    );
}
