# Documentation site

This directory is the [VitePress](https://vitepress.dev) source for the Mortar documentation site, published to GitHub Pages at `https://nbellowe.github.io/mortar/`.

## Structure

```
index.md             — landing page
users/               — end-user docs (plain language, task-oriented)
operators/           — operator docs (installation, config, plugins)
contributors/        — contributor docs (dev setup, architecture, plugin development)
.vitepress/config.ts — site config
```

Contributor reference docs that aren't part of the rendered site (CONTRIBUTING.md, CODE_OF_CONDUCT.md, etc.) live at the repo root.

## Authoring rule of thumb

- **Users** docs: focus on tasks, avoid homelab jargon
- **Operators** docs: explain config fields, deployment, and plugin setup for technical readers
- **Contributors** docs: dev setup, architecture, and extension points

## Running locally

```bash
cd docs/site
npm install
npm run dev
```

Open `http://localhost:5173/mortar/`.

## Deployment

Pushing to `main` triggers the `docs.yml` GitHub Actions workflow, which builds the site and deploys it to GitHub Pages. No manual steps required.
