import { BACKEND_URL } from "../config";
import type { components, paths } from "./generated/swagger-types";

type LoginPayload = paths["/v1/auth/login"]["post"]["requestBody"]["content"]["application/json"];
type RegisterPayload = paths["/v1/auth/register"]["post"]["requestBody"]["content"]["application/json"];
type LoginResponse = paths["/v1/auth/login"]["post"]["responses"][200]["content"]["application/json"];
type ApiErrorPayload = components["schemas"]["TransportError"];
type VerifyLinkBody = components["schemas"]["VerifyLinkRequest"];

export type ApiError = Error & {
    /** HTTP статус из backend (code в TransportError или response.status) */
    code?: number;
    /** request/transaction id для логов */
    rid?: string;
};

const defaultFetchOptions: RequestInit = {
    credentials: "include",
};

async function handleResponseError(response: Response): Promise<never> {
    const statusCode = response.status;
    let transportError: ApiErrorPayload | null = null;
    try {
        transportError = (await response.json()) as ApiErrorPayload;
    } catch {
        // backend вернул не JSON — используем fallback ниже
    }

    const backendCode = transportError?.code ?? statusCode;
    const rid = transportError?.transaction_id;

    const message =
        (transportError?.message && String(transportError.message)) ||
        `Unknown error. ${backendCode}`;

    const error: ApiError = new Error(message);
    error.code = backendCode;
    if (rid) {
        error.rid = rid;
    }

    throw error;
}

async function postJson<TPayload, TResponse>(
    path: string,
    payload: TPayload,
    options: RequestInit = {}
): Promise<TResponse> {
    const response = await fetch(`${BACKEND_URL}${path}`, {
        method: "POST",
        ...defaultFetchOptions,
        headers: {
            "Content-Type": "application/json",
            ...options.headers,
        },
        body: JSON.stringify(payload),
        ...options,
    });

    if (response.ok) {
        const contentType = response.headers.get("content-type") ?? "";
        if (response.status === 204 || !contentType.includes("application/json")) {
            return {} as TResponse;
        }
        return (await response.json()) as TResponse;
    }

    return handleResponseError(response);
}

export function loginRequest(payload: LoginPayload): Promise<LoginResponse> {
    return postJson<LoginPayload, LoginResponse>("/v1/auth/login", payload);
}

export function registerRequest(payload: RegisterPayload): Promise<void> {
    return postJson<RegisterPayload, void>("/v1/auth/register", payload);
}

/** GET verify-link: подтверждение email по токену из письма (token + email в query). Успех — 204. */
export function verifyLinkRequest(token: string, email: string): Promise<void> {
    const params = new URLSearchParams({
        token,
        email,
    });
    return getNoContent(`/v1/auth/verify-link?${params.toString()}`);
}

/** GET verify-link: подтверждение email по коду из формы (token + email в query). Успех — 204. */
export function verifyLinkByCodeRequest(token: string, email: string): Promise<void> {
    const params = new URLSearchParams({
        token,
        email,
    });
    return getNoContent(`/v1/auth/verify-link?${params.toString()}`);
}

async function getNoContent(path: string): Promise<void> {
    const response = await fetch(`${BACKEND_URL}${path}`, {
        method: "GET",
        ...defaultFetchOptions,
    });
    if (response.ok || response.status === 204) {
        return;
    }
    return handleResponseError(response);
}

/** POST send-code: переотправка ссылки верификации на email. Успех — 204. */
export function sendCodeRequest(email: string): Promise<void> {
    return postJson<{ email: string }, void>("/v1/auth/send-code", { email });
}
export function refreshRequest(refreshToken?: string): Promise<LoginResponse> {
    return postJson<{ refreshToken?: string }, LoginResponse>(
        "/v1/auth/refresh",
        { refreshToken: refreshToken ?? "" }
    );
}

/** POST logout: инвалидация refresh-токена на бэкенде. Успех — 204. */
export function logoutRequest(refreshToken?: string): Promise<void> {
    return postJson<{ refreshToken?: string }, void>(
        "/v1/auth/logout",
        { refreshToken: refreshToken ?? "" }
    );
}

/** POST forgot-password: запрос ссылки на восстановление пароля. Успех — 202 (Accepted). */
export function forgotPasswordRequest(payload: { email: string }): Promise<void> {
    return postJson<{ email: string }, void>("/v1/auth/forgot-password", payload);
}

/** POST reset-password: сброс пароля по токену из письма. Успех — 204. */
export function resetPasswordConfirmRequest(payload: { token: string; new_password: string }): Promise<void> {
    return postJson<{ token: string; new_password: string }, void>("/v1/auth/reset-password", payload);
}

/** PATCH change-password: смена пароля авторизованным пользователем. Успех — 204. */
export function changePasswordRequest(payload: { old_password: string; new_password: string }): Promise<void> {
    return postJson<{ old_password: string; new_password: string }, void>("/v1/auth/change-password", payload, {
        method: "PATCH",
    });
}
