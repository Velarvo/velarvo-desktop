<script lang="ts">
  import TopBar from './TopBar.svelte'
  import Sidebar from './Sidebar.svelte'

  let collapsed = false

  const toggleSidebar = () => {
    collapsed = !collapsed
  }
</script>

<div class="flex flex-col h-screen bg-background overflow-hidden">
  <TopBar onToggle={toggleSidebar} />

  <div class="flex flex-1 overflow-hidden">
    <Sidebar {collapsed} />

    <main
      class="relative flex min-w-0 flex-1 flex-col overflow-hidden bg-surface-canvas"
    >
      <slot />
    </main>
  </div>
</div>
