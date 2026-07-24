export interface Status {
  version?: string
  usageAccess: boolean
  batteryFree: boolean
  enabled: boolean
  url: string
  token: string
  chat: string
  summary: string
  pending: number
}

export interface AppUsage {
  package: string
  label: string
  category: string
  foreground_seconds: number
  launch_count: number
}

export interface Snapshot {
  window_start: string
  window_end: string
  foreground_app?: string
  foreground_category?: string
  screen_on: boolean
  apps: AppUsage[]
  distracting_seconds: number
  today: {
    date: string
    total_foreground_seconds: number
    unlocks: number
    apps: AppUsage[]
  }
}

interface BridgeApi {
  status(): string
  snapshot(): string
  sampleNow(): void
  save(url: string, token: string, chat: string): void
  stop(): void
  openUsageAccess(): void
  openBattery(): void
}

declare global {
  interface Window {
    Bridge?: BridgeApi
    refresh?: () => void
  }
}

const fallback: Status = {
  usageAccess: false,
  batteryFree: false,
  enabled: false,
  url: '',
  token: '',
  chat: '',
  summary: 'Приложение открыто вне телефона — данных нет',
  pending: 0,
}

export function readStatus(): Status {
  if (!window.Bridge) return fallback
  try {
    return JSON.parse(window.Bridge.status()) as Status
  } catch {
    return fallback
  }
}

export function readSnapshot(): Snapshot | null {
  if (!window.Bridge) return null
  try {
    const raw = window.Bridge.snapshot()
    if (!raw) return null
    return JSON.parse(raw) as Snapshot
  } catch {
    return null
  }
}

export function formatDuration(seconds: number): string {
  const total = Math.max(0, Math.round(seconds))
  const h = Math.floor(total / 3600)
  const m = Math.floor((total % 3600) / 60)
  if (h > 0) return `${h} ч ${m} мин`
  if (m > 0) return `${m} мин`
  return `${total} с`
}

export const bridge = {
  save: (url: string, token: string, chat: string) => window.Bridge?.save(url, token, chat),
  stop: () => window.Bridge?.stop(),
  sampleNow: () => window.Bridge?.sampleNow(),
  openUsageAccess: () => window.Bridge?.openUsageAccess(),
  openBattery: () => window.Bridge?.openBattery(),
}
