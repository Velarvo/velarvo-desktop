<script lang="ts">
  import EditorPaneNode from './EditorPaneNode.svelte'
  import { countGroups, editorState } from '@/lib/editorLayout'

  $: root = $editorState.root
  $: onlyOneGroup = countGroups(root) <= 1
</script>

<div class="flex h-full min-h-0 w-full flex-col bg-surface-canvas">
  <EditorPaneNode node={root} canClose={!onlyOneGroup} />
</div>
