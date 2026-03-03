const DEFAULT_BACKEND_URL = "http://localhost:8081";

const rawBackendUrl = import.meta.env.VITE_BACKEND_URL?.trim();

export const BACKEND_URL =
    rawBackendUrl && rawBackendUrl.length > 0
        ? rawBackendUrl.replace(/\/+$/, "")
        : DEFAULT_BACKEND_URL;

