import { ApiError } from '@/api/client'
import { useModerationWarningStore } from '@/stores/useModerationWarningStore'

/**
 * Checks if an error is an automated moderation violation and, if so,
 * opens the moderation warning modal.
 *
 * Returns true when the error was handled as a moderation violation.
 */
export function handleModerationError(error: unknown): boolean {
  if (!(error instanceof ApiError)) return false
  if (error.code !== 'MODERATION_VIOLATION') return false

  useModerationWarningStore.getState().show({
    message: error.message,
    strikes: error.strikes ?? 1,
    isBanned: error.isBanned ?? false,
  })

  return true
}
