<script lang="ts">
  import { Globe2 } from 'lucide-svelte'
  import SectionHero from '@/components/ui/SectionHero.svelte'
  import {
    currentLanguage,
    LANGUAGES,
    setLanguage,
    translate,
  } from '@/lib/i18n'
</script>

<section class="space-y-4">
  <header>
    <h3 class="text-xl font-semibold text-white">
      {$translate('settings.preferences.title')}
    </h3>
    <p class="mt-1 max-w-xl text-sm leading-6 text-muted-foreground">
      {$translate('settings.preferences.subtitle')}
    </p>
  </header>

  <SectionHero
    icon={Globe2}
    title={$translate('settings.preferences.languageTitle')}
    description={$translate('settings.preferences.languageDescription')}
  >
    <div class="flex flex-wrap gap-2">
      {#each LANGUAGES as language (language.code)}
        {@const isActive = $currentLanguage === language.code}
        <button
          type="button"
          class={`inline-flex items-center gap-2 rounded-xl border px-3 py-2 text-sm font-medium transition ${
            isActive
              ? 'border-primary/30 bg-primary/10 text-primary'
              : 'border-white/10 bg-white/4 text-white/80 hover:border-white/15 hover:bg-white/[0.07] hover:text-white'
          }`}
          on:click={() => setLanguage(language.code)}
        >
          <span class="text-[11px] font-semibold uppercase tracking-[0.12em]">
            {language.flag}
          </span>
          <span>{language.label}</span>
        </button>
      {/each}
    </div>
  </SectionHero>
</section>
