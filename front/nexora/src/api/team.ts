import { BACKEND_URL } from "../config";
import { fetchWithAuthRetry } from "./base";
import type { ProfileResponse } from "./profile";

export interface UsersListResponse {
    users: ProfileResponse[];
}

export interface CreateStaffRequest {
    email: string;
    full_name: string;
    role_id: number;
}

export async function listTeam(): Promise<UsersListResponse> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/users`, { method: "GET" });
    if (!res.ok) {
        throw new Error("Failed to list team members");
    }
    return res.json() as Promise<UsersListResponse>;
}

export async function createStaff(data: CreateStaffRequest): Promise<ProfileResponse> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/users`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    });
    if (!res.ok) {
        throw new Error("Failed to create staff member");
    }
    return res.json() as Promise<ProfileResponse>;
}

export async function deleteUser(id: string): Promise<void> {
    const res = await fetchWithAuthRetry(`${BACKEND_URL}/v1/users/${id}`, {
        method: "DELETE",
    });
    if (!res.ok) {
        throw new Error("Failed to delete user");
    }
}
