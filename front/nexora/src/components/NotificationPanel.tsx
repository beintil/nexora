import React from "react";
import { useNotifications } from "../hooks/useNotifications";
import { Bell, Check, Info, Users as TeamIcon, Phone, X } from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";
import { timeAgo } from "@/lib/time";

interface NotificationPanelProps {
    isOpen: boolean;
    onClose: () => void;
}

const iconByType = {
    system: Info,
    team: TeamIcon,
    call: Phone,
};

export const NotificationPanel: React.FC<NotificationPanelProps> = ({ isOpen, onClose }) => {
    const { notifications, markRead, unreadCount } = useNotifications();

    if (!isOpen) return null;

    return (
        <AnimatePresence>
            <motion.div
                initial={{ opacity: 0, y: 10, scale: 0.95 }}
                animate={{ opacity: 1, y: 0, scale: 1 }}
                exit={{ opacity: 0, y: 10, scale: 0.95 }}
                className="absolute right-0 top-14 z-50 w-80 rounded-lg border border-border bg-card p-0 shadow-xl ring-1 ring-black/5"
            >
                <div className="flex items-center justify-between border-b border-border p-3">
                    <h3 className="text-sm font-semibold">Уведомления</h3>
                    {unreadCount > 0 && (
                        <span className="rounded-full bg-primary/10 px-2 py-0.5 text-[10px] font-medium text-primary">
                            {unreadCount} новых
                        </span>
                    )}
                    <button onClick={onClose} className="rounded-md p-1 hover:bg-muted">
                        <X className="size-4 text-muted-foreground" />
                    </button>
                </div>

                <div className="max-h-[400px] overflow-y-auto overflow-x-hidden p-1 scrollbar-thin scrollbar-thumb-border">
                    {notifications.length === 0 ? (
                        <div className="flex flex-col items-center justify-center py-8 text-center">
                            <Bell className="mb-2 size-8 text-muted-foreground/20" />
                            <p className="text-xs text-muted-foreground">Уведомлений пока нет</p>
                        </div>
                    ) : (
                        notifications.map((n) => {
                            const Icon = iconByType[n.type] || Info;
                            return (
                                <div
                                    key={n.id}
                                    className={cn(
                                        "group relative flex gap-3 rounded-md p-3 transition-colors hover:bg-muted/50",
                                        !n.is_read && "bg-primary/5"
                                    )}
                                >
                                    <div className={cn(
                                        "flex size-8 shrink-0 items-center justify-center rounded-full",
                                        n.type === 'system' ? "bg-blue-500/10 text-blue-500" :
                                        n.type === 'team' ? "bg-purple-500/10 text-purple-500" :
                                        "bg-green-500/10 text-green-500"
                                    )}>
                                        <Icon className="size-4" />
                                    </div>
                                    <div className="flex flex-1 flex-col gap-1 overflow-hidden">
                                        <div className="flex items-start justify-between gap-2">
                                            <p className="text-xs font-semibold leading-tight text-foreground truncate">{n.title}</p>
                                            <span className="shrink-0 text-[10px] text-muted-foreground">
                                                {timeAgo(new Date(n.created_at))}
                                            </span>
                                        </div>
                                        <p className="text-[11px] leading-snug text-muted-foreground line-clamp-2">{n.message}</p>
                                    </div>
                                    {!n.is_read && (
                                        <button
                                            onClick={() => markRead(n.id)}
                                            className="absolute right-2 top-8 opacity-0 transition-opacity group-hover:opacity-100 hover:text-primary"
                                            title="Отметить как прочитанное"
                                        >
                                            <Check className="size-3" />
                                        </button>
                                    )}
                                </div>
                            );
                        })
                    )}
                </div>

                {notifications.length > 0 && (
                    <div className="border-t border-border p-2">
                        <button className="w-full rounded-md py-1.5 text-center text-[11px] font-medium text-muted-foreground hover:bg-muted hover:text-foreground transition-colors">
                            Показать все
                        </button>
                    </div>
                )}
            </motion.div>
        </AnimatePresence>
    );
};
