import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'discord.go',
  description: 'A typed Go library for building Discord applications',
  cleanUrls: true,
  lastUpdated: true,

  head: [
    ['meta', { name: 'theme-color', content: '#5865f2' }],
    ['link', { rel: 'icon', href: '/favicon.svg', type: 'image/svg+xml' }],
  ],

  srcDir: '.',
  outDir: '.vitepress/dist',

  // Ignore dead links to .go source files, example code, and other non-docs content
  ignoreDeadLinks: true,

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'discord.go',

    nav: [
      { text: 'Guide', link: '/' },
      { text: 'High-Level API', link: '/high-level/' },
      { text: 'Low-Level API', link: '/low-level/' },
      { text: 'Examples', link: '/examples/' },
      { text: 'pkg.go.dev', link: 'https://pkg.go.dev/github.com/discord-go/discord.go' },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/discord-go/discord.go' },
    ],

    search: {
      provider: 'local',
      options: {
        translations: {
          button: {
            buttonText: 'Search docs',
            buttonAriaLabel: 'Search documentation',
          },
          modal: {
            noResultsText: 'No results found',
            resetButtonTitle: 'Clear search',
            footer: {
              selectText: 'to select',
              navigateText: 'to navigate',
            },
          },
        },
      },
    },

    outline: {
      label: 'On This Page',
      level: [2, 3],
    },

    docFooter: {
      prev: 'Previous',
      next: 'Next',
    },

    sidebar: [
      {
        text: 'Introduction',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/' },
        ],
      },
      {
        text: 'Getting Started',
        collapsed: false,
        items: [
          { text: 'Overview', link: '/examples/setup/' },
          { text: 'App Setup', link: '/examples/setup/app-setup' },
          { text: 'Adding Your App', link: '/examples/setup/adding-your-app' },
          { text: 'Installation', link: '/examples/setup/installation' },
          { text: 'Basic Client', link: '/examples/setup/basic-client' },
          { text: 'Linter', link: '/examples/setup/linter' },
        ],
      },
      {
        text: 'Creating Your App',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/commands/' },
          { text: 'Project Setup', link: '/examples/commands/project-setup' },
          { text: 'Main File', link: '/examples/commands/main-file' },
          { text: 'Creating Commands', link: '/examples/commands/creating-commands' },
          { text: 'Handling Commands', link: '/examples/commands/handling-commands' },
          { text: 'Handling Events', link: '/examples/commands/handling-events' },
          { text: 'Slash Commands', link: '/examples/commands/slash-commands' },
          { text: 'Parsing Options', link: '/examples/commands/parsing-options' },
          { text: 'Command Responses', link: '/examples/commands/command-responses' },
          { text: 'Deploying Commands', link: '/examples/commands/deploying-commands' },
          { text: 'Reloading Commands', link: '/examples/commands/reloading-commands' },
          { text: 'Deleting Commands', link: '/examples/commands/deleting-commands' },
          { text: 'Slash Command Permissions', link: '/examples/commands/slash-permissions' },
          { text: 'Cooldowns', link: '/examples/commands/cooldowns' },
          { text: 'Context Menus', link: '/examples/commands/context-menus' },
          { text: 'Advanced Command Creation', link: '/examples/commands/advanced-command-creation' },
          { text: 'Autocomplete', link: '/examples/commands/autocomplete' },
          { text: 'Moderation', link: '/examples/commands/moderation' },
        ],
      },
      {
        text: 'Interactions',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/interactions/' },
          { text: 'Interactions', link: '/examples/interactions/interactions' },
          { text: 'Buttons', link: '/examples/interactions/buttons' },
          { text: 'Action Rows', link: '/examples/interactions/action-rows' },
          { text: 'Select Menus', link: '/examples/interactions/select-menus' },
          { text: 'Modals', link: '/examples/interactions/modals' },
          { text: 'Display Components', link: '/examples/interactions/display-components' },
          { text: 'Components V2', link: '/examples/interactions/components-v2' },
        ],
      },
      {
        text: 'More To Know',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/more-to-know/' },
          { text: 'Gateway', link: '/examples/more-to-know/gateway' },
          { text: 'Gateway Intents', link: '/examples/more-to-know/gateway-intents' },
          { text: 'Audit Logs', link: '/examples/more-to-know/audit-logs' },
          { text: 'Collectors', link: '/examples/more-to-know/collectors' },
          { text: 'Embeds', link: '/examples/more-to-know/embeds' },
          { text: 'Formatters', link: '/examples/more-to-know/formatters' },
          { text: 'Partials And Cache', link: '/examples/more-to-know/partials-cache' },
          { text: 'Permissions', link: '/examples/more-to-know/permissions' },
          { text: 'Reactions', link: '/examples/more-to-know/reactions' },
          { text: 'Threads', link: '/examples/more-to-know/threads' },
          { text: 'Webhooks', link: '/examples/more-to-know/webhooks' },
          { text: 'Canvas Alternatives', link: '/examples/more-to-know/canvas-alternatives' },
          { text: 'Common Errors', link: '/examples/more-to-know/common-errors' },
        ],
      },
      {
        text: 'Persistence',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/persistence/' },
          { text: 'Keyv-Compatible Persistence', link: '/examples/persistence/keyv' },
          { text: 'Sequelize-Model Persistence', link: '/examples/persistence/sequelize' },
        ],
      },
      {
        text: 'Advanced',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/advanced/' },
          { text: 'Full Template', link: '/examples/advanced/full-template' },
          { text: 'OAuth2 Authorization Code Flow', link: '/examples/advanced/oauth2' },
          { text: 'Gateway Sharding', link: '/examples/advanced/sharding' },
        ],
      },
      {
        text: 'Voice',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/voice/' },
        ],
      },
      {
        text: 'Templates',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/templates/' },
        ],
      },
      {
        text: 'Runnable Code',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/examples/code/' },
        ],
      },
      {
        text: 'High-Level API',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/high-level/' },
          { text: 'Bot Client', link: '/high-level/client' },
          { text: 'Configuration', link: '/high-level/configuration' },
          { text: 'Commands And Routing', link: '/high-level/commands' },
          { text: 'Interaction Responses And Data', link: '/high-level/interactions' },
          { text: 'Components And Components V2', link: '/high-level/components' },
          { text: 'Buttons', link: '/high-level/buttons' },
          { text: 'Modals And Text Inputs', link: '/high-level/modals' },
          { text: 'Collectors And Jobs', link: '/high-level/collectors' },
          { text: 'Permissions And Middleware', link: '/high-level/permissions' },
          { text: 'Lifecycle And Shutdown', link: '/high-level/lifecycle' },
          { text: 'Presence And Latency', link: '/high-level/presence' },
          { text: 'Caching', link: '/high-level/caching' },
          { text: 'Resource Access', link: '/high-level/resources' },
          { text: 'Embeds And Rich Messages', link: '/high-level/embeds' },
          { text: 'Voice State Control', link: '/high-level/voice' },
          { text: 'Errors And Recovery', link: '/high-level/errors' },
        ],
      },
      {
        text: 'Low-Level API',
        collapsed: true,
        items: [
          { text: 'Overview', link: '/low-level/' },
          { text: 'Client', link: '/low-level/client/' },
          { text: 'REST', link: '/low-level/rest/' },
          { text: '  REST Requests', link: '/low-level/rest/requests' },
          { text: '  Endpoint Groups', link: '/low-level/rest/endpoints' },
          { text: '  REST Rate Limits', link: '/low-level/rest/ratelimits' },
          { text: '  Multipart Uploads', link: '/low-level/rest/uploads' },
          { text: 'Gateway', link: '/low-level/gateway/' },
          { text: '  Gateway Events', link: '/low-level/gateway/events' },
          { text: '  Heartbeats', link: '/low-level/gateway/heartbeat' },
          { text: '  Shards', link: '/low-level/gateway/shards' },
          { text: 'HTTP', link: '/low-level/http/' },
          { text: 'Ratelimit', link: '/low-level/ratelimit/' },
          { text: 'Cache', link: '/low-level/cache/' },
          { text: 'Storage', link: '/low-level/storage/' },
          { text: 'Intents', link: '/low-level/intents/' },
          { text: 'Events', link: '/low-level/events/' },
          { text: 'Snowflakes', link: '/low-level/snowflake/' },
          { text: 'JSON', link: '/low-level/json/' },
          { text: 'Resource Models', link: '/low-level/models/' },
          { text: 'Applications', link: '/low-level/application/' },
          { text: 'Audit Logs', link: '/low-level/auditlog/' },
          { text: 'CDN', link: '/low-level/cdn/' },
          { text: 'Channels', link: '/low-level/channels/' },
          { text: 'Components', link: '/low-level/components/' },
          { text: 'Emojis And Stickers', link: '/low-level/emojis/' },
          { text: 'Guilds', link: '/low-level/guilds/' },
          { text: 'Interactions', link: '/low-level/interactions/' },
          { text: 'Messages', link: '/low-level/messages/' },
          { text: 'OAuth2', link: '/low-level/oauth2/' },
          { text: 'Permissions', link: '/low-level/permissions/' },
          { text: 'Roles', link: '/low-level/roles/' },
          { text: 'Users', link: '/low-level/users/' },
          { text: 'Voice', link: '/low-level/voice/' },
          { text: 'Webhooks', link: '/low-level/webhook/' },
        ],
      },
    ],

    footer: {
      message: 'Released under the Apache License 2.0.',
      copyright: 'Copyright © 2026 discord.go contributors',
    },
  },

  markdown: {
    lineNumbers: true,
    config: (md) => {
      // Rewrite relative .md links to work with VitePress routing
      md.core.ruler.after('inline', 'fix-relative-links', (state) => {
        state.tokens.forEach((token) => {
          if (token.type === 'inline' && token.children) {
            token.children.forEach((child) => {
              if (child.type === 'link_open') {
                const href = child.attrs.find((a) => a[0] === 'href')
                if (href && href[1]) {
                  let url = href[1]
                  // Only fix relative .md links (not http, not anchors, not .go files)
                  if (url.endsWith('.md') && !url.startsWith('http') && !url.startsWith('#')) {
                    // Remove .md extension
                    url = url.replace(/\.md$/, '')
                    // Handle README and index links -> directory index
                    if (url.endsWith('/README')) {
                      url = url.replace(/\/README$/, '/')
                    } else if (url === 'README') {
                      url = './'
                    } else if (url.endsWith('/index')) {
                      url = url.replace(/\/index$/, '/')
                    } else if (url === 'index') {
                      url = './'
                    }
                    href[1] = url
                  }
                }
              }
            })
          }
        })
      })
    },
  },
})
