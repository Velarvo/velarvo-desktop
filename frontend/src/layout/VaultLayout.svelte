<script lang="ts">
  import type { Snippet } from 'svelte'
  import { ShieldCheck } from '@lucide/svelte'
  import logoUrl from '@/assets/logos/main.png'
  import { t } from '@/lib/i18n.svelte'
  interface Props {
    children?: Snippet
  }

  let { children }: Props = $props()
</script>

<div class="min-h-screen w-full flex bg-background overflow-hidden">
  <div class="drag-region w-full absolute top-0 h-10 z-50"></div>

  <div
    class="hidden lg:flex lg:w-[45%] xl:w-[55%] relative flex-col justify-between overflow-hidden bg-surface-base border-r border-border p-10 xl:p-14"
  >
    <div class="hero-glow pointer-events-none absolute inset-0"></div>

    <div class="relative z-10 flex items-center">
      <img src={logoUrl} alt="Velarvo logo" class="h-17 w-17 rounded-lg p-1" />
      <span class="text-2xl font-bold tracking-tight text-white">Velarvo</span>
    </div>

    <div class="relative z-10 space-y-6">
      <h1
        class="text-4xl xl:text-5xl font-bold leading-tight tracking-tight text-white"
      >
        {t('vaultLayout.heroLine1')}
        <br />
        <span class="text-primary">{t('vaultLayout.heroLine2')}</span>
        <br />
        {t('vaultLayout.heroLine3Prefix')}
        <span class="text-primary">
          {t('vaultLayout.heroLine3Accent')}
        </span>
      </h1>
      <p class="max-w-sm text-base text-muted-foreground leading-relaxed">
        {t('vaultLayout.description')}
      </p>
    </div>
  </div>

  <div
    class="flex flex-1 w-[65%] flex-col items-center justify-center min-h-screen px-4 py-10 sm:px-8 lg:px-12 overflow-y-auto"
  >
    <div class="lg:hidden mb-10 flex w-full max-w-sm items-center">
      <div class="flex items-center gap-2">
        <img
          src={logoUrl}
          alt="Velarvo logo"
          class="h-7 w-7 rounded-lg object-cover"
        />
        <span class="text-base font-bold tracking-tight text-white">
          Velarvo
        </span>
      </div>
    </div>

    <div class="w-full max-w-sm">
      {@render children?.()}
      <p
        class="mt-8 flex items-center gap-1.5 text-[11px] text-muted-foreground/60"
      >
        <ShieldCheck class="h-3.5 w-3.5" />
        {t('common.e2eEncrypted')}
      </p>
    </div>
  </div>
</div>

<style>
  .hero-glow {
    background:
      radial-gradient(
        circle 420px at 120px 120px,
        rgb(var(--color-primary-rgb) / 0.1),
        transparent
      ),
      radial-gradient(
        circle 360px at calc(100% - 200px) calc(100% - 70px),
        rgb(139 92 246 / 0.08),
        transparent
      );
  }
</style>
