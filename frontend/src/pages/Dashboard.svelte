<script lang="ts">
  import AppLayout from '@/layout/AppLayout.svelte'
  import FileExplorer from '@/components/file-explorer/FileExplorer.svelte'
  import {
    fileSystemExplorerColumns,
    fileSystemExplorerDataSource,
    fileSystemExplorerLabels,
  } from '@/components/file-explorer/filesystem'

  const dashboardExplorer = {
    dataSource: fileSystemExplorerDataSource,
    columns: fileSystemExplorerColumns,
    labels: fileSystemExplorerLabels,
  }
</script>

<AppLayout>
  <div class="h-full w-full max-w-full p-3 sm:p-4">
    <div class="h-full w-full max-w-full mx-auto flex flex-col">
      <div class="flex-1 min-h-0 w-full max-w-full">
        <FileExplorer {...dashboardExplorer} />
      </div>
    </div>
  </div>
</AppLayout>
