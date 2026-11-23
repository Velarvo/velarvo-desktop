<script lang="ts">
  import { fade, scale } from 'svelte/transition'
  import {
    Circle,
    CircleCheck,
    CircleX,
    Eye,
    LoaderCircle,
    RotateCw,
    Server,
  } from '@lucide/svelte'
  import Button from '@/components/ui/Button.svelte'
  import OsIcon from '@/components/ui/OsIcon.svelte'
  import { t } from '@/lib/i18n.svelte'
  import { prefersReducedMotion } from 'svelte/motion'
  import {
    emptyConnectProgress,
    SSHConnectPhase,
    SSH_CONNECT_PHASES,
    SSHTerminalStatus,
    type SSHConnection,
    type SSHConnectProgress,
  } from '@/types/ssh'

  interface Props {
    status: SSHTerminalStatus
    errorMessage?: string
    connection?: SSHConnection | null
    progress?: SSHConnectProgress
    phaseLabels?: Partial<Record<SSHConnectPhase, [string, string]>>
    dismissible?: boolean
    onReconnect?: () => void
  }

  let {
    status,
    errorMessage = '',
    connection = null,
    progress = emptyConnectProgress(),
    phaseLabels = {},
    dismissible = true,
    onReconnect,
  }: Props = $props()

  let dismissed = $state(false)

  let isConnecting = $derived(status === SSHTerminalStatus.Connecting)
  let isError = $derived(status === SSHTerminalStatus.Error)
  let visible = $derived(
    status !== SSHTerminalStatus.Open && !(dismissible && dismissed),
  )

  $effect.pre(() => {
    if (status) {
      dismissed = false
    }
  })

  let elapsed = $state(0)

  $effect(() => {
    if (!isConnecting) return
    elapsed = 0
    const timer = setInterval(() => {
      elapsed += 100
    }, 100)
    return () => clearInterval(timer)
  })

  const DEFAULT_PHASE_LABEL: Record<SSHConnectPhase, [string, string]> = {
    [SSHConnectPhase.Channel]: ['ssh.phaseChannel', 'Secure channel'],
    [SSHConnectPhase.Session]: ['ssh.phaseSession', 'Shell session'],
  }

  type PhaseState = 'done' | 'active' | 'failed' | 'pending'

  const phaseState = (
    phase: SSHConnectPhase,
    value: SSHConnectProgress,
    connecting: boolean,
  ): PhaseState => {
    if (value.failedAt === phase) return 'failed'
    if (value.timings[phase] !== undefined) return 'done'
    if (value.current === phase) return 'active'

    const untouched = value.current === null && !value.timings[phase]
    if (connecting && untouched && phase === SSH_CONNECT_PHASES[0]) {
      return 'active'
    }
    return 'pending'
  }

  const formatMs = (ms: number): string =>
    ms < 1000 ? `${Math.round(ms)} ms` : `${(ms / 1000).toFixed(1)} s`

  let steps = $derived(
    SSH_CONNECT_PHASES.map((phase) => ({
      phase,
      state: phaseState(phase, progress, isConnecting),
      timing: progress.timings[phase],
      label: phaseLabels[phase] ?? DEFAULT_PHASE_LABEL[phase],
    })),
  )

  let doneCount = $derived(steps.filter((step) => step.state === 'done').length)
  let reachedCount = $derived(doneCount + (progress.failedAt !== null ? 1 : 0))

  let fillPercent = $derived(
    Math.round((reachedCount / SSH_CONNECT_PHASES.length) * 100),
  )
  let hasActiveStep = $derived(steps.some((step) => step.state === 'active'))

  let tone = $derived(isError ? 'error' : isConnecting ? 'live' : 'idle')

  let title = $derived(
    connection?.name || connection?.host || t('ssh.terminal', 'Terminal'),
  )

  let target = $derived(
    connection
      ? `${connection.username}@${connection.host}:${connection.port}`
      : '',
  )

  let headline = $derived(
    isError
      ? t('ssh.connectionError', 'Connection error')
      : isConnecting
        ? t('ssh.connecting', 'Connecting…')
        : t('ssh.closed', 'Closed'),
  )

  let hint = $derived(
    !isConnecting && !isError
      ? t(
          'ssh.closedHint',
          'The remote shell ended. Reconnect to start a new session.',
        )
      : '',
  )
</script>

{#if visible}
  <div
    class="absolute inset-0 z-10 grid place-items-center p-6 backdrop-blur-[3px]"
    style="background: color-mix(in srgb, var(--color-surface-canvas) 82%, transparent)"
    role="status"
    aria-live="polite"
    aria-busy={isConnecting}
    transition:fade={{ duration: prefersReducedMotion.current ? 0 : 140 }}
  >
    <div
      class="card card-{tone} w-full max-w-[27rem] overflow-hidden rounded-lg border border-border bg-surface-raised shadow-xl shadow-black/50"
      transition:scale={{
        duration: prefersReducedMotion.current ? 0 : 160,
        start: 0.97,
        opacity: 0,
      }}
    >
      <div class="relative h-1 w-full bg-white/8">
        <div
          class="h-full rounded-r-full transition-[width] duration-500 ease-out {isError
            ? 'bg-red-400'
            : 'bg-primary'}"
          style="width: {fillPercent}%"
        ></div>
        {#if hasActiveStep && !prefersReducedMotion.current}
          <span class="shimmer" aria-hidden="true"></span>
        {/if}
      </div>

      <div class="p-6">
        <div class="flex items-center gap-3.5">
          <span class="tile tile-{tone}" aria-hidden="true">
            {#if connection?.os}
              <OsIcon os={connection.os} size={24} />
            {:else}
              <Server size={23} strokeWidth={1.8} />
            {/if}
          </span>

          <span class="min-w-0 flex-1">
            <span class="block truncate text-base font-semibold text-white">
              {title}
            </span>
            {#if target}
              <span
                class="mt-1 block truncate font-mono text-xs text-muted-foreground"
              >
                {target}
              </span>
            {/if}
          </span>
        </div>

        <div class="mt-6 flex items-baseline justify-between gap-3">
          <span
            class="flex items-center gap-2.5 text-[13px] font-semibold {isError
              ? 'text-red-400'
              : isConnecting
                ? 'text-primary'
                : 'text-muted-foreground'}"
          >
            {#if isConnecting}
              <span class="pip" aria-hidden="true"></span>
            {/if}
            {headline}
          </span>
          {#if isConnecting}
            <span class="font-mono text-xs tabular-nums text-muted-foreground">
              {formatMs(elapsed)}
            </span>
          {/if}
        </div>

        <ul class="mt-4 space-y-0.5">
          {#each steps as step (step.phase)}
            <li
              class="flex items-center gap-3 rounded-lg px-2.5 py-2 {step.state ===
              'active'
                ? 'bg-white/[0.04]'
                : ''}"
            >
              <span class="marker" aria-hidden="true">
                {#if step.state === 'done'}
                  <CircleCheck size={17} strokeWidth={2} class="text-primary" />
                {:else if step.state === 'failed'}
                  <CircleX size={17} strokeWidth={2} class="text-red-400" />
                {:else if step.state === 'active' && !prefersReducedMotion.current}
                  <LoaderCircle
                    size={17}
                    strokeWidth={2.2}
                    class="animate-spin text-primary"
                  />
                {:else if step.state === 'active'}
                  <Circle size={17} strokeWidth={2.4} class="text-primary" />
                {:else}
                  <Circle size={17} strokeWidth={1.75} class="text-white/15" />
                {/if}
              </span>
              <span
                class="min-w-0 flex-1 truncate text-[13px] {step.state ===
                'pending'
                  ? 'text-muted-foreground/60'
                  : step.state === 'failed'
                    ? 'text-red-300'
                    : 'text-white/85'}"
              >
                {t(...step.label)}
              </span>
              {#if step.timing !== undefined}
                <span
                  class="shrink-0 font-mono text-[11px] tabular-nums text-muted-foreground"
                >
                  {formatMs(step.timing)}
                </span>
              {/if}
            </li>
          {/each}
        </ul>

        {#if isError && errorMessage}
          <p
            class="mt-4 max-h-28 overflow-y-auto rounded-lg border border-red-400/25 bg-red-500/10 px-3 py-2.5 font-mono text-[11px] leading-relaxed break-words select-text text-red-200"
          >
            {errorMessage}
          </p>
        {:else if hint}
          <p class="mt-4 text-[12.5px] leading-relaxed text-muted-foreground">
            {hint}
          </p>
        {/if}

        {#if !isConnecting}
          <div class="mt-6 flex items-center gap-2.5">
            <Button size="md" fullWidth onclick={() => onReconnect?.()}>
              <RotateCw size={15} strokeWidth={2.2} />
              {t('ssh.reconnect', 'Reconnect')}
            </Button>
            {#if dismissible}
              <Button
                variant="secondary"
                size="md"
                onclick={() => (dismissed = true)}
              >
                <Eye size={15} strokeWidth={2} />
                {t('ssh.viewOutput', 'Output')}
              </Button>
            {/if}
          </div>
        {/if}
      </div>
    </div>
  </div>
{:else if dismissible && status !== SSHTerminalStatus.Open}
  <button
    type="button"
    class="absolute right-3 bottom-3 z-10 inline-flex items-center gap-2 rounded-full border border-border bg-surface-raised/95 py-2 pr-4 pl-3.5 text-xs font-medium text-muted-foreground shadow-lg shadow-black/50 backdrop-blur-sm transition hover:border-white/20 hover:text-white"
    transition:fade={{ duration: prefersReducedMotion.current ? 0 : 140 }}
    onclick={() => (dismissed = false)}
  >
    <span
      class="h-1.5 w-1.5 rounded-full {isError
        ? 'bg-red-400'
        : 'bg-muted-foreground'}"
    ></span>
    {headline}
  </button>
{/if}

<style>
  .card {
    background-image: linear-gradient(
      to bottom,
      rgb(var(--color-primary-rgb) / 0.05),
      transparent 40%
    );
  }

  .card-error {
    background-image: linear-gradient(
      to bottom,
      rgb(248 113 113 / 0.07),
      transparent 40%
    );
  }

  .tile {
    display: grid;
    flex: none;
    place-items: center;
    width: 3rem;
    height: 3rem;
    border: 1px solid var(--color-border);
    border-radius: 0.875rem;
    background: rgb(255 255 255 / 0.04);
    color: var(--color-muted-foreground);
  }

  .tile-live {
    border-color: rgb(var(--color-primary-rgb) / 0.35);
    background: rgb(var(--color-primary-rgb) / 0.12);
    color: var(--color-primary);
  }

  .tile-error {
    border-color: rgb(248 113 113 / 0.4);
    background: rgb(239 68 68 / 0.12);
    color: rgb(252 165 165);
  }

  .pip {
    width: 0.45rem;
    height: 0.45rem;
    border-radius: 9999px;
    background: var(--color-primary);
    box-shadow: 0 0 0 0 rgb(var(--color-primary-rgb) / 0.5);
    animation: pulse 1.6s ease-out infinite;
  }

  .shimmer {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      90deg,
      transparent,
      rgb(var(--color-primary-rgb) / 0.9),
      transparent
    );
    animation: sweep 1.5s cubic-bezier(0.45, 0, 0.25, 1) infinite;
  }

  .marker {
    display: grid;
    flex: none;
    place-items: center;
    width: 1.15rem;
    height: 1.15rem;
  }

  @keyframes sweep {
    from {
      transform: translateX(-100%);
    }
    to {
      transform: translateX(100%);
    }
  }

  @keyframes pulse {
    70% {
      box-shadow: 0 0 0 0.35rem rgb(var(--color-primary-rgb) / 0);
    }
    100% {
      box-shadow: 0 0 0 0 rgb(var(--color-primary-rgb) / 0);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .pip,
    .shimmer {
      animation: none;
    }
  }
</style>
