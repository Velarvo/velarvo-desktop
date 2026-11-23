export interface ReversibleAction {
  label: string
  undo: () => Promise<void>
  redo: () => Promise<void>
}

type Direction = 'undo' | 'redo'

export class SFTPHistory {
  #stacks = $state.raw<Record<Direction, ReversibleAction[]>>({
    undo: [],
    redo: [],
  })
  #busy = false

  canUndo = $derived(this.#stacks.undo.length > 0)
  canRedo = $derived(this.#stacks.redo.length > 0)
  undoLabel = $derived(this.#stacks.undo.at(-1)?.label ?? '')
  redoLabel = $derived(this.#stacks.redo.at(-1)?.label ?? '')

  push = (action: ReversibleAction) => {
    this.#stacks = { undo: [...this.#stacks.undo, action], redo: [] }
  }

  undo = () => this.#run('undo', 'redo')
  redo = () => this.#run('redo', 'undo')

  clear = () => {
    this.#stacks = { undo: [], redo: [] }
  }

  async #run(from: Direction, to: Direction) {
    if (this.#busy) return

    const action = this.#stacks[from].at(-1)
    if (!action) return
    this.#stacks = { ...this.#stacks, [from]: this.#stacks[from].slice(0, -1) }

    this.#busy = true
    try {
      await (from === 'undo' ? action.undo() : action.redo())
      this.#stacks = { ...this.#stacks, [to]: [...this.#stacks[to], action] }
    } catch (error) {
      this.#stacks = {
        ...this.#stacks,
        [from]: [...this.#stacks[from], action],
      }
      throw error
    } finally {
      this.#busy = false
    }
  }
}
