import { useEffect, useState } from 'react'
import {
  Activity,
  BarChart3,
  BatteryCharging,
  CheckCircle2,
  Clock,
  RefreshCw,
  Send,
  Square,
  XCircle,
  Zap,
} from 'lucide-react'
import {
  bridge,
  formatDuration,
  readSnapshot,
  readStatus,
  type AppUsage,
  type Snapshot,
  type Status,
} from './bridge'

const CATEGORIES: Record<string, { name: string; color: string; distracting: boolean }> = {
  work: { name: 'Работа', color: '#007D9C', distracting: false },
  productivity: { name: 'Продуктивность', color: '#00ADD8', distracting: false },
  communication: { name: 'Общение', color: '#10B981', distracting: false },
  reading: { name: 'Чтение', color: '#5DC9E2', distracting: false },
  social: { name: 'Соцсети', color: '#EF4444', distracting: true },
  video: { name: 'Видео', color: '#F59E0B', distracting: true },
  games: { name: 'Игры', color: '#A855F7', distracting: true },
  other: { name: 'Прочее', color: '#94A3B8', distracting: false },
}

function categoryMeta(id: string) {
  return CATEGORIES[id] ?? CATEGORIES.other
}

function StatsCard({ snapshot }: { snapshot: Snapshot | null }) {
  const apps = snapshot?.today.apps ?? []
  const byCategory = new Map<string, number>()
  for (const a of apps) {
    byCategory.set(a.category, (byCategory.get(a.category) ?? 0) + a.foreground_seconds)
  }
  const rows = Array.from(byCategory.entries())
    .map(([id, seconds]) => ({ id, seconds, ...categoryMeta(id) }))
    .sort((x, y) => y.seconds - x.seconds)
  const max = rows.reduce((m, r) => Math.max(m, r.seconds), 0)
  const total = snapshot?.today.total_foreground_seconds ?? 0
  const distracting = rows.filter((r) => r.distracting).reduce((s, r) => s + r.seconds, 0)
  const topApps: AppUsage[] = [...apps].sort((a, b) => b.foreground_seconds - a.foreground_seconds).slice(0, 6)

  return (
    <Card title="Статистика за сегодня" icon={<BarChart3 className="h-5 w-5" />}>
      {rows.length === 0 ? (
        <p className="text-sm text-muted">
          Данных пока нет. Нажми «Замерить сейчас» или подожди ближайший замер.
        </p>
      ) : (
        <div className="flex flex-col gap-4">
          <div className="flex items-baseline justify-between">
            <span className="text-2xl font-semibold text-ink">{formatDuration(total)}</span>
            <span className="text-sm text-muted">разблокировок: {snapshot?.today.unlocks ?? 0}</span>
          </div>
          <div className="text-sm text-muted">
            Отвлекающих: <span className="font-medium text-danger">{formatDuration(distracting)}</span>
            {total > 0 && ` (${Math.round((distracting / total) * 100)}%)`}
          </div>

          <div className="flex flex-col gap-2.5">
            {rows.map((r) => (
              <div key={r.id}>
                <div className="mb-1 flex items-center justify-between text-sm">
                  <span className="flex items-center gap-2 text-ink">
                    <span className="h-3 w-3 rounded-sm" style={{ backgroundColor: r.color }} />
                    {r.name}
                  </span>
                  <span className="text-muted">{formatDuration(r.seconds)}</span>
                </div>
                <div className="h-2.5 w-full overflow-hidden rounded-full bg-slate-100">
                  <div
                    className="h-full rounded-full"
                    style={{ width: `${max > 0 ? (r.seconds / max) * 100 : 0}%`, backgroundColor: r.color }}
                  />
                </div>
              </div>
            ))}
          </div>

          <div className="border-t border-slate-100 pt-3">
            <div className="mb-2 text-sm font-medium text-ink">Приложения</div>
            <div className="flex flex-col gap-1.5">
              {topApps.map((a) => (
                <div key={a.package} className="flex items-center justify-between text-sm">
                  <span className="flex items-center gap-2 truncate text-ink">
                    <span
                      className="h-2.5 w-2.5 shrink-0 rounded-sm"
                      style={{ backgroundColor: categoryMeta(a.category).color }}
                    />
                    <span className="truncate">{a.label}</span>
                  </span>
                  <span className="shrink-0 text-muted">{formatDuration(a.foreground_seconds)}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </Card>
  )
}

function StatusRow({
  title,
  ok,
  okText,
  failText,
  action,
  onFix,
}: {
  title: string
  ok: boolean
  okText: string
  failText: string
  action: string
  onFix: () => void
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-3">
      <div className="flex items-start gap-3">
        {ok ? (
          <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-success" />
        ) : (
          <XCircle className="mt-0.5 h-5 w-5 shrink-0 text-danger" />
        )}
        <div>
          <div className="text-ink">{title}</div>
          <div className={ok ? 'text-sm text-muted' : 'text-sm text-danger'}>{ok ? okText : failText}</div>
        </div>
      </div>
      {!ok && (
        <button
          onClick={onFix}
          className="shrink-0 rounded-card border border-primary px-3 py-1.5 text-sm text-primary-dark transition hover:bg-primary-light/10"
        >
          {action}
        </button>
      )}
    </div>
  )
}

function Card({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <section className="rounded-card bg-surface p-5 shadow-sm ring-1 ring-slate-200">
      <h2 className="mb-3 flex items-center gap-2 text-base font-semibold text-ink">
        <span className="text-primary-dark">{icon}</span>
        {title}
      </h2>
      {children}
    </section>
  )
}

export default function App() {
  const [status, setStatus] = useState<Status>(readStatus)
  const [snapshot, setSnapshot] = useState<Snapshot | null>(readSnapshot)
  const [url, setUrl] = useState('')
  const [token, setToken] = useState('')
  const [chat, setChat] = useState('')
  const [saved, setSaved] = useState(false)
  const [busy, setBusy] = useState(false)

  const refresh = () => {
    const next = readStatus()
    setStatus(next)
    setSnapshot(readSnapshot())
    return next
  }

  useEffect(() => {
    const initial = refresh()
    setUrl(initial.url)
    setToken(initial.token)
    setChat(initial.chat)
    window.refresh = () => refresh()
    const timer = window.setInterval(refresh, 5000)
    return () => window.clearInterval(timer)
  }, [])

  const input =
    'w-full rounded-card border border-slate-300 px-3 py-2 text-ink outline-none transition focus:border-primary'

  return (
    <div className="min-h-screen bg-canvas px-5 py-8 font-sans text-ink">
      <div className="mx-auto flex max-w-md flex-col gap-5">
        <header>
          <h1 className="text-2xl font-semibold">
            Focus <span className="text-sm font-normal text-muted">{status.version ?? ''}</span>
          </h1>
          <p className="text-sm text-muted">Замер использования приложений раз в 10 минут</p>
        </header>

        <Card title="Состояние" icon={<Activity className="h-5 w-5" />}>
          <div className="divide-y divide-slate-100">
            <StatusRow
              title="Доступ к истории использования"
              ok={status.usageAccess}
              okText="выдан"
              failText="без него замеры невозможны"
              action="Выдать"
              onFix={bridge.openUsageAccess}
            />
            <StatusRow
              title="Экономия батареи"
              ok={status.batteryFree}
              okText="отключена"
              failText="система будет выгружать сбор"
              action="Отключить"
              onFix={bridge.openBattery}
            />
            <div className="flex items-center justify-between py-3">
              <div className="flex items-center gap-3">
                <BatteryCharging className={`h-5 w-5 ${status.enabled ? 'text-success' : 'text-muted'}`} />
                <span>{status.enabled ? 'Сбор включён' : 'Сбор выключен'}</span>
              </div>
              {status.enabled && (
                <button
                  onClick={() => {
                    bridge.stop()
                    refresh()
                  }}
                  className="flex items-center gap-1.5 rounded-card border border-slate-300 px-3 py-1.5 text-sm text-muted transition hover:border-danger hover:text-danger"
                >
                  <Square className="h-4 w-4" />
                  Выключить
                </button>
              )}
            </div>
          </div>
        </Card>

        <Card title="Куда отправлять" icon={<Send className="h-5 w-5" />}>
          <div className="flex flex-col gap-3">
            <input
              className={input}
              value={url}
              onChange={(e) => {
                setUrl(e.target.value)
                setSaved(false)
              }}
              placeholder="https://host:8083"
              inputMode="url"
              autoCapitalize="none"
              autoCorrect="off"
              spellCheck={false}
            />
            <input
              className={input}
              value={token}
              onChange={(e) => {
                setToken(e.target.value)
                setSaved(false)
              }}
              placeholder="Токен доступа"
              autoCapitalize="none"
              autoCorrect="off"
              spellCheck={false}
            />
            <input
              className={input}
              value={chat}
              onChange={(e) => {
                setChat(e.target.value.replace(/[^\d-]/g, ''))
                setSaved(false)
              }}
              placeholder="chat_id"
              inputMode="numeric"
            />
            <div className="flex items-center gap-3">
              <button
                onClick={() => {
                  bridge.save(url, token, chat)
                  setSaved(true)
                  refresh()
                }}
                className="rounded-card bg-primary px-4 py-2 font-medium text-white transition hover:bg-primary-dark"
              >
                Сохранить и включить сбор
              </button>
              {saved && <span className="text-sm text-success">сохранено</span>}
            </div>
            <p className="text-sm text-muted">
              Пока адрес пустой, замеры копятся на телефоне и уйдут, как только адрес появится.
            </p>
          </div>
        </Card>

        <StatsCard snapshot={snapshot} />

        <Card title="Последний замер" icon={<Clock className="h-5 w-5" />}>
          <pre className="whitespace-pre-wrap font-sans text-sm leading-relaxed text-ink">{status.summary}</pre>
          <div className="mt-4 flex items-center justify-between gap-2">
            <span className="text-sm text-muted">В очереди на отправку: {status.pending}</span>
            <div className="flex items-center gap-2">
              <button
                disabled={busy}
                onClick={() => {
                  setBusy(true)
                  bridge.sampleNow()
                  window.setTimeout(() => {
                    refresh()
                    setBusy(false)
                  }, 1500)
                }}
                className="flex items-center gap-1.5 rounded-card bg-primary px-3 py-1.5 text-sm text-white transition hover:bg-primary-dark disabled:opacity-60"
              >
                <Zap className="h-4 w-4" />
                {busy ? 'Замеряю…' : 'Замерить сейчас'}
              </button>
              <button
                onClick={refresh}
                className="flex items-center gap-1.5 rounded-card border border-slate-300 px-3 py-1.5 text-sm text-muted transition hover:border-primary hover:text-primary-dark"
              >
                <RefreshCw className="h-4 w-4" />
                Обновить
              </button>
            </div>
          </div>
        </Card>

        <Card title="Чтобы система не выгружала сбор" icon={<BatteryCharging className="h-5 w-5" />}>
          <ol className="flex list-decimal flex-col gap-1.5 pl-5 text-sm leading-relaxed text-muted">
            <li>Разрешить автозапуск приложения.</li>
            <li>Батарея — «Не оптимизировать», отключить глубокую оптимизацию и оптимизацию в режиме сна.</li>
            <li>Закрепить приложение в списке недавних.</li>
            <li>На OnePlus настройка батареи иногда возвращается сама — проверяйте её здесь.</li>
          </ol>
        </Card>
      </div>
    </div>
  )
}
