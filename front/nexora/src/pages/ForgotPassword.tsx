import { useState } from "react"
import type { FormEvent } from "react"
import { forgotPasswordRequest, type ApiError } from "../api/auth"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import AuthLayout from "../layouts/AuthLayout"

interface ForgotPasswordProps {
    onBack: () => void
}

export default function ForgotPassword({ onBack }: ForgotPasswordProps) {
    const [email, setEmail] = useState("")
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState("")
    const [success, setSuccess] = useState("")

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault()
        setError("")
        setSuccess("")

        if (!email) {
            setError("Please enter your email address.")
            return
        }

        try {
            setIsLoading(true)
            await forgotPasswordRequest({ email })
            setSuccess("If an account exists for that email, we have sent a password reset link.")
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
            title="Reset Password" 
            subtitle="Enter your email to receive a reset link"
            onBack={onBack}
        >
            {success ? (
                <div className="space-y-6">
                    <div className="p-4 rounded-xl bg-emerald-50 border border-emerald-100 text-emerald-700 text-sm leading-relaxed">
                        {success}
                    </div>
                    <Button 
                        onClick={onBack}
                        className="w-full h-12 text-base font-medium rounded-xl"
                    >
                        Back to Login
                    </Button>
                </div>
            ) : (
                <form className="space-y-5" onSubmit={handleSubmit}>
                    <div className="space-y-4">
                        <Input
                            type="email"
                            placeholder="Email address"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            className="h-12 bg-background/50 border-input px-4 text-base"
                            disabled={isLoading}
                        />
                    </div>

                    {error && (
                        <div className="flex gap-3 rounded-xl border border-destructive/20 bg-destructive/10 px-4 py-3.5 text-sm text-destructive shadow-sm">
                            <p className="leading-snug">{error}</p>
                        </div>
                    )}

                    <div className="space-y-3 pt-2">
                        <Button
                            type="submit"
                            disabled={isLoading}
                            className="w-full h-12 text-base font-medium rounded-xl"
                        >
                            {isLoading ? "Sending..." : "Send Reset Link"}
                        </Button>
                        <Button
                            type="button"
                            variant="ghost"
                            onClick={onBack}
                            className="w-full h-12 text-base font-medium rounded-xl"
                            disabled={isLoading}
                        >
                            Back to Login
                        </Button>
                    </div>
                </form>
            )}
        </AuthLayout>
    )
}
