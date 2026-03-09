/**
 * Конфигурация пунктов меню приложения.
 * Визуальная часть (лейблы, пути, иконки) задаётся только на фронте.
 */

export interface MenuItem {
  id: string;
  label: string;
  path: string;
  icon: string;
}

/** Пункты меню по умолчанию (для навигации после входа). */
export const DEFAULT_MENU_ITEMS: MenuItem[] = [
  { id: "dashboard", label: "Главная", path: "/dashboard", icon: "home" },
  { id: "dashboards", label: "Дашборды", path: "/dashboards", icon: "layout" },
  { id: "calls", label: "Звонки", path: "/calls", icon: "phone" },
  { id: "analytics", label: "Аналитика", path: "/analytics", icon: "chart" },
  { id: "team", label: "Команда", path: "/team", icon: "users" },
  { id: "settings", label: "Настройки", path: "/settings", icon: "settings" },
];
