import { useEffect, useState, useRef, type FormEvent } from "react"
import { useSearchParams, useNavigate, useLocation } from "react-router-dom"
import { verifyLinkRequest, verifyLinkByCodeRequest, sendCodeRequest, type ApiError } from "../api/auth"
import { CheckCircle, XCircle, Loader2, Mail } from "lucide-react"

const REDIRECT_SECONDS = 5

export default function VerifyLinkPage() {
    const [searchParams] = useSearchParams()
    const navigate = useNavigate()
    const token = searchParams.get("token")
    const emailFromQuery = searchParams.get("email")?.trim() ?? ""
    const location = useLocation()
    const emailFromState = (location.state as { email?: string } | null)?.email?.trim() ?? ""
    const email = emailFromQuery || emailFromState

    const [status, setStatus] = useState<"loading" | "success" | "error" | "no_token">("loading")
    const [errorMessage, setErrorMessage] = useState("")
    const [countdown, setCountdown] = useState(REDIRECT_SECONDS)

    const [resendLoading, setResendLoading] = useState(false)
    const [resendSuccess, setResendSuccess] = useState(false)
    const [resendError, setResendError] = useState("")

    const [code, setCode] = useState("")
    const [codeError, setCodeError] = useState("")
    const [codeLoading, setCodeLoading] = useState(false)

    const didVerifyToken = useRef<string | null>(null)

    useEffect(() => {
        if (!token || token.trim() === "") {
            setStatus("no_token")
            return
        }
        if (!email) {
            setStatus("no_token")
            return
        }
        if (didVerifyToken.current === token) return
        didVerifyToken.current = token

        verifyLinkRequest(token, email)
            .then(() => {
                setStatus("success")
            })
            .catch((err) => {
                if (err instanceof Error) {
                    const apiErr = err as ApiError
                    setStatus("error")
                    setErrorMessage(apiErr.message || "Ошибка подтверждения. Ссылка могла истечь или уже была использована.")
                }
            })

        return () => {}
    }, [token])

    useEffect(() => {
        if (status !== "success") return
        const t = setInterval(() => {
            setCountdown((prev) => {
                if (prev <= 1) {
                    clearInterval(t)
                    navigate("/", { state: { openAuth: "login" } })
                    return 0
                }
                return prev - 1
            })
        }, 1000)
        return () => clearInterval(t)
    }, [status, navigate])

    async function handleResendClick() {
        if (!email) return
        setResendError("")
        setResendSuccess(false)
        try {
            setResendLoading(true)
            await sendCodeRequest(email)
            setResendSuccess(true)
            // После успешной переотправки возвращаемся к вводу кода
            setStatus("no_token")
            setErrorMessage("")
            setCode("")
            setCodeError("")
        } catch (err) {
            if (err instanceof Error) {
                const apiErr = err as ApiError
                setResendError(apiErr.message || "Не удалось отправить ссылку.")
            }
        } finally {
            setResendLoading(false)
        }
    }

    async function handleCodeSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setCodeError("")

        const trimmed = code.trim()
        if (!trimmed) {
            setCodeError("Введите код из письма.")
            return
        }
        if (!/^[0-9]{6}$/.test(trimmed)) {
            setCodeError("Код должен состоять из 6 цифр.")
            return
        }

        if (!email) {
            setCodeError("Не удалось определить email. Перейдите по ссылке из письма или зарегистрируйтесь заново.")
            return
        }

        try {
            setCodeLoading(true)
            await verifyLinkByCodeRequest(trimmed, email)
            setStatus("success")
        } catch (err) {
            if (err instanceof Error) {
                const apiErr = err as ApiError
                setStatus("error")
                setErrorMessage(apiErr.message || "Ошибка подтверждения кода. Код мог истечь или уже быть использован.")
            }
        } finally {
            setCodeLoading(false)
        }
    }

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-100 to-slate-200 flex items-center justify-center px-6 py-12">
            <div className="w-full max-w-md rounded-[32px] overflow-hidden shadow-2xl bg-white p-10 text-center">
                {status === "loading" && (
                    <>
                        <div className="flex justify-center mb-6">
                            <Loader2 className="h-16 w-16 text-slate-400 animate-spin" />
                        </div>
                        <h1 className="text-2xl font-semibold text-slate-900 mb-2">Подтверждение аккаунта</h1>
                        <p className="text-slate-600 text-sm">Подождите...</p>
                    </>
                )}

                {status === "no_token" && (
                    <>
                        <div className="flex justify-center mb-6">
                            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-slate-100">
                                <Mail className="h-10 w-10 text-slate-600" />
                            </div>
                        </div>
                        <h1 className="text-2xl font-semibold text-slate-900 mb-2">Подтвердите аккаунт кодом</h1>
                        <p className="text-slate-600 text-sm mb-6">
                            Введите 6-значный код из письма или перейдите по ссылке в письме.
                        </p>

                        <form className="space-y-4 text-left" onSubmit={handleCodeSubmit}>
                            <div className="space-y-2">
                                <label className="block text-xs font-medium text-slate-500 uppercase tracking-wide">
                                    Код подтверждения
                                </label>
                                <input
                                    type="text"
                                    inputMode="numeric"
                                    pattern="[0-9]*"
                                    maxLength={6}
                                    autoComplete="one-time-code"
                                    value={code}
                                    onChange={(e) => {
                                        const value = e.target.value.replace(/\D/g, "").slice(0, 6)
                                        setCode(value)
                                    }}
                                    className="w-full px-4 py-3 rounded-2xl border border-slate-300 text-center text-lg tracking-[0.35em] font-semibold outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-900 bg-slate-50"
                                    placeholder="••••••"
                                />
                                {codeError && <p className="text-xs text-rose-600">{codeError}</p>}
                            </div>
                            <button
                                type="submit"
                                disabled={codeLoading || code.length !== 6}
                                className="w-full py-3 rounded-2xl bg-slate-900 text-white font-medium hover:opacity-90 transition disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                                {codeLoading ? "Проверяем код..." : "Подтвердить код"}
                            </button>
                        </form>

                        <div className="mt-8 pt-6 border-t border-slate-200">
                            <p className="text-slate-600 text-sm font-medium mb-3 flex items-center justify-center gap-2">
                                <Mail className="h-4 w-4" />
                                Не пришло письмо или ссылка истекла?
                            </p>
                            {email ? (
                                <div className="space-y-3">
                                    {resendError && <p className="text-sm text-rose-600">{resendError}</p>}
                                    {resendSuccess && (
                                        <p className="text-sm text-green-600">Новая ссылка отправлена. Проверьте почту.</p>
                                    )}
                                    <button
                                        type="button"
                                        onClick={handleResendClick}
                                        disabled={resendLoading}
                                        className="w-full py-3 rounded-2xl border border-slate-300 bg-white text-slate-700 font-medium hover:bg-slate-50 transition disabled:opacity-60 disabled:cursor-not-allowed"
                                    >
                                        {resendLoading ? "Отправка..." : "Отправить новую ссылку"}
                                    </button>
                                </div>
                            ) : (
                                <p className="text-slate-500 text-xs">
                                    Авторизуйтесь и запросите новую ссылку для подтверждения.
                                </p>
                            )}
                            <button
                                type="button"
                                onClick={() => navigate("/", { state: { openAuth: "login" } })}
                                className="mt-4 w-full py-3 rounded-2xl border border-slate-300 bg-white text-slate-700 font-medium hover:bg-slate-50 transition"
                            >
                                Вернуться к входу
                            </button>
                        </div>
                    </>
                )}

                {status === "success" && (
                    <>
                        <div className="flex justify-center mb-6">
                            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-green-100">
                                <CheckCircle className="h-12 w-12 text-green-600" />
                            </div>
                        </div>
                        <h1 className="text-2xl font-semibold text-slate-900 mb-2">Подтверждение завершено</h1>
                        <p className="text-slate-600 text-sm mb-6">
                            Ваш аккаунт подтверждён. Войдите, чтобы продолжить.
                        </p>
                        <p className="text-slate-500 text-sm">
                            Перенаправление на страницу входа через <span className="font-semibold text-slate-700">{countdown}</span> {countdown === 1 ? "секунду" : countdown < 5 ? "секунды" : "секунд"}...
                        </p>
                        <button
                            type="button"
                            onClick={() => navigate("/", { state: { openAuth: "login" } })}
                            className="mt-6 w-full py-4 rounded-2xl bg-slate-900 text-white font-medium hover:opacity-90 transition"
                        >
                            Sign in now
                        </button>
                    </>
                )}

                {status === "error" && (
                    <>
                        <div className="flex justify-center mb-6">
                            <div className="flex h-20 w-20 items-center justify-center rounded-full bg-rose-100">
                                <XCircle className="h-12 w-12 text-rose-600" />
                            </div>
                        </div>
                        <h1 className="text-2xl font-semibold text-slate-900 mb-2">Ошибка подтверждения</h1>
                        <p className="text-slate-600 text-sm mb-6">{errorMessage}</p>
                        <button
                            type="button"
                            onClick={() => navigate("/", { state: { openAuth: "login" } })}
                            className="w-full py-4 rounded-2xl bg-slate-900 text-white font-medium hover:opacity-90 transition"
                        >
                            Go to sign in
                        </button>

                        <div className="mt-8 pt-6 border-t border-slate-200">
                            <p className="text-slate-600 text-sm font-medium mb-3 flex items-center justify-center gap-2">
                                <Mail className="h-4 w-4" />
                                Не пришло письмо или ссылка истекла?
                            </p>
                            {email ? (
                                <div className="space-y-3">
                                    {resendError && <p className="text-sm text-rose-600">{resendError}</p>}
                                    {resendSuccess && (
                                        <p className="text-sm text-green-600">Новая ссылка отправлена. Проверьте почту.</p>
                                    )}
                                    <button
                                        type="button"
                                        onClick={handleResendClick}
                                        disabled={resendLoading}
                                        className="w-full py-3 rounded-2xl border border-slate-300 bg-white text-slate-700 font-medium hover:bg-slate-50 transition disabled:opacity-60 disabled:cursor-not-allowed"
                                    >
                                        {resendLoading ? "Отправка..." : "Отправить новую ссылку"}
                                    </button>
                                </div>
                            ) : (
                                <p className="text-slate-500 text-xs">Перейдите по ссылке из письма, чтобы запросить новую ссылку для подтверждения.</p>
                            )}
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
