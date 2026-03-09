import { useState, useEffect, useCallback } from "react";
import { listNotifications, markAsRead, type NotificationResponse } from "../api/notification";

export function useNotifications() {
    const [notifications, setNotifications] = useState<NotificationResponse[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchNotifications = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await listNotifications();
            setNotifications(data.notifications || []);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load notifications");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchNotifications();
        // Поллинг уведомлений каждые 30 секунд
        const interval = setInterval(fetchNotifications, 30000);
        return () => clearInterval(interval);
    }, [fetchNotifications]);

    const markRead = async (id: string) => {
        try {
            await markAsRead(id);
            setNotifications(prev => 
                prev.map(n => n.id === id ? { ...n, is_read: true } : n)
            );
        } catch (err) {
            console.error("Failed to mark notification as read", err);
        }
    };

    const unreadCount = notifications.filter(n => !n.is_read).length;

    return { notifications, unreadCount, loading, error, refetch: fetchNotifications, markRead };
}
