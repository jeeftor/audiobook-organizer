export const siteBaseURL = 'https://jeeftor.github.io/audiobook-organizer/'

export const pages = [
  {
    source: 'docs/index.md',
    output: 'index.html',
    title: 'Audiobook Organizer for Audiobookshelf',
    group: 'Start',
  },
  {
    source: 'docs/INSTALLATION.md',
    output: 'installation.html',
    title: 'Installation',
    group: 'Start',
  },
  {
    source: 'docs/getting-started.md',
    output: 'getting-started.html',
    title: 'Getting Started',
    group: 'Start',
  },
  {
    source: 'docs/interfaces.md',
    output: 'interfaces.html',
    title: 'Choose An Interface',
    group: 'Start',
  },
  {
    source: 'docs/GUI.md',
    output: 'web-ui.html',
    title: 'Local Web UI',
    group: 'Interfaces',
  },
  {
    source: 'docs/CLI.md',
    output: 'cli.html',
    title: 'CLI',
    group: 'Interfaces',
  },
  {
    source: 'docs/TUI.md',
    output: 'tui.html',
    title: 'TUI',
    group: 'Interfaces',
  },
  {
    source: 'docs/organize.md',
    output: 'organize.html',
    title: 'Organize',
    group: 'Workflows',
  },
  {
    source: 'docs/RENAME_FEATURE.md',
    output: 'rename.html',
    title: 'Rename Files',
    group: 'Workflows',
  },
  {
    source: 'docs/explore-metadata.md',
    output: 'explore-metadata.html',
    title: 'Explore Metadata',
    group: 'Workflows',
  },
  {
    source: 'docs/audiobookshelf.md',
    output: 'audiobookshelf.html',
    title: 'Audiobookshelf',
    group: 'Workflows',
  },
  {
    source: 'docs/safety-and-undo.md',
    output: 'safety-and-undo.html',
    title: 'Safety And Undo',
    group: 'Workflows',
  },
  {
    source: 'docs/METADATA.md',
    output: 'metadata.html',
    title: 'Metadata Sources',
    group: 'Reference',
  },
  {
    source: 'docs/METADATA_COMMAND.md',
    output: 'metadata-command.html',
    title: 'Metadata Command',
    group: 'Reference',
  },
  {
    source: 'docs/LAYOUTS.md',
    output: 'layouts.html',
    title: 'Layouts',
    group: 'Reference',
  },
  {
    source: 'docs/CONFIGURATION.md',
    output: 'configuration.html',
    title: 'Configuration',
    group: 'Reference',
  },
  {
    source: 'CHANGELOG.md',
    output: 'changelog.html',
    title: 'Changelog',
    group: 'Reference',
  },
  {
    source: 'docs/troubleshooting.md',
    output: 'troubleshooting.html',
    title: 'Troubleshooting',
    group: 'Reference',
  },
  {
    source: 'docs/development/docs-visuals.md',
    output: 'development/docs-visuals.html',
    title: 'Docs Visuals',
    group: 'Development',
  },
  {
    source: 'docs/GUI_TESTING.md',
    output: 'development/web-ui-testing.html',
    title: 'Web UI Testing',
    group: 'Development',
  },
]

export const requiredGeneratedAssets = [
  'web-ui/web-ui-metadata-json-preview.png',
  'web-ui/web-ui-metadata-json-review.png',
  'cli/cli-help.png',
  'cli/cli-organize-run.gif',
  'cli/cli-rename-preview.gif',
  'tui/tui-organize-preview.gif',
  'tui/tui-organize-preview.png',
]

export const requiredHomepageGeneratedAssets = [
  'assets/generated/web-ui/web-ui-metadata-json-preview.png',
  'assets/generated/cli/cli-organize-run.gif',
  'assets/generated/cli/cli-rename-preview.gif',
  'assets/generated/tui/tui-organize-preview.gif',
]
