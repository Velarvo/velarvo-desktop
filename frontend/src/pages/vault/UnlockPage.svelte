<script lang="ts">
  import { onMount } from 'svelte'
  import { ArrowRight, LoaderCircle, LockKeyhole } from '@lucide/svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import Input from '@/components/ui/Input.svelte'
  import VaultLayout from '@/layout/VaultLayout.svelte'
  import { describeError } from '@/lib/errors'
  import { t } from '@/lib/i18n.svelte'
  import { navigate } from '@/lib/router.svelte'
  import { unlockVault } from '@/lib/vault.svelte'

  let masterPassword = $state('')
  let errorMessage = $state('')
  let isSubmitting = $state(false)
  let passwordInput = $state<HTMLInputElement | null>(null)

  onMount(() => passwordInput?.focus())

  const handleUnlock = async (event: SubmitEvent) => {
    event.preventDefault()
    if (isSubmitting) return
    errorMessage = ''

    if (!masterPassword.trim()) {
      errorMessage = t('vault.unlock.validation.passwordRequired')
      passwordInput?.focus()
      return
    }

    isSubmitting = true
    try {
      await unlockVault(masterPassword)
      masterPassword = ''
      navigate('/dashboard')
    } catch (error) {
      errorMessage = describeError(error)
      passwordInput?.select()
    } finally {
      isSubmitting = false
    }
  }
</script>

<VaultLayout>
  <form onsubmit={handleUnlock} novalidate>
    <h1 class="text-xl font-semibold tracking-tight text-white">
      {t('vault.unlock.title')}
    </h1>
    <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">
      {t('vault.unlock.subtitle')}
    </p>

    <label for="masterPassword" class="sr-only">
      {t('vault.unlock.masterPassword')}
    </label>
    <Input
      id="masterPassword"
      type="password"
      size="lg"
      class="mt-6"
      autocomplete="current-password"
      placeholder={t('vault.unlock.passwordPlaceholder')}
      invalid={Boolean(errorMessage)}
      bind:ref={passwordInput}
      bind:value={masterPassword}
      oninput={() => (errorMessage = '')}
    >
      {#snippet leading()}
        <LockKeyhole class="h-4 w-4" />
      {/snippet}
      {#snippet trailing()}
        <IconButton
          type="submit"
          variant="primary"
          icon={isSubmitting ? LoaderCircle : ArrowRight}
          iconClass={isSubmitting ? 'animate-spin' : ''}
          label={isSubmitting
            ? t('vault.unlock.submitting')
            : t('vault.unlock.submit')}
          disabled={isSubmitting}
        />
      {/snippet}
    </Input>

    {#if errorMessage}
      <p class="mt-2 text-xs text-red-300" role="alert">{errorMessage}</p>
    {/if}
  </form>
</VaultLayout>
