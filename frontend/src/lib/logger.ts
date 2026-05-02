import {
  LogDebug,
  LogError,
  LogInfo,
  LogWarning,
} from '../../wailsjs/runtime/runtime'

type LogLevel = 'debug' | 'info' | 'warn' | 'error'

type ConsoleMethod = (...args: unknown[]) => void

const nativeConsole = {
  debug: console.debug.bind(console),
  info: console.info.bind(console),
  log: console.log.bind(console),
  warn: console.warn.bind(console),
  error: console.error.bind(console),
}

let bridgeInstalled = false

function serialize(args: unknown[]): string {
  return args
    .map((value) => {
      if (typeof value === 'string') {
        return value
      }

      if (value instanceof Error) {
        return value.stack ?? value.message
      }

      try {
        return JSON.stringify(value)
      } catch {
        return String(value)
      }
    })
    .join(' ')
}

function emit(level: LogLevel, args: unknown[]) {
  const message = serialize(args)

  if (typeof window === 'undefined' || !('runtime' in window)) {
    return
  }

  switch (level) {
    case 'debug':
      LogDebug(message)
      return
    case 'info':
      LogInfo(message)
      return
    case 'warn':
      LogWarning(message)
      return
    case 'error':
      LogError(message)
      return
  }
}

function bridge(method: ConsoleMethod, level: LogLevel): ConsoleMethod {
  return (...args: unknown[]) => {
    method(...args)
    emit(level, args)
  }
}

export const logger = {
  debug: (...args: unknown[]) => {
    nativeConsole.debug(...args)
    emit('debug', args)
  },
  info: (...args: unknown[]) => {
    nativeConsole.info(...args)
    emit('info', args)
  },
  warn: (...args: unknown[]) => {
    nativeConsole.warn(...args)
    emit('warn', args)
  },
  error: (...args: unknown[]) => {
    nativeConsole.error(...args)
    emit('error', args)
  },
}

export function installGlobalLoggerBridge() {
  if (bridgeInstalled) {
    return
  }

  console.debug = bridge(nativeConsole.debug, 'debug')
  console.info = bridge(nativeConsole.info, 'info')
  console.log = bridge(nativeConsole.log, 'info')
  console.warn = bridge(nativeConsole.warn, 'warn')
  console.error = bridge(nativeConsole.error, 'error')

  window.addEventListener('error', (event) => {
    logger.error('Unhandled error', event.error ?? event.message)
  })

  window.addEventListener('unhandledrejection', (event) => {
    logger.error('Unhandled promise rejection', event.reason)
  })

  bridgeInstalled = true
}
