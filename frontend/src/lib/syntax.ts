import type { HLJSApi, LanguageFn } from 'highlight.js'

export interface HighlightedCode {
  html: string
  language: string
}

const MAX_HIGHLIGHT_LENGTH = 200_000

const MIN_AUTO_RELEVANCE = 5

const BY_EXTENSION: Record<string, string> = {
  bash: 'bash',
  c: 'c',
  cc: 'cpp',
  cfg: 'ini',
  cjs: 'javascript',
  conf: 'ini',
  cpp: 'cpp',
  cs: 'csharp',
  css: 'css',
  cts: 'typescript',
  cxx: 'cpp',
  diff: 'diff',
  editorconfig: 'ini',
  env: 'ini',
  gemfile: 'ruby',
  go: 'go',
  gql: 'graphql',
  graphql: 'graphql',
  h: 'c',
  hh: 'cpp',
  hpp: 'cpp',
  htaccess: 'apache',
  htm: 'xml',
  html: 'xml',
  hxx: 'cpp',
  ini: 'ini',
  java: 'java',
  javascript: 'javascript',
  js: 'javascript',
  json: 'json',
  json5: 'json',
  jsonc: 'json',
  jsx: 'javascript',
  kt: 'kotlin',
  kts: 'kotlin',
  ksh: 'bash',
  less: 'less',
  log: 'plaintext',
  lua: 'lua',
  m: 'objectivec',
  markdown: 'markdown',
  md: 'markdown',
  mdx: 'markdown',
  mjs: 'javascript',
  mk: 'makefile',
  mm: 'objectivec',
  mts: 'typescript',
  patch: 'diff',
  php: 'php',
  phtml: 'php',
  pl: 'perl',
  plist: 'xml',
  pm: 'perl',
  properties: 'ini',
  ps1: 'powershell',
  psm1: 'powershell',
  py: 'python',
  pyi: 'python',
  pyw: 'python',
  r: 'r',
  rake: 'ruby',
  rb: 'ruby',
  rs: 'rust',
  sass: 'scss',
  scss: 'scss',
  sh: 'bash',
  sql: 'sql',
  svelte: 'xml',
  svg: 'xml',
  swift: 'swift',
  text: 'plaintext',
  toml: 'ini',
  ts: 'typescript',
  tsx: 'typescript',
  txt: 'plaintext',
  vb: 'vbnet',
  vue: 'xml',
  xml: 'xml',
  xsl: 'xml',
  yaml: 'yaml',
  yml: 'yaml',
  zsh: 'bash',
}

const BY_FILENAME: Record<string, string> = {
  '.bash_aliases': 'bash',
  '.bash_profile': 'bash',
  '.bashrc': 'bash',
  '.dockerignore': 'ini',
  '.env': 'ini',
  '.gitconfig': 'ini',
  '.gitignore': 'ini',
  '.npmrc': 'ini',
  '.profile': 'bash',
  '.zprofile': 'bash',
  '.zshrc': 'bash',
  containerfile: 'dockerfile',
  dockerfile: 'dockerfile',
  gemfile: 'ruby',
  gnumakefile: 'makefile',
  known_hosts: 'plaintext',
  makefile: 'makefile',
  'nginx.conf': 'nginx',
  rakefile: 'ruby',
}

const EXTRA_GRAMMARS: Record<string, () => Promise<{ default: LanguageFn }>> = {
  apache: () => import('highlight.js/lib/languages/apache'),
  dockerfile: () => import('highlight.js/lib/languages/dockerfile'),
  nginx: () => import('highlight.js/lib/languages/nginx'),
  powershell: () => import('highlight.js/lib/languages/powershell'),
}

const languageForFile = (name: string): string | null => {
  const base = name.trim().toLowerCase()
  if (!base) return null

  const byName = BY_FILENAME[base]
  if (byName) return byName

  const segments = base.split('.')
  for (let index = 1; index < segments.length; index += 1) {
    const candidate = BY_EXTENSION[segments[index]]
    if (candidate) return candidate
  }

  if (base.startsWith('.') && segments.length === 2) {
    return BY_EXTENSION[segments[1]] ?? null
  }
  return null
}

export const lineNumberGutter = (code: string): string => {
  const lines = code.split('\n')
  if (lines.length > 1 && lines[lines.length - 1] === '') lines.pop()
  return Array.from({ length: lines.length }, (_, index) => index + 1).join(
    '\n',
  )
}

export const escapeHtml = (value: string): string =>
  value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')

let enginePromise: Promise<HLJSApi> | null = null

const engine = (): Promise<HLJSApi> => {
  enginePromise ??= import('highlight.js/lib/common').then(
    (module) => module.default,
  )
  return enginePromise
}

const ensureLanguage = async (hljs: HLJSApi, id: string): Promise<boolean> => {
  if (hljs.getLanguage(id)) return true

  const load = EXTRA_GRAMMARS[id]
  if (!load) return false

  try {
    const grammar = await load()
    hljs.registerLanguage(id, grammar.default)
    return true
  } catch {
    return false
  }
}

const plain = (code: string): HighlightedCode => ({
  html: escapeHtml(code),
  language: '',
})

export const highlightCode = async (
  name: string,
  code: string,
): Promise<HighlightedCode> => {
  if (!code || code.length > MAX_HIGHLIGHT_LENGTH) return plain(code)

  let hljs: HLJSApi
  try {
    hljs = await engine()
  } catch {
    return plain(code)
  }

  const id = languageForFile(name)
  if (id === 'plaintext') return plain(code)

  try {
    if (id && (await ensureLanguage(hljs, id))) {
      const result = hljs.highlight(code, {
        language: id,
        ignoreIllegals: true,
      })
      return {
        html: result.value,
        language: hljs.getLanguage(id)?.name ?? id,
      }
    }

    const guess = hljs.highlightAuto(code)
    if (!guess.language || guess.relevance < MIN_AUTO_RELEVANCE) {
      return plain(code)
    }
    return {
      html: guess.value,
      language: hljs.getLanguage(guess.language)?.name ?? guess.language,
    }
  } catch {
    return plain(code)
  }
}
