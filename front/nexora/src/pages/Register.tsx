import { useState } from "react"
import type { FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { registerRequest, type ApiError } from "../api/auth"
import { useAuth } from "../context/AuthContext"
import { BACKEND_URL } from "../config"

type RegisterPageProps = {
    onBack: () => void
    onSwitch: () => void
}

export default function Register({ onBack, onSwitch }: RegisterPageProps) {
    const navigate = useNavigate()
    const { setAccessToken } = useAuth()
    const [companyName, setCompanyName] = useState("")
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState("")
    const [errorRid, setErrorRid] = useState("")
    const [errorCode, setErrorCode] = useState<number | undefined>()
    const [success, setSuccess] = useState("")
    const [showVerificationModal, setShowVerificationModal] = useState(false)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError("")
        setErrorRid("")
        setErrorCode(undefined)
        setSuccess("")

        if (!companyName || !email || !password) {
            setError("Заполните все поля.")
            return
        }

        try {
            setIsLoading(true)
            const data = await registerRequest({ companyName, email, password })
            setAccessToken(data.accessToken)
            setSuccess("Аккаунт успешно создан.")
            setShowVerificationModal(true)
        } catch (submitError) {
            if (submitError instanceof Error) {
                const apiError = submitError as ApiError
                setError(apiError.message || "Не удалось создать аккаунт.")
                setErrorCode(apiError.code)
                setErrorRid(apiError.code === 500 && apiError.rid ? apiError.rid : "")
            } else {
                setError("Не удалось создать аккаунт.")
            }
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-100 to-slate-200 flex items-center justify-center px-6 py-12">
            <div className="w-full max-w-5xl grid md:grid-cols-2 rounded-[32px] overflow-hidden shadow-2xl bg-white">

                <div className="hidden md:flex flex-col justify-between bg-gradient-to-br from-slate-900 via-slate-800 to-slate-700 text-white p-12">
                    <div>
                        <button onClick={onBack} className="text-3xl font-semibold tracking-tight">
                            Nexora
                        </button>
                        <p className="mt-8 text-slate-300 leading-relaxed text-lg">
                            Стройте умные процессы и не упускайте возможности.
                        </p>
                    </div>

                    <div className="space-y-4 text-slate-400 text-sm">
                        <p>• Надёжная облачная архитектура</p>
                        <p>• Автоматическое SMS-восстановление</p>
                        <p>• Масштабируемость для бизнеса</p>
                    </div>
                </div>

                <div className="p-10 md:p-14 flex flex-col justify-center">
                    <h2 className="text-3xl font-semibold mb-2">Создайте аккаунт</h2>
                    <p className="text-slate-500 mb-8">Начните работу с Nexora уже сегодня</p>

                    <form className="space-y-5" onSubmit={handleSubmit}>
                        <input
                            type="text"
                            placeholder="Название компании"
                            value={companyName}
                            onChange={(event) => setCompanyName(event.target.value)}
                            className="w-full px-5 py-4 rounded-2xl border border-slate-300 bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-900 transition"
                        />
                        <input
                            type="email"
                            placeholder="Эл. почта"
                            value={email}
                            onChange={(event) => setEmail(event.target.value)}
                            className="w-full px-5 py-4 rounded-2xl border border-slate-300 bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-900 transition"
                        />
                        <input
                            type="password"
                            placeholder="Пароль"
                            value={password}
                            onChange={(event) => setPassword(event.target.value)}
                            className="w-full px-5 py-4 rounded-2xl border border-slate-300 bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-900 transition"
                        />
                        {error && (
                            <div className="flex gap-3 rounded-2xl border border-rose-200/80 bg-rose-50/90 px-4 py-3.5 text-sm text-rose-900 shadow-sm">
                                <span className="mt-0.5 shrink-0 text-rose-500" aria-hidden>
                                    <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                                        <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                                    </svg>
                                </span>
                                <div className="min-w-0 flex-1 space-y-1">
                                    <p className="leading-snug">{error}</p>
                                    {errorRid && (
                                        <p className="text-xs text-rose-700/80">
                                            ID запроса: <span className="font-mono tracking-tight">{errorRid}</span>
                                        </p>
                                    )}
                                </div>
                            </div>
                        )}
                        {success && <p className="text-sm text-green-600">{success}</p>}
                        {showVerificationModal && (
                            <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
                                <div className="w-full max-w-md rounded-2xl bg-white p-8 shadow-xl">
                                    <div className="mb-6 flex justify-center">
                                        <div className="flex h-14 w-14 items-center justify-center rounded-full bg-green-100">
                                            <svg className="h-7 w-7 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                                            </svg>
                                        </div>
                                    </div>
                                    <h3 className="text-xl font-semibold text-center text-slate-900 mb-2">Проверьте почту</h3>
                                    <p className="text-slate-600 text-center text-sm leading-relaxed mb-6">
                                        Мы отправили ссылку для подтверждения на <span className="font-medium text-slate-800">{email}</span>. Откройте письмо и перейдите по ссылке, чтобы завершить регистрацию.
                                    </p>
                                    <p className="text-slate-500 text-center text-xs mb-4">
                                        Не пришло письмо?{" "}
                                        <button
                                            type="button"
                                            onClick={() => { setShowVerificationModal(false); navigate("/verify-link", { state: { email } }); }}
                                            className="text-slate-900 font-medium hover:underline"
                                        >
                                            Запросить новую ссылку
                                        </button>
                                    </p>
                                    <button
                                        type="button"
                                        onClick={() => { setShowVerificationModal(false); onSwitch(); }}
                                        className="w-full py-4 rounded-2xl bg-slate-900 text-white font-medium hover:opacity-90 transition"
                                    >
                                        OK, go to sign in
                                    </button>
                                </div>
                            </div>
                        )}
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full py-4 rounded-2xl bg-slate-900 text-white font-medium hover:opacity-90 transition disabled:opacity-60 disabled:cursor-not-allowed"
                        >
                            {isLoading ? "Creating Account..." : "Create Account"}
                        </button>

                        <div className="relative my-4">
                            <div className="absolute inset-0 flex items-center">
                                <div className="w-full border-t border-slate-200" />
                            </div>
                            <div className="relative flex justify-center text-sm">
                                <span className="px-3 bg-white text-slate-500">или войдите через</span>
                            </div>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                            <a
                                href={`${BACKEND_URL}/v1/auth/google`}
                                className="flex items-center justify-center gap-2 py-3 px-4 rounded-2xl border border-slate-300 bg-white text-slate-700 font-medium hover:bg-slate-50 transition"
                            >
                                <svg className="w-5 h-5" viewBox="0 0 24 24">
                                    <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                                    <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                                    <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                                    <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
                                </svg>
                                Google
                            </a>
                            <a
                                href={`${BACKEND_URL}/v1/auth/apple`}
                                className="flex items-center justify-center gap-2 py-3 px-4 rounded-2xl border border-slate-300 bg-white text-slate-700 font-medium hover:bg-slate-50 transition"
                            >
                                <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-1.18 1.62-2.09 3.21-3.32 4.89zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/>
                                </svg>
                                Apple
                            </a>
                        </div>
                    </form>

                    <div className="text-sm text-slate-500 mt-6">
                        Уже есть аккаунт?{" "}
                        <button onClick={onSwitch} className="text-slate-900 font-medium hover:underline">
                            Sign in
                        </button>
                    </div>
                </div>

            </div>
        </div>
    )
}
