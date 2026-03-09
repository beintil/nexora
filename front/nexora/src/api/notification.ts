import { BACKEND_URL } from "../config";
import { fetchWithAuthRetry } from "./base";

export interface NotificationResponse {
    id: string;
    user_id: string;
    company_id: string;
    type: "system" | "team" | "call";
    title: string;
    message: string;
    is_read: boolean;
    created_at: string;
}

export interface NotificationsListResponse {
    notifications: NotificationResponse[];
}

export async function listNotifications(): Promise<NotificationsListResponse> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/notifications`, { method: "GET" });
    if (!res.ok) {
        throw new Error("Failed to list notifications");
    }
    return res.json() as Promise<NotificationsListResponse>;
}

export async function markAsRead(id: string): Promise<void> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/notifications/${id}/read`, {
        method: "POST",
    });
    if (!res.ok) {
        throw new Error("Failed to mark notification as read");
    }
}
