/**
 * scopeweaver-cli Static Landing Page Interaction Script
 * - Bilingual language toggle (EN / KO) with document title & ARIA updates
 * - Shell tab switching (POSIX sh vs. PowerShell) with keyboard arrow support
 * - Clipboard copy with textarea fallback and screen-reader feedback
 * - Zero external dependencies, zero tracking
 */
(function () {
  "use strict";

  var htmlEl = document.documentElement;
  var langButtons = document.querySelectorAll("[data-set-lang]");
  var shellTabs = document.querySelectorAll("[data-shell-tab]");
  var shellPanels = document.querySelectorAll("[data-shell-panel]");
  var copyButtons = document.querySelectorAll("[data-copy-target]");
  var announcer = document.getElementById("sr-announcer");

  var TITLES = {
    en: "scopeweaver-cli — Standalone Agent-Native CLI Edition of ScopeWeaver",
    ko: "scopeweaver-cli — 에이전트 친화적 독립 실행형 ScopeWeaver CLI"
  };

  var COPY_LABELS = {
    en: {
      copied: "Copied",
      failed: "Copy manually",
      failedAnnounce: "Clipboard unavailable. Select and copy the command below.",
      announce: "Commands copied to clipboard."
    },
    ko: {
      copied: "복사됨",
      failed: "직접 복사",
      failedAnnounce: "클립보드를 사용할 수 없습니다. 아래 명령을 선택해 직접 복사하세요.",
      announce: "클립보드에 명령이 복사되었습니다."
    }
  };

  var activeLang = "en";
  var activeShell = "posix";

  function setLanguage(lang) {
    if (lang !== "en" && lang !== "ko") {
      lang = "en";
    }
    activeLang = lang;
    htmlEl.setAttribute("lang", lang);
    htmlEl.setAttribute("data-lang", lang);
    document.title = TITLES[lang] || TITLES.en;

    for (var i = 0; i < langButtons.length; i++) {
      var btn = langButtons[i];
      var isMatch = btn.getAttribute("data-set-lang") === lang;
      btn.classList.toggle("is-active", isMatch);
      btn.setAttribute("aria-pressed", isMatch ? "true" : "false");
    }

    try {
      window.localStorage.setItem("scopeweaver_cli_lang", lang);
    } catch (_) {
      /* Ignore storage restrictions in sandboxed/fileviews */
    }
  }

  function setShellTab(shellKey, focusTab) {
    if (shellKey !== "posix" && shellKey !== "pwsh") {
      shellKey = "posix";
    }
    activeShell = shellKey;

    for (var i = 0; i < shellTabs.length; i++) {
      var tab = shellTabs[i];
      var isSelected = tab.getAttribute("data-shell-tab") === shellKey;
      tab.classList.toggle("is-active", isSelected);
      tab.setAttribute("aria-selected", isSelected ? "true" : "false");
      tab.setAttribute("tabindex", isSelected ? "0" : "-1");
      if (isSelected && focusTab) {
        tab.focus();
      }
    }

    for (var j = 0; j < shellPanels.length; j++) {
      var panel = shellPanels[j];
      var isMatch = panel.getAttribute("data-shell-panel") === shellKey;
      if (isMatch) {
        panel.removeAttribute("hidden");
      } else {
        panel.setAttribute("hidden", "");
      }
    }
  }

  function resolveCopyElement(targetId) {
    if (targetId === "active-hero-code") {
      targetId = activeShell === "pwsh" ? "hero-code-pwsh" : "hero-code-posix";
    } else if (targetId === "active-task-json") {
      targetId = activeLang === "ko" ? "task-json-ko" : "task-json-en";
    }
    return document.getElementById(targetId);
  }

  function resolveCopyText(targetId) {
    var element = resolveCopyElement(targetId);
    return element ? element.textContent.trim() : "";
  }

  function selectManualCopy(targetId) {
    var element = resolveCopyElement(targetId);
    if (!element) return;
    element.setAttribute("tabindex", "-1");
    element.focus({ preventScroll: true });
    var selection = window.getSelection();
    if (selection) {
      var range = document.createRange();
      range.selectNodeContents(element);
      selection.removeAllRanges();
      selection.addRange(range);
    }
  }

  function fallbackCopyText(text) {
    var textarea = document.createElement("textarea");
    textarea.value = text;
    textarea.setAttribute("readonly", "");
    textarea.style.position = "fixed";
    textarea.style.top = "-9999px";
    textarea.style.left = "-9999px";
    document.body.appendChild(textarea);
    textarea.focus();
    textarea.select();

    var ok = false;
    try {
      ok = document.execCommand("copy");
    } catch (_) {
      ok = false;
    }
    document.body.removeChild(textarea);
    return ok;
  }

  function copyTextToClipboard(text, callback) {
    if (navigator.clipboard && typeof navigator.clipboard.writeText === "function") {
      navigator.clipboard.writeText(text).then(
        function () {
          callback(true);
        },
        function () {
          callback(fallbackCopyText(text));
        }
      );
    } else {
      callback(fallbackCopyText(text));
    }
  }

  function showCopyFeedback(button, success) {
    var labels = button.querySelectorAll(".copy-btn__label");
    for (var i = 0; i < labels.length; i++) {
      var lbl = labels[i];
      if (!lbl.hasAttribute("data-orig-text")) {
        lbl.setAttribute("data-orig-text", lbl.textContent);
      }
      var langKey = lbl.getAttribute("data-i18n") === "ko" ? "ko" : "en";
      lbl.textContent = success ? COPY_LABELS[langKey].copied : COPY_LABELS[langKey].failed;
    }

    button.classList.toggle("is-copied", success);
    if (announcer) {
      announcer.textContent = success ? COPY_LABELS[activeLang].announce : COPY_LABELS[activeLang].failedAnnounce;
    }

    if (button._copyTimer) {
      window.clearTimeout(button._copyTimer);
    }
    button._copyTimer = window.setTimeout(function () {
      button.classList.remove("is-copied");
      for (var j = 0; j < labels.length; j++) {
        var orig = labels[j].getAttribute("data-orig-text");
        if (orig) {
          labels[j].textContent = orig;
        }
      }
    }, 1800);
  }

  /* Bind Language Switch Buttons */
  for (var i = 0; i < langButtons.length; i++) {
    langButtons[i].addEventListener("click", function (e) {
      var lang = e.currentTarget.getAttribute("data-set-lang");
      setLanguage(lang);
      if (lang === "ko" && activeShell === "posix") {
        /* Keep user shell choice independent, or user can switch manually */
      }
    });
  }

  /* Bind Shell Tab Buttons + Arrow Key Navigation */
  for (var k = 0; k < shellTabs.length; k++) {
    shellTabs[k].addEventListener("click", function (e) {
      var shellKey = e.currentTarget.getAttribute("data-shell-tab");
      setShellTab(shellKey, false);
    });

    shellTabs[k].addEventListener("keydown", function (e) {
      if (e.key === "ArrowRight" || e.key === "ArrowLeft") {
        e.preventDefault();
        var nextShell = activeShell === "posix" ? "pwsh" : "posix";
        setShellTab(nextShell, true);
      }
    });
  }

  /* Bind Copy Buttons */
  for (var m = 0; m < copyButtons.length; m++) {
    copyButtons[m].addEventListener("click", function (e) {
      var btn = e.currentTarget;
      var targetId = btn.getAttribute("data-copy-target");
      var text = resolveCopyText(targetId);
      if (!text) {
        return;
      }
      copyTextToClipboard(text, function (ok) {
        if (!ok) selectManualCopy(targetId);
        showCopyFeedback(btn, ok);
      });
    });
  }

  /* Initialize saved or URL-specified language */
  (function init() {
    var initialLang = "en";
    try {
      var params = new URLSearchParams(window.location.search);
      var urlLang = params.get("lang");
      var storedLang = window.localStorage.getItem("scopeweaver_cli_lang");
      if (urlLang === "ko" || urlLang === "en") {
        initialLang = urlLang;
      } else if (storedLang === "ko" || storedLang === "en") {
        initialLang = storedLang;
      }
    } catch (_) {
      initialLang = "en";
    }
    setLanguage(initialLang);
    setShellTab("posix", false);
  })();
})();
