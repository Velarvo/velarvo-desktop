import { fireEvent, render, screen, waitFor } from '@testing-library/svelte'
import { describe, expect, it, vi } from 'vitest'
import FileExplorer from '@/components/file-explorer/FileExplorer.svelte'
import type {
  ExplorerDataSource,
  ExplorerNode,
} from '@/components/file-explorer/types'

const rootNodes: ExplorerNode[] = [
  {
    id: '/fixture-root/Documents',
    name: 'Documents',
    path: '/fixture-root/Documents',
    type: 'directory',
    isContainer: true,
  },
  {
    id: '/fixture-root/archive.zip',
    name: 'archive.zip',
    path: '/fixture-root/archive.zip',
    type: 'file',
    size: 1024,
  },
]

const documentsNodes: ExplorerNode[] = [
  {
    id: '/fixture-root/Documents/notes.txt',
    name: 'notes.txt',
    path: '/fixture-root/Documents/notes.txt',
    type: 'file',
    size: 512,
  },
]

const createDataSource = () => {
  const loadNodes = vi.fn(async (path: string) => {
    if (path === '/fixture-root/Documents') {
      return documentsNodes
    }

    return rootNodes
  })

  const dataSource: ExplorerDataSource = {
    getInitialPath: vi.fn(async () => '/fixture-root'),
    loadNodes,
  }

  return dataSource
}

describe('FileExplorer', () => {
  it('loads and renders the initial directory', async () => {
    const dataSource = createDataSource()

    render(FileExplorer, {
      props: {
        dataSource,
      },
    })

    expect(await screen.findByText('Documents')).toBeTruthy()
    expect(screen.getByText('archive.zip')).toBeTruthy()
    expect(dataSource.loadNodes).toHaveBeenCalledWith('/fixture-root')
  })

  it('navigates into a container node', async () => {
    const dataSource = createDataSource()

    render(FileExplorer, {
      props: {
        dataSource,
      },
    })

    const documentsButton = await screen.findByRole('button', {
      name: 'Documents',
    })
    await fireEvent.click(documentsButton)

    expect(await screen.findByText('notes.txt')).toBeTruthy()
    expect(dataSource.loadNodes).toHaveBeenCalledWith('/fixture-root/Documents')
  })

  it('filters visible nodes by search query', async () => {
    const dataSource = createDataSource()

    render(FileExplorer, {
      props: {
        dataSource,
      },
    })

    await screen.findByText('Documents')

    const searchInput = screen.getByPlaceholderText('Search...')
    await fireEvent.input(searchInput, { target: { value: 'archive' } })

    await waitFor(() => {
      expect(screen.queryByText('Documents')).toBeNull()
      expect(screen.getByText('archive.zip')).toBeTruthy()
    })
  })
})
