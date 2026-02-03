/**
 * Logger utility for conditional logging based on environment
 * Only logs in development mode, silent in production
 */

const isDevelopment = import.meta.env.MODE === 'development'

type LogLevel = 'log' | 'info' | 'warn' | 'error' | 'debug'

class Logger {
  private logWithLevel(level: LogLevel, ...args: unknown[]): void {
    if (isDevelopment) {
      console[level](...args)
    }
  }

  /**
   * General purpose logging
   */
  log(...args: unknown[]): void {
    this.logWithLevel('log', ...args)
  }

  /**
   * Informational messages
   */
  info(...args: unknown[]): void {
    this.logWithLevel('info', ...args)
  }

  /**
   * Warning messages
   */
  warn(...args: unknown[]): void {
    this.logWithLevel('warn', ...args)
  }

  /**
   * Error messages - always logged even in production
   */
  error(...args: unknown[]): void {
    // Errors should be logged in production too
    console.error(...args)
  }

  /**
   * Debug messages - only in development
   */
  debug(...args: unknown[]): void {
    this.logWithLevel('debug', ...args)
  }

  /**
   * Group logging for better organization
   */
  group(label: string, callback: () => void): void {
    if (isDevelopment) {
      console.group(label)
      callback()
      console.groupEnd()
    }
  }

  /**
   * Table logging for structured data
   */
  table(data: unknown): void {
    if (isDevelopment && console.table) {
      console.table(data)
    }
  }

  /**
   * Performance timing
   */
  time(label: string): void {
    if (isDevelopment) {
      console.time(label)
    }
  }

  timeEnd(label: string): void {
    if (isDevelopment) {
      console.timeEnd(label)
    }
  }
}

// Export singleton instance
export const logger = new Logger()

// Export default for import convenience
export default logger
