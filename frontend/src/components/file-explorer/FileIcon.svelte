<script lang="ts">
  import {
    IconArchive,
    IconBrandCss3,
    IconBrandDocker,
    IconBrandGit,
    IconBrandGithub,
    IconBrandGolang,
    IconBrandGraphql,
    IconBrandHtml5,
    IconBrandJavascript,
    IconBrandNodejs,
    IconBrandNpm,
    IconBrandPhp,
    IconBrandPrisma,
    IconBrandPython,
    IconBrandReact,
    IconBrandRust,
    IconBrandSass,
    IconBrandSvelte,
    IconBrandTailwind,
    IconBrandTypescript,
    IconBrandVite,
    IconBrandVue,
    IconDatabase,
    IconFile,
    IconFileCode,
    IconFileDescription,
    IconFileMusic,
    IconFileSettings,
    IconFileSpreadsheet,
    IconFileText,
    IconFolder,
    IconBraces,
    IconKey,
    IconLock,
    IconMarkdown,
    IconMovie,
    IconPackage,
    IconPhoto,
    IconTerminal,
    IconFileTypePdf,
    IconFileTypeXml,
  } from '@tabler/icons-svelte'
  import { getFileIconDescriptor, type FileIconKey } from './fileIcons'

  export let name = ''
  export let isContainer = false
  export let className = ''

  const iconComponents = {
    archive: IconArchive,
    audio: IconFileMusic,
    code: IconFileCode,
    config: IconFileSettings,
    css: IconBrandCss3,
    database: IconDatabase,
    docker: IconBrandDocker,
    document: IconFileDescription,
    file: IconFile,
    folder: IconFolder,
    git: IconBrandGit,
    github: IconBrandGithub,
    go: IconBrandGolang,
    graphql: IconBrandGraphql,
    html: IconBrandHtml5,
    image: IconPhoto,
    javascript: IconBrandJavascript,
    json: IconBraces,
    key: IconKey,
    lock: IconLock,
    markdown: IconMarkdown,
    node: IconBrandNodejs,
    npm: IconBrandNpm,
    package: IconPackage,
    pdf: IconFileTypePdf,
    php: IconBrandPhp,
    prisma: IconBrandPrisma,
    python: IconBrandPython,
    react: IconBrandReact,
    rust: IconBrandRust,
    sass: IconBrandSass,
    spreadsheet: IconFileSpreadsheet,
    svelte: IconBrandSvelte,
    tailwind: IconBrandTailwind,
    terminal: IconTerminal,
    text: IconFileText,
    typescript: IconBrandTypescript,
    video: IconMovie,
    vite: IconBrandVite,
    vue: IconBrandVue,
    xml: IconFileTypeXml,
    yaml: IconFileSettings,
  } satisfies Record<FileIconKey, typeof IconFile>

  $: descriptor = getFileIconDescriptor(name, isContainer)
  $: Icon = iconComponents[descriptor.key]
</script>

<span
  class="inline-flex h-5 w-5 shrink-0 items-center justify-center {className}"
  style={`color: ${descriptor.color}; background: none; border: none;`}
  aria-hidden="true"
  data-file-icon={descriptor.key}
>
  <svelte:component this={Icon} class="h-5 w-5" stroke={1.8} />
</span>
