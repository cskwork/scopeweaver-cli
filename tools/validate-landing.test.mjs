import assert from "node:assert/strict";
import { cpSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";
import test from "node:test";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "..");

test("single-page landing validation rejects broken bilingual content, controls and assets", async (t) => {
  const cases = [
    ["valid single-page EN/KO", null, null],
    ["missing KO sibling", (html) => html.replace('<span data-i18n="ko" lang="ko">본문으로 건너뛰기</span>', ""), /missing matching adjacent EN\/KO sibling/],
    ["balanced counts with wrong sibling", (html) => html.replace('<span data-i18n="ko" lang="ko">본문으로 건너뛰기</span>', '<div><span data-i18n="ko" lang="ko">본문으로 건너뛰기</span></div>'), /missing matching adjacent EN\/KO sibling/],
    ["missing asset", (html) => html.replace('href="scope-mark.svg"', 'href="missing.svg"'), /missing asset -> missing.svg/],
    ["asset escapes into sibling folder", (html) => html.replace('href="scope-mark.svg"', 'href="../landing-other/asset.svg"'), /reference escapes landing/],
    ["absolute asset path", (html) => html.replace('href="scope-mark.svg"', 'href="/scope-mark.svg"'), /absolute root asset path/],
    ["bad tab target", (html) => html.replace('aria-controls="panel-posix"', 'aria-controls="missing-panel"'), /shell tab and panel are linked: tab-posix/],
    ["duplicate ID", (html) => html.replace('id="tab-posix"', 'id="main-content"'), /duplicate id: main-content/],
    ["bad copy target", (html) => html.replace('data-copy-target="agent-commands-code"', 'data-copy-target="missing-code"'), /copy control has bilingual labels and targets: missing-code/],
    ["missing KO lang", (html) => html.replace('data-i18n="ko" lang="ko"', 'data-i18n="ko"'), /Korean data-i18n content missing lang=ko/],
    ["invalid language switch state", (html) => html.replace('data-set-lang="ko" aria-pressed="false"', 'data-set-lang="ko" aria-pressed="true"'), /accessible EN\/KO language switch and initial state/],
    ["orphan fragment link", (html) => html.replace('href="#main-content"', 'href="#missing-anchor"'), /anchor target exists: #missing-anchor/],
    ["figure missing demo caption", (html) => html.replace('</main>', '<figure><img src="scope-mark.svg" alt="Demo"></figure></main>'), /figure has a demo caption/],
    ["figure with demo caption", (html) => html.replace('</main>', '<figure><img src="scope-mark.svg" alt="Demo"><figcaption class="demo">Illustrative demo</figcaption></figure></main>'), null],
    ["planning metadata", (html) => html.replace('</main>', '<p>THESIS: internal planning</p></main>'), /planning metadata leaked/],
    ["missing CSS font", null, /missing asset -> assets\/fonts\/missing.woff2/, (landing) => {
      const file = join(landing, "styles.css");
      writeFileSync(file, readFileSync(file, "utf8").replace("schibsted-grotesk-latin-wght.woff2", "missing.woff2"));
    }],
    ["missing nojekyll", null, /\.nojekyll present/, (landing) => rmSync(join(landing, ".nojekyll"))],
    ["missing page", null, /missing page:/, (landing) => rmSync(join(landing, "index.html"))],
  ];

  for (const [name, mutate, failure, mutateFiles] of cases) {
    await t.test(name, () => {
      const fixture = mkdtempSync(join(tmpdir(), "scopeweaver-landing-"));
      try {
        const landing = join(fixture, "landing");
        cpSync(join(root, "landing"), landing, { recursive: true });
        mkdirSync(join(fixture, "tools"));
        cpSync(join(root, "tools", "validate-landing.mjs"), join(fixture, "tools", "validate-landing.mjs"));
        // Every fixture starts valid independently of the production marker's presence.
        writeFileSync(join(landing, ".nojekyll"), "");
        mkdirSync(join(fixture, "landing-other"));
        writeFileSync(join(fixture, "landing-other", "asset.svg"), "<svg/>");
        if (mutate) {
          const file = join(landing, "index.html");
          const original = readFileSync(file, "utf8");
          const changed = mutate(original);
          assert.notEqual(changed, original, "fixture mutation must match the current page");
          writeFileSync(file, changed);
        }
        mutateFiles?.(landing);
        const result = spawnSync(process.execPath, [join(fixture, "tools", "validate-landing.mjs")], { encoding: "utf8" });
        assert.ifError(result.error);
        const output = result.stdout + result.stderr;
        assert.equal(result.status, failure ? 1 : 0, output);
        if (failure) assert.match(result.stderr, failure, output);
        assert.doesNotMatch(output, /TypeError|ENOENT|ReferenceError/, output);
      } finally {
        rmSync(fixture, { recursive: true, force: true });
      }
    });
  }
});
