<script lang="ts">
  import { ImageUp, LoaderCircle, X } from '@lucide/svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import Field from '@/components/ui/Field.svelte'
  import Input from '@/components/ui/Input.svelte'
  import Modal from '@/components/ui/Modal.svelte'
  import { describeError } from '@/lib/errors'
  import { t } from '@/lib/i18n.svelte'
  import {
    ICON_ACCEPT,
    prepareProjectIcon,
    type PreparedIcon,
  } from '@/lib/projectIcon'
  import { createProject, type Project } from '@/lib/projects.svelte'

  interface Props {
    open: boolean
    onClose: () => void
    onCreated?: (detail: { project: Project }) => void
  }

  let { open, onClose, onCreated }: Props = $props()

  let name = $state('')
  let icon = $state<PreparedIcon | null>(null)
  let preparingIcon = $state(false)
  let submitting = $state(false)
  let errorMessage = $state('')
  let fileInput = $state<HTMLInputElement>()

  let busy = $derived(submitting || preparingIcon)
  let initial = $derived((name.trim()[0] ?? 'P').toUpperCase())

  const setIcon = (next: PreparedIcon | null) => {
    if (icon) URL.revokeObjectURL(icon.previewUrl)
    icon = next
  }

  $effect.pre(() => {
    if (open) return
    name = ''
    errorMessage = ''
    setIcon(null)
  })

  const iconErrors: Record<string, string> = {
    unsupported: 'projects.create.iconUnsupported',
    tooLarge: 'projects.create.iconTooLarge',
    canvas: 'projects.create.iconCanvasError',
  }

  const chooseIcon = async (event: Event) => {
    const input = event.currentTarget as HTMLInputElement
    const file = input.files?.[0]
    input.value = ''
    if (!file) return

    preparingIcon = true
    errorMessage = ''
    try {
      setIcon(await prepareProjectIcon(file))
    } catch (error) {
      const reason = error instanceof Error ? error.message : ''
      errorMessage = t(iconErrors[reason] ?? 'projects.create.iconFailed')
    } finally {
      preparingIcon = false
    }
  }

  const submit = async () => {
    if (busy) return
    if (!name.trim()) {
      errorMessage = t('projects.create.validation.nameRequired')
      return
    }

    submitting = true
    errorMessage = ''
    try {
      const project = await createProject({
        name: name.trim(),
        icon: icon ? { mime: icon.mime, data: icon.data } : undefined,
      })
      onCreated?.({ project })
      onClose()
    } catch (error) {
      errorMessage = describeError(error)
    } finally {
      submitting = false
    }
  }
</script>

<Modal
  {open}
  title={t('projects.create.title')}
  description={t('projects.create.subtitle')}
  dismissible={!busy}
  onsubmit={submit}
  {onClose}
>
  <div class="flex items-start gap-4">
    <div class="flex shrink-0 flex-col items-center gap-1.5 pt-1">
      <button
        type="button"
        class="group relative flex h-16 w-16 items-center justify-center overflow-hidden rounded-lg bg-primary text-2xl font-black text-swatch-foreground shadow-[inset_0_1px_0_rgba(255,255,255,0.18)] disabled:cursor-not-allowed"
        aria-label={t('projects.create.chooseIcon')}
        title={t('projects.create.chooseIcon')}
        disabled={busy}
        onclick={() => fileInput?.click()}
      >
        {#if icon}
          <img
            class="h-full w-full object-cover"
            src={icon.previewUrl}
            alt=""
          />
        {:else}
          {initial}
        {/if}
        <span
          class="absolute inset-0 flex items-center justify-center bg-black/55 text-white transition-opacity {preparingIcon
            ? 'opacity-100'
            : 'opacity-0 group-hover:opacity-100'}"
        >
          {#if preparingIcon}
            <LoaderCircle class="h-5 w-5 animate-spin" />
          {:else}
            <ImageUp class="h-5 w-5" />
          {/if}
        </span>
      </button>

      {#if icon}
        <button
          type="button"
          class="inline-flex items-center gap-1 text-xs text-muted-foreground transition hover:text-white"
          disabled={busy}
          onclick={() => setIcon(null)}
        >
          <X class="h-3 w-3" />
          {t('projects.create.removeIcon')}
        </button>
      {/if}
    </div>

    <div class="min-w-0 flex-1">
      <Field
        label={t('projects.create.name')}
        for="project-name"
        hint={t('projects.create.iconHint')}
      >
        <Input
          id="project-name"
          placeholder={t('projects.create.namePlaceholder')}
          autocomplete="off"
          disabled={submitting}
          bind:value={name}
        />
      </Field>
    </div>
  </div>

  <input
    bind:this={fileInput}
    class="hidden"
    type="file"
    accept={ICON_ACCEPT}
    onchange={chooseIcon}
  />

  {#if errorMessage}
    <div class="mt-4">
      <Alert variant="destructive" message={errorMessage} />
    </div>
  {/if}

  {#snippet footer()}
    <Button variant="ghost" size="sm" disabled={busy} onclick={onClose}>
      {t('common.cancel')}
    </Button>
    <Button
      type="submit"
      size="sm"
      loading={submitting}
      disabled={preparingIcon}
    >
      {submitting ? t('projects.create.creating') : t('projects.create.create')}
    </Button>
  {/snippet}
</Modal>
