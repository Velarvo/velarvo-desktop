<script lang="ts">
  import { onMount } from 'svelte'
  import { navigate, router } from '@/lib/router.svelte'
  import { bootstrapVault, vault } from '@/lib/vault.svelte'
  import OnboardingPage from '@/pages/vault/OnboardingPage.svelte'
  import UnlockPage from '@/pages/vault/UnlockPage.svelte'
  import Dashboard from '@/pages/Dashboard.svelte'
  import { bootstrapLanguage, t } from '@/lib/i18n.svelte'

  type View = 'loading' | 'onboarding' | 'unlock' | 'dashboard'

  const APP_ROUTES = new Set(['/dashboard'])

  const ROUTE_FOR_VIEW: Record<Exclude<View, 'loading'>, string> = {
    onboarding: '/',
    unlock: '/unlock',
    dashboard: '/dashboard',
  }

  onMount(() => {
    void Promise.all([bootstrapLanguage(), bootstrapVault()])
  })

  let view = $derived(
    ((): View => {
      if (vault.loading) return 'loading'
      if (!vault.isSetup) return 'onboarding'
      if (!vault.isUnlocked) return 'unlock'
      return 'dashboard'
    })(),
  )

  $effect.pre(() => {
    if (view !== 'loading') {
      if (view === 'dashboard') {
        if (!APP_ROUTES.has(router.path)) {
          navigate('/dashboard')
        }
      } else {
        const target = ROUTE_FOR_VIEW[view]
        if (router.path !== target) {
          navigate(target)
        }
      }
    }
  })
</script>

{#if view === 'loading'}
  <div
    class="flex h-screen items-center justify-center bg-background"
    role="status"
    aria-live="polite"
  >
    <div class="flex items-center gap-3 text-sm text-muted-foreground">
      <span
        class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-primary/30 border-t-primary"
        aria-hidden="true"
      ></span>
      <span>{t('common.loading')}</span>
    </div>
  </div>
{:else if view === 'dashboard'}
  <Dashboard />
{:else if view === 'unlock'}
  <UnlockPage />
{:else}
  <OnboardingPage />
{/if}
