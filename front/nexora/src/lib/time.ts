/**
 * Форматирует дату в относительное время (например, "5 мин. назад").
 */
export function timeAgo(date: Date | string): string {
    const d = typeof date === 'string' ? new Date(date) : date;
    const seconds = Math.floor((new Date().getTime() - d.getTime()) / 1000);
    
    if (seconds < 60) return "только что";
    
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes} мин. назад`;
    
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} ч. назад`;
    
    const days = Math.floor(hours / 24);
    if (days < 7) return `${days} дн. назад`;
    
    return d.toLocaleDateString("ru-RU", { day: "numeric", month: "short" });
}
