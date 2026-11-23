import { Service as Files, type Preview } from '@bindings/files'
import { SFTP } from '@bindings/sshconn'

export type FilePreview = Preview

export const previewImageSource = (item: FilePreview): string =>
  `data:${item.mediaType || 'application/octet-stream'};base64,${item.data ?? ''}`

export const previewLocalFile = (path: string) => Files.Preview(path)

export const previewRemoteFile = (connectionId: string, path: string) =>
  SFTP.Preview(connectionId, path)
