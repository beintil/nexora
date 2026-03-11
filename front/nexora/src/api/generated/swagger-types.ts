export interface paths {
    "/webhook/twilio/voice/status": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Twilio Voice status callback (webhook)
         * @description Twilio StatusCallback for call-progress-events. Тело запроса — JSON (TwilioVoiceStatusCallbackForm).
         */
        post: operations["twilioVoiceStatusCallback"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
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
    "/v1/auth/refresh": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Refresh access and refresh tokens
         * @description Accepts refresh token in body. Returns new access and refresh tokens; backend also sets refresh token in httpOnly cookie.
         */
        post: operations["authRefresh"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/auth/logout": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Log out and invalidate refresh token
         * @description Accepts refresh token in body (or from cookie). Invalidates the token in Redis and clears the cookie. Idempotent when token is missing or already invalid.
         */
        post: operations["authLogout"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/auth/send-code": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Send verification code to email or phone */
        post: operations["authSendCode"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/auth/verify-link": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Verify link (transition by link from email/SMS)
         * @description Token (UUID) передаётся в query. Ссылка приходит в сообщении.
         */
        get: operations["authVerifyLinkGet"];
        put?: never;
        /** Verify link by token in body */
        post: operations["authVerifyLinkPost"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/profile": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** Get current user profile */
        get: operations["profileGet"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        /** Update profile (e.g. full name) */
        patch: operations["profileUpdate"];
        trace?: never;
    };
    "/v1/calls": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * List company calls
         * @description Returns paginated list of calls for the current company with optional filters. Body with more than one parameter.
         */
        post: operations["callsList"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/calls/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /**
         * Get call tree by ID
         * @description Returns full call tree with events and details for the given call ID (must belong to current company).
         */
        get: operations["callGetByID"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/metrics/calls": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /**
         * Get call metrics for period
         * @description Returns aggregated call metrics and timeseries for the current company. date_from and date_to in body are required.
         */
        post: operations["metricsGetCallMetrics"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/company/telephony": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List company telephony */
        get: operations["companyTelephonyList"];
        put?: never;
        /** Attach telephony to company */
        post: operations["companyTelephonyAttach"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/company/telephony/dictionary": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        /** List telephony dictionary */
        get: operations["companyTelephonyDictionary"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/company/telephony/{id}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post?: never;
        /** Detach telephony from company */
        delete: operations["companyTelephonyDetach"];
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/profile/avatar": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        /** Upload avatar */
        post: operations["profileUploadAvatar"];
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
            email?: string;
            phone?: string;
            /** Format: password */
            password: string;
            fullName?: string;
        };
        LoginRequest: {
            /** @description Email or phone number */
            login: string;
            /** Format: password */
            password: string;
        };
        /** @description At least one of email or phone is required */
        SendCodeRequest: {
            /** Format: email */
            email?: string;
            phone?: string;
        };
        VerifyLinkRequest: {
            /**
             * Format: uuid
             * @description UUID из ссылки подтверждения
             */
            token: string;
        };
        RefreshRequest: {
            refreshToken: string;
        };
        LoginResponse: {
            accessToken: string;
            refreshToken: string;
        };
        TwilioVoiceStatusCallbackForm: {
            CallSid?: string;
            ParentCallSid?: string;
            AccountSid?: string;
            From?: string;
            To?: string;
            CallStatus?: string;
            Direction?: string;
            ApiVersion?: string;
            CallerName?: string;
            ForwardedFrom?: string;
            CallbackSource?: string;
            SequenceNumber?: string;
            Timestamp?: string;
            CallDuration?: string;
            Duration?: string;
            SipResponseCode?: string;
            RecordingSid?: string;
            RecordingUrl?: string;
            RecordingDuration?: string;
            Called?: string;
            CalledCity?: string;
            CalledCountry?: string;
            CalledState?: string;
            CalledZip?: string;
            Caller?: string;
            CallerCity?: string;
            CallerCountry?: string;
            CallerState?: string;
            CallerZip?: string;
            FromCity?: string;
            FromCountry?: string;
            FromState?: string;
            FromZip?: string;
            ToCity?: string;
            ToCountry?: string;
            ToState?: string;
            ToZip?: string;
        };
        ProfileResponse: {
            /** Format: uuid */
            id?: string;
            /** Format: uuid */
            company_id?: string;
            company_name?: string;
            /** Format: int16 */
            role_id?: number;
            /** Format: email */
            email?: string;
            phone?: string;
            full_name?: string;
            /** @description Полная ссылка на аватар в хранилище */
            avatar_url?: string;
            /** @description Идентификатор файла аватара в хранилище (UUID) */
            avatar_id?: string;
            /** Format: date-time */
            created_at?: string;
            /** Format: date-time */
            updated_at?: string;
        };
        UpdateProfileRequest: {
            full_name?: string;
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
             * @example validation error
             */
            message: string;
            /** @description Additional error details (optional) */
            details?: string;
            /**
             * Format: int32
             * @description HTTP status code (e.g., 400, 409, 500)
             * @example 400
             */
            code: number;
            /**
             * Format: uuid
             * @description Unique transaction ID for tracing
             */
            transaction_id?: string;
        };
        /** @description Pagination parameters for list requests (page is 1-based, sent by client) */
        PaginationParams: {
            /**
             * Format: int32
             * @description Page size
             */
            limit: number;
            /**
             * Format: int32
             * @description Offset for pagination
             */
            offset: number;
            /**
             * Format: int32
             * @description Current page (1-based)
             */
            page: number;
        };
        /** @description Pagination meta in list responses; total is always set by backend */
        PaginationMeta: {
            /** Format: int32 */
            limit: number;
            /** Format: int32 */
            offset: number;
            /**
             * Format: int32
             * @description Current page (1-based)
             */
            page: number;
            /**
             * Format: int32
             * @description Total number of items
             */
            total: number;
        };
        CompanyTelephonyItem: {
            /** Format: uuid */
            id?: string;
            telephony_name?: string;
            external_account_id?: string;
            /** Format: date-time */
            created_at?: string;
        };
        CompanyTelephonyListResponse: {
            items: components["schemas"]["CompanyTelephonyItem"][];
        };
        CompanyTelephonyCreateRequest: {
            telephony_name: string;
            external_account_id: string;
        };
        TelephonyDictionaryItem: {
            /** Format: int64 */
            id?: number;
            name?: string;
        };
        TelephonyDictionaryResponse: {
            items: components["schemas"]["TelephonyDictionaryItem"][];
        };
        CallMetricsRequest: {
            /**
             * Format: date
             * @description Start date (YYYY-MM-DD).
             */
            date_from: string;
            /**
             * Format: date
             * @description End date (YYYY-MM-DD).
             */
            date_to: string;
        };
        CallMetricsSummary: {
            /** Format: int32 */
            total?: number;
            /** Format: int32 */
            answered?: number;
            /** Format: int32 */
            missed?: number;
            by_direction?: {
                [key: string]: number;
            };
        };
        CallMetricsPoint: {
            /** Format: date */
            date?: string;
            /** Format: int32 */
            total?: number;
            /** Format: int32 */
            answered?: number;
            /** Format: int32 */
            missed?: number;
        };
        CallMetricsResponse: {
            summary: components["schemas"]["CallMetricsSummary"];
            timeseries: components["schemas"]["CallMetricsPoint"][];
        };
        CallsListRequest: {
            page: components["schemas"]["PaginationParams"];
            /**
             * Format: date
             * @description Filter from date (YYYY-MM-DD)
             */
            date_from?: string;
            /**
             * Format: date
             * @description Filter to date (YYYY-MM-DD)
             */
            date_to?: string;
            /** @enum {string} */
            direction?: "call_direction_inbound" | "call_direction_outbound_api" | "call_direction_outbound_dial";
            /** @description Last call event status filter */
            status?: string;
            /** Format: uuid */
            company_telephony_id?: string;
        };
        CallsListItem: {
            /** Format: uuid */
            id?: string;
            /** Format: uuid */
            company_telephony_id?: string;
            from_number?: string;
            to_number?: string;
            /** @enum {string} */
            direction?: "call_direction_inbound" | "call_direction_outbound_api" | "call_direction_outbound_dial";
            /** Format: date-time */
            created_at?: string;
            /** Format: date-time */
            updated_at?: string;
            /** @description Last call event status */
            last_status?: string;
            /** @description True if the call has child calls (can be expanded) */
            has_children?: boolean;
        };
        CallsListResponse: {
            items: components["schemas"]["CallsListItem"][];
            meta: components["schemas"]["PaginationMeta"];
        };
        CallEventResponse: {
            /** Format: uuid */
            id?: string;
            status?: string;
            /** Format: date-time */
            timestamp?: string;
        };
        CallDetailsResponse: {
            recording_sid?: string;
            recording_url?: string;
            /** Format: int32 */
            recording_duration?: number;
            from_country?: string;
            from_city?: string;
            to_country?: string;
            to_city?: string;
            carrier?: string;
            trunk?: string;
        };
        CallResponse: {
            /** Format: uuid */
            id?: string;
            /** Format: uuid */
            company_telephony_id?: string;
            /** Format: uuid */
            parent_call_id?: string;
            external_call_id?: string;
            external_parent_call_id?: string;
            from_number?: string;
            to_number?: string;
            /** @enum {string} */
            direction?: "call_direction_inbound" | "call_direction_outbound_api" | "call_direction_outbound_dial";
            /** Format: date-time */
            created_at?: string;
            /** Format: date-time */
            updated_at?: string;
            events?: components["schemas"]["CallEventResponse"][];
            details?: components["schemas"]["CallDetailsResponse"];
        };
        CallTreeResponse: {
            call?: components["schemas"]["CallResponse"];
            children?: components["schemas"]["CallTreeResponse"][];
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
    twilioVoiceStatusCallback: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["TwilioVoiceStatusCallbackForm"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
        };
    };
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
            /** @description Successfully registered. No body; confirm email then login. */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
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
    authRefresh: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["RefreshRequest"];
            };
        };
        responses: {
            /** @description New tokens */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["LoginResponse"];
                };
            };
            /** @description Invalid or expired refresh token */
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
    authLogout: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["RefreshRequest"];
            };
        };
        responses: {
            /** @description Logged out */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
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
    authSendCode: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["SendCodeRequest"];
            };
        };
        responses: {
            /** @description Code sent */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Bad request (invalid email/phone or missing both) */
            400: {
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
    authVerifyLinkGet: {
        parameters: {
            query: {
                token: string;
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Link verified */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Invalid, expired or too many attempts */
            400: {
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
    authVerifyLinkPost: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["VerifyLinkRequest"];
            };
        };
        responses: {
            /** @description Link verified */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Invalid, expired or too many attempts */
            400: {
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
    profileGet: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Profile data */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProfileResponse"];
                };
            };
            /** @description Unauthorized */
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
    profileUpdate: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["UpdateProfileRequest"];
            };
        };
        responses: {
            /** @description Updated profile */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProfileResponse"];
                };
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
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
    callsList: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["CallsListRequest"];
            };
        };
        responses: {
            /** @description List of calls with pagination meta */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CallsListResponse"];
                };
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
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
    callGetByID: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Call tree with events and details */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CallTreeResponse"];
                };
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Call not found */
            404: {
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
    metricsGetCallMetrics: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: {
            content: {
                "application/json": components["schemas"]["CallMetricsRequest"];
            };
        };
        responses: {
            /** @description Call metrics summary and timeseries */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CallMetricsResponse"];
                };
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
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
    companyTelephonyList: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description List of connected telephonies */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CompanyTelephonyListResponse"];
                };
            };
            /** @description Unauthorized */
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
    companyTelephonyAttach: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["CompanyTelephonyCreateRequest"];
            };
        };
        responses: {
            /** @description Created */
            201: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["CompanyTelephonyItem"];
                };
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Conflict (already attached) */
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
    companyTelephonyDictionary: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description Available telephony providers */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TelephonyDictionaryResponse"];
                };
            };
            /** @description Unauthorized */
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
    companyTelephonyDetach: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                id: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description No content */
            204: {
                headers: {
                    [name: string]: unknown;
                };
                content?: never;
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
            401: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Not found */
            404: {
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
    profileUploadAvatar: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "multipart/form-data": {
                    /** Format: binary */
                    avatar: string;
                };
            };
        };
        responses: {
            /** @description Profile with new avatar URL */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ProfileResponse"];
                };
            };
            /** @description Bad request */
            400: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["TransportError"];
                };
            };
            /** @description Unauthorized */
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
