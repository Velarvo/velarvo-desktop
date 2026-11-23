<script lang="ts">
  import { FileQuestion, LoaderCircle, TriangleAlert } from '@lucide/svelte'
  import FileIcon from '@/components/file-explorer/FileIcon.svelte'
  import { formatFileSize } from '@/components/file-explorer/utils'
  import Modal from '@/components/ui/Modal.svelte'
  import { previewImageSource, type FilePreview } from '@/lib/filePreview'
  import { t } from '@/lib/i18n.svelte'
  import {
    escapeHtml,
    highlightCode,
    lineNumberGutter,
    type HighlightedCode,
  } from '@/lib/syntax'

  interface Props {
    open?: boolean
    origin?: string
    name?: string
    loading?: boolean
    error?: string
    item?: FilePreview | null
    onClose?: () => void
  }

  let {
    open = false,
    origin = '',
    name = '',
    loading = false,
    error = '',
    item = null,
    onClose,
  }: Props = $props()

  let highlighted = $state<HighlightedCode | null>(null)
  let highlightRequest = 0

  const colourize = async (current: FilePreview | null) => {
    const request = ++highlightRequest

    if (!current || current.kind !== 'text') {
      highlighted = null
      return
    }

    const code = current.text ?? ''
    highlighted = { html: escapeHtml(code), language: '' }

    const result = await highlightCode(current.name, code)
    if (request === highlightRequest) highlighted = result
  }

  $effect(() => {
    void colourize(item)
  })

  let gutter = $derived(
    item?.kind === 'text' ? lineNumberGutter(item.text ?? '') : '',
  )

  let subtitle = $derived(item?.path ?? '')
  let meta = $derived(
    item
      ? [
          formatFileSize(item.size),
          highlighted?.language || item.mediaType,
          origin,
        ]
          .filter(Boolean)
          .join(' · ')
      : origin,
  )
</script>

<!-- eslint-disable svelte/no-at-html-tags -->

<Modal
  {open}
  title={name}
  description={subtitle}
  size="xl"
  flush
  onClose={() => onClose?.()}
>
  {#snippet icon()}
    <FileIcon {name} />
  {/snippet}

  <div class="flex max-h-[70vh] min-h-[16rem] flex-col">
    {#if loading}
      <div
        class="flex flex-1 items-center justify-center gap-2 px-5 py-10 text-sm text-muted-foreground"
      >
        <LoaderCircle class="h-4 w-4 animate-spin" />
        {t('sftp.previewLoading', 'Reading the file...')}
      </div>
    {:else if error}
      <div
        class="flex flex-1 flex-col items-center justify-center gap-2 px-5 py-10 text-center"
      >
        <TriangleAlert class="h-5 w-5 text-red-400" />
        <p class="max-w-lg text-sm text-red-400">{error}</p>
      </div>
    {:else if item?.kind === 'text'}
      <div class="min-h-0 flex-1 overflow-auto bg-surface-canvas">
        <div class="flex min-h-full w-max min-w-full items-stretch">
          <pre
            aria-hidden="true"
            class="sticky left-0 z-10 shrink-0 border-r border-line-subtle bg-surface-canvas py-4 pr-3 pl-5 text-right font-mono text-xs leading-relaxed text-fg-faint select-none">{gutter}</pre>
          <pre
            class="code-preview py-4 pr-5 pl-4 font-mono text-xs leading-relaxed whitespace-pre text-white/90 select-text"><code
              >{@html highlighted?.html ?? escapeHtml(item.text ?? '')}</code
            ></pre>
        </div>
      </div>
    {:else if item?.kind === 'image'}
      <div class="min-h-0 flex-1 overflow-auto p-5">
        <img
          src={previewImageSource(item)}
          alt={name}
          class="mx-auto max-h-full rounded-lg border border-line bg-surface-canvas object-contain"
        />
      </div>
    {:else}
      <div
        class="flex flex-1 flex-col items-center justify-center gap-2 px-5 py-10 text-center"
      >
        <FileQuestion class="h-6 w-6 text-muted-foreground" />
        <p class="text-sm text-muted-foreground">
          {#if item?.reason === 'tooLarge'}
            {t(
              'sftp.previewTooLarge',
              'This file is too large to preview. Transfer it to open it.',
            )}
          {:else}
            {t(
              'sftp.previewBinary',
              'This file has no preview. Transfer it to open it.',
            )}
          {/if}
        </p>
      </div>
    {/if}

    {#if item && !loading && !error}
      <div
        class="flex shrink-0 items-center gap-2 border-t border-line-subtle px-5 py-2.5 text-[11px] text-muted-foreground"
      >
        <span class="truncate">{meta}</span>
        {#if item.truncated}
          <span class="ml-auto shrink-0 text-primary">
            {t('sftp.previewTruncated', 'Showing the beginning of the file')}
          </span>
        {/if}
      </div>
    {/if}
  </div>
</Modal>
