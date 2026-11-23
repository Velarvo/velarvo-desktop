<script lang="ts">
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import ColorPicker from '@/components/ui/ColorPicker.svelte'
  import Field from '@/components/ui/Field.svelte'
  import Input from '@/components/ui/Input.svelte'
  import Modal from '@/components/ui/Modal.svelte'
  import { DEFAULT_ACCENT_COLOR } from '@/lib/colors'
  import { describeError } from '@/lib/errors'
  import { t } from '@/lib/i18n.svelte'
  import { createWorkspace, type Workspace } from '@/lib/workspaces.svelte'

  interface Props {
    open: boolean
    projectId?: string | null
    projectName?: string
    onClose: () => void
    onCreated?: (detail: { workspace: Workspace }) => void
  }

  let {
    open,
    projectId = null,
    projectName = '',
    onClose,
    onCreated,
  }: Props = $props()

  let name = $state('')
  let color = $state<string>(DEFAULT_ACCENT_COLOR)
  let submitting = $state(false)
  let errorMessage = $state('')

  $effect.pre(() => {
    if (open) return
    name = ''
    color = DEFAULT_ACCENT_COLOR
    errorMessage = ''
  })

  const submit = async () => {
    if (submitting) return
    const normalized = name.trim().toLowerCase()
    if (!projectId) {
      errorMessage = t('workspaces.create.validation.projectRequired')
      return
    }
    if (!normalized) {
      errorMessage = t('workspaces.create.validation.nameRequired')
      return
    }

    submitting = true
    errorMessage = ''
    try {
      const workspace = await createWorkspace({
        projectId,
        name: normalized,
        color,
      })
      onCreated?.({ workspace })
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
  title={t('workspaces.create.title')}
  description={t('workspaces.create.subtitle')}
  dismissible={!submitting}
  onsubmit={submit}
  {onClose}
>
  <div class="space-y-5">
    <Field
      label={t('workspaces.create.name')}
      for="workspace-name"
      hint={projectName
        ? `${t('workspaces.create.projectContext')} ${projectName}`
        : ''}
    >
      <Input
        id="workspace-name"
        placeholder={t('workspaces.create.namePlaceholder')}
        autocomplete="off"
        disabled={submitting}
        bind:value={name}
      />
    </Field>

    <ColorPicker
      bind:value={color}
      label={t('workspaces.create.color')}
      disabled={submitting}
    />

    {#if errorMessage}
      <Alert variant="destructive" message={errorMessage} />
    {/if}
  </div>

  {#snippet footer()}
    <Button variant="ghost" size="sm" disabled={submitting} onclick={onClose}>
      {t('common.cancel')}
    </Button>
    <Button type="submit" size="sm" loading={submitting}>
      {submitting
        ? t('workspaces.create.creating')
        : t('workspaces.create.create')}
    </Button>
  {/snippet}
</Modal>
