import React, { useState } from "react";
import { useTeam } from "../hooks/useTeam";
import { useProfileContext } from "../context/ProfileContext";
import { roleLabel } from "../api/profile";
import { 
    Users, 
    UserPlus, 
    Trash2, 
    Mail, 
    Shield, 
    Search,
    Loader2,
    Plus,
    X,
    CheckCircle2
} from "lucide-react";
import { motion, AnimatePresence } from "framer-motion";
import { cn } from "@/lib/utils";

export default function TeamPage() {
    const { members, loading, error, addMember, removeMember } = useTeam();
    const { profile } = useProfileContext();
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState("");
    
    // Form state
    const [newMember, setNewMember] = useState({
        email: "",
        full_name: "",
        role_id: 3 // Default to Manager
    });
    const [isSubmitting, setIsSubmitting] = useState(false);

    const isOwner = profile?.role_id === 2; // Role Owner

    const filteredMembers = members.filter(m => 
        m.full_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
        m.email?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const handleAddMember = async (e: React.FormEvent) => {
        e.preventDefault();
        setIsSubmitting(true);
        try {
            await addMember(newMember);
            setIsAddModalOpen(false);
            setNewMember({ email: "", full_name: "", role_id: 3 });
        } catch (err) {
            console.error(err);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="flex flex-col gap-8 pb-10">
            {/* Header section with Glassmorphism */}
            <div className="relative overflow-hidden rounded-2xl bg-gradient-to-br from-primary/10 via-background to-secondary/10 p-8 border border-border/50">
                <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
                    <div>
                        <h1 className="text-3xl font-bold tracking-tight text-foreground">Управление командой</h1>
                        <p className="mt-2 text-muted-foreground max-w-xl">
                            Добавляйте новых сотрудников, управляйте ролями и контролируйте доступ к вашей платформе Nexora.
                        </p>
                    </div>
                    {isOwner && (
                        <button 
                            onClick={() => setIsAddModalOpen(true)}
                            className="flex items-center gap-2 rounded-xl bg-primary px-5 py-2.5 text-sm font-semibold text-primary-foreground shadow-lg shadow-primary/20 hover:bg-primary/90 transition-all hover:scale-[1.02] active:scale-[0.98]"
                        >
                            <UserPlus className="size-4" />
                            Добавить коллегу
                        </button>
                    )}
                </div>
                {/* Decorative elements */}
                <div className="absolute -right-20 -top-20 size-64 rounded-full bg-primary/5 blur-3xl" />
                <div className="absolute -left-20 -bottom-20 size-64 rounded-full bg-secondary/10 blur-3xl" />
            </div>

            {/* Main Content Area */}
            <div className="flex flex-col gap-4">
                <div className="flex h-12 items-center justify-between gap-4 px-2">
                    <div className="relative flex-1 max-w-sm">
                        <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                        <input 
                            type="text" 
                            placeholder="Поиск по имени или email..."
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                            className="h-10 w-full rounded-xl border border-border bg-background pl-10 pr-4 text-sm transition-all focus:border-primary/50 focus:ring-4 focus:ring-primary/5 outline-none"
                        />
                    </div>
                    <div className="flex items-center gap-2 text-xs text-muted-foreground whitespace-nowrap">
                        <Users className="size-4" />
                        Всего участников: <span className="font-semibold text-foreground">{members.length}</span>
                    </div>
                </div>

                {error && (
                    <div className="rounded-xl bg-red-500/10 p-4 text-sm text-red-500 border border-red-500/20">
                        {error}
                    </div>
                )}

                {loading ? (
                    <div className="flex h-64 flex-col items-center justify-center rounded-2xl border border-dashed border-border bg-muted/20">
                        <Loader2 className="size-10 animate-spin text-primary/40" />
                        <p className="mt-4 text-sm text-muted-foreground">Загрузка списка команды...</p>
                    </div>
                ) : (
                    <motion.div 
                        variants={{
                            hidden: { opacity: 0 },
                            show: {
                                opacity: 1,
                                transition: {
                                    staggerChildren: 0.05
                                }
                            }
                        }}
                        initial="hidden"
                        animate="show"
                        className="grid gap-4 md:grid-cols-2 lg:grid-cols-3"
                    >
                        <AnimatePresence mode="popLayout">
                            {filteredMembers.map((member) => (
                                <motion.div
                                    key={member.id}
                                    layout
                                    variants={{
                                        hidden: { opacity: 0, y: 20, scale: 0.95 },
                                        show: { 
                                            opacity: 1, 
                                            y: 0, 
                                            scale: 1,
                                            transition: { type: "spring", stiffness: 300, damping: 30 }
                                        }
                                    }}
                                    initial="hidden"
                                    animate="show"
                                    exit={{ opacity: 0, scale: 0.95, transition: { duration: 0.2 } }}
                                    className="group relative flex flex-col overflow-hidden rounded-2xl border border-border bg-card p-5 transition-all hover:shadow-xl hover:shadow-primary/5 hover:border-primary/20"
                                >
                                    <div className="flex items-start justify-between">
                                        <div className="size-12 rounded-2xl bg-gradient-to-br from-primary/20 to-secondary/20 p-0.5 shadow-inner">
                                            <div className="flex size-full items-center justify-center rounded-xl bg-card">
                                                {member.avatar_url ? (
                                                    <img src={member.avatar_url} alt="" className="size-full rounded-xl object-cover" />
                                                ) : (
                                                    <span className="text-xl font-bold text-primary/40">
                                                        {(member.full_name || member.email || "U")[0].toUpperCase()}
                                                    </span>
                                                )}
                                            </div>
                                        </div>
                                        <div className={cn(
                                            "rounded-full px-2.5 py-1 text-[10px] font-bold uppercase tracking-wider",
                                            member.role_id === 2 ? "bg-red-500/10 text-red-500" :
                                            member.role_id === 0 ? "bg-blue-500/10 text-blue-500" :
                                            "bg-indigo-500/10 text-indigo-500"
                                        )}>
                                            {roleLabel(member.role_id)}
                                        </div>
                                    </div>

                                    <div className="mt-4 flex-1">
                                        <h3 className="font-bold text-foreground truncate">{member.full_name || "Без имени"}</h3>
                                        <div className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
                                            <Mail className="size-3" />
                                            <span className="truncate">{member.email}</span>
                                        </div>
                                    </div>

                                    <div className="mt-6 flex items-center justify-between pt-4 border-t border-border/50">
                                        <div className="flex items-center gap-1.5 text-[10px] text-muted-foreground">
                                            <CheckCircle2 className="size-3 text-green-500" />
                                            Активен
                                        </div>
                                        {isOwner && member.id !== profile?.id && (
                                            <button 
                                                onClick={() => removeMember(member.id)}
                                                className="rounded-lg p-2 text-muted-foreground hover:bg-red-500/10 hover:text-red-500 transition-colors"
                                                title="Удалить пользователя"
                                            >
                                                <Trash2 className="size-4" />
                                            </button>
                                        )}
                                    </div>
                                </motion.div>
                            ))}
                        </AnimatePresence>

                        {filteredMembers.length === 0 && searchQuery && (
                            <div className="col-span-full flex flex-col items-center justify-center py-20 text-center">
                                <Search className="mb-4 size-12 text-muted-foreground/20" />
                                <h3 className="text-lg font-semibold">Участники не найдены</h3>
                                <p className="text-sm text-muted-foreground">По вашему запросу "{searchQuery}" ничего не найдено.</p>
                            </div>
                        )}
                    </motion.div>
                )}
            </div>

            {/* Add Member Modal */}
            <AnimatePresence>
                {isAddModalOpen && (
                    <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-background/80 backdrop-blur-sm">
                        <motion.div
                            initial={{ opacity: 0, scale: 0.9, y: 20 }}
                            animate={{ opacity: 1, scale: 1, y: 0 }}
                            exit={{ opacity: 0, scale: 0.9, y: 20 }}
                            className="w-full max-w-md overflow-hidden rounded-3xl border border-border bg-card shadow-2xl"
                        >
                            <div className="relative p-8">
                                <button 
                                    onClick={() => setIsAddModalOpen(false)}
                                    className="absolute right-6 top-6 rounded-xl p-2 text-muted-foreground hover:bg-muted"
                                >
                                    <X className="size-5" />
                                </button>
                                
                                <div className="flex flex-col items-center text-center">
                                    <div className="mb-4 flex size-16 items-center justify-center rounded-3xl bg-primary/10 text-primary">
                                        <UserPlus className="size-8" />
                                    </div>
                                    <h2 className="text-2xl font-bold">Новый коллега</h2>
                                    <p className="mt-2 text-sm text-muted-foreground">Пригласите нового сотрудника в ваше рабочее пространство.</p>
                                </div>

                                <form onSubmit={handleAddMember} className="mt-8 flex flex-col gap-5">
                                    <div className="space-y-2">
                                        <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground ml-1">Имя и фамилия</label>
                                        <div className="relative">
                                            <Users className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                            <input 
                                                required
                                                type="text" 
                                                value={newMember.full_name}
                                                onChange={(e) => setNewMember({...newMember, full_name: e.target.value})}
                                                placeholder="Александр Иванов"
                                                className="h-11 w-full rounded-xl border border-border bg-background pl-10 pr-4 text-sm outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/5 transition-all"
                                            />
                                        </div>
                                    </div>

                                    <div className="space-y-2">
                                        <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground ml-1">Электронная почта</label>
                                        <div className="relative">
                                            <Mail className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                            <input 
                                                required
                                                type="email" 
                                                value={newMember.email}
                                                onChange={(e) => setNewMember({...newMember, email: e.target.value})}
                                                placeholder="alex@company.com"
                                                className="h-11 w-full rounded-xl border border-border bg-background pl-10 pr-4 text-sm outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/5 transition-all"
                                            />
                                        </div>
                                    </div>

                                    <div className="space-y-2">
                                        <label className="text-xs font-semibold uppercase tracking-wider text-muted-foreground ml-1">Роль в системе</label>
                                        <div className="relative">
                                            <Shield className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                                            <select 
                                                value={newMember.role_id}
                                                onChange={(e) => setNewMember({...newMember, role_id: Number(e.target.value)})}
                                                className="h-11 w-full appearance-none rounded-xl border border-border bg-background pl-10 pr-4 text-sm outline-none focus:border-primary/50 focus:ring-4 focus:ring-primary/5 transition-all"
                                            >
                                                <option value={3}>Менеджер</option>
                                                <option value={2}>Владелец</option>
                                            </select>
                                        </div>
                                    </div>

                                    <button 
                                        disabled={isSubmitting}
                                        type="submit"
                                        className="mt-4 flex h-12 w-full items-center justify-center gap-2 rounded-2xl bg-primary font-bold text-primary-foreground shadow-lg shadow-primary/20 hover:bg-primary/90 transition-all disabled:opacity-50"
                                    >
                                        {isSubmitting ? <Loader2 className="size-5 animate-spin" /> : <Plus className="size-5" />}
                                        Создать аккаунт
                                    </button>
                                </form>
                            </div>
                        </motion.div>
                    </div>
                )}
            </AnimatePresence>
        </div>
    );
}
