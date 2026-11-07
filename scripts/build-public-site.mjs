#!/usr/bin/env node
/**
 * Build static public site (dockyrd.io) from site/source.html.
 * Output: site/index.html, site/assets/*, site/docs/*.html, robots.txt, sitemap.xml
 */
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { dockyrdLogoLink } from "./public-site-logo.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const siteDir = path.join(root, "site");
const sourcePath = path.join(siteDir, "source.html");
const assetsDir = path.join(siteDir, "assets");
const docsOutDir = path.join(siteDir, "docs");
const fontsSrcSans = path.join(
  root,
  "web/node_modules/geist/dist/fonts/geist-sans/Geist-Variable.woff2",
);
const fontsSrcMono = path.join(
  root,
  "web/node_modules/geist/dist/fonts/geist-mono/GeistMono-Variable.woff2",
);

const pkg = JSON.parse(fs.readFileSync(path.join(root, "package.json"), "utf8"));
const VERSION = pkg.version ?? "0.0.0";

const SITE_ORIGIN = "https://dockyrd.io";
const REPO = "Joeydekruis/dockyrd";
const REPO_URL = `https://github.com/${REPO}`;
const INSTALL_CMD = `curl -fsSL https://raw.githubusercontent.com/${REPO}/development/scripts/install.sh | sudo sh`;
const UNINSTALL_CMD = `curl -fsSL https://raw.githubusercontent.com/${REPO}/development/scripts/uninstall.sh | sudo sh`;

const DOC_USER_MAP = {
  introduction: null,
  installation: "install.md",
  "quick-start": "first-app.md",
  nextjs: "nextjs.md",
  "static-sites": null,
  databases: null,
  "object-storage": null,
  domains: "domains-dns.md",
  variables: null,
  updates: null,
  cli: null,
  troubleshooting: "troubleshooting.md",
  architecture: null,
  security: null,
};

function readSource() {
  if (!fs.existsSync(sourcePath)) {
    console.error("Missing site/source.html");
    process.exit(1);
  }
  let html = fs.readFileSync(sourcePath, "utf8");
  html = html.replaceAll("https://get.dockyrd.dev/uninstall", UNINSTALL_CMD.split(" | ")[0].replace("curl -fsSL ", ""));
  html = html.replaceAll("https://get.dockyrd.dev", INSTALL_CMD.split(" | ")[0].replace("curl -fsSL ", ""));
  html = html.replaceAll(
    "curl -fsSL https://raw.githubusercontent.com/Joeydekruis/dockyrd/development/scripts/install.sh | sudo sh",
    INSTALL_CMD,
  );
  html = html.replaceAll("curl -fsSL https://get.dockyrd.dev | sudo sh", INSTALL_CMD);
  html = html.replaceAll(
    "curl -fsSL https://get.dockyrd.dev/uninstall | sudo sh",
    UNINSTALL_CMD,
  );
  html = html.replaceAll("/usr/local/bin/dockyrd", "/usr/local/bin/dockyrd");
  html = html.replaceAll("sudo dockyrd ", "sudo dockyrd ");
  html = html.replaceAll("<code>dockyrd ", "<code>dockyrd ");
  html = html.replaceAll(">dockyrd ", ">dockyrd ");
  html = html.replaceAll("/etc/dockyrd/", "/etc/dockyrd/");
  html = html.replaceAll("/var/lib/dockyrd/", "/var/lib/dockyrd/");
  html = html.replaceAll("var REPO = 'dockyrd/dockyrd'", `var REPO = '${REPO}'`);
  html = html.replaceAll("v0.1 · preview", `v${VERSION}`);
  html = html.replaceAll(
    '<span class="status warn" style="font-size:inherit">v0.1</span>',
    `<span class="status warn" style="font-size:inherit">v${VERSION}</span>`,
  );
  html = html.replaceAll(
    '<link rel="preconnect" href="https://fonts.googleapis.com">',
    "",
  );
  html = html.replaceAll(
    '<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>',
    "",
  );
  html = html.replace(
    /<link rel="stylesheet" href="https:\/\/fonts\.googleapis\.com[^>]+>/g,
    "",
  );
  html = html.replace(
    "Go to <b>Settings → Git providers</b>",
    "Go to <b>Settings → Integrations</b>",
  );
  return html;
}

function extractBlock(html, startMarker, endMarker) {
  const start = html.indexOf(startMarker);
  const end = html.indexOf(endMarker, start);
  if (start < 0 || end < 0) throw new Error(`Block not found: ${startMarker}`);
  return html.slice(start + startMarker.length, end);
}

function copyFonts() {
  fs.mkdirSync(path.join(assetsDir, "fonts"), { recursive: true });
  for (const [src, name] of [
    [fontsSrcSans, "Geist-Variable.woff2"],
    [fontsSrcMono, "GeistMono-Variable.woff2"],
  ]) {
    if (!fs.existsSync(src)) {
      console.warn(`Font missing (${src}); run pnpm --dir web install`);
      continue;
    }
    fs.copyFileSync(src, path.join(assetsDir, "fonts", name));
  }
}

const fontFace = `
@font-face{font-family:"Geist";font-style:normal;font-weight:100 900;font-display:swap;src:url("fonts/Geist-Variable.woff2") format("woff2")}
@font-face{font-family:"Geist Mono";font-style:normal;font-weight:100 900;font-display:swap;src:url("fonts/GeistMono-Variable.woff2") format("woff2")}
`;

/** depth 0 = site/index.html, 1 = site/docs/*.html */
function assetsPrefix(depth) {
  return depth === 0 ? "assets/" : "../assets/";
}

function homeHref(depth) {
  return depth === 0 ? "./" : "../";
}

function docHref(depth, page) {
  return depth === 0 ? `docs/${page}.html` : `${page}.html`;
}

function extractArticles(html) {
  const start = html.indexOf('<div id="docs-pages">');
  const end = html.indexOf('<nav class="pager"', start);
  if (start < 0 || end < 0) throw new Error("docs-pages block not found");
  const pagesBlock = html.slice(start, end);
  const re =
    /<article class="doc" data-page="([^"]+)" data-section="([^"]+)" data-title="([^"]+)">([\s\S]*?)<\/article>/g;
  const articles = [];
  let m;
  while ((m = re.exec(pagesBlock)) !== null) {
    articles.push({
      page: m[1],
      section: m[2],
      title: m[3],
      body: m[4],
    });
  }
  if (!articles.length) throw new Error("No doc articles found in source.html");
  return articles;
}

function extractHome(html) {
  return extractBlock(html, "<!-- ======================================================= HOME ======== -->", "<!-- ======================================================= DOCS ======== -->");
}

function extractSvg(html) {
  return extractBlock(html, "<!-- ============ SVG sprite ============ -->", "<div class=\"banner\"");
}

function extractStyles(html) {
  const css = extractBlock(html, "/* Layout: sticky top bar", "</style>");
  return fontFace + "/* Layout: sticky top bar" + css;
}

function searchDialogHtml() {
  return `<dialog class="search" id="searchDlg" aria-label="Search documentation">
  <div class="search-field">
    <label for="searchInput"><svg width="18" height="18" aria-hidden="true"><use href="#i-search"/></svg><span hidden>Search</span></label>
    <input id="searchInput" type="search" placeholder="Search the docs" autocomplete="off" spellcheck="false">
    <kbd>Esc</kbd>
  </div>
  <ul class="results" id="results" role="listbox"></ul>
</dialog>`;
}

function pageUrl(slug) {
  return slug ? `${SITE_ORIGIN}/docs/${slug}.html` : `${SITE_ORIGIN}/`;
}

function rewriteDocLinks(html, depth) {
  const toDoc = (page) => docHref(depth, page);
  let out = html
    .replaceAll('href="#docs-', 'href="__DOC__')
    .replace(/href="__DOC__([^"#]+)"/g, (_, slug) => `href="${toDoc(slug)}"`)
    .replaceAll('href="#home"', `href="${homeHref(depth)}"`)
    .replaceAll('href="#docs-introduction"', `href="${toDoc("introduction")}"`);
  return out;
}

function headMeta({ title, description, canonical, ogType = "website", depth = 0 }) {
  const t = title.includes("Dockyrd") ? title : `${title} · Dockyrd`;
  const desc = description.replace(/"/g, "&quot;");
  const assets = assetsPrefix(depth);
  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover">
<title>${t}</title>
<meta name="description" content="${desc}">
<link rel="canonical" href="${canonical}">
<meta property="og:type" content="${ogType}">
<meta property="og:site_name" content="Dockyrd">
<meta property="og:title" content="${t}">
<meta property="og:description" content="${desc}">
<meta property="og:url" content="${canonical}">
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="${t}">
<meta name="twitter:description" content="${desc}">
<link rel="icon" href="${assets}favicon.svg" type="image/svg+xml">
<link rel="stylesheet" href="${assets}site.css?v=${VERSION}">
</head>
<body>`;
}

function shellHeader(activeNav, depth = 0) {
  const navItem = (key, href, label) => {
    const cur = activeNav === key ? ' aria-current="page"' : "";
    return `<a href="${href}" data-nav="${key}"${cur}>${label}</a>`;
  };
  const home = homeHref(depth);
  return `${extractSvgPlaceholder()}
<div class="banner" role="region" aria-label="Announcement">
  <div class="banner-in">
    <span class="tag">Open source</span>
    <span>One lightweight binary on your own server. No lock-in, no mess.</span>
    <a href="${docHref(depth, "introduction")}">Read the docs →</a>
  </div>
</div>
<header class="topbar">
  <div class="topbar-in">
    <button class="menu-btn" type="button" id="menuBtn" aria-label="Open documentation menu" aria-expanded="false" aria-controls="sidebar"><svg width="20" height="20"><use href="#i-menu"/></svg></button>
    ${dockyrdLogoLink(home, "site-logo-h")}
    <nav class="topnav" aria-label="Primary">
      ${navItem("home", home, "Product")}
      ${navItem("docs", docHref(depth, "introduction"), "Docs")}
      ${navItem("cli", docHref(depth, "cli"), "CLI")}
      ${navItem("arch", docHref(depth, "architecture"), "Architecture")}
    </nav>
    <div class="top-right">
      <button class="search-btn" type="button" id="searchBtn" aria-label="Search documentation">
        <svg width="16" height="16" aria-hidden="true"><use href="#i-search"/></svg><span>Search docs</span><kbd>⌘K</kbd>
      </button>
      <span class="version">v${VERSION}</span>
      <a class="btn btn-outline gh-btn" href="${REPO_URL}" target="_blank" rel="noopener noreferrer" aria-label="Star Dockyrd on GitHub">
        <svg width="16" height="16" aria-hidden="true"><use href="#i-github"/></svg>
        <span class="gh-label">Star</span>
        <span class="star-count" data-stars></span>
      </a>
    </div>
  </div>
</header>`;
}

let svgCache = "";
function extractSvgPlaceholder() {
  return svgCache;
}

function shellFooter(depth = 0) {
  const home = homeHref(depth);
  return `<footer class="site">
  <div class="in">
    ${dockyrdLogoLink(home, "site-logo-f", 104)}
    <nav aria-label="Footer">
      <a href="${docHref(depth, "introduction")}">Docs</a>
      <a href="${docHref(depth, "installation")}">Install</a>
      <a href="${docHref(depth, "cli")}">CLI</a>
      <a href="${REPO_URL}" target="_blank" rel="noopener noreferrer">GitHub</a>
      <a href="${REPO_URL}/releases" target="_blank" rel="noopener noreferrer">Releases</a>
      <a href="${docHref(depth, "security")}">Security</a>
    </nav>
    <span>Open source under Apache-2.0</span>
  </div>
</footer>`;
}

function buildSidebarNav(articles, currentPage, depth = 1) {
  const groups = new Map();
  for (const a of articles) {
    if (!groups.has(a.section)) groups.set(a.section, []);
    groups.get(a.section).push(a);
  }
  let html = "";
  for (const [section, items] of groups) {
    html += `<div class="group"><span class="eyebrow">${section}</span>`;
    for (const item of items) {
      const cur = item.page === currentPage ? ' aria-current="page"' : "";
      html += `<a href="${docHref(depth, item.page)}"${cur}>${item.title}</a>`;
    }
    html += "</div>";
  }
  return html;
}

function buildToc(articleHtml) {
  const re = /<h2 id="([^"]+)">([^<]+)/g;
  let html = "";
  let m;
  while ((m = re.exec(articleHtml)) !== null) {
    html += `<li><a href="#${m[1]}">${m[1].includes("-") ? m[2] : m[2]}</a></li>`;
  }
  return html ? `<ol>${html}</ol>` : "";
}

function editLink(page) {
  const md = DOC_USER_MAP[page];
  const href = md
    ? `${REPO_URL}/edit/development/docs/user/${md}`
    : `${REPO_URL}/tree/development/site/source.html`;
  return `<a href="${href}" target="_blank" rel="noopener noreferrer">Edit this page on GitHub</a>`;
}

function buildDocPage(article, articles, index) {
  const depth = 1;
  const assets = assetsPrefix(depth);
  const body = rewriteDocLinks(article.body, depth);
  const prev = index > 0 ? articles[index - 1] : null;
  const next = index < articles.length - 1 ? articles[index + 1] : null;
  let pager = '<nav class="pager" aria-label="Previous and next page">';
  if (prev) {
    pager += `<a class="prev" href="${docHref(depth, prev.page)}"><span class="eyebrow">Previous</span><span>${prev.title}</span></a>`;
  }
  if (next) {
    pager += `<a class="next" href="${docHref(depth, next.page)}"><span class="eyebrow">Next</span><span>${next.title}</span></a>`;
  }
  pager += "</nav>";

  const leadMatch = body.match(/<p class="lead">([\s\S]*?)<\/p>/);
  const description = leadMatch
    ? leadMatch[1].replace(/<[^>]+>/g, "").trim()
    : `${article.title} — Dockyrd documentation`;

  const navKey =
    article.page === "cli"
      ? "cli"
      : article.page === "architecture"
        ? "arch"
        : "docs";

  return `${headMeta({
    title: article.title,
    description,
    canonical: pageUrl(article.page),
    ogType: "article",
    depth,
  })}
${shellHeader(navKey, depth)}
<div class="docs">
  <aside class="sidebar" id="sidebar" aria-label="Documentation">
    <nav id="docnav">${buildSidebarNav(articles, article.page, depth)}</nav>
    <div class="side-foot">
      <span class="status warn">Pre-release</span>
      <span style="color:var(--ink-2)">These docs track v${VERSION}. Details can change before the next release.</span>
    </div>
  </aside>
  <main class="article" id="main">
    <div class="crumbs"><span class="eyebrow">${article.section}</span></div>
    <article class="doc">${body}</article>
    ${pager}
    <div class="edit">${editLink(article.page)}<span>Apache-2.0 licensed</span></div>
  </main>
  <aside class="toc" aria-label="On this page"><span class="eyebrow">On this page</span>${buildToc(body)}</aside>
</div>
${shellFooter(depth)}
${searchDialogHtml()}
<script src="${assets}site-config.js"></script>
<script src="${assets}site.js"></script>
</body></html>`;
}

function buildHome(homeRaw) {
  const depth = 0;
  const assets = assetsPrefix(depth);
  let home = homeRaw.replace('<main id="view-home">', "<main>");
  home = rewriteDocLinks(home, depth);
  home = home.replace(
    `<code><span class="p">$ </span>${INSTALL_CMD.replace(/&/g, "&amp;")}</code>`,
    `<code><span class="p">$ </span>${INSTALL_CMD}</code>`,
  );
  home = home.replaceAll(
    `data-copy="${INSTALL_CMD}"`,
    `data-copy="${INSTALL_CMD}"`,
  );

  const jsonLd = `<script type="application/ld+json">${JSON.stringify({
    "@context": "https://schema.org",
    "@type": "SoftwareApplication",
    name: "Dockyrd",
    applicationCategory: "DeveloperApplication",
    operatingSystem: "Linux",
    offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
    downloadUrl: `${REPO_URL}/releases`,
    softwareHelp: pageUrl("introduction"),
    license: "https://www.apache.org/licenses/LICENSE-2.0",
  })}</script>`;

  return `${headMeta({
    title: "Dockyrd — Self-hosted deploy platform",
    description:
      "Dockyrd is an open source, self-hosted deploy platform. One Go binary, one install command, apps, static sites, databases and object storage on your own server.",
    canonical: `${SITE_ORIGIN}/`,
    depth,
  })}
${jsonLd}
${shellHeader("home", depth)}
${home}
${shellFooter(depth)}
${searchDialogHtml()}
<script src="${assets}site-config.js"></script>
<script src="${assets}site.js"></script>
</body></html>`;
}

function writeSiteAssets(articles) {
  const searchIndex = articles.map((a) => ({
    href: `docs/${a.page}.html`,
    title: a.title,
    sec: a.section,
    text: "",
  }));

  const config = `window.DOCKYRD_SITE=${JSON.stringify({
    repo: REPO,
    repoUrl: REPO_URL,
    installCmd: INSTALL_CMD,
    searchIndex,
  })};\n`;

  fs.writeFileSync(path.join(assetsDir, "site-config.js"), config);
  fs.copyFileSync(
    path.join(siteDir, "site.runtime.js"),
    path.join(assetsDir, "site.js"),
  );
}

function main() {
  const html = readSource();
  svgCache = extractSvg(html).trim();
  const articles = extractArticles(html);
  const homeRaw = extractHome(html);
  const css = extractStyles(html);

  fs.mkdirSync(assetsDir, { recursive: true });
  fs.mkdirSync(docsOutDir, { recursive: true });
  copyFonts();

  fs.writeFileSync(path.join(assetsDir, "site.css"), css);
  fs.copyFileSync(
    path.join(root, "web/public/favicon.svg"),
    path.join(assetsDir, "favicon.svg"),
  );

  writeSiteAssets(articles);

  fs.writeFileSync(path.join(siteDir, "index.html"), buildHome(homeRaw));

  for (let i = 0; i < articles.length; i++) {
    const out = path.join(docsOutDir, `${articles[i].page}.html`);
    fs.writeFileSync(out, buildDocPage(articles[i], articles, i));
  }

  fs.writeFileSync(
    path.join(docsOutDir, "index.html"),
    `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta http-equiv="refresh" content="0;url=introduction.html"><link rel="canonical" href="${SITE_ORIGIN}/docs/introduction.html"><title>Dockyrd Docs</title></head><body><p><a href="introduction.html">Dockyrd documentation</a></p></body></html>`,
  );

  const urls = [`${SITE_ORIGIN}/`, ...articles.map((a) => pageUrl(a.page))];
  fs.writeFileSync(
    path.join(siteDir, "sitemap.xml"),
    `<?xml version="1.0" encoding="UTF-8"?>\n<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">\n${urls.map((u) => `  <url><loc>${u}</loc></url>`).join("\n")}\n</urlset>\n`,
  );

  fs.writeFileSync(
    path.join(siteDir, "robots.txt"),
    `User-agent: *\nAllow: /\nSitemap: ${SITE_ORIGIN}/sitemap.xml\n`,
  );

  console.log(`Built public site (${articles.length} doc pages, v${VERSION}) → site/`);
}

main();
