import { useState, useRef, useEffect, useCallback, type FormEvent } from "react"
import { Link } from "react-router-dom"
import Cropper from "react-easy-crop"
import type { Area } from "react-easy-crop"
import "react-easy-crop/react-easy-crop.css"
import { useProfileContext } from "../context/ProfileContext"
import { updateProfile, uploadAvatar } from "../api/profile"
import { getCroppedImg } from "../utils/cropImage"
import { ArrowLeft, Loader2, User, Camera, CheckCircle2, AlertCircle } from "lucide-react"

const ALLOWED_AVATAR_TYPES = ["image/jpeg", "image/png", "image/webp"]

export default function ProfileEditPage() {
    const { profile, loading: profileLoading, refetch } = useProfileContext()
    const [fullName, setFullName] = useState(profile?.full_name ?? "")
    const [avatarFile, setAvatarFile] = useState<File | null>(null)
    const [avatarPreview, setAvatarPreview] = useState<string | null>(profile?.avatar_url ?? null)
    const [saving, setSaving] = useState(false)
    const [success, setSuccess] = useState(false)
    const [error, setError] = useState("")
    const [avatarPreviewError, setAvatarPreviewError] = useState(false)
    const fileInputRef = useRef<HTMLInputElement>(null)

    const [cropModalOpen, setCropModalOpen] = useState(false)
    const [cropImageUrl, setCropImageUrl] = useState<string | null>(null)
    const [cropOriginalFile, setCropOriginalFile] = useState<File | null>(null)
    const [crop, setCrop] = useState({ x: 0, y: 0 })
    const [zoom, setZoom] = useState(1)
    const [rotation, setRotation] = useState(0)
    const [lastCroppedAreaPixels, setLastCroppedAreaPixels] = useState<Area | null>(null)
    const [cropApplying, setCropApplying] = useState(false)

    useEffect(() => {
        setAvatarPreviewError(false)
    }, [avatarPreview])

    useEffect(() => {
        if (!profile) return
        setFullName(profile.full_name ?? "")
        setAvatarPreview(profile.avatar_url ?? null)
    }, [profile?.id, profile?.full_name, profile?.avatar_url])

    const processFile = (file: File | null) => {
        setError("")
        if (!file) {
            setAvatarFile(null)
            setAvatarPreview(profile?.avatar_url ?? null)
            return
        }
        if (!ALLOWED_AVATAR_TYPES.includes(file.type)) {
            setError("Выберите изображение: JPEG, PNG или WebP.")
            return
        }
        if (file.size > 5 * 1024 * 1024) {
            setError("Размер файла не более 5 МБ.")
            return
        }
        setCropOriginalFile(file)
        setCropImageUrl(URL.createObjectURL(file))
        setCrop({ x: 0, y: 0 })
        setZoom(1)
        setRotation(0)
        setLastCroppedAreaPixels(null)
        setCropModalOpen(true)
    }

    const onFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        processFile(e.target.files?.[0] ?? null)
    }

    const onDrop = (e: React.DragEvent) => {
        e.preventDefault()
        processFile(e.dataTransfer.files?.[0] ?? null)
    }
    const onDragOver = (e: React.DragEvent) => e.preventDefault()

    const onCropComplete = useCallback((_: Area, croppedAreaPixels: Area) => {
        setLastCroppedAreaPixels(croppedAreaPixels)
    }, [])

    const closeCropModal = useCallback(() => {
        setCropModalOpen(false)
        if (cropImageUrl) URL.revokeObjectURL(cropImageUrl)
        setCropImageUrl(null)
        setCropOriginalFile(null)
        setLastCroppedAreaPixels(null)
        if (fileInputRef.current) fileInputRef.current.value = ""
    }, [cropImageUrl])

    const applyCrop = useCallback(async () => {
        if (!cropImageUrl || !cropOriginalFile || !lastCroppedAreaPixels) return
        setCropApplying(true)
        try {
            const mimeType = cropOriginalFile.type || "image/jpeg"
            const blob = await getCroppedImg(cropImageUrl, lastCroppedAreaPixels, rotation, mimeType)
            const file = new File([blob], cropOriginalFile.name, { type: mimeType })
            setAvatarFile(file)
            setAvatarPreview(URL.createObjectURL(blob))
            closeCropModal()
        } catch {
            setError("Не удалось обработать изображение.")
        } finally {
            setCropApplying(false)
        }
    }, [cropImageUrl, cropOriginalFile, lastCroppedAreaPixels, rotation, closeCropModal])

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault()
        setError("")
        setSuccess(false)
        setSaving(true)
        try {
            const nameTrimmed = fullName.trim()
            if (nameTrimmed !== (profile?.full_name ?? "")) {
                await updateProfile({ full_name: nameTrimmed || undefined })
            }
            if (avatarFile) {
                await uploadAvatar(avatarFile)
                setAvatarFile(null)
                if (fileInputRef.current) fileInputRef.current.value = ""
            }
            await refetch()
            setSuccess(true)
            setAvatarFile(null)
            if (fileInputRef.current) fileInputRef.current.value = ""
        } catch {
            setError("Не удалось сохранить изменения.")
        } finally {
            setSaving(false)
        }
    }

    if (profileLoading && !profile) {
        return (
            <div className="flex min-h-[40vh] flex-1 items-center justify-center bg-slate-50/50">
                <Loader2 className="h-8 w-8 animate-spin text-slate-400" />
            </div>
        )
    }

    if (!profile) {
        return (
            <div className="flex flex-1 flex-col items-center justify-center gap-4 bg-slate-50/50 px-6 py-12">
                <p className="text-sm text-slate-500">Профиль не загружен.</p>
                <Link
                    to="/dashboard"
                    className="text-sm font-medium text-slate-700 transition hover:text-slate-900"
                >
                    На главную
                </Link>
            </div>
        )
    }

    return (
        <div className="min-h-0 flex-1 bg-[#f5f4f2]">
            <div className="mx-auto max-w-xl px-4 py-10 sm:px-6">
                <Link
                    to="/dashboard"
                    className="inline-flex items-center gap-2 text-sm text-slate-500 transition hover:text-slate-800"
                >
                    <ArrowLeft className="h-4 w-4" />
                    Назад
                </Link>

                <form onSubmit={handleSubmit} className="mt-10">
                    {/* Один блок: аватар + имя в одном потоке */}
                    <div className="rounded-2xl bg-white/90 p-8 shadow-[0_2px_20px_rgba(0,0,0,0.06)] backdrop-blur-sm">
                        <div className="flex flex-col items-center gap-8 sm:flex-row sm:items-start sm:gap-10">
                            <button
                                type="button"
                                onClick={() => fileInputRef.current?.click()}
                                onDrop={onDrop}
                                onDragOver={onDragOver}
                                className="group relative flex h-32 w-32 shrink-0 overflow-hidden rounded-full ring-2 ring-slate-200/80 ring-offset-4 ring-offset-white transition hover:ring-slate-300 focus:outline-none focus:ring-2 focus:ring-slate-400 focus:ring-offset-2"
                            >
                                {avatarPreview && !avatarPreviewError ? (
                                    <img
                                        src={avatarPreview}
                                        alt=""
                                        className="h-full w-full object-cover"
                                        onError={() => setAvatarPreviewError(true)}
                                    />
                                ) : (
                                    <span className="flex h-full w-full items-center justify-center bg-slate-100">
                                        <User className="h-14 w-14 text-slate-400" />
                                    </span>
                                )}
                                <span className="absolute inset-0 flex items-center justify-center rounded-full bg-black/40 opacity-0 transition group-hover:opacity-100">
                                    <Camera className="h-8 w-8 text-white" />
                                </span>
                            </button>
                            <input
                                ref={fileInputRef}
                                type="file"
                                accept={ALLOWED_AVATAR_TYPES.join(",")}
                                onChange={onFileChange}
                                className="hidden"
                            />
                            <div className="min-w-0 flex-1 text-center sm:text-left">
                                <input
                                    id="full_name"
                                    type="text"
                                    value={fullName}
                                    onChange={(e) => setFullName(e.target.value)}
                                    placeholder="Ваше имя"
                                    className="w-full border-0 border-b-2 border-slate-200 bg-transparent pb-2 text-xl font-medium text-slate-900 placeholder-slate-400 transition focus:border-slate-400 focus:outline-none"
                                    maxLength={255}
                                />
                                <p className="mt-1 text-xs text-slate-400">{fullName.length}/255</p>
                                {profile?.email && (
                                    <p className="mt-3 text-sm text-slate-500">{profile.email}</p>
                                )}
                                <p className="mt-4 text-sm text-slate-500">
                                    Нажмите на фото, чтобы заменить. JPEG, PNG или WebP, до 5 МБ.
                                </p>
                            </div>
                        </div>
                    </div>

                    {error && (
                        <div className="mt-6 flex items-center gap-3 rounded-xl bg-red-50/90 px-4 py-3 text-sm text-red-800">
                            <AlertCircle className="h-4 w-4 shrink-0" />
                            {error}
                        </div>
                    )}
                    {success && (
                        <div className="mt-6 flex items-center gap-3 rounded-xl bg-emerald-50/90 px-4 py-3 text-sm text-emerald-800">
                            <CheckCircle2 className="h-4 w-4 shrink-0" />
                            Сохранено.
                        </div>
                    )}

                    <div className="mt-8 flex justify-end">
                        <button
                            type="submit"
                            disabled={saving}
                            className="rounded-full bg-slate-900 px-8 py-3 text-sm font-medium text-white transition hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-slate-500 focus:ring-offset-2 disabled:opacity-60"
                        >
                            {saving ? (
                                <>
                                    <Loader2 className="mr-2 inline h-4 w-4 animate-spin" />
                                    Сохранение...
                                </>
                            ) : (
                                "Сохранить"
                            )}
                        </button>
                    </div>
                </form>
            </div>

            {cropModalOpen && cropImageUrl && (
                <div className="fixed inset-0 z-50 flex flex-col bg-slate-900/95 backdrop-blur-sm">
                    <div className="flex flex-1 flex-col items-center justify-center p-6">
                        <p className="mb-5 text-sm font-medium text-slate-300">Настройте кадр: перемещайте, масштабируйте, поворачивайте</p>
                        <div className="relative h-[70vmin] w-[70vmin] max-h-[380px] max-w-[380px] rounded-full overflow-hidden bg-slate-800 shadow-2xl ring-1 ring-white/10">
                            <Cropper
                                image={cropImageUrl}
                                crop={crop}
                                zoom={zoom}
                                rotation={rotation}
                                aspect={1}
                                cropShape="round"
                                showGrid={false}
                                onCropChange={setCrop}
                                onZoomChange={setZoom}
                                onRotationChange={setRotation}
                                onCropComplete={onCropComplete}
                                onCropAreaChange={(_, pixels) => setLastCroppedAreaPixels(pixels)}
                                objectFit="contain"
                                minZoom={0.5}
                                maxZoom={3}
                            />
                        </div>
                        <div className="mt-6 flex w-full max-w-[380px] flex-col gap-4">
                            <div>
                                <label className="mb-1.5 block text-xs font-medium text-slate-400">Масштаб</label>
                                <input
                                    type="range"
                                    min={0.5}
                                    max={3}
                                    step={0.1}
                                    value={zoom}
                                    onChange={(e) => setZoom(Number(e.target.value))}
                                    className="w-full accent-slate-400"
                                />
                            </div>
                            <div>
                                <label className="mb-1.5 block text-xs font-medium text-slate-400">Поворот</label>
                                <input
                                    type="range"
                                    min={0}
                                    max={360}
                                    step={1}
                                    value={rotation}
                                    onChange={(e) => setRotation(Number(e.target.value))}
                                    className="w-full accent-slate-400"
                                />
                            </div>
                        </div>
                        <div className="mt-8 flex gap-3">
                            <button
                                type="button"
                                onClick={closeCropModal}
                                className="rounded-lg border border-slate-500 bg-transparent px-4 py-2.5 text-sm font-medium text-slate-300 transition hover:bg-white/10 hover:text-white"
                            >
                                Отмена
                            </button>
                            <button
                                type="button"
                                onClick={applyCrop}
                                disabled={cropApplying || !lastCroppedAreaPixels}
                                className="rounded-lg bg-white px-4 py-2.5 text-sm font-semibold text-slate-900 shadow-lg transition hover:bg-slate-100 disabled:opacity-50"
                            >
                                {cropApplying ? (
                                    <span className="inline-flex items-center gap-2">
                                        <Loader2 className="h-4 w-4 animate-spin" />
                                        Применяем...
                                    </span>
                                ) : (
                                    "Применить"
                                )}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    )
}
