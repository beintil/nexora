/**
 * Creates an image element from URL (object URL or data URL).
 */
function createImage(url: string): Promise<HTMLImageElement> {
    return new Promise((resolve, reject) => {
        const img = new Image()
        img.onload = () => resolve(img)
        img.onerror = reject
        img.src = url
    })
}

/**
 * Returns the bounding box size of an image after rotation (degrees).
 */
function getRotatedSize(width: number, height: number, rotation: number): { width: number; height: number } {
    const rad = (rotation * Math.PI) / 180
    return {
        width: Math.abs(Math.cos(rad) * width) + Math.abs(Math.sin(rad) * height),
        height: Math.abs(Math.sin(rad) * width) + Math.abs(Math.cos(rad) * height),
    }
}

export type CropArea = {
    x: number
    y: number
    width: number
    height: number
}

/**
 * Returns a cropped image as Blob. Handles rotation: the image is drawn rotated
 * into the rotated bounding box, then the crop area (in that space) is extracted.
 *
 * @param imageSrc - Object URL or data URL of the image
 * @param pixelCrop - Crop area in pixels (in rotated media space, as from react-easy-crop)
 * @param rotation - Rotation in degrees
 * @param mimeType - Output MIME type (default image/jpeg)
 */
export async function getCroppedImg(
    imageSrc: string,
    pixelCrop: CropArea,
    rotation: number,
    mimeType: string = "image/jpeg"
): Promise<Blob> {
    const img = await createImage(imageSrc)
    const { naturalWidth, naturalHeight } = img

    const rotatedSize = getRotatedSize(naturalWidth, naturalHeight, rotation)

    const canvas1 = document.createElement("canvas")
    canvas1.width = rotatedSize.width
    canvas1.height = rotatedSize.height
    const ctx1 = canvas1.getContext("2d")
    if (!ctx1) throw new Error("Canvas 2d not available")

    const cx = rotatedSize.width / 2
    const cy = rotatedSize.height / 2
    ctx1.translate(cx, cy)
    ctx1.rotate((rotation * Math.PI) / 180)
    ctx1.translate(-naturalWidth / 2, -naturalHeight / 2)
    ctx1.drawImage(img, 0, 0, naturalWidth, naturalHeight, 0, 0, naturalWidth, naturalHeight)

    const canvas2 = document.createElement("canvas")
    canvas2.width = pixelCrop.width
    canvas2.height = pixelCrop.height
    const ctx2 = canvas2.getContext("2d")
    if (!ctx2) throw new Error("Canvas 2d not available")

    ctx2.drawImage(
        canvas1,
        pixelCrop.x,
        pixelCrop.y,
        pixelCrop.width,
        pixelCrop.height,
        0,
        0,
        pixelCrop.width,
        pixelCrop.height
    )

    return new Promise((resolve, reject) => {
        canvas2.toBlob(
            (blob) => (blob ? resolve(blob) : reject(new Error("toBlob failed"))),
            mimeType,
            0.92
        )
    })
}
