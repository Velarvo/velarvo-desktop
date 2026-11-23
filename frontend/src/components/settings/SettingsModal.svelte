<script lang="ts">
  import { LockKeyhole, SlidersHorizontal } from '@lucide/svelte'
  import Modal from '@/components/ui/Modal.svelte'
  import { t } from '@/lib/i18n.svelte'
  import PreferencesSection from './PreferencesSection.svelte'
  import SecuritySection from './SecuritySection.svelte'

  interface Props {
    open: boolean
    onClose: () => void
  }

  let { open, onClose }: Props = $props()

  const sections = [
    { id: 'security', icon: LockKeyhole },
    { id: 'preferences', icon: SlidersHorizontal },
  ] as const

  let active = $state<(typeof sections)[number]['id']>('security')
  let busy = $state(false)

  $effect.pre(() => {
    if (!open) active = 'security'
  })
</script>

<Modal
  {open}
  title={t('settings.title')}
  size="lg"
  dismissible={!busy}
  flush
  {onClose}
>
  <div class="grid h-[26rem] grid-cols-[11rem_minmax(0,1fr)]">
    <nav class="space-y-0.5 border-r border-line-subtle bg-fill-subtle p-2">
      {#each sections as section (section.id)}
        <button
          type="button"
          class="flex w-full items-center gap-2.5 rounded-md px-2.5 py-2 text-left text-sm transition {active ===
          section.id
            ? 'bg-primary/10 text-primary'
            : 'text-fg-muted hover:bg-fill hover:text-white'}"
          aria-current={active === section.id ? 'page' : undefined}
          disabled={busy}
          onclick={() => (active = section.id)}
        >
          <section.icon class="h-4 w-4" />
          {t(`settings.${section.id}.nav`)}
        </button>
      {/each}
    </nav>

    <section class="overflow-y-auto px-6 py-5">
      <h3 class="text-sm font-semibold text-white">
        {t(`settings.${active}.title`)}
      </h3>
      <p class="mt-0.5 text-sm text-muted-foreground">
        {t(`settings.${active}.subtitle`)}
      </p>

      <div class="mt-3">
        {#if active === 'security'}
          <SecuritySection onBusy={(value) => (busy = value)} {onClose} />
        {:else}
          <PreferencesSection />
        {/if}
      </div>
    </section>
  </div>
</Modal>
