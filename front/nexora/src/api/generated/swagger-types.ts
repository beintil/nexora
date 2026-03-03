export interface paths {
    "/v1/auth/register": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Register a new company account
         * @description Register a new company account with email/password credentials.
         */
        post: operations["authRegister"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/auth/login": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Sign in with email and password */
        post: operations["authLogin"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        RegisterRequest: {
            companyName: string;
            /** Format: email */
            email: string;
            /** Format: password */
            password: string;
        };
        LoginRequest: {
            /** Format: email */
            email: string;
            /** Format: password */
            password: string;
        };
        LoginResponse: {
            /** @description Access token (placeholder for now). */
            accessToken: string;
            /** @description Refresh token (placeholder for now). */
            refreshToken: string;
        };
        RegisterResponse: {
            /** @description Access token (placeholder for now). */
            accessToken: string;
            /** @description Refresh token (placeholder for now). */
            refreshToken: string;
        };
        /** @description Error response for API requests */
        TransportError: {
            /**
             * @description Error type identifier
             * @example ServiceErrorUserAlreadyExists
             */
            error?: string;
            /**
             * @description Human-readable error message
             * @example user already exists: test@example.com
             */
            message: string;
            /**
             * @description Additional error details (optional)
             * @example Check email or public_id uniqueness
             */
            details?: string;
            /**
             * Format: int32
             * @description HTTP status code (e.g., 400, 409, 500)
             * @example 409
             */
            code: number;
            /**
             * Format: uuid
             * @description Unique transaction ID for tracing
             * @example 123e4567-e89b-12d3-a456-426614174000
             */
            transaction_id?: string;
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type $defs = Record<string, never>;
export interface operations {
    authRegister: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RegisterRequest"];
            };
        };
        responses: {
            /** @description Successfully registered */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["RegisterResponse"];
                };
            };
            /** @description Bad request (validation error) */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Account already exists */
            409: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Internal server error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
        };
    };
    authLogin: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["LoginRequest"];
            };
        };
        responses: {
            /** @description Login success */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LoginResponse"];
                };
            };
            /** @description Bad request (validation error) */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Invalid credentials */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Internal server error */
            500: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
        };
    };
}
