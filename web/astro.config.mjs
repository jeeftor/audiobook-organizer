import { defineConfig } from 'astro/config'
import starlight from '@astrojs/starlight'

export default defineConfig({
  site: 'https://jeeftor.github.io',
  base: '/audiobook-organizer',
  outDir: '../output/docs-starlight',
  integrations: [
    starlight({
      title: 'Audiobook Organizer',
      description:
        'Audiobook Organizer docs for Audiobookshelf, metadata.json, embedded audiobook metadata, safe previews, renames, and undo workflows.',
      logo: {
        src: './src/assets/docs-logo.png',
        alt: 'Audiobook Organizer',
      },
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/jeeftor/audiobook-organizer',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/jeeftor/audiobook-organizer/edit/master/web/',
      },
      head: [
        {
          tag: 'script',
          content: `try {
  if (localStorage.getItem('starlight-theme') === null) {
    localStorage.setItem('starlight-theme', 'dark');
  }
} catch {}`,
        },
      ],
      disable404Route: true,
      customCss: ['./src/styles/starlight.css'],
      sidebar: [
        {
          label: 'Start',
          items: [
            { label: 'Overview', slug: 'index' },
            { label: 'Getting Started', slug: 'getting-started' },
            { label: 'Installation', slug: 'installation' },
            { label: 'Choose An Interface', slug: 'interfaces' },
          ],
        },
        {
          label: 'Interfaces',
          items: [
            { label: 'Local Web UI', slug: 'web-ui' },
            { label: 'CLI', slug: 'cli' },
            { label: 'TUI', slug: 'tui' },
          ],
        },
        {
          label: 'Workflows',
          items: [
            { label: 'Organize', slug: 'organize' },
            { label: 'Rename Files', slug: 'rename' },
            { label: 'Explore Metadata', slug: 'explore-metadata' },
            { label: 'Audiobookshelf', slug: 'audiobookshelf' },
            { label: 'Safety And Undo', slug: 'safety-and-undo' },
          ],
        },
        {
          label: 'Reference',
          items: [
            { label: 'Metadata Sources', slug: 'metadata' },
            { label: 'Metadata Command', slug: 'metadata-command' },
            { label: 'Layouts', slug: 'layouts' },
            { label: 'Configuration', slug: 'configuration' },
            { label: 'Changelog', slug: 'changelog' },
            { label: 'Troubleshooting', slug: 'troubleshooting' },
          ],
        },
        {
          label: 'Development',
          items: [
            { label: 'Docs Visuals', slug: 'development/docs-visuals' },
            { label: 'Web UI Testing', slug: 'development/web-ui-testing' },
          ],
        },
      ],
    }),
  ],
})
