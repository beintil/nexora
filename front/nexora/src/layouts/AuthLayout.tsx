import type { ReactNode } from "react"

interface AuthLayoutProps {
    children: ReactNode
    title: string
    subtitle: string
    onBack?: () => void
    brandName?: string
}

export default function AuthLayout({
    children,
    title,
    subtitle,
    onBack,
    brandName = "Nexora"
}: AuthLayoutProps) {
    return (
        <div className="min-h-screen bg-muted/20 flex items-center justify-center p-4 sm:p-6 md:p-12 font-sans">
            <div className="w-full max-w-5xl grid md:grid-cols-2 rounded-3xl overflow-hidden shadow-premium bg-card border border-border/50">
                
                {/* Left Side: Brand & Features */}
                <div className="hidden md:flex flex-col justify-between bg-primary text-primary-foreground p-12 relative overflow-hidden">
                    <div className="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none">
                        <div className="absolute top-[-20%] left-[-10%] w-[60%] h-[60%] rounded-full bg-white/5 blur-3xl mix-blend-overlay" />
                        <div className="absolute bottom-[-20%] right-[-10%] w-[60%] h-[60%] rounded-full bg-blue-500/20 blur-3xl mix-blend-overlay" />
                    </div>

                    <div className="relative z-10">
                        {onBack ? (
                            <button 
                                onClick={onBack} 
                                className="text-3xl font-bold tracking-tight hover:opacity-80 transition-opacity"
                            >
                                {brandName}
                            </button>
                        ) : (
                            <div className="text-3xl font-bold tracking-tight">
                                {brandName}
                            </div>
                        )}
                        <p className="mt-8 text-primary-foreground/80 leading-relaxed text-lg max-w-[280px]">
                            Secure infrastructure and automation for modern business.
                        </p>
                    </div>

                    <div className="relative z-10 space-y-5 text-primary-foreground/70 text-sm font-medium">
                        <p className="flex items-center gap-3">
                            <span className="flex size-6 items-center justify-center rounded-full bg-primary-foreground/10 text-primary-foreground">✓</span>
                            Enterprise-grade security
                        </p>
                        <p className="flex items-center gap-3">
                            <span className="flex size-6 items-center justify-center rounded-full bg-primary-foreground/10 text-primary-foreground">✓</span>
                            Scalable cloud architecture
                        </p>
                        <p className="flex items-center gap-3">
                            <span className="flex size-6 items-center justify-center rounded-full bg-primary-foreground/10 text-primary-foreground">✓</span>
                            24/7 Monitoring & Support
                        </p>
                    </div>
                </div>

                {/* Right Side: Form */}
                <div className="p-8 md:p-14 flex flex-col justify-center bg-card">
                    <div className="mb-8">
                        <h2 className="text-3xl font-bold tracking-tight text-foreground mb-2">{title}</h2>
                        <p className="text-muted-foreground">{subtitle}</p>
                    </div>
                    {children}
                </div>
            </div>
        </div>
    )
}
