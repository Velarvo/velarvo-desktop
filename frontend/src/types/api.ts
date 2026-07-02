export interface APIResponse<T = unknown> {
  success: boolean
  code: string
  message?: string
  params?: Record<string, string>
  error?: string
  statusCode?: number
  data?: T
}
