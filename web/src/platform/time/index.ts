export enum Duration {
  millisecond = 1,
  second = millisecond * 1000,
  minute = second * 60,
  hour = minute * 60,
}

interface DurationParams {
  milliseconds?: number
  seconds?: number
  minutes?: number
  hours?: number
}

export function duration(duration: DurationParams): number {
  let d = duration.milliseconds || 0
  d += duration.seconds || 0 * Duration.second
  d += duration.minutes || 0 * Duration.minute
  d += duration.hours || 0 * Duration.hour
  return d
}
