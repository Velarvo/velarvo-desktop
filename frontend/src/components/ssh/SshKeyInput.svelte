<script lang="ts">
  import { tick } from 'svelte'
  import {
    Check,
    ChevronDown,
    ClipboardPaste,
    Hash,
    KeyRound,
    Plus,
    Search,
    ShieldCheck,
    TriangleAlert,
  } from '@lucide/svelte'
  import FloatingDropdown from '@/components/ui/FloatingDropdown.svelte'
  import { t } from '@/lib/i18n.svelte'
  import { isPrivateKeyText, sshKeyAlgorithmLabel } from '@/lib/sshPresentation'
  import type { SSHKey } from '@/types/ssh'

  interface Props {
    keys?: SSHKey[]
    value?: string
    selectedKeyId?: string
    selectedKey?: SSHKey | null
    placeholder?: string
    label?: string
    busy?: boolean
    disabled?: boolean
    onCreate?: (detail: { label: string }) => void
    onImport?: () => void
    onSelect?: (detail: { key: SSHKey }) => void
    onDraftChange?: (detail: { value: string }) => void
  }

  let {
    keys = [],
    value = $bindable(''),
    selectedKeyId = $bindable(''),
    selectedKey = null,
    placeholder = t('ssh.keyPlaceholder', 'Search keys or paste a private key'),
    label = t('ssh.sshKey', 'SSH Key'),
    busy = false,
    disabled = false,
    onCreate,
    onImport,
    onSelect,
    onDraftChange,
  }: Props = $props()

  let open = $state(false)
  let activeIndex = $state(0)
  let inputElement = $state<HTMLTextAreaElement | null>(null)
  let fieldElement = $state<HTMLDivElement | null>(null)
  const inputId = `ssh-key-input-${Math.random().toString(36).slice(2, 10)}`
  const listboxId = `${inputId}-listbox`
  const optionId = (index: number) => `${inputId}-opt-${index}`

  const scrollActiveIntoView = async () => {
    await tick()
    const active = document.getElementById(optionId(activeIndex))
    active?.scrollIntoView({ block: 'nearest' })
  }

  const setActiveIndex = (index: number) => {
    activeIndex = index
    void scrollActiveIntoView()
  }

  const openPicker = () => {
    if (disabled || open) return
    open = true
    const selectedIndex = filteredKeys.findIndex(
      (key) => key.id === selectedKeyId,
    )
    activeIndex = Math.max(selectedIndex, 0)
    void scrollActiveIntoView()
  }

  const updateValue = (next: string) => {
    value = next
    if (selected && next !== selected.label) {
      selectedKeyId = ''
    }
    activeIndex = 0
    if (!open) open = true
    onDraftChange?.({ value: next })
  }

  const commitKey = (key: SSHKey) => {
    selectedKeyId = key.id
    value = key.label
    open = false
    onSelect?.({ key })
    void tick().then(() => inputElement?.focus())
  }

  const commitCreate = () => {
    open = false
    onCreate?.({ label: createLabel })
  }

  const commitImport = () => {
    if (!looksLikePrivateKey || busy) return
    open = false
    onImport?.()
  }

  const activateIndex = (index: number) => {
    if (showImport) {
      commitImport()
      return
    }

    if (showPublicWarning) return

    if (index < filteredKeys.length) {
      commitKey(filteredKeys[index])
      return
    }

    commitCreate()
  }

  const handleKeydown = (event: KeyboardEvent) => {
    if (event.isComposing || event.keyCode === 229) return

    if (!open && (event.key === 'ArrowDown' || event.key === 'ArrowUp')) {
      openPicker()
      event.preventDefault()
      return
    }

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault()
        setActiveIndex(optionCount ? (activeIndex + 1) % optionCount : 0)
        break
      case 'ArrowUp':
        event.preventDefault()
        setActiveIndex(
          optionCount ? (activeIndex - 1 + optionCount) % optionCount : 0,
        )
        break
      case 'Home':
        if (open && optionCount) {
          event.preventDefault()
          setActiveIndex(0)
        }
        break
      case 'End':
        if (open && optionCount) {
          event.preventDefault()
          setActiveIndex(optionCount - 1)
        }
        break
      case 'Enter':
        event.preventDefault()
        if (open && optionCount) activateIndex(activeIndex)
        else openPicker()
        break
      case 'Escape':
        if (open) {
          event.preventDefault()
          event.stopPropagation()
          open = false
        }
        break
      case 'Tab':
        open = false
        break
    }
  }

  let normalizedValue = $derived(value.trim())
  let selected = $derived(
    selectedKey ?? keys.find((key) => key.id === selectedKeyId) ?? null,
  )
  let isShowingSelection = $derived(
    Boolean(selected) && normalizedValue === selected?.label,
  )
  let normalizedSearch = $derived(
    isShowingSelection ? '' : normalizedValue.toLowerCase(),
  )
  let looksLikePrivateKey = $derived(isPrivateKeyText(normalizedValue))
  let publicKeyParts = $derived(normalizedValue.split(/\s+/))
  let publicKeyAlgorithm = $derived(
    /^(ssh-(rsa|ed25519)|ecdsa-sha2-[^\s]+)$/i.test(publicKeyParts[0] ?? '')
      ? publicKeyParts[0]
      : '',
  )
  let looksLikePublicKey = $derived(
    Boolean(publicKeyAlgorithm) && (publicKeyParts[1]?.length ?? 0) > 20,
  )
  let publicKeyComment = $derived(
    looksLikePublicKey ? publicKeyParts.slice(2).join(' ') : '',
  )
  let filteredKeys = $derived(
    normalizedSearch
      ? keys.filter((key) => {
          const haystack =
            `${key.label} ${key.keyType} ${key.fingerprintSha256} ${key.publicKey}`.toLowerCase()
          return haystack.includes(normalizedSearch)
        })
      : keys,
  )
  let exactMatch = $derived(
    normalizedSearch.length > 0
      ? (keys.find(
          (key) =>
            key.label.toLowerCase() === normalizedSearch ||
            key.fingerprintSha256.toLowerCase() === normalizedSearch ||
            key.publicKey.trim() === normalizedValue,
        ) ?? null)
      : null,
  )
  let showImport = $derived(looksLikePrivateKey)
  let showPublicWarning = $derived(looksLikePublicKey && !looksLikePrivateKey)
  let showCreate = $derived(!showImport && !showPublicWarning)
  let createLabel = $derived(
    normalizedSearch && !exactMatch ? normalizedValue : '',
  )
  let createIndex = $derived(filteredKeys.length)
  let optionCount = $derived(
    showImport || showPublicWarning ? 1 : filteredKeys.length + 1,
  )

  $effect.pre(() => {
    if (activeIndex >= optionCount) activeIndex = Math.max(optionCount - 1, 0)
  })

  export const focus = () => {
    openPicker()
    void tick().then(() => inputElement?.focus())
  }
</script>

<div class="flex flex-col gap-1.5">
  <label
    for={inputId}
    class={label ? 'text-xs font-medium text-muted-foreground' : 'sr-only'}
  >
    {label || t('ssh.sshKey', 'SSH Key')}
  </label>

  <div
    bind:this={fieldElement}
    class={`flex min-h-9 items-center gap-2 rounded-md border bg-fill pl-3 pr-1.5 transition focus-within:bg-fill-hover ${
      open
        ? 'border-primary/50'
        : 'border-border hover:border-line-strong focus-within:border-primary/50'
    }`}
  >
    <span class="shrink-0 text-muted-foreground" aria-hidden="true">
      {#if looksLikePrivateKey}
        <ShieldCheck class="h-4 w-4 text-primary" />
      {:else if showPublicWarning}
        <TriangleAlert class="h-4 w-4 text-warning" />
      {:else if selectedKeyId}
        <KeyRound class="h-4 w-4 text-primary" />
      {:else}
        <Search class="h-4 w-4" />
      {/if}
    </span>

    <textarea
      bind:this={inputElement}
      rows="1"
      role="combobox"
      aria-expanded={open}
      aria-controls={listboxId}
      aria-autocomplete="list"
      aria-activedescendant={open && optionCount
        ? optionId(activeIndex)
        : undefined}
      id={inputId}
      autocomplete="off"
      autocorrect="off"
      autocapitalize="off"
      spellcheck="false"
      {disabled}
      {placeholder}
      {value}
      class={`h-9 min-w-0 flex-1 resize-none overflow-hidden bg-transparent py-2 text-sm leading-5 text-white outline-none placeholder:text-muted-foreground/60 disabled:cursor-not-allowed disabled:opacity-60 ${
        looksLikePrivateKey || showPublicWarning ? 'font-mono' : ''
      }`}
      onfocus={(event) => {
        openPicker()
        if (isShowingSelection) event.currentTarget.select()
      }}
      onclick={openPicker}
      oninput={(event) =>
        updateValue((event.currentTarget as HTMLTextAreaElement).value)}
      onkeydown={handleKeydown}></textarea>

    {#if looksLikePrivateKey}
      <span
        class="shrink-0 rounded-md bg-primary/10 px-2 py-0.5 text-xs font-semibold text-primary"
      >
        {t('ssh.privateKey', 'Private key')}
      </span>
    {:else if publicKeyAlgorithm}
      <span
        class="max-w-28 shrink-0 truncate rounded-md bg-warning/10 px-2 py-0.5 font-mono text-xs font-semibold text-warning"
      >
        {publicKeyAlgorithm}
      </span>
    {:else if selected && selectedKeyId}
      <span
        class="shrink-0 rounded-md bg-muted px-1.5 py-0.5 font-mono text-[10px] font-medium text-muted-foreground"
      >
        {sshKeyAlgorithmLabel(selected.keyType)}
      </span>
    {/if}

    <button
      type="button"
      tabindex="-1"
      {disabled}
      aria-label={t('ssh.showSavedKeys', 'Show saved keys')}
      class="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition hover:bg-fill-hover hover:text-white"
      onmousedown={(event) => {
        event.preventDefault()
        if (open) {
          open = false
        } else {
          inputElement?.focus()
          openPicker()
        }
      }}
    >
      <ChevronDown class={`h-4 w-4 transition ${open ? 'rotate-180' : ''}`} />
    </button>
  </div>

  <FloatingDropdown
    open={open && !disabled}
    anchorElement={fieldElement}
    placement="bottom-start"
    matchAnchorWidth
    offset={6}
    group="ssh-key-picker"
    panelClass="flex flex-col overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-xl shadow-black/40"
    onClose={() => (open = false)}
  >
    <ul
      id={listboxId}
      role="listbox"
      aria-label={t('ssh.savedKeys', 'Saved keys')}
      class="max-h-64 min-h-0 flex-1 overflow-y-auto p-1"
    >
      {#if showImport}
        <li
          id={optionId(0)}
          role="option"
          aria-selected="false"
          data-active={activeIndex === 0}
          class={`flex cursor-pointer items-center gap-3 rounded px-2.5 py-2 text-sm ${
            activeIndex === 0 ? 'bg-fill-hover' : ''
          }`}
          onmouseenter={() => setActiveIndex(0)}
          onmousedown={(event) => {
            event.preventDefault()
            void commitImport()
          }}
        >
          <span
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary"
          >
            <ClipboardPaste class="h-4 w-4" />
          </span>
          <span class="flex min-w-0 flex-col">
            <span class="truncate font-semibold text-white">
              {t('ssh.importPastedKey', 'Import pasted private key')}
            </span>
            <span class="truncate text-xs text-muted-foreground">
              {t(
                'ssh.importKeyHint',
                'Private key will be encrypted in the vault.',
              )}
            </span>
          </span>
        </li>
      {:else if showPublicWarning}
        <li
          id={optionId(0)}
          role="option"
          aria-disabled="true"
          aria-selected="false"
          data-active={activeIndex === 0}
          class="flex items-start gap-3 rounded-md border border-warning/25 bg-warning/10 px-3 py-2.5"
        >
          <TriangleAlert class="mt-0.5 h-4 w-4 shrink-0 text-warning" />
          <span class="min-w-0">
            <span class="block truncate text-sm font-semibold text-white">
              {t('ssh.publicKeyOnly', 'Public key only')}
            </span>
            <span class="mt-0.5 block truncate text-xs text-muted-foreground">
              {publicKeyComment ||
                t(
                  'ssh.publicKeyOnlyHint',
                  'Paste the private key to create a usable SSH credential.',
                )}
            </span>
          </span>
        </li>
      {:else}
        <li
          role="presentation"
          class="flex items-center justify-between px-2.5 pb-1 pt-1.5 text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground"
        >
          <span>{t('ssh.savedKeys', 'Saved keys')}</span>
          <span class="font-mono tracking-normal">
            {normalizedSearch
              ? `${filteredKeys.length}/${keys.length}`
              : keys.length}
          </span>
        </li>

        {#each filteredKeys as key, index (key.id)}
          {@const isActive = index === activeIndex}
          {@const isSelected = key.id === selectedKeyId}
          <li
            id={optionId(index)}
            role="option"
            aria-selected={isSelected}
            data-active={isActive}
            class={`flex cursor-pointer items-center gap-3 rounded px-2.5 py-2 text-sm ${
              isActive ? 'bg-fill-hover' : ''
            }`}
            onmouseenter={() => setActiveIndex(index)}
            onmousedown={(event) => {
              event.preventDefault()
              commitKey(key)
            }}
          >
            <span
              class={`flex h-8 w-8 shrink-0 items-center justify-center rounded-md ${
                isSelected
                  ? 'bg-primary/15 text-primary'
                  : 'bg-muted text-muted-foreground'
              }`}
            >
              <KeyRound class="h-4 w-4" />
            </span>
            <span class="flex min-w-0 flex-1 flex-col">
              <span class="flex min-w-0 items-center gap-2">
                <span class="truncate font-semibold text-white">
                  {key.label}
                </span>
                <span
                  class="shrink-0 rounded bg-muted px-1.5 py-0.5 font-mono text-[10px] font-medium text-muted-foreground"
                >
                  {sshKeyAlgorithmLabel(key.keyType)}
                </span>
                {#if key.hasPassphrase}
                  <span
                    class="shrink-0 text-[10px] font-medium text-muted-foreground"
                  >
                    {t('ssh.passphraseProtected', 'Passphrase')}
                  </span>
                {/if}
              </span>
              <span
                class="mt-0.5 flex min-w-0 items-center gap-1 font-mono text-[11px] text-muted-foreground"
              >
                <Hash class="h-3 w-3 shrink-0 opacity-60" />
                <span class="truncate">{key.fingerprintSha256}</span>
              </span>
            </span>
            {#if isSelected}
              <Check class="ml-auto h-4 w-4 shrink-0 text-primary" />
            {/if}
          </li>
        {/each}

        {#if filteredKeys.length === 0}
          <li
            role="presentation"
            class="flex flex-col items-center gap-1 px-3 py-5 text-center"
          >
            <span class="text-sm font-medium text-white">
              {normalizedSearch
                ? t('ssh.noMatchingKeys', 'No saved key matches')
                : t('ssh.noSavedKeys', 'No saved keys yet.')}
            </span>
            <span class="text-xs text-muted-foreground">
              {t(
                'ssh.keyEmptyHint',
                'Create a new key below, or paste a private key to import it.',
              )}
            </span>
          </li>
        {/if}
      {/if}
    </ul>

    {#if showCreate}
      <div
        id={optionId(createIndex)}
        role="option"
        tabindex="-1"
        aria-selected="false"
        data-active={activeIndex === createIndex}
        class={`flex w-full shrink-0 cursor-pointer items-center gap-3 border-t border-border px-3.5 py-2.5 text-left text-sm transition ${
          activeIndex === createIndex ? 'bg-fill-hover' : 'bg-fill-subtle'
        }`}
        onmouseenter={() => setActiveIndex(createIndex)}
        onmousedown={(event) => {
          event.preventDefault()
          commitCreate()
        }}
      >
        <span
          class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-dashed border-primary/40 text-primary"
        >
          <Plus class="h-4 w-4" />
        </span>
        <span class="flex min-w-0 flex-1 flex-col">
          <span class="truncate font-semibold text-white">
            {createLabel
              ? `${t('ssh.createKeyNamed', 'Create key')} "${createLabel}"`
              : t('ssh.createNewKey', 'Create new key')}
          </span>
          <span class="truncate text-xs text-muted-foreground">
            {t(
              'ssh.createKeyHint',
              'Generate an Ed25519, RSA or ECDSA key pair',
            )}
          </span>
        </span>
      </div>
    {/if}
  </FloatingDropdown>

  {#if showPublicWarning}
    <p class="flex items-center gap-1.5 text-xs text-warning">
      <TriangleAlert class="h-3.5 w-3.5" />
      {t(
        'ssh.publicKeyOnlyHint',
        'Paste the private key to create a usable SSH credential.',
      )}
    </p>
  {/if}
</div>
