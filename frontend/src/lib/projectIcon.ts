const ICON_SIZE = 192
const MAX_ICON_BYTES = 256 * 1024

export const ICON_ACCEPT = 'image/png,image/jpeg,image/webp'

export interface PreparedIcon {
  mime: 'image/webp'
  data: string
  previewUrl: string
}

const loadImage = async (url: string): Promise<HTMLImageElement> => {
  const image = new Image()
  image.src = url
  await image.decode()
  return image
}

const toBase64 = async (blob: Blob): Promise<string> => {
  const bytes = new Uint8Array(await blob.arrayBuffer())
  let binary = ''
  for (let i = 0; i < bytes.length; i += 0x8000) {
    binary += String.fromCharCode(...bytes.subarray(i, i + 0x8000))
  }
  return btoa(binary)
}

export const prepareProjectIcon = async (file: File): Promise<PreparedIcon> => {
  if (!ICON_ACCEPT.split(',').includes(file.type)) {
    throw new Error('unsupported')
  }

  const sourceUrl = URL.createObjectURL(file)
  try {
    const image = await loadImage(sourceUrl)
    const canvas = document.createElement('canvas')
    canvas.width = canvas.height = ICON_SIZE
    const context = canvas.getContext('2d')
    if (!context) throw new Error('canvas')

    const scale = Math.max(ICON_SIZE / image.width, ICON_SIZE / image.height)
    const width = image.width * scale
    const height = image.height * scale
    context.drawImage(
      image,
      (ICON_SIZE - width) / 2,
      (ICON_SIZE - height) / 2,
      width,
      height,
    )

    for (const quality of [0.86, 0.72, 0.58]) {
      const blob = await new Promise<Blob | null>((resolve) =>
        canvas.toBlob(resolve, 'image/webp', quality),
      )
      if (blob && blob.size <= MAX_ICON_BYTES) {
        return {
          mime: 'image/webp',
          data: await toBase64(blob),
          previewUrl: URL.createObjectURL(blob),
        }
      }
    }
    throw new Error('tooLarge')
  } finally {
    URL.revokeObjectURL(sourceUrl)
  }
}
