<script lang="ts">
  import { createEventDispatcher, tick } from 'svelte'
  import { Layers } from 'lucide-svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import ColorPicker from '@/components/ui/ColorPicker.svelte'
  import ModalShell from '@/components/ui/ModalShell.svelte'
  import { DEFAULT_ACCENT_COLOR } from '@/lib/colors'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import { createWorkspace, type Workspace } from '@/lib/workspaces'

  export let open = false
  export let projectId: string | null = null
  export let projectName = ''

  const dispatch = createEventDispatcher<{
    close: void
    created: { workspace: Workspace }
  }>()

  let name = ''
  let color: string = DEFAULT_ACCENT_COLOR
  let isSubmitting = false
  let errorMessage = ''
  let nameInput: HTMLInputElement | null = null

  const onNameInput = () => {
    name = name.toLowerCase()
  }

  $: normalizedName = name.trim().toLowerCase()
  $: previewName =
    normalizedName || $translate('workspaces.create.previewFallback')

  const reset = () => {
    name = ''
    color = DEFAULT_ACCENT_COLOR
    isSubmitting = false
    errorMessage = ''
  }

  const close = () => {
    if (isSubmitting) return
    dispatch('close')
  }

  const handleSubmit = async () => {
    if (isSubmitting) {
      return
    }

    if (!projectId) {
      errorMessage = t('workspaces.create.validation.projectRequired')
      return
    }

    if (!normalizedName) {
      errorMessage = t('workspaces.create.validation.nameRequired')
      return
    }

    isSubmitting = true
    errorMessage = ''

    const resp = await createWorkspace({
      projectId,
      name: normalizedName,
      color,
    })

    try {
      if (!resp.success || !resp.data) {
        errorMessage = getResponseMessage(
          resp,
          t('errors.WORKSPACE_CREATE_FAILED'),
        )
        return
      }

      dispatch('created', { workspace: resp.data })
      reset()
      dispatch('close')
    } finally {
      isSubmitting = false
    }
  }

  const focusNameField = async () => {
    await tick()
    nameInput?.focus()
    nameInput?.select()
  }

  $: if (!open) {
    reset()
  }

  $: if (open) {
    void focusNameField()
  }
</script>

<ModalShell
  {open}
  title={$translate('workspaces.create.title')}
  subtitle={$translate('workspaces.create.subtitle')}
  closeLabel={$translate('workspaces.create.close')}
  titleId="workspace-create-title"
  descriptionId="workspace-create-description"
  maxWidthClass="max-w-3xl"
  bodyClass="relative"
  closeDisabled={isSubmitting}
  on:close={close}
>
  <svelte:fragment slot="icon">
    <div
      class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-primary/20 bg-primary/10 text-primary"
    >
      <Layers class="h-5 w-5" />
    </div>
  </svelte:fragment>

  <form on:submit|preventDefault={handleSubmit}>
    <div
      class="grid gap-5 px-6 py-6 sm:px-8 md:grid-cols-[minmax(0,1fr)_17rem]"
    >
      <div class="space-y-4">
        {#if projectName}
          <p
            class="rounded-lg border border-white/8 bg-white/[0.025] px-3 py-2 text-xs text-muted-foreground"
          >
            {$translate('workspaces.create.projectContext')}
            <span class="font-semibold text-white/80">{projectName}</span>
          </p>
        {/if}

        <div class="space-y-2">
          <label
            for="workspace-name"
            class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
          >
            {$translate('workspaces.create.name')}
          </label>
          <input
            id="workspace-name"
            bind:this={nameInput}
            class="h-11 w-full rounded-xl border border-white/10 bg-surface-field px-4 text-sm text-white outline-none transition placeholder:text-white/25 focus:border-primary/35 focus:bg-surface-field-focus"
            placeholder={$translate('workspaces.create.namePlaceholder')}
            bind:value={name}
            on:input={onNameInput}
            autocomplete="off"
          />
        </div>

        <ColorPicker
          bind:value={color}
          label={$translate('workspaces.create.color')}
          disabled={isSubmitting}
        />

        {#if errorMessage}
          <Alert variant="destructive" message={errorMessage} />
        {/if}
      </div>

      <aside
        class="flex min-h-full flex-col rounded-xl border border-white/10 bg-white/[0.025] p-4"
      >
        <p
          class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
        >
          {$translate('workspaces.create.preview')}
        </p>

        <div
          class="mt-3 flex items-center gap-3 rounded-lg border border-white/10 bg-surface-field p-3"
        >
          <span
            class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-md shadow-[inset_0_1px_0_rgba(255,255,255,0.18)]"
            style={`background:${color}; color:var(--color-swatch-foreground);`}
          >
            <Layers class="h-4 w-4" />
          </span>
          <span class="truncate text-sm font-semibold text-white">
            {previewName}
          </span>
        </div>

        <div class="mt-4 h-px bg-white/8"></div>

        <div
          class="mt-4 space-y-3 rounded-lg border border-white/8 bg-surface-base/70 p-3"
        >
          <div class="flex items-center gap-2">
            <span
              class="h-2.5 w-2.5 rounded-full"
              style={`background:${color};`}
            ></span>
            <div class="h-2 w-24 rounded-full bg-white/10"></div>
          </div>
          <div class="space-y-2">
            <div class="h-2 rounded-full bg-white/7"></div>
            <div class="h-2 w-2/3 rounded-full bg-white/7"></div>
          </div>
        </div>
      </aside>
    </div>

    <div
      class="flex flex-col-reverse gap-3 border-t border-white/8 px-6 py-5 sm:flex-row sm:items-center sm:justify-end sm:px-8"
    >
      <Button
        variant="outline"
        disabled={isSubmitting}
        on:click={close}
        class="w-full sm:w-auto"
      >
        {$translate('common.cancel')}
      </Button>

      <Button
        type="submit"
        variant="primary-glow"
        loading={isSubmitting}
        class="w-full sm:w-auto sm:min-w-40"
      >
        {isSubmitting
          ? $translate('workspaces.create.creating')
          : $translate('workspaces.create.create')}
      </Button>
    </div>
  </form>
</ModalShell>
