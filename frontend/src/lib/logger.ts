import { Frontend as GoLog } from '@bindings/logger'

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

const native = {
  debug: console.debug.bind(console),
  info: console.info.bind(console),
  log: console.log.bind(console),
  warn: console.warn.bind(console),
  error: console.error.bind(console),
}

const serialize = (args: unknown[]): string =>
  args
    .map((value) => {
      if (typeof value === 'string') return value
      if (value instanceof Error) {
        return value.stack ? `${value.message}\n${value.stack}` : value.message
      }
      try {
        return JSON.stringify(value)
      } catch {
        return String(value)
      }
    })
    .join(' ')

const forward = (level: LogLevel, args: unknown[]): void => {
  GoLog.Write(level, serialize(args)).catch(() => {})
}

const logTo =
  (level: LogLevel, print: (...args: unknown[]) => void) =>
  (...args: unknown[]) => {
    print(...args)
    forward(level, args)
  }

export const logger = {
  debug: logTo('debug', native.debug),
  info: logTo('info', native.info),
  warn: logTo('warn', native.warn),
  error: logTo('error', native.error),
}

let installed = false

export const installGlobalLoggerBridge = (): void => {
  if (installed) return
  installed = true

  console.debug = logger.debug
  console.info = logger.info
  console.log = logTo('info', native.log)
  console.warn = logger.warn
  console.error = logger.error

  window.addEventListener('error', (event) => {
    logger.error('Unhandled error', event.error ?? event.message)
  })
  window.addEventListener('unhandledrejection', (event) => {
    logger.error('Unhandled promise rejection', event.reason)
  })
}
