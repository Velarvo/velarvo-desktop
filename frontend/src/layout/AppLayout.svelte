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

    <main class="relative flex-1 overflow-auto bg-[#07070b]">
      <div
        class="pointer-events-none absolute top-0 inset-x-0 h-48 bg-primary/[0.03] blur-3xl"
      ></div>
      <div class="relative z-10 p-5">
        <slot />
      </div>
    </main>
  </div>
</div>
