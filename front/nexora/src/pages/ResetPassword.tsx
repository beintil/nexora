import { useState, useEffect } from "react"
import type { FormEvent } from "react"
import { useLocation, useNavigate } from "react-router-dom"
import { resetPasswordConfirmRequest, type ApiError } from "../api/auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import AuthLayout from "../layouts/AuthLayout"

export default function ResetPassword() {
    const location = useLocation()
    const navigate = useNavigate()
    const [token, setToken] = useState("")
    const [newPassword, setNewPassword] = useState("")
    const [confirmPassword, setConfirmPassword] = useState("")
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState("")
    const [success, setSuccess] = useState("")

    useEffect(() => {
        const params = new URLSearchParams(location.search)
        const t = params.get("token")
        if (t) {
            setToken(t)
        } else {
            setError("Invalid or missing reset token.")
        }
    }, [location])

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError("")
        setSuccess("")

        if (!newPassword || !confirmPassword) {
            setError("Please fill in all fields.")
            return
        }

        if (newPassword !== confirmPassword) {
            setError("Passwords do not match.")
            return
        }

        if (!token) {
            setError("Invalid reset token.")
            return
        }

        try {
            setIsLoading(true)
            await resetPasswordConfirmRequest({ 
                token, 
                new_password: newPassword as any 
            })
            setSuccess("Your password has been successfully reset.")
            setTimeout(() => {
                navigate("/", { state: { openAuth: "login" } })
            }, 3000)
        } catch (submitError) {
            if (submitError instanceof Error) {
                const apiError = submitError as ApiError
                setError(apiError.message || "An error occurred. Please try again.")
            } else {
                setError("An error occurred. Please try again.")
            }
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <AuthLayout 
            title="New Password" 
            subtitle="Enter your new password below"
            onBack={() => navigate("/", { state: { openAuth: "login" } })}
        >
            {success ? (
                <div className="space-y-6">
                    <div className="p-4 rounded-xl bg-emerald-50 border border-emerald-100 text-emerald-700 text-sm leading-relaxed">
                        {success} Redirecting to login...
                    </div>
                </div>
            ) : (
                <form className="space-y-5" onSubmit={handleSubmit}>
                    <div className="space-y-4">
                        <Input
                            type="password"
                            placeholder="New Password"
                            value={newPassword}
                            onChange={(e) => setNewPassword(e.target.value)}
                            className="h-12 bg-background/50 border-input px-4 text-base"
                            disabled={isLoading || (!!error && !token)}
                        />
                        <Input
                            type="password"
                            placeholder="Confirm New Password"
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            className="h-12 bg-background/50 border-input px-4 text-base"
                            disabled={isLoading || (!!error && !token)}
                        />
                    </div>

                    {error && (
                        <div className="flex gap-3 rounded-xl border border-destructive/20 bg-destructive/10 px-4 py-3.5 text-sm text-destructive shadow-sm">
                            <p className="leading-snug">{error}</p>
                        </div>
                    )}

                    <Button
                        type="submit"
                        disabled={isLoading || !token}
                        className="w-full h-12 text-base font-medium rounded-xl"
                    >
                        {isLoading ? "Resetting..." : "Reset Password"}
                    </Button>
                </form>
            )}
        </AuthLayout>
    )
}
