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
    <div class="w-full max-w-full mx-auto flex flex-col gap-4">
      <div class="flex flex-col lg:flex-row w-full max-w-full gap-3 lg:gap-4">
        <FileExplorer
          {...dashboardExplorer}
          className="w-full lg:w-1/2"
          viewportClassName="overflow-auto"
          viewportMaxHeight="calc(100vh - 14rem)"
        />
        <FileExplorer
          {...dashboardExplorer}
          className="w-full lg:w-1/2"
          viewportClassName="overflow-auto"
          viewportMaxHeight="calc(100vh - 14rem)"
        />
      </div>
    </div>
  </div>
</AppLayout>
