export interface APIResponse<T = unknown> {
  success: boolean
  code: string
  message?: string
  error?: string
  statusCode?: number
  data?: T
}
