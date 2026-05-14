export type BookRow = {
  id: number
  title: string
  author: string
  series: string
  source: string
  currentPath: string
  proposedPath: string
  status: 'ready' | 'warning'
  statusText: string
}

export type JobEvent = {
  time: string
  level: 'ok' | 'warn' | 'info'
  event: string
  detail: string
}

export const rows: BookRow[] = [
  {
    id: 1,
    title: 'Project Hail Mary',
    author: 'Andy Weir',
    series: '-',
    source: 'm4b',
    currentPath: 'Project Hail Mary / Andy Weir',
    proposedPath: 'Andy Weir/Project Hail Mary (2021)/Project Hail Mary.m4b',
    status: 'ready',
    statusText: 'Ready',
  },
  {
    id: 2,
    title: 'The Will of the Many',
    author: 'James Islington',
    series: 'Licanius Trilogy Book 1',
    source: 'm4b',
    currentPath: 'The Will of the Many / James Islington',
    proposedPath: 'James Islington/The Will of the Many (2023)/The Will of the Many.m4b',
    status: 'ready',
    statusText: 'Ready',
  },
  {
    id: 3,
    title: 'Dune',
    author: 'Frank Herbert',
    series: 'Dune Book 1',
    source: 'm4b',
    currentPath: 'Dune / Frank Herbert',
    proposedPath: 'Frank Herbert/Dune (1965)/Dune.m4b',
    status: 'ready',
    statusText: 'Ready',
  },
  {
    id: 4,
    title: 'The Name of the Wind',
    author: 'Patrick Rothfuss',
    series: 'The Kingkiller Chronicle Book 1',
    source: 'm4b',
    currentPath: 'The Name of the Wind / Patrick Rothfuss',
    proposedPath: 'Patrick Rothfuss/The Kingkiller Chronicle Book 1/The Name of the Wind.m4b',
    status: 'ready',
    statusText: 'Ready',
  },
  {
    id: 5,
    title: '1984',
    author: 'George Orwell',
    series: '-',
    source: 'mp3',
    currentPath: '1984 / George Orwell',
    proposedPath: 'George Orwell/1984 (1949)/1984.mp3',
    status: 'ready',
    statusText: 'Ready',
  },
  {
    id: 6,
    title: 'The Hobbit',
    author: 'J.R.R. Tolkien',
    series: 'The Hobbit',
    source: 'abs',
    currentPath: 'The Hobbit / JRR Tolkien',
    proposedPath: 'J.R.R. Tolkien/The Hobbit (1937)/The Hobbit.m4b',
    status: 'warning',
    statusText: 'Missing metadata',
  },
]

export const jobEvents: JobEvent[] = [
  { time: '10:24:31', level: 'ok', event: 'Scan completed', detail: 'Found 20 audiobooks in 1.42s' },
  { time: '10:24:31', level: 'ok', event: 'Metadata lookup completed', detail: 'Matched 18 of 20 items' },
  { time: '10:24:28', level: 'warn', event: 'Metadata not found', detail: 'The Hobbit by J.R.R. Tolkien' },
  { time: '10:24:27', level: 'warn', event: 'Multiple matches found', detail: '1984 by George Orwell, 3 matches' },
  { time: '10:24:25', level: 'info', event: 'Connecting to Audiobookshelf', detail: 'http://localhost:13378' },
  { time: '10:24:24', level: 'ok', event: 'Path mapping validated', detail: 'All 3 test paths are accessible' },
]
