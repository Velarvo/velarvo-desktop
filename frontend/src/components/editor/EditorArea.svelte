<script lang="ts">
  import EditorPaneNode from './EditorPaneNode.svelte'
  import { countGroups, editor } from '@/lib/editorLayout.svelte'

  let root = $derived(editor.layout.root)
  let onlyOneGroup = $derived(countGroups(root) <= 1)
</script>

<div class="flex h-full min-h-0 w-full flex-col bg-surface-canvas">
  <EditorPaneNode node={root} canClose={!onlyOneGroup} />
</div>
