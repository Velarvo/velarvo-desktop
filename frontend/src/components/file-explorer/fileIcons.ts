export type FileIconKey =
  | 'archive'
  | 'audio'
  | 'code'
  | 'config'
  | 'css'
  | 'database'
  | 'docker'
  | 'document'
  | 'file'
  | 'folder'
  | 'git'
  | 'github'
  | 'go'
  | 'graphql'
  | 'html'
  | 'image'
  | 'javascript'
  | 'json'
  | 'key'
  | 'lock'
  | 'markdown'
  | 'node'
  | 'npm'
  | 'package'
  | 'pdf'
  | 'php'
  | 'prisma'
  | 'python'
  | 'react'
  | 'rust'
  | 'sass'
  | 'spreadsheet'
  | 'svelte'
  | 'tailwind'
  | 'terminal'
  | 'text'
  | 'typescript'
  | 'video'
  | 'vite'
  | 'vue'
  | 'xml'
  | 'yaml'

export interface FileIconDescriptor {
  key: FileIconKey
  label: string
  color: string
  background: string
  border: string
}

const defaultIcon: FileIconDescriptor = {
  key: 'file',
  label: 'File',
  color: '#94a3b8',
  background: 'rgba(148, 163, 184, 0.08)',
  border: 'rgba(148, 163, 184, 0.16)',
}

const folderIcon: FileIconDescriptor = {
  key: 'folder',
  label: 'Folder',
  color: '#12ce90',
  background: 'rgba(18, 206, 144, 0.12)',
  border: 'rgba(18, 206, 144, 0.24)',
}

const icons = {
  archive: ['Archive', '#f59e0b'],
  audio: ['Audio', '#c084fc'],
  code: ['Code', '#93c5fd'],
  config: ['Config', '#a3a3a3'],
  css: ['CSS', '#38bdf8'],
  database: ['Database', '#fbbf24'],
  docker: ['Docker', '#60a5fa'],
  document: ['Document', '#60a5fa'],
  file: ['File', '#94a3b8'],
  folder: ['Folder', '#12ce90'],
  git: ['Git', '#f97316'],
  github: ['GitHub', '#e5e7eb'],
  go: ['Go', '#22d3ee'],
  graphql: ['GraphQL', '#ec4899'],
  html: ['HTML', '#fb7185'],
  image: ['Image', '#34d399'],
  javascript: ['JavaScript', '#facc15'],
  json: ['JSON', '#fbbf24'],
  key: ['Key', '#facc15'],
  lock: ['Lock file', '#a78bfa'],
  markdown: ['Markdown', '#93c5fd'],
  node: ['Node.js', '#86efac'],
  npm: ['npm', '#ef4444'],
  package: ['Package', '#f59e0b'],
  pdf: ['PDF', '#f87171'],
  php: ['PHP', '#818cf8'],
  prisma: ['Prisma', '#5eead4'],
  python: ['Python', '#fbbf24'],
  react: ['React', '#67e8f9'],
  rust: ['Rust', '#fb923c'],
  sass: ['Sass', '#f472b6'],
  spreadsheet: ['Spreadsheet', '#4ade80'],
  svelte: ['Svelte', '#fb923c'],
  tailwind: ['Tailwind CSS', '#67e8f9'],
  terminal: ['Shell script', '#a3e635'],
  text: ['Text', '#cbd5e1'],
  typescript: ['TypeScript', '#60a5fa'],
  video: ['Video', '#fb7185'],
  vite: ['Vite', '#c084fc'],
  vue: ['Vue', '#34d399'],
  xml: ['XML', '#fb923c'],
  yaml: ['YAML', '#fbbf24'],
} satisfies Record<FileIconKey, [string, string]>

const makeIcon = (key: FileIconKey): FileIconDescriptor => {
  if (key === 'folder') return folderIcon
  if (key === 'file') return defaultIcon
  const icon = icons[key as keyof typeof icons]
  if (!icon) return defaultIcon
  const [label, color] = icon

  return {
    key,
    label,
    color,
    background: `${color}1f`,
    border: `${color}33`,
  }
}

const descriptorByKey = Object.fromEntries(
  (Object.keys(icons) as FileIconKey[]).map((key) => [key, makeIcon(key)]),
) as Record<FileIconKey, FileIconDescriptor>

const exactNameIcons = {
  '.env': 'key',
  '.env.development': 'key',
  '.env.local': 'key',
  '.env.production': 'key',
  '.gitattributes': 'git',
  '.gitignore': 'git',
  '.npmrc': 'npm',
  '.prettierrc': 'config',
  'bun.lock': 'lock',
  'cargo.lock': 'lock',
  'cargo.toml': 'rust',
  'compose.yaml': 'docker',
  'compose.yml': 'docker',
  'docker-compose.yaml': 'docker',
  'docker-compose.yml': 'docker',
  dockerfile: 'docker',
  'eslint.config.js': 'config',
  'eslint.config.mjs': 'config',
  'eslint.config.ts': 'config',
  gemfile: 'package',
  'go.mod': 'go',
  'go.sum': 'go',
  'jsconfig.json': 'config',
  license: 'document',
  makefile: 'terminal',
  'package-lock.json': 'lock',
  'package.json': 'npm',
  'pnpm-lock.yaml': 'lock',
  'postcss.config.js': 'config',
  'postcss.config.mjs': 'config',
  'pyproject.toml': 'python',
  readme: 'markdown',
  'readme.md': 'markdown',
  'requirements.txt': 'python',
  'schema.prisma': 'prisma',
  'svelte.config.js': 'svelte',
  'svelte.config.ts': 'svelte',
  'tailwind.config.js': 'tailwind',
  'tailwind.config.ts': 'tailwind',
  'tsconfig.json': 'config',
  'vite.config.js': 'vite',
  'vite.config.mjs': 'vite',
  'vite.config.ts': 'vite',
  'yarn.lock': 'lock',
} satisfies Record<string, FileIconKey>

const extensionIcons = {
  '7z': 'archive',
  avi: 'video',
  bash: 'terminal',
  bmp: 'image',
  bz2: 'archive',
  c: 'code',
  cjs: 'javascript',
  cpp: 'code',
  cs: 'code',
  css: 'css',
  csv: 'spreadsheet',
  doc: 'document',
  docx: 'document',
  fish: 'terminal',
  gif: 'image',
  go: 'go',
  gz: 'archive',
  h: 'code',
  hpp: 'code',
  htm: 'html',
  html: 'html',
  ico: 'image',
  java: 'code',
  jpeg: 'image',
  jpg: 'image',
  js: 'javascript',
  json: 'json',
  jsx: 'react',
  key: 'key',
  kt: 'code',
  less: 'css',
  lock: 'lock',
  log: 'text',
  md: 'markdown',
  mdx: 'react',
  mjs: 'javascript',
  mov: 'video',
  mp3: 'audio',
  mp4: 'video',
  ogg: 'audio',
  pdf: 'pdf',
  php: 'php',
  png: 'image',
  ppt: 'document',
  pptx: 'document',
  ps1: 'terminal',
  py: 'python',
  rar: 'archive',
  rb: 'code',
  rs: 'rust',
  sass: 'sass',
  scss: 'sass',
  sh: 'terminal',
  sql: 'database',
  sqlite: 'database',
  svg: 'xml',
  svelte: 'svelte',
  swift: 'code',
  tar: 'archive',
  toml: 'config',
  ts: 'typescript',
  tsx: 'react',
  txt: 'text',
  vue: 'vue',
  wav: 'audio',
  webm: 'video',
  webp: 'image',
  xls: 'spreadsheet',
  xlsx: 'spreadsheet',
  xml: 'xml',
  yaml: 'yaml',
  yml: 'yaml',
  zip: 'archive',
  zsh: 'terminal',
} satisfies Record<string, FileIconKey>

export const getFileExtension = (filename: string) => {
  const normalized = filename.trim().toLowerCase()
  const dotIndex = normalized.lastIndexOf('.')

  if (dotIndex <= 0 || dotIndex === normalized.length - 1) {
    return ''
  }

  return normalized.slice(dotIndex + 1)
}

export const getFileIconDescriptor = (
  filename: string,
  isContainer = false,
): FileIconDescriptor => {
  if (isContainer) {
    return folderIcon
  }

  const normalized = filename.trim().toLowerCase()
  const exactIcon = exactNameIcons[normalized as keyof typeof exactNameIcons]
  if (exactIcon) {
    return descriptorByKey[exactIcon]
  }

  if (normalized.startsWith('.env.')) {
    return descriptorByKey.key
  }

  if (normalized.startsWith('dockerfile.')) {
    return descriptorByKey.docker
  }

  const extension = getFileExtension(normalized)
  const extensionIcon = extensionIcons[extension as keyof typeof extensionIcons]
  if (extensionIcon) {
    return descriptorByKey[extensionIcon]
  }

  return defaultIcon
}
