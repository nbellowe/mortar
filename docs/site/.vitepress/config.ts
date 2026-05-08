import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Mortar',
  description: 'One front door for your homelab media stack',
  base: '/mortar/',

  themeConfig: {
    logo: '/logo.svg',

    nav: [
      { text: 'For Users', link: '/users/' },
      { text: 'For Operators', link: '/operators/' },
      { text: 'For Contributors', link: '/contributors/' },
    ],

    sidebar: {
      '/users/': [
        {
          text: 'Using Mortar',
          items: [
            { text: 'Overview', link: '/users/' },
            { text: 'Requesting media', link: '/users/requesting' },
            { text: 'Browsing your library', link: '/users/browsing' },
          ],
        },
      ],
      '/operators/': [
        {
          text: 'Running Mortar',
          items: [
            { text: 'Getting started', link: '/operators/' },
            { text: 'Installation', link: '/operators/installation' },
            { text: 'Configuration', link: '/operators/configuration' },
            { text: 'Plugins', link: '/operators/plugins' },
          ],
        },
      ],
      '/contributors/': [
        {
          text: 'Contributing',
          items: [
            { text: 'Development setup', link: '/contributors/' },
            { text: 'Architecture', link: '/contributors/architecture' },
            { text: 'Writing a plugin', link: '/contributors/plugin-development' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/nbellowe/mortar' },
    ],

    footer: {
      message: 'Released under the MIT License.',
    },

    search: {
      provider: 'local',
    },
  },
})
