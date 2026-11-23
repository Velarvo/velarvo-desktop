export class Transient<T> {
  current: T
  readonly #idle: T
  readonly #ms: number
  #timer: ReturnType<typeof setTimeout> | null = null

  constructor(ms: number, idle: T) {
    this.#ms = ms
    this.#idle = idle
    this.current = $state(idle)
  }

  show = (value: T) => {
    this.#stop()
    this.current = value
    this.#timer = setTimeout(() => {
      this.#timer = null
      this.current = this.#idle
    }, this.#ms)
  }

  clear = () => {
    this.#stop()
    this.current = this.#idle
  }

  #stop() {
    if (this.#timer) clearTimeout(this.#timer)
    this.#timer = null
  }
}

export class TransientFlag extends Transient<boolean> {
  constructor(ms: number) {
    super(ms, false)
  }

  raise = () => this.show(true)
}
