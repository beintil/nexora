import { lazy, Suspense } from "react"
import { Routes, Route, Navigate } from "react-router-dom"
import { Loader2 } from "lucide-react"

import MainPage from "./pages/MainPage"
import VerifyLinkPage from "./pages/VerifyLinkPage"
import ResetPassword from "./pages/ResetPassword"
import AppLayout from "./layouts/AppLayout"
import DashboardPage from "./pages/DashboardPage"
import CallsPage from "./pages/CallsPage"

const DashboardsListPage = lazy(() => import("./pages/DashboardsListPage"))
const CallDetailPage = lazy(() => import("./pages/CallDetailPage"))
const SettingsPage = lazy(() => import("./pages/SettingsPage"))
const AnalyticsPage = lazy(() => import("./pages/AnalyticsPage"))
const ProfileEditPage = lazy(() => import("./pages/ProfileEditPage"))
const TeamPage = lazy(() => import("./pages/TeamPage"))

const PageLoader = () => (
    <div className="flex h-screen w-screen items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
    </div>
)

function App() {
    return (
        <Suspense fallback={<PageLoader />}>
            <Routes>
                <Route path="/" element={<MainPage />} />
                <Route path="/verify-link" element={<VerifyLinkPage />} />
                <Route path="/reset-password" element={<ResetPassword />} />
                <Route element={<AppLayout />}>
                    <Route path="/dashboard" element={<DashboardPage />} />
                    <Route path="/dashboards" element={<DashboardsListPage />} />
                    <Route path="/calls" element={<CallsPage />} />
                    <Route path="/calls/:id" element={<CallDetailPage />} />
                    <Route path="/settings" element={<SettingsPage />} />
                    <Route path="/analytics" element={<AnalyticsPage />} />
                    <Route path="/profile" element={<ProfileEditPage />} />
                    <Route path="/team" element={<TeamPage />} />
                </Route>
                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </Suspense>
    )
}

export default App
