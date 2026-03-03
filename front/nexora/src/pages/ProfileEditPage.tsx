import { useState, useRef, useEffect, useCallback, type FormEvent } from "react"
import { Link } from "react-router-dom"
import Cropper from "react-easy-crop"
import type { Area } from "react-easy-crop"
import "react-easy-crop/react-easy-crop.css"
import { useProfileContext } from "../context/ProfileContext"
import { updateProfile, uploadAvatar } from "../api/profile"
import { getCroppedImg } from "../utils/cropImage"
import { ArrowLeft, Loader2, User } from "lucide-react"

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

    const onFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0]
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
            <div className="p-8 flex items-center justify-center min-h-[200px]" style={{ backgroundColor: "var(--theme-bg-page)" }}>
                <Loader2 className="h-8 w-8 animate-spin" style={{ color: "var(--theme-text-muted)" }} />
            </div>
        )
    }

    if (!profile) {
        return (
            <div className="p-8" style={{ backgroundColor: "var(--theme-bg-page)" }}>
                <p style={{ color: "var(--theme-text-muted)" }}>Профиль не загружен.</p>
                <Link to="/dashboard" className="font-medium hover:underline mt-2 inline-block" style={{ color: "var(--theme-text)" }}>
                    На главную
                </Link>
            </div>
        )
    }

    return (
        <div className="p-8 max-w-lg" style={{ backgroundColor: "var(--theme-bg-page)" }}>
            <Link
                to="/dashboard"
                className="inline-flex items-center gap-2 font-medium mb-6 hover:underline"
                style={{ color: "var(--theme-text-muted)" }}
            >
                <ArrowLeft className="h-4 w-4" />
                Назад
            </Link>
            <h1 className="text-2xl font-semibold mb-2" style={{ color: "var(--theme-text)" }}>Редактирование профиля</h1>
            <p className="text-sm mb-8" style={{ color: "var(--theme-text-muted)" }}>
                Укажите ФИО и при необходимости загрузите новое фото профиля.
            </p>

            <form onSubmit={handleSubmit} className="space-y-6">
                <div>
                    <label htmlFor="full_name" className="block text-sm font-medium mb-2" style={{ color: "var(--theme-text)" }}>
                        ФИО
                    </label>
                    <input
                        id="full_name"
                        type="text"
                        value={fullName}
                        onChange={(e) => setFullName(e.target.value)}
                        placeholder="Иван Иванов"
                        className="input-theme w-full px-4 py-3 rounded-xl border focus:outline-none"
                        style={{
                            backgroundColor: "var(--theme-input-bg)",
                            borderColor: "var(--theme-border)",
                            color: "var(--theme-text)",
                        }}
                        maxLength={255}
                    />
                </div>

                <div>
                    <label className="block text-sm font-medium mb-2" style={{ color: "var(--theme-text)" }}>Аватар</label>
                    <div className="flex items-center gap-4">
                        <button
                            type="button"
                            onClick={() => fileInputRef.current?.click()}
                            className="flex items-center justify-center w-24 h-24 rounded-full transition overflow-hidden"
                            style={{ backgroundColor: "var(--theme-hover)", color: "var(--theme-text-muted)" }}
                        >
                            {avatarPreview && !avatarPreviewError ? (
                                <img
                                    src={avatarPreview}
                                    alt=""
                                    className="w-full h-full object-cover"
                                    onError={() => setAvatarPreviewError(true)}
                                />
                            ) : (
                                <User className="h-10 w-10" />
                            )}
                        </button>
                        <input
                            ref={fileInputRef}
                            type="file"
                            accept={ALLOWED_AVATAR_TYPES.join(",")}
                            onChange={onFileChange}
                            className="hidden"
                        />
                        <div className="text-sm" style={{ color: "var(--theme-text-muted)" }}>
                            Нажмите на круг, чтобы выбрать изображение (JPEG, PNG или WebP, до 5 МБ).
                        </div>
                    </div>
                </div>

                {error && (
                    <div className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
                        {error}
                    </div>
                )}
                {success && (
                    <div className="rounded-xl border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-800">
                        Изменения сохранены.
                    </div>
                )}

                <button
                    type="submit"
                    disabled={saving}
                    className="w-full py-3 rounded-xl font-medium transition disabled:opacity-60 disabled:cursor-not-allowed min-h-[48px]"
                    style={{
                        backgroundColor: "var(--theme-btn-primary-bg)",
                        color: "var(--theme-btn-primary-text)",
                    }}
                >
                    {saving ? (
                        <span className="inline-flex items-center gap-2">
                            <Loader2 className="h-4 w-4 animate-spin" />
                            Сохранение...
                        </span>
                    ) : (
                        "Сохранить"
                    )}
                </button>
            </form>

            {cropModalOpen && cropImageUrl && (
                <div className="fixed inset-0 z-50 flex flex-col bg-black/80">
                    <div className="flex flex-1 flex-col items-center justify-center p-4">
                        <p className="mb-4 text-white">Расположите фото: двигайте, приближайте, поворачивайте</p>
                        <div className="relative h-[70vmin] w-[70vmin] max-h-[400px] max-w-[400px] rounded-full overflow-hidden bg-slate-800">
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
                        <div className="mt-6 flex w-full max-w-[400px] flex-col gap-4">
                            <div>
                                <label className="mb-1 block text-xs text-slate-300">Масштаб</label>
                                <input
                                    type="range"
                                    min={0.5}
                                    max={3}
                                    step={0.1}
                                    value={zoom}
                                    onChange={(e) => setZoom(Number(e.target.value))}
                                    className="w-full accent-slate-100"
                                />
                            </div>
                            <div>
                                <label className="mb-1 block text-xs text-slate-300">Поворот</label>
                                <input
                                    type="range"
                                    min={0}
                                    max={360}
                                    step={1}
                                    value={rotation}
                                    onChange={(e) => setRotation(Number(e.target.value))}
                                    className="w-full accent-slate-100"
                                />
                            </div>
                        </div>
                        <div className="mt-6 flex gap-3">
                            <button
                                type="button"
                                onClick={closeCropModal}
                                className="rounded-xl border border-slate-400 bg-transparent px-5 py-2.5 text-white hover:bg-white/10"
                            >
                                Отмена
                            </button>
                            <button
                                type="button"
                                onClick={applyCrop}
                                disabled={cropApplying || !lastCroppedAreaPixels}
                                className="rounded-xl bg-white px-5 py-2.5 text-slate-900 font-medium hover:opacity-90 disabled:opacity-50"
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
