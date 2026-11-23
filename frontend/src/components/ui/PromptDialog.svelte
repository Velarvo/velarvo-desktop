<script lang="ts">
  import { tick } from 'svelte'
  import Button from '@/components/ui/Button.svelte'
  import Field from '@/components/ui/Field.svelte'
  import Input from '@/components/ui/Input.svelte'
  import Modal from '@/components/ui/Modal.svelte'
  import { t } from '@/lib/i18n.svelte'

  interface Props {
    open?: boolean
    title?: string
    subtitle?: string
    label?: string
    value?: string
    confirmLabel?: string
    selectionEnd?: number | null
    onSubmit?: (detail: { value: string }) => void
    onClose?: () => void
  }

  let {
    open = false,
    title = '',
    subtitle = '',
    label = '',
    value = $bindable(''),
    confirmLabel = '',
    selectionEnd = null,
    onSubmit,
    onClose,
  }: Props = $props()

  let input = $state<HTMLInputElement | null>(null)

  const focusInput = async () => {
    await tick()
    input?.focus()
    input?.setSelectionRange(0, selectionEnd ?? value.length)
  }

  $effect(() => {
    if (open) void focusInput()
  })

  const submit = () => {
    const trimmed = value.trim()
    if (!trimmed) return
    onSubmit?.({ value: trimmed })
  }
</script>

<Modal
  {open}
  {title}
  description={subtitle}
  size="sm"
  onsubmit={submit}
  onClose={() => onClose?.()}
>
  <Field
    label={label || title}
    for="prompt-dialog-input"
    class={label ? '' : '[&>label]:sr-only'}
  >
    <Input
      id="prompt-dialog-input"
      bind:ref={input}
      bind:value
      autocomplete="off"
      spellcheck="false"
    />
  </Field>

  {#snippet footer()}
    <Button variant="ghost" size="sm" onclick={() => onClose?.()}>
      {t('common.cancel')}
    </Button>
    <Button type="submit" size="sm" disabled={!value.trim()}>
      {confirmLabel || t('common.confirm')}
    </Button>
  {/snippet}
</Modal>
