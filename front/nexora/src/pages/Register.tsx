import { useState } from "react"
import type { FormEvent } from "react"
import { useNavigate } from "react-router-dom"
import { registerRequest, type ApiError } from "../api/auth"
import { BACKEND_URL } from "../config"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import AuthLayout from "../layouts/AuthLayout"

type RegisterPageProps = {
    onBack: () => void
    onSwitch: () => void
}

export default function Register({ onBack, onSwitch }: RegisterPageProps) {
    const navigate = useNavigate()
    const [companyName, setCompanyName] = useState("")
    const [email, setEmail] = useState("")
    const [password, setPassword] = useState("")
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState("")
    const [errorRid, setErrorRid] = useState("")
    const [success, setSuccess] = useState("")
    const [showVerificationModal, setShowVerificationModal] = useState(false)

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError("")
        setErrorRid("")
        setSuccess("")

        if (!companyName || !email || !password) {
            setError("Заполните все поля.")
            return
        }

        try {
            setIsLoading(true)
            await registerRequest({ companyName, email, password })
            setSuccess("Аккаунт создан. Подтвердите почту и войдите.")
            setShowVerificationModal(true)
        } catch (submitError) {
            if (submitError instanceof Error) {
                const apiError = submitError as ApiError
                setError(apiError.message || "Не удалось создать аккаунт.")
                setErrorRid(apiError.code === 500 && apiError.rid ? apiError.rid : "")
            } else {
                setError("Не удалось создать аккаунт.")
            }
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <AuthLayout 
            title="Создайте аккаунт" 
            subtitle="Начните работу с Nexora уже сегодня"
            onBack={onBack}
        >
            <form className="space-y-5" onSubmit={handleSubmit}>
                <div className="space-y-4">
                    <Input
                        type="text"
                        placeholder="Название компании"
                        value={companyName}
                        onChange={(e) => setCompanyName(e.target.value)}
                        className="h-12 bg-background/50 border-input px-4 text-base"
                    />
                    <Input
                        type="email"
                        placeholder="Эл. почта"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="h-12 bg-background/50 border-input px-4 text-base"
                    />
                    <Input
                        type="password"
                        placeholder="Пароль"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="h-12 bg-background/50 border-input px-4 text-base"
                    />
                </div>

                {error && (
                    <div className="flex gap-3 rounded-xl border border-destructive/20 bg-destructive/10 px-4 py-3.5 text-sm text-destructive shadow-sm">
                        <span className="mt-0.5 shrink-0" aria-hidden>
                            <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
                                <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                            </svg>
                        </span>
                        <div className="min-w-0 flex-1 space-y-1">
                            <p className="leading-snug">{error}</p>
                            {errorRid && (
                                <p className="text-xs opacity-80">
                                    ID запроса: <span className="font-mono tracking-tight">{errorRid}</span>
                                </p>
                            )}
                        </div>
                    </div>
                )}
                {success && <p className="text-sm text-emerald-600 font-medium">{success}</p>}
                
                {showVerificationModal && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
                        <div className="w-full max-w-md rounded-3xl bg-card border border-border p-8 shadow-2xl">
                            <div className="mb-6 flex justify-center">
                                <div className="flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/10 text-emerald-500">
                                    <svg className="h-8 w-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                                    </svg>
                                </div>
                            </div>
                            <h3 className="text-xl font-bold text-center text-foreground mb-3">Проверьте почту</h3>
                            <p className="text-muted-foreground text-center text-sm leading-relaxed mb-6">
                                Мы отправили ссылку для подтверждения на <span className="font-semibold text-foreground">{email}</span>. Откройте письмо и перейдите по ссылке, чтобы завершить регистрацию.
                            </p>
                            <p className="text-muted-foreground text-center text-xs mb-6">
                                Не пришло письмо?{" "}
                                <button
                                    type="button"
                                    onClick={() => { setShowVerificationModal(false); navigate("/verify-link", { state: { email } }); }}
                                    className="text-primary font-semibold hover:underline"
                                >
                                    Запросить новую ссылку
                                </button>
                            </p>
                            <Button
                                type="button"
                                onClick={() => { setShowVerificationModal(false); onSwitch(); }}
                                className="w-full h-12 text-base font-medium rounded-xl"
                            >
                                OK, перейти ко входу
                            </Button>
                        </div>
                    </div>
                )}

                <Button
                    type="submit"
                    disabled={isLoading}
                    className="w-full h-12 text-base font-medium rounded-xl mt-2"
                >
                    {isLoading ? "Creating Account..." : "Create Account"}
                </Button>

                <div className="relative my-8">
                    <div className="absolute inset-0 flex items-center">
                        <div className="w-full border-t border-border" />
                    </div>
                    <div className="relative flex justify-center text-xs uppercase tracking-wider font-medium">
                        <span className="px-4 bg-card text-muted-foreground">или зарегистрируйтесь через</span>
                    </div>
                </div>

                <div className="grid grid-cols-2 gap-4">
                    <a
                        href={`${BACKEND_URL}/v1/auth/google/start`}
                        className="flex items-center justify-center gap-2 h-11 px-4 rounded-xl border border-input bg-background/50 text-foreground font-medium hover:bg-accent hover:text-accent-foreground transition-colors shadow-sm"
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
                        href={`${BACKEND_URL}/v1/auth/apple/start`}
                        className="flex items-center justify-center gap-2 h-11 px-4 rounded-xl border border-input bg-background/50 text-foreground font-medium hover:bg-accent hover:text-accent-foreground transition-colors shadow-sm"
                    >
                        <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
                            <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-1.18 1.62-2.09 3.21-3.32 4.89zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z"/>
                        </svg>
                        Apple
                    </a>
                </div>
            </form>

            <div className="text-sm text-center text-muted-foreground mt-8">
                Уже есть аккаунт?{" "}
                <button onClick={onSwitch} className="text-foreground font-semibold hover:underline transition-all">
                    Sign in
                </button>
            </div>
        </AuthLayout>
    )
}
