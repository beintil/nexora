import { Routes, Route, Navigate } from "react-router-dom"
import MainPage from "./pages/MainPage"
import VerifyLinkPage from "./pages/VerifyLinkPage"
import AppLayout from "./layouts/AppLayout"
import DashboardPage from "./pages/DashboardPage"
import DashboardsListPage from "./pages/DashboardsListPage"
import CallsPage from "./pages/CallsPage"
import SettingsPage from "./pages/SettingsPage"
import AnalyticsPage from "./pages/AnalyticsPage"
import ProfileEditPage from "./pages/ProfileEditPage"

function App() {
    return (
        <>
            <Routes>
                <Route path="/" element={<MainPage />} />
                <Route path="/verify-link" element={<VerifyLinkPage />} />
                <Route element={<AppLayout />}>
                    <Route path="/dashboard" element={<DashboardPage />} />
                    <Route path="/dashboards" element={<DashboardsListPage />} />
                    <Route path="/calls" element={<CallsPage />} />
                    <Route path="/settings" element={<SettingsPage />} />
                    <Route path="/analytics" element={<AnalyticsPage />} />
                    <Route path="/profile" element={<ProfileEditPage />} />
                </Route>
                <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
        </>
    )
}

export default App
