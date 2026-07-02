<script lang="ts">
  import { Terminal } from 'lucide-svelte'
  import {
    siAlpinelinux,
    siApple,
    siArchlinux,
    siCentos,
    siDebian,
    siFedora,
    siFreebsd,
    siLinux,
    siRedhat,
    siUbuntu,
  } from 'simple-icons'

  export let os: string = ''
  export let size = 17

  const BRAND: Record<string, { path: string }> = {
    ubuntu: siUbuntu,
    debian: siDebian,
    alpine: siAlpinelinux,
    fedora: siFedora,
    centos: siCentos,
    rhel: siRedhat,
    arch: siArchlinux,
    linux: siLinux,
    macos: siApple,
    freebsd: siFreebsd,
  }

  $: brand = BRAND[os]
</script>

{#if os === 'windows'}
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="currentColor"
    aria-hidden="true"
  >
    <rect x="3" y="3" width="8" height="8" rx="1" />
    <rect x="13" y="3" width="8" height="8" rx="1" />
    <rect x="3" y="13" width="8" height="8" rx="1" />
    <rect x="13" y="13" width="8" height="8" rx="1" />
  </svg>
{:else if brand}
  <svg
    width={size}
    height={size}
    viewBox="0 0 24 24"
    fill="currentColor"
    aria-hidden="true"
  >
    <path d={brand.path} />
  </svg>
{:else}
  <Terminal {size} strokeWidth={2.1} />
{/if}
