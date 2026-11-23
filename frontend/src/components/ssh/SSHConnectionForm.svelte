<script lang="ts">
  import { tooltip } from '@/components/ui/tooltip'
  import { onDestroy, onMount, tick } from 'svelte'
  import { Check, Copy, Info, Plus, X } from '@lucide/svelte'
  import Alert from '@/components/ui/Alert.svelte'
  import Button from '@/components/ui/Button.svelte'
  import Field from '@/components/ui/Field.svelte'
  import FormSection from '@/components/ui/FormSection.svelte'
  import IconButton from '@/components/ui/IconButton.svelte'
  import Input from '@/components/ui/Input.svelte'
  import Select, { type SelectOption } from '@/components/ui/Select.svelte'
  import SshKeyInput from '@/components/ssh/SshKeyInput.svelte'
  import { copyText } from '@/lib/clipboard'
  import { describeError } from '@/lib/errors'
  import { isPrivateKeyText } from '@/lib/sshPresentation'
  import { t } from '@/lib/i18n.svelte'
  import { TransientFlag } from '@/lib/transient.svelte'
  import {
    createSSHConnection,
    createSSHKey,
    importSSHKey,
    refreshSSHKeys,
    selectSSHConnection,
    updateSSHConnection,
    sshState,
  } from '@/lib/sshConnections.svelte'
  import {
    closeTab,
    EditorTabKind,
    replaceTab,
    updateTab,
  } from '@/lib/editorLayout.svelte'
  import {
    isSSHConnected,
    SSH_DEFAULT_PORT,
    SSHKeyType,
    type SSHConnection,
    type SSHConnectionInput,
    type SSHKey,
  } from '@/types/ssh'

  interface Props {
    tabId: string
    groupId: string
    connectionId?: string | null
  }

  let { tabId, groupId, connectionId = $bindable(null) }: Props = $props()

  let name = $state('')
  let host = $state('')
  let port = $state(SSH_DEFAULT_PORT)
  let username = $state('')
  let password = $state('')

  let passwordEnabled = $state(true)
  let keyEnabled = $state(false)

  let selectedKeyId = $state('')
  let keyLabel = $state('')
  let keyQuery = $state('')
  let keyType = $state(SSHKeyType.Ed25519)
  let keyPassphrase = $state('')
  let keyPicker = $state<ReturnType<typeof SshKeyInput>>()
  let hostInput = $state<HTMLInputElement | null>(null)
  let userInput = $state<HTMLInputElement | null>(null)
  let passwordInput = $state<HTMLInputElement | null>(null)
  let attempted = $state(false)
  let showKeyCreatePanel = $state(false)
  let keyBusy = $state(false)
  let lastCreatedKey = $state<SSHKey | null>(null)
  let showPublicKey = $state(false)
  const copiedPublicKey = new TransientFlag(2200)

  const isMac =
    typeof navigator !== 'undefined' && /Mac/i.test(navigator.platform)
  const modKey = isMac ? '⌘' : 'Ctrl'

  const justSaved = new TransientFlag(2500)
  let busy = $state(false)
  let errorMessage = $state('')

  const keyTypeOptions = [
    { value: SSHKeyType.Ed25519, label: 'Ed25519' },
    { value: SSHKeyType.RSA4096, label: 'RSA 4096' },
    { value: SSHKeyType.ECDSA384, label: 'ECDSA P-384' },
  ] satisfies SelectOption[]

  let hydratedFor: string | null | undefined = undefined

  onMount(() => {
    void refreshSSHKeys()
  })

  onDestroy(() => {
    justSaved.clear()
    copiedPublicKey.clear()
  })

  const hydrate = (id: string | null) => {
    if (id === hydratedFor) return
    hydratedFor = id
    if (!id) return

    const connection = sshState.connections.find((item) => item.id === id)
    if (!connection) return

    name = connection.name
    host = connection.host
    port = connection.port
    username = connection.username
    passwordEnabled = connection.hasPassword
    keyEnabled = Boolean(connection.keyId)
    selectedKeyId = connection.keyId ?? ''
    keyQuery = connection.keyLabel ?? ''
    if (!passwordEnabled && !keyEnabled) passwordEnabled = true
  }

  const addPassword = () => {
    passwordEnabled = true
    errorMessage = ''
    void tick().then(() => passwordInput?.focus())
  }

  const removePassword = () => {
    passwordEnabled = false
    password = ''
  }

  const addKeyCredential = () => {
    keyEnabled = true
    errorMessage = ''
    void tick().then(() => keyPicker?.focus())
  }

  const removeKeyCredential = () => {
    keyEnabled = false
    showKeyCreatePanel = false
    selectedKeyId = ''
    keyQuery = ''
    keyLabel = ''
    keyPassphrase = ''
    lastCreatedKey = null
    showPublicKey = false
  }

  const buildInput = (): SSHConnectionInput => ({
    name,
    host,
    port: Number(port) || SSH_DEFAULT_PORT,
    username,
    password: passwordEnabled ? password : '',
    clearPassword: !passwordEnabled,
    keyId: keyEnabled ? selectedKeyId : '',
  })

  const persist = async (): Promise<SSHConnection> => {
    const input = buildInput()
    let saved: SSHConnection

    if (connectionId) {
      saved = await updateSSHConnection(connectionId, input)
      updateTab(tabId, { title: saved.name })
    } else {
      saved = await createSSHConnection(input)
      connectionId = saved.id
      updateTab(tabId, {
        title: saved.name,
        data: { connectionId: saved.id },
      })
    }

    selectSSHConnection(saved.id)
    password = ''
    return saved
  }

  const submit = async (
    onSaved: (connection: SSHConnection) => void,
  ): Promise<void> => {
    if (busy || keyBusy) return
    if (missingField) {
      attempted = true
      focusMissing()
      return
    }
    busy = true
    errorMessage = ''
    try {
      onSaved(await persist())
    } catch (error) {
      errorMessage = describeError(error)
    } finally {
      busy = false
    }
  }

  const save = () => submit(() => justSaved.raise())

  const connect = () =>
    submit((connection) =>
      replaceTab(tabId, {
        title: connection.name || connection.host,
        kind: EditorTabKind.SSHTerminal,
        data: { connectionId: connection.id },
      }),
    )

  const adoptKey = async (produce: () => Promise<SSHKey>) => {
    if (keyBusy) return
    keyBusy = true
    errorMessage = ''
    try {
      const key = await produce()
      lastCreatedKey = key
      selectedKeyId = key.id
      keyQuery = key.label
      keyEnabled = true
      keyPassphrase = ''
      keyLabel = ''
      showKeyCreatePanel = false
    } catch (error) {
      errorMessage = describeError(error)
    } finally {
      keyBusy = false
    }
  }

  const generateKey = () =>
    adoptKey(() =>
      createSSHKey({
        label: keyLabel.trim() || defaultKeyLabel,
        keyType,
        passphrase: keyPassphrase,
      }),
    )

  const importPastedKey = () => {
    if (!looksLikePrivateKey) return
    return adoptKey(() =>
      importSSHKey({
        label: keyLabel.trim() || `${host || 'Imported'} SSH key`,
        privateKey: normalizedKeyQuery,
        passphrase: keyPassphrase,
      }),
    )
  }

  const selectKey = (key: SSHKey) => {
    selectedKeyId = key.id
    keyQuery = key.label
    keyEnabled = true
    lastCreatedKey = null
    showPublicKey = false
    showKeyCreatePanel = false
  }

  const handleKeyDraftChange = () => {
    keyEnabled = true
    if (selectedKey && keyQuery !== selectedKey.label) {
      selectedKeyId = ''
      lastCreatedKey = null
    }
  }

  const startKeyCreation = (label: string) => {
    keyLabel = label
    keyPassphrase = ''
    showKeyCreatePanel = true
    lastCreatedKey = null
  }

  const cancelKeyCreation = () => {
    showKeyCreatePanel = false
    keyLabel = ''
    keyPassphrase = ''
  }

  const focusMissing = () => {
    if (missingField === 'host') hostInput?.focus()
    else if (missingField === 'user') userInput?.focus()
    else if (keyEnabled) keyPicker?.focus()
    else if (passwordEnabled) passwordInput?.focus()
    else addKeyCredential()
  }

  const handleShortcut = (event: KeyboardEvent) => {
    if (!(event.metaKey || event.ctrlKey) || event.isComposing) return
    if (event.key === 'Enter') {
      event.preventDefault()
      void connect()
    } else if (event.key.toLowerCase() === 's') {
      event.preventDefault()
      void save()
    }
  }

  const copyPublicKey = async () => {
    if (publicKeyValue && (await copyText(publicKeyValue))) {
      copiedPublicKey.raise()
    }
  }

  const cancel = () => closeTab(groupId, tabId)
  $effect.pre(() => {
    hydrate(connectionId)
  })
  let storeConnection = $derived(
    connectionId !== null
      ? (sshState.connections.find((item) => item.id === connectionId) ?? null)
      : null,
  )
  let isEditing = $derived(connectionId !== null)
  let hasStoredPassword = $derived(storeConnection?.hasPassword ?? false)
  let isConnected = $derived(
    storeConnection ? isSSHConnected(storeConnection) : false,
  )
  let selectedKey = $derived(
    sshState.keys.find((item) => item.id === selectedKeyId) ??
      (storeConnection?.keyId === selectedKeyId
        ? ({
            id: selectedKeyId,
            workspaceId: storeConnection.workspaceId,
            label: storeConnection.keyLabel ?? '',
            keyType: SSHKeyType.Ed25519,
            publicKey: '',
            fingerprintSha256: storeConnection.keyFingerprint ?? '',
            hasPassphrase: false,
            createdAt: 0,
          } satisfies SSHKey)
        : null),
  )
  let normalizedKeyQuery = $derived(keyQuery.trim())
  let looksLikePrivateKey = $derived(isPrivateKeyText(normalizedKeyQuery))
  let publicKeyValue = $derived(
    selectedKey?.publicKey || lastCreatedKey?.publicKey || '',
  )
  let publicKeyFingerprint = $derived(
    selectedKey?.fingerprintSha256 || lastCreatedKey?.fingerprintSha256 || '',
  )
  let passwordReady = $derived(
    passwordEnabled &&
      (password.trim().length > 0 || (isEditing && hasStoredPassword)),
  )
  let keyReady = $derived(keyEnabled && selectedKeyId.trim().length > 0)
  let defaultKeyLabel = $derived(`${host.trim() || 'SSH'} key`)
  let missingField = $derived<'host' | 'user' | 'auth' | null>(
    !host.trim()
      ? 'host'
      : !username.trim()
        ? 'user'
        : !passwordReady && !keyReady
          ? 'auth'
          : null,
  )
  let showMissing = $derived(attempted ? missingField : null)
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="@container flex h-full min-h-0 w-full flex-col bg-surface-canvas"
  onkeydown={handleShortcut}
>
  <div class="min-h-0 flex-1 overflow-y-auto">
    <div class="mx-auto w-full max-w-xl px-5 py-7 @2xl:py-10">
      {#if errorMessage}
        <Alert variant="destructive" message={errorMessage} />
        <div class="h-6"></div>
      {/if}

      <FormSection title={t('ssh.sectionConnection', 'Connection')}>
        <Field
          layout="inline"
          label={t('ssh.fieldName', 'Name')}
          for="ssh-name"
        >
          <Input
            id="ssh-name"
            bind:value={name}
            placeholder={host.trim() || 'Production API'}
          />
        </Field>

        <Field
          layout="inline"
          label={t('ssh.fieldHost', 'Host')}
          for="ssh-host"
          invalid={showMissing === 'host'}
        >
          <Input
            id="ssh-host"
            mono
            bind:ref={hostInput}
            bind:value={host}
            placeholder="api-01.production.internal"
            autocomplete="off"
            autocapitalize="off"
            spellcheck="false"
            invalid={showMissing === 'host'}
          />
        </Field>

        <Field
          layout="inline"
          label={t('ssh.fieldPort', 'Port')}
          for="ssh-port"
        >
          <Input
            id="ssh-port"
            mono
            type="number"
            min="1"
            max="65535"
            class="w-28"
            bind:value={port}
            placeholder={String(SSH_DEFAULT_PORT)}
          />
        </Field>

        <Field
          layout="inline"
          label={t('ssh.fieldUser', 'Username')}
          for="ssh-user"
          invalid={showMissing === 'user'}
        >
          <Input
            id="ssh-user"
            mono
            bind:ref={userInput}
            bind:value={username}
            placeholder="root"
            autocomplete="off"
            autocapitalize="off"
            spellcheck="false"
            invalid={showMissing === 'user'}
          />
        </Field>
      </FormSection>

      <div class="my-7 h-px bg-border"></div>

      <FormSection title={t('ssh.sectionAuth', 'Authentication')}>
        {#if keyEnabled}
          <Field
            layout="inline"
            label={t('ssh.fieldKey', 'SSH key')}
            invalid={showMissing === 'auth'}
          >
            <div class="flex items-start gap-1">
              <div class="min-w-0 flex-1">
                <SshKeyInput
                  label=""
                  keys={sshState.keys}
                  placeholder={t(
                    'ssh.keyPlaceholderShort',
                    'Choose or create a key',
                  )}
                  bind:value={keyQuery}
                  bind:selectedKeyId
                  {selectedKey}
                  busy={keyBusy}
                  bind:this={keyPicker}
                  onSelect={(detail) => selectKey(detail.key)}
                  onCreate={(detail) => startKeyCreation(detail.label)}
                  onImport={importPastedKey}
                  onDraftChange={handleKeyDraftChange}
                />
              </div>
              <IconButton
                icon={X}
                class="mt-0.5"
                label={t('ssh.removeKey', 'Remove key')}
                onclick={removeKeyCredential}
              />
            </div>

            {#if selectedKey && !showKeyCreatePanel}
              <dl
                class="mr-9 mt-2 divide-y divide-border overflow-hidden rounded-md border border-border bg-fill-subtle text-xs"
              >
                <div
                  class="grid grid-cols-[5.5rem_minmax(0,1fr)] items-center gap-3 px-3 py-2"
                >
                  <dt class="text-muted-foreground">
                    {t('ssh.fingerprint', 'Fingerprint')}
                  </dt>
                  <dd class="flex min-w-0 items-center gap-2">
                    <span
                      class="min-w-0 truncate font-mono text-white/85"
                      {@attach tooltip({
                        content: publicKeyFingerprint,
                        overflowOnly: true,
                      })}
                    >
                      {publicKeyFingerprint}
                    </span>
                    {#if selectedKey.hasPassphrase}
                      <span
                        class="shrink-0 rounded border border-border px-1.5 py-px text-[10px] text-muted-foreground"
                      >
                        {t('ssh.passphraseProtected', 'Passphrase')}
                      </span>
                    {/if}
                  </dd>
                </div>

                {#if publicKeyValue}
                  <div
                    class="grid grid-cols-[5.5rem_minmax(0,1fr)_auto] items-start gap-3 px-3 py-2"
                  >
                    <dt class="flex h-6 items-center text-muted-foreground">
                      {t('ssh.publicKey', 'Public key')}
                    </dt>
                    <dd class="min-w-0">
                      <button
                        type="button"
                        class={`block w-full min-w-0 py-1 text-left font-mono text-white/85 transition hover:text-white ${
                          showPublicKey
                            ? 'break-all leading-relaxed'
                            : 'truncate'
                        }`}
                        aria-expanded={showPublicKey}
                        {@attach tooltip({
                          content: showPublicKey
                            ? t('ssh.collapsePublicKey', 'Collapse')
                            : t('ssh.expandPublicKey', 'Show the whole key'),
                        })}
                        onclick={() => (showPublicKey = !showPublicKey)}
                      >
                        {publicKeyValue}
                      </button>
                    </dd>
                    <Button
                      variant="secondary"
                      size="xs"
                      onclick={copyPublicKey}
                    >
                      {#if copiedPublicKey.current}
                        <Check class="h-3 w-3 text-primary" />
                        {t('common.copied', 'Copied')}
                      {:else}
                        <Copy class="h-3 w-3" />
                        {t('common.copy', 'Copy')}
                      {/if}
                    </Button>
                  </div>
                {/if}

                {#if lastCreatedKey}
                  <p
                    class="flex items-start gap-2 bg-primary/5 px-3 py-2 text-primary/90"
                  >
                    <Info class="mt-px h-3.5 w-3.5 shrink-0" />
                    <span>
                      {t(
                        'ssh.authorizedKeysHint',
                        'Add the public key to ~/.ssh/authorized_keys on the server.',
                      )}
                    </span>
                  </p>
                {/if}
              </dl>
            {/if}

            {#if showKeyCreatePanel || looksLikePrivateKey}
              <div
                class="mr-9 mt-2.5 space-y-3 rounded-md border border-border bg-fill-subtle p-3"
              >
                <p class="text-sm font-medium text-white">
                  {showKeyCreatePanel
                    ? t('ssh.newKeyTitle', 'New key')
                    : t('ssh.importKeyTitle', 'Import pasted key')}
                </p>

                <Field label={t('ssh.keyName', 'Name')} for="ssh-key-name">
                  <Input
                    id="ssh-key-name"
                    bind:value={keyLabel}
                    placeholder={defaultKeyLabel}
                  />
                </Field>

                {#if showKeyCreatePanel}
                  <Field label={t('ssh.keyType', 'Type')}>
                    <div class="max-w-44">
                      <Select bind:value={keyType} options={keyTypeOptions} />
                    </div>
                  </Field>
                {/if}

                <Field
                  label={t('ssh.keyPassphrase', 'Passphrase')}
                  for="ssh-key-passphrase"
                  optional
                >
                  <Input
                    id="ssh-key-passphrase"
                    type="password"
                    autocomplete="off"
                    bind:value={keyPassphrase}
                  />
                </Field>

                <div class="flex justify-end gap-2 pt-1">
                  <Button
                    variant="ghost"
                    size="sm"
                    onclick={showKeyCreatePanel
                      ? cancelKeyCreation
                      : () => (keyQuery = '')}
                  >
                    {t('common.cancel', 'Cancel')}
                  </Button>
                  <Button
                    size="sm"
                    loading={keyBusy}
                    onclick={showKeyCreatePanel ? generateKey : importPastedKey}
                  >
                    {showKeyCreatePanel
                      ? t('ssh.generate', 'Generate')
                      : t('ssh.import', 'Import')}
                  </Button>
                </div>
              </div>
            {/if}
          </Field>
        {/if}

        {#if passwordEnabled}
          <Field
            layout="inline"
            label={t('ssh.fieldPassword', 'Password')}
            for="ssh-password"
            invalid={showMissing === 'auth' && !keyEnabled}
          >
            <div class="flex items-center gap-1">
              <Input
                id="ssh-password"
                type="password"
                autocomplete="off"
                bind:ref={passwordInput}
                bind:value={password}
                placeholder={hasStoredPassword
                  ? t('ssh.passwordKeep', '•••••••• (unchanged)')
                  : ''}
              />
              <IconButton
                icon={X}
                label={t('ssh.removePassword', 'Remove password')}
                onclick={removePassword}
              />
            </div>
          </Field>
        {/if}

        {#if !keyEnabled || !passwordEnabled}
          <div class="flex items-center gap-1 @md:pl-[8.5rem]">
            {#if !keyEnabled}
              <Button
                variant="ghost"
                size="sm"
                class={`-ml-3 ${showMissing === 'auth' ? 'text-red-300!' : ''}`}
                onclick={addKeyCredential}
              >
                <Plus class="h-4 w-4" />
                {t('ssh.addKey', 'SSH key')}
              </Button>
            {/if}
            {#if !passwordEnabled}
              <Button
                variant="ghost"
                size="sm"
                class={`${keyEnabled ? '-ml-3' : ''} ${
                  showMissing === 'auth' && !keyEnabled ? 'text-red-300!' : ''
                }`}
                onclick={addPassword}
              >
                <Plus class="h-4 w-4" />
                {t('ssh.addPassword', 'Password')}
              </Button>
            {/if}
          </div>
        {/if}
      </FormSection>
    </div>
  </div>

  <footer
    class="flex shrink-0 items-center gap-2 border-t border-border bg-surface-base px-4 py-2.5"
  >
    {#if isConnected || justSaved.current}
      <span class="flex items-center gap-2 text-xs text-primary">
        <span class="h-1.5 w-1.5 rounded-full bg-primary"></span>
        {isConnected
          ? t('ssh.connected', 'Connected')
          : t('ssh.saved', 'Saved')}
      </span>
    {/if}

    <Button variant="ghost" size="sm" class="ml-auto" onclick={cancel}>
      {t('common.cancel', 'Cancel')}
    </Button>
    <Button
      variant="secondary"
      size="sm"
      {@attach tooltip({
        content: t('common.save', 'Save'),
        shortcut: `${modKey}S`,
      })}
      disabled={busy || keyBusy}
      onclick={save}
    >
      {t('common.save', 'Save')}
    </Button>
    <Button
      size="sm"
      shortcut={`${modKey}↵`}
      disabled={busy || keyBusy}
      onclick={connect}
    >
      {t('ssh.connect', 'Connect')}
    </Button>
  </footer>
</div>
