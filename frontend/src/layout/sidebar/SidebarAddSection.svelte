<script lang="ts">
  import { Sparkles } from '@lucide/svelte'
  import Button from '@/components/ui/Button.svelte'
  import Input from '@/components/ui/Input.svelte'
  import { t } from '@/lib/i18n.svelte'

  interface Props {
    onAddCustom?: (detail: { label: string }) => void
  }

  let { onAddCustom }: Props = $props()

  let customLabel = $state('')

  const handleAddCustom = () => {
    const label = customLabel.trim()
    if (!label) return
    onAddCustom?.({ label })
    customLabel = ''
  }
</script>

<div class="space-y-2">
  <div
    class="px-1 pb-1 text-[11px] uppercase tracking-wider text-muted-foreground"
  >
    {t('sidebar.addSectionPanel.title')}
  </div>

  <div class="space-y-2">
    <div
      class="flex items-center gap-1.5 text-[11px] uppercase tracking-wider text-muted-foreground"
    >
      <Sparkles size={12} class="text-primary" />
      {t('sidebar.addSectionPanel.customSection')}
    </div>

    <div class="flex items-center gap-1.5">
      <Input
        size="sm"
        aria-label={t('sidebar.addSectionPanel.customSection')}
        placeholder={t('sidebar.addSectionPanel.customSectionPlaceholder')}
        bind:value={customLabel}
        onkeydown={(event) => event.key === 'Enter' && handleAddCustom()}
      />
      <Button variant="secondary" size="sm" onclick={handleAddCustom}>
        {t('sidebar.addSectionPanel.add')}
      </Button>
    </div>
  </div>
</div>
