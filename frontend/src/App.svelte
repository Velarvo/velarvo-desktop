<script lang="ts">
  import { onMount } from 'svelte'
  import { route, navigate } from '@/lib/router'
  import { bootstrapAuth, isLoadingAuth } from '@/lib/auth'
  import {
    bootstrapVault,
    isLoadingVault,
    isVaultSetup,
    isVaultUnlocked,
  } from '@/lib/vault'
  import OnboardingPage from '@/pages/auth/OnboardingPage.svelte'
  import UnlockPage from '@/pages/auth/UnlockPage.svelte'
  import Dashboard from '@/pages/Dashboard.svelte'
  import { bootstrapLanguage, translate } from '@/lib/i18n'

  type View = 'loading' | 'onboarding' | 'unlock' | 'dashboard'

  const APP_ROUTES = new Set(['/dashboard'])

  const ROUTE_FOR_VIEW: Record<Exclude<View, 'loading'>, string> = {
    onboarding: '/',
    unlock: '/unlock',
    dashboard: '/dashboard',
  }

  onMount(() => {
    void Promise.all([bootstrapLanguage(), bootstrapAuth(), bootstrapVault()])
  })

  $: isBootstrapping = $isLoadingAuth || $isLoadingVault

  $: view = ((): View => {
    if (isBootstrapping) return 'loading'
    if (!$isVaultSetup) return 'onboarding'
    if (!$isVaultUnlocked) return 'unlock'
    return 'dashboard'
  })()

  $: if (view !== 'loading') {
    if (view === 'dashboard') {
      if (!APP_ROUTES.has($route)) {
        navigate('/dashboard')
      }
    } else {
      const target = ROUTE_FOR_VIEW[view]
      if ($route !== target) {
        navigate(target)
      }
    }
  }
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
      <span>{$translate('common.loading')}</span>
    </div>
  </div>
{:else if view === 'dashboard'}
  <Dashboard />
{:else if view === 'unlock'}
  <UnlockPage />
{:else}
  <OnboardingPage />
{/if}
