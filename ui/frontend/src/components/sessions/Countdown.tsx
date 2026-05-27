import { useEffect, useState } from 'react'

interface CountdownProps {
  totalSecs: number
  startedAt: string // ISO timestamp
}

export function Countdown({ totalSecs, startedAt }: CountdownProps) {
  const [remaining, setRemaining] = useState(() => calcRemaining(totalSecs, startedAt))

  useEffect(() => {
    setRemaining(calcRemaining(totalSecs, startedAt))

    const id = setInterval(() => {
      setRemaining(calcRemaining(totalSecs, startedAt))
    }, 500)

    return () => clearInterval(id)
  }, [totalSecs, startedAt])

  const expired = remaining <= 0
  const display = formatSecs(Math.max(0, remaining))

  return (
    <span
      className={
        expired
          ? 'font-mono font-bold text-destructive animate-pulse'
          : 'font-mono font-semibold tabular-nums'
      }
    >
      {display}
    </span>
  )
}

function calcRemaining(totalSecs: number, startedAt: string): number {
  const elapsedMs = Date.now() - new Date(startedAt).getTime()
  return totalSecs - Math.floor(elapsedMs / 1000)
}

function formatSecs(secs: number): string {
  const m = Math.floor(secs / 60)
  const s = secs % 60
  return `${m}:${String(s).padStart(2, '0')}`
}
