<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { LockKeyhole, Settings2 } from 'lucide-svelte'
  import ModalShell from '@/components/ui/ModalShell.svelte'
  import { translate } from '@/lib/i18n'
  import PreferencesSection from './sections/PreferencesSection.svelte'
  import SecuritySection from './sections/SecuritySection.svelte'

  export let open = false

  const dispatch = createEventDispatcher<{ close: void }>()

  const sections = [
    {
      id: 'security' as const,
      navKey: 'settings.security.nav',
      icon: LockKeyhole,
      component: SecuritySection,
    },
    {
      id: 'preferences' as const,
      navKey: 'settings.preferences.nav',
      icon: Settings2,
      component: PreferencesSection,
    },
  ]

  type SectionId = (typeof sections)[number]['id']

  let activeSection: SectionId = sections[0].id
  let sectionBusy = false

  $: activeConfig =
    sections.find((section) => section.id === activeSection) ?? sections[0]

  const close = () => {
    if (sectionBusy) return
    dispatch('close')
  }

  const setActive = (id: SectionId) => {
    if (id === activeSection) return
    activeSection = id
    sectionBusy = false
  }

  $: if (!open) {
    activeSection = sections[0].id
    sectionBusy = false
  }
</script>

<ModalShell
  {open}
  title={$translate('settings.title')}
  subtitle={$translate('settings.subtitle')}
  closeLabel={$translate('settings.security.close')}
  titleId="settings-title"
  descriptionId="settings-description"
  maxWidthClass="max-w-5xl"
  bodyClass="grid h-[min(40rem,80vh)] lg:grid-cols-[14rem_minmax(0,1fr)]"
  closeDisabled={sectionBusy}
  on:close={close}
>
  <svelte:fragment slot="icon">
    <div
      class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl border border-primary/20 bg-primary/10 text-primary"
    >
      <Settings2 class="h-5 w-5" />
    </div>
  </svelte:fragment>

  <aside
    class="overflow-y-auto border-b border-white/8 bg-white/[0.025] p-3 lg:border-b-0 lg:border-r lg:border-white/8"
  >
    <p
      class="px-2 py-2 text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground"
    >
      {$translate('settings.navTitle')}
    </p>

    <div class="mt-1 space-y-1">
      {#each sections as section (section.id)}
        <button
          type="button"
          class={`flex w-full items-center gap-3 rounded-xl border px-3 py-3 text-left text-sm transition ${
            activeSection === section.id
              ? 'border-primary/25 bg-primary/10 text-primary shadow-[inset_0_0_0_1px_rgb(var(--color-primary-rgb)_/_0.05)]'
              : 'border-transparent text-white/80 hover:border-white/8 hover:bg-white/[0.04] hover:text-white'
          }`}
          on:click={() => setActive(section.id)}
        >
          <svelte:component this={section.icon} class="h-4 w-4" />
          <span>{$translate(section.navKey)}</span>
        </button>
      {/each}
    </div>
  </aside>

  <main class="min-h-0 overflow-y-auto p-5">
    <svelte:component
      this={activeConfig.component}
      on:close={close}
      on:busy={(event: CustomEvent<boolean>) => (sectionBusy = event.detail)}
    />
  </main>
</ModalShell>
