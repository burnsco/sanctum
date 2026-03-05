import { create } from 'zustand'

export interface ModerationWarningData {
  message: string
  strikes: number
  isBanned: boolean
}

interface ModerationWarningState {
  open: boolean
  data: ModerationWarningData | null
  show: (data: ModerationWarningData) => void
  dismiss: () => void
}

export const useModerationWarningStore = create<ModerationWarningState>()(
  set => ({
    open: false,
    data: null,
    show: (data: ModerationWarningData) => set({ open: true, data }),
    dismiss: () => set({ open: false, data: null }),
  })
)
