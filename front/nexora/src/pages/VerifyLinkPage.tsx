import { useEffect, useState, useRef } from "react"
import { useSearchParams, useNavigate, useLocation } from "react-router-dom"
import { verifyLinkRequest, sendCodeRequest, type ApiError } from "../api/auth"
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

    const didVerifyToken = useRef<string | null>(null)

    useEffect(() => {
        if (!token || token.trim() === "") {
            setStatus("no_token")
            return
        }
        if (didVerifyToken.current === token) return
        didVerifyToken.current = token

        verifyLinkRequest(token)
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
        } catch (err) {
            if (err instanceof Error) {
                const apiErr = err as ApiError
                setResendError(apiErr.message || "Не удалось отправить ссылку.")
            }
        } finally {
            setResendLoading(false)
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
                        <h1 className="text-2xl font-semibold text-slate-900 mb-2">Запросить новую ссылку для подтверждения</h1>
                        <p className="text-slate-600 text-sm mb-6">
                            {email
                                ? "Нажмите кнопку ниже, чтобы получить новую ссылку на почту."
                                : "Перейдите по ссылке из письма для повторного запроса или войдите в аккаунт."}
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
                                    className="w-full py-4 rounded-2xl bg-slate-900 text-white font-medium hover:opacity-90 transition disabled:opacity-60 disabled:cursor-not-allowed"
                                >
                                    {resendLoading ? "Отправка..." : "Отправить новую ссылку"}
                                </button>
                            </div>
                        ) : null}
                        <button
                            type="button"
                            onClick={() => navigate("/", { state: { openAuth: "login" } })}
                            className={`w-full py-3 rounded-2xl border border-slate-300 bg-white text-slate-700 font-medium hover:bg-slate-50 transition ${email ? "mt-4" : ""}`}
                        >
                            Вернуться к входу
                        </button>
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
