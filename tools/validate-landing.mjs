#!/usr/bin/env node
/* Integrity check for the bilingual, single-page ScopeWeaver landing page.
   Lives outside landing/ so it is never published. */

import { readFileSync, existsSync, readdirSync, statSync, realpathSync } from "node:fs";
import { resolve, dirname, join, relative, isAbsolute, sep } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const landing = join(root, "landing");
let errors = 0;
let checks = 0;
const fail = (m) => { errors++; console.error("  ✗ " + m); };
const check = (condition, m) => {
  if (condition) { checks++; console.log("  ✓ " + m); }
  else fail(m);
};

const file = join(landing, "index.html");
if (!existsSync(file)) {
  fail("missing page: " + file);
  process.exit(1);
}
const html = readFileSync(file, "utf8");

// shortcut: a static, explicitly closed HTML document; use a DOM parser if the page adopts optional closing tags.
const document = { children: [] };
const stack = [document];
const elements = [];
const voidTags = new Set("area base br col embed hr img input link meta param source track wbr".split(" "));
for (const match of html.matchAll(/<!--[\s\S]*?-->|<![^>]*>|<\/?([a-z][\w-]*)\b([^>]*)>/gi)) {
  if (!match[1]) continue;
  const tag = match[1].toLowerCase();
  if (match[0].startsWith("</")) {
    if (stack.length === 1 || stack.at(-1).tag !== tag) fail(`unbalanced closing tag: ${tag}`);
    else stack.pop();
    continue;
  }
  const attrs = Object.fromEntries([...match[2].matchAll(/([^\s=/>]+)(?:\s*=\s*(?:"([^"]*)"|'([^']*)'|([^\s>]+)))?/g)]
    .map((m) => [m[1].toLowerCase(), m[2] ?? m[3] ?? m[4] ?? ""]));
  const parent = stack.at(-1);
  const element = { tag, attrs, parent, children: [] };
  parent.children.push(element);
  elements.push(element);
  if (!voidTags.has(tag) && !match[0].endsWith("/>")) stack.push(element);
}
check(stack.length === 1, "HTML elements are balanced");
const select = (attr, value) => elements.filter((el) => value === undefined ? attr in el.attrs : el.attrs[attr] === value);
const hasClass = (el, name) => (el.attrs.class || "").split(/\s+/).includes(name);
const htmlEl = elements.find((el) => el.tag === "html");
check(htmlEl?.attrs.lang === "en" && htmlEl.attrs["data-lang"] === "en", "initial html lang and data-lang are en");

/* --- 1. Local HTML/CSS assets stay under the project-pages folder. --- */
const contained = (target) => {
  const path = relative(realpathSync(landing), target);
  return path === "" || (path !== ".." && !path.startsWith(".." + sep) && !isAbsolute(path));
};
const scanned = new Set();
const checkRef = (ref, base) => {
  if (/^(https?:|mailto:|data:|\/\/)/i.test(ref) || ref.startsWith("#")) return;
  if (ref.startsWith("/")) { fail(`absolute root asset path: ${ref}`); return; }
  let clean;
  try { clean = decodeURIComponent(ref.replace(/[#?].*$/, "")); }
  catch { fail(`invalid URL encoding: ${ref}`); return; }
  if (!clean) return;
  const target = resolve(base, clean);
  if (!contained(target)) { fail(`reference escapes landing/: ${ref}`); return; }
  if (!existsSync(target)) { fail(`missing asset -> ${ref}`); return; }
  if (!contained(realpathSync(target))) { fail(`reference escapes landing/ through symlink: ${ref}`); return; }
  if (/\.css$/i.test(target) && !scanned.has(target)) {
    scanned.add(target);
    for (const match of readFileSync(target, "utf8").matchAll(/url\(\s*["']?([^"')]+)["']?\s*\)/g)) checkRef(match[1].trim(), dirname(target));
  }
};
const assetErrors = errors;
for (const el of elements) {
  for (const attr of ["src", "href"]) if (el.attrs[attr]) checkRef(el.attrs[attr], landing);
  if (el.attrs.srcset) for (const ref of el.attrs.srcset.split(",")) checkRef(ref.trim().split(/\s+/)[0], landing);
}
check(errors === assetErrors, "all HTML and CSS references resolve inside landing/");

/* --- 2. Shared anchors and paired EN/KO content replace two-page parity. --- */
const byId = new Map();
for (const el of select("id")) {
  const id = el.attrs.id;
  if (byId.has(id)) fail(`duplicate id: ${id}`);
  else byId.set(id, el);
}
for (const el of select("href")) {
  if (el.attrs.href.startsWith("#")) check(byId.has(el.attrs.href.slice(1)), `anchor target exists: ${el.attrs.href}`);
}
const localized = select("data-i18n");
check(localized.length > 0, "bilingual data-i18n content exists");
for (const el of localized) {
  const lang = el.attrs["data-i18n"];
  const siblings = el.parent.children;
  const peer = siblings[siblings.indexOf(el) + (lang === "en" ? 1 : -1)];
  if (!["en", "ko"].includes(lang) || peer?.attrs["data-i18n"] !== (lang === "en" ? "ko" : "en") || peer.tag !== el.tag)
    fail(`data-i18n ${lang}: missing matching adjacent EN/KO sibling (${el.tag})`);
  if (lang === "ko" && el.attrs.lang !== "ko") fail("Korean data-i18n content missing lang=ko");
}
const switches = select("data-set-lang");
check(switches.length === 2 && ["en", "ko"].every((lang) => switches.some((el) =>
  el.tag === "button" && el.attrs.type === "button" && el.attrs["data-set-lang"] === lang &&
  el.attrs["aria-pressed"] === (lang === "en" ? "true" : "false") &&
  el.parent.attrs.role === "group" && el.parent.attrs["aria-label"])), "accessible EN/KO language switch and initial state");

/* --- 3. Controls resolve to real, accessible targets in the shared page. --- */
const tabs = select("role", "tab");
const panels = select("role", "tabpanel");
const tablists = select("role", "tablist");
check(tabs.length === 2 && panels.length === 2 && tablists.length === 1, "shell tablist has two tabs and panels");
check(tablists.every((el) => el.attrs["aria-label"]), "shell tablist is labelled");
for (const tab of tabs) {
  const panel = byId.get(tab.attrs["aria-controls"]);
  check(tab.tag === "button" && tab.attrs.type === "button" && tab.parent.attrs.role === "tablist" &&
    panel?.attrs.role === "tabpanel" && panel.attrs["aria-labelledby"] === tab.attrs.id &&
    ["posix", "pwsh"].includes(tab.attrs["data-shell-tab"]) && panel.attrs["data-shell-panel"] === tab.attrs["data-shell-tab"],
    `shell tab and panel are linked: ${tab.attrs.id}`);
  check(["true", "false"].includes(tab.attrs["aria-selected"]) &&
    panel && ("hidden" in panel.attrs) === (tab.attrs["aria-selected"] === "false"), `shell visibility matches selection: ${tab.attrs.id}`);
}
check(tabs.filter((el) => el.attrs["aria-selected"] === "true").length === 1 &&
  panels.every((panel) => tabs.some((tab) => tab.attrs["aria-controls"] === panel.attrs.id)), "one selected tab and no orphan panels");
const copyTargets = {
  "active-hero-code": ["hero-code-posix", "hero-code-pwsh"],
  "active-task-json": ["task-json-en", "task-json-ko"],
};
const copies = select("data-copy-target");
check(copies.length > 0, "copy controls exist");
for (const button of copies) {
  const target = button.attrs["data-copy-target"];
  check(button.tag === "button" && button.attrs.type === "button" && hasClass(button, "copy-btn") &&
    (copyTargets[target] || [target]).every((id) => byId.has(id)) &&
    ["en", "ko"].every((lang) => button.children.some((el) => hasClass(el, "copy-btn__label") && el.attrs["data-i18n"] === lang)),
    `copy control has bilingual labels and targets: ${target}`);
}
const announcer = byId.get("sr-announcer");
check(announcer?.attrs["aria-live"] === "polite" && announcer.attrs["aria-atomic"] === "true", "copy feedback has a live announcer");
check(elements.some((el) => el.tag === "script" && el.attrs.src === "script.js"), "interaction script is loaded");
check(elements.some((el) => el.tag === "link" && el.attrs.rel === "stylesheet" && el.attrs.href === "styles.css"), "language and layout stylesheet is loaded");

/* --- 4. Every screenshot figure carries its own demo caption. --- */
const figures = elements.filter((el) => el.tag === "figure");
for (const figure of figures)
  check(figure.children.some((el) => el.tag === "figcaption" && hasClass(el, "demo")), "figure has a demo caption");
check(true, `${figures.length} screenshot figures checked for demo captions`);

/* --- 5. No planning/seed/contract metadata shipped in landing/. --- */
const banned = [/THESIS:/, /OWN-WORLD:/, /surface seed/i, /Direction contract/i, /impeccable:product-schema/];
const metadataErrors = errors;
const walk = (dir) => {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (!contained(realpathSync(path))) { fail(`shipped path escapes landing/: ${path}`); continue; }
    if (statSync(path).isDirectory()) walk(path);
    else if (/\.(html|css|js|txt|xml|json|md)$/.test(entry)) {
      const text = readFileSync(path, "utf8");
      for (const re of banned) if (re.test(text)) fail(`planning metadata leaked into ${path} (${re})`);
    }
  }
};
walk(landing);
check(errors === metadataErrors, "no planning/seed/contract metadata found in landing/");
check(existsSync(join(landing, ".nojekyll")), ".nojekyll present");

console.log(`\n${checks} checks passed, ${errors} error(s).`);
process.exit(errors ? 1 : 0);
