import { useState, useEffect, useCallback } from "react";
import { listTeam, createStaff, deleteUser, type CreateStaffRequest } from "../api/team";
import type { ProfileResponse } from "../api/profile";

export function useTeam() {
    const [members, setMembers] = useState<ProfileResponse[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    const fetchTeam = useCallback(async () => {
        setLoading(true);
        setError(null);
        try {
            const data = await listTeam();
            setMembers(data.users);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load team");
        } finally {
            setLoading(false);
        }
    }, []);

    useEffect(() => {
        fetchTeam();
    }, [fetchTeam]);

    const addMember = async (data: CreateStaffRequest) => {
        setError(null);
        try {
            const newUser = await createStaff(data);
            setMembers(prev => [...prev, newUser]);
            return newUser;
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to add member");
            throw err;
        }
    };

    const removeMember = async (id: string) => {
        setError(null);
        try {
            await deleteUser(id);
            setMembers(prev => prev.filter(m => m.id !== id));
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to delete member");
            throw err;
        }
    };

    return { members, loading, error, refetch: fetchTeam, addMember, removeMember };
}
