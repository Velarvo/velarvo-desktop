<script lang="ts">
  import { onMount } from 'svelte'
  import { route, navigate } from '@/lib/router'
  import { bootstrapAuth, isAuth, isLoadingAuth } from '@/lib/auth'
  import LoginPage from '@/pages/auth/LoginPage.svelte'
  import RegisterPage from '@/pages/auth/RegisterPage.svelte'
  import OnboardingPage from '@/pages/auth/OnboardingPage.svelte'
  import Dashboard from '@/pages/Dashboard.svelte'

  onMount(async () => {
    await bootstrapAuth()
  })

  $: path = $route || '/'
  $: showDashboard = path === '/dashboard'
  $: showLogin = path === '/login'
  $: showRegister = path === '/register'
  $: showOnboarding = path === '/'

  $: if (!$isLoadingAuth && !$isAuth && showDashboard) {
    navigate('/login')
  }
</script>

{#if showDashboard}
  <Dashboard />
{:else if showLogin}
  <LoginPage />
{:else if showRegister}
  <RegisterPage />
{:else if showOnboarding}
  <OnboardingPage />
{:else}
  <OnboardingPage />
{/if}
