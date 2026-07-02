<script lang="ts">
  import { createEventDispatcher } from 'svelte'
  import { Sparkles } from 'lucide-svelte'
  import { translate } from '@/lib/i18n'

  const dispatch = createEventDispatcher<{
    addCustom: { label: string }
  }>()

  let customLabel = ''

  const handleAddCustom = () => {
    const label = customLabel.trim()
    if (!label) return
    dispatch('addCustom', { label })
    customLabel = ''
  }
</script>

<div class="space-y-2">
  <div
    class="px-1 pb-1 text-[11px] uppercase tracking-wider text-muted-foreground"
  >
    {$translate('sidebar.addSectionPanel.title')}
  </div>

  <div class="space-y-2">
    <div
      class="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground"
    >
      <Sparkles size={12} class="text-primary" />
      {$translate('sidebar.addSectionPanel.customSection')}
    </div>

    <div class="flex items-center gap-1.5">
      <input
        class="h-8 w-full rounded-md border border-border bg-background px-2 text-xs text-white"
        placeholder={$translate(
          'sidebar.addSectionPanel.customSectionPlaceholder',
        )}
        bind:value={customLabel}
        on:keydown={(event) => event.key === 'Enter' && handleAddCustom()}
      />
      <button
        type="button"
        class="h-8 shrink-0 rounded-md border border-border px-2 text-xs text-white hover:bg-white/5"
        on:click={handleAddCustom}
        >{$translate('sidebar.addSectionPanel.add')}</button
      >
    </div>
  </div>
</div>
