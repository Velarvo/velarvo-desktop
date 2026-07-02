<script lang="ts">
  import logoUrl from '@/assets/logos/main.png'
  import { translate } from '@/lib/i18n'
</script>

<div class="min-h-screen w-full flex bg-background select-none overflow-hidden">
  <div class="drag-region w-full absolute top-0 h-10 z-50"></div>

  <div
    class="hidden lg:flex lg:w-[45%] xl:w-[55%] relative flex-col justify-between overflow-hidden bg-surface-base border-r border-border p-10 xl:p-14"
  >
    <div class="pointer-events-none absolute inset-0">
      <div
        class="absolute -left-32 -top-32 h-[500px] w-[500px] rounded-full bg-primary/10 blur-[120px]"
      ></div>
      <div
        class="absolute -bottom-32 right-0 h-[400px] w-[400px] rounded-full bg-violet-500/8 blur-[120px]"
      ></div>
    </div>

    <div class="relative z-10 flex items-center">
      <img src={logoUrl} alt="Velarvo logo" class="h-17 w-17 rounded-lg p-1" />
      <span class="text-2xl font-bold tracking-tight text-white">Velarvo</span>
    </div>

    <div class="relative z-10 space-y-6">
      <h1
        class="text-4xl xl:text-5xl font-bold leading-tight tracking-tight text-white"
      >
        {$translate('authLayout.heroLine1')}
        <br />
        <span class="text-primary">{$translate('authLayout.heroLine2')}</span>
        <br />
        {$translate('authLayout.heroLine3Prefix')}
        <span class="text-primary"
          >{$translate('authLayout.heroLine3Accent')}</span
        >
      </h1>
      <p class="max-w-sm text-base text-muted-foreground leading-relaxed">
        {$translate('authLayout.description')}
      </p>
    </div>
  </div>

  <div
    class="flex flex-1 w-[65%] flex-col items-center justify-center min-h-screen px-4 py-10 sm:px-8 lg:px-12 overflow-y-auto"
  >
    <div
      class="lg:hidden flex items-center justify-between mb-8 self-start w-full max-w-[500px]"
    >
      <div class="flex items-center gap-2">
        <img
          src={logoUrl}
          alt="Velarvo logo"
          class="h-7 w-7 rounded-lg object-cover"
        />
        <span class="text-base font-bold tracking-tight text-white"
          >Velarvo</span
        >
      </div>
    </div>

    <div class="w-full max-w-[500px]"><slot /></div>
    <p class="mt-6 text-center text-[11px] text-muted-foreground/60">
      {$translate('common.e2eEncrypted')}
    </p>
  </div>
</div>
