<script lang="ts">
  import { createEventDispatcher, onDestroy, tick } from 'svelte'
  import { ImagePlus, LoaderCircle, Trash2, Upload } from 'lucide-svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import ModalShell from '@/components/ui/ModalShell.svelte'
  import { getResponseMessage } from '@/lib/errors'
  import { t, translate } from '@/lib/i18n'
  import {
    createProject,
    type Project,
    type ProjectIconInput,
  } from '@/lib/projects'

  export let open = false

  const dispatch = createEventDispatcher<{
    close: void
    created: { project: Project }
  }>()

  const iconSize = 192

  let name = ''
  let selectedIcon: ProjectIconInput | null = null
  let iconPreviewUrl = ''
  let iconFileName = ''
  let isPreparingIcon = false
  let isSubmitting = false
  let errorMessage = ''
  let iconError = ''
  let nameInput: HTMLInputElement | null = null
  let fileInput: HTMLInputElement | null = null

  $: normalizedName = name.trim()
  $: previewName =
    normalizedName || $translate('projects.create.previewFallback')
  $: previewInitial = (normalizedName[0] ?? 'P').toUpperCase()

  const revokeIconPreview = () => {
    if (!iconPreviewUrl) return
    URL.revokeObjectURL(iconPreviewUrl)
    iconPreviewUrl = ''
  }

  const reset = () => {
    name = ''
    selectedIcon = null
    iconFileName = ''
    isPreparingIcon = false
    isSubmitting = false
    errorMessage = ''
    iconError = ''
    revokeIconPreview()
    if (fileInput) {
      fileInput.value = ''
    }
  }

  const close = () => {
    if (isSubmitting || isPreparingIcon) return
    dispatch('close')
  }

  const blobToBytes = async (blob: Blob) => {
    return Array.from(new Uint8Array(await blob.arrayBuffer()))
  }

  const loadImage = async (url: string) => {
    const image = new Image()
    image.decoding = 'async'
    image.src = url

    await image.decode()
    return image
  }

  const canvasToBlob = async (canvas: HTMLCanvasElement, quality: number) => {
    return new Promise<Blob | null>((resolve) => {
      canvas.toBlob(resolve, 'image/webp', quality)
    })
  }

  const makeSquareIcon = async (file: File) => {
    const sourceUrl = URL.createObjectURL(file)

    try {
      const image = await loadImage(sourceUrl)
      const canvas = document.createElement('canvas')
      canvas.width = iconSize
      canvas.height = iconSize

      const context = canvas.getContext('2d')
      if (!context) {
        throw new Error(t('projects.create.iconCanvasError'))
      }

      const scale = Math.max(iconSize / image.width, iconSize / image.height)
      const width = image.width * scale
      const height = image.height * scale

      context.drawImage(
        image,
        (iconSize - width) / 2,
        (iconSize - height) / 2,
        width,
        height,
      )

      for (const quality of [0.86, 0.72, 0.58]) {
        const blob = await canvasToBlob(canvas, quality)
        if (blob && blob.size <= 256 * 1024) {
          return blob
        }
      }

      throw new Error(t('projects.create.iconTooLarge'))
    } finally {
      URL.revokeObjectURL(sourceUrl)
    }
  }

  const handleIconFileChange = async (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    iconError = ''
    errorMessage = ''

    if (!file.type.startsWith('image/')) {
      iconError = t('projects.create.iconUnsupported')
      target.value = ''
      return
    }

    isPreparingIcon = true

    try {
      const blob = await makeSquareIcon(file)
      revokeIconPreview()
      iconPreviewUrl = URL.createObjectURL(blob)
      iconFileName = file.name
      selectedIcon = {
        mime: 'image/webp',
        data: await blobToBytes(blob),
      }
    } catch (error) {
      iconError =
        error instanceof Error
          ? error.message
          : $translate('projects.create.iconFailed')
      selectedIcon = null
      iconFileName = ''
      revokeIconPreview()
      target.value = ''
    } finally {
      isPreparingIcon = false
    }
  }

  const clearIcon = () => {
    selectedIcon = null
    iconFileName = ''
    iconError = ''
    revokeIconPreview()
    if (fileInput) {
      fileInput.value = ''
    }
  }

  const handleSubmit = async () => {
    if (isSubmitting || isPreparingIcon) {
      return
    }

    if (!normalizedName) {
      errorMessage = t('projects.create.validation.nameRequired')
      return
    }

    isSubmitting = true
    errorMessage = ''

    const resp = await createProject({
      name: normalizedName,
      icon: selectedIcon ?? undefined,
    })

    try {
      if (!resp.success || !resp.data) {
        errorMessage = getResponseMessage(
          resp,
          t('errors.PROJECT_CREATE_FAILED'),
        )
        return
      }

      dispatch('created', { project: resp.data })
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

  onDestroy(() => {
    revokeIconPreview()
  })
</script>

<ModalShell
  {open}
  title={$translate('projects.create.title')}
  subtitle={$translate('projects.create.subtitle')}
  closeLabel={$translate('projects.create.close')}
  titleId="project-create-title"
  descriptionId="project-create-description"
  maxWidthClass="max-w-3xl"
  bodyClass="relative"
  closeDisabled={isSubmitting || isPreparingIcon}
  on:close={close}
>
  <svelte:fragment slot="icon">
    <div
      class="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-primary/20 bg-primary/10 text-primary"
    >
      <ImagePlus class="h-5 w-5" />
    </div>
  </svelte:fragment>

  <form on:submit|preventDefault={handleSubmit}>
    <div
      class="grid gap-5 px-6 py-6 sm:px-8 md:grid-cols-[minmax(0,1fr)_17rem]"
    >
      <div class="space-y-4">
        <div class="space-y-2">
          <label
            for="project-name"
            class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
          >
            {$translate('projects.create.name')}
          </label>
          <input
            id="project-name"
            bind:this={nameInput}
            class="h-11 w-full rounded-xl border border-white/10 bg-surface-field px-4 text-sm text-white outline-none transition placeholder:text-white/25 focus:border-primary/35 focus:bg-surface-field-focus"
            placeholder={$translate('projects.create.namePlaceholder')}
            bind:value={name}
            autocomplete="off"
          />
        </div>

        <div class="rounded-xl border border-white/10 bg-white/[0.025] p-4">
          <div class="flex items-center justify-between gap-3">
            <div class="min-w-0">
              <p
                class="text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
              >
                {$translate('projects.create.icon')}
              </p>
              <p class="mt-1 truncate text-xs text-muted-foreground">
                {iconFileName || $translate('projects.create.iconHint')}
              </p>
            </div>

            {#if selectedIcon}
              <button
                type="button"
                class="inline-flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04] text-muted-foreground transition hover:border-white/15 hover:bg-white/[0.07] hover:text-white"
                aria-label={$translate('projects.create.removeIcon')}
                on:click={clearIcon}
              >
                <Trash2 class="h-4 w-4" />
              </button>
            {/if}
          </div>

          <input
            bind:this={fileInput}
            class="hidden"
            type="file"
            accept="image/png,image/jpeg,image/webp"
            on:change={handleIconFileChange}
          />

          <button
            type="button"
            class="mt-4 flex h-36 w-full items-center justify-center rounded-lg border border-dashed border-white/12 bg-surface-field transition hover:border-primary/25 hover:bg-surface-field-focus disabled:cursor-not-allowed disabled:opacity-60"
            disabled={isPreparingIcon || isSubmitting}
            on:click={() => fileInput?.click()}
          >
            {#if isPreparingIcon}
              <span
                class="inline-flex items-center gap-2 text-sm text-white/80"
              >
                <LoaderCircle class="h-4 w-4 animate-spin" />
                {$translate('projects.create.preparingIcon')}
              </span>
            {:else if iconPreviewUrl}
              <span
                class="flex h-24 w-24 items-center justify-center rounded-xl border border-white/10 bg-surface-base p-1 shadow-[0_14px_34px_rgba(0,0,0,0.32)]"
              >
                <img
                  class="h-full w-full rounded-lg object-cover"
                  src={iconPreviewUrl}
                  alt=""
                />
              </span>
            {:else}
              <span
                class="inline-flex flex-col items-center gap-2 text-sm text-muted-foreground"
              >
                <span
                  class="inline-flex h-11 w-11 items-center justify-center rounded-lg border border-white/10 bg-white/[0.04] text-white/80"
                >
                  <Upload class="h-5 w-5" />
                </span>
                {$translate('projects.create.chooseIcon')}
              </span>
            {/if}
          </button>

          {#if iconError}
            <p class="mt-3 text-xs leading-5 text-red-200">{iconError}</p>
          {/if}
        </div>

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
          {$translate('projects.create.preview')}
        </p>

        <div
          class="mt-3 flex items-center gap-3 rounded-lg border border-white/10 bg-surface-field p-3"
        >
          <span
            class="inline-flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-md text-sm font-black shadow-[inset_0_1px_0_rgba(255,255,255,0.18)]"
            style="background:var(--color-primary); color:var(--color-swatch-foreground);"
          >
            {#if iconPreviewUrl}
              <img
                class="h-full w-full object-cover"
                src={iconPreviewUrl}
                alt=""
              />
            {:else}
              {previewInitial}
            {/if}
          </span>
          <span class="truncate text-sm font-semibold text-white">
            {previewName}
          </span>
        </div>

        <div class="mt-4 h-px bg-white/8"></div>

        <div
          class="mt-4 rounded-lg border border-white/8 bg-surface-base/70 p-3"
        >
          <div class="h-2 w-24 rounded-full bg-white/10"></div>
          <div class="mt-3 space-y-2">
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
        disabled={isSubmitting || isPreparingIcon}
        on:click={close}
        class="w-full sm:w-auto"
      >
        {$translate('common.cancel')}
      </Button>

      <Button
        type="submit"
        variant="primary-glow"
        loading={isSubmitting}
        disabled={isPreparingIcon}
        class="w-full sm:w-auto sm:min-w-40"
      >
        {isSubmitting
          ? $translate('projects.create.creating')
          : $translate('projects.create.create')}
      </Button>
    </div>
  </form>
</ModalShell>
