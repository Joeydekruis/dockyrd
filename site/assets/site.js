(function () {
  var cfg = window.DOCKYRD_SITE || {};
  var REPO = cfg.repo || "Joeydekruis/dockyrd";
  var REPO_URL = cfg.repoUrl || "https://github.com/" + REPO;

  document.querySelectorAll("[data-gh]").forEach(function (a) {
    a.href = a.dataset.gh === "releases" ? REPO_URL + "/releases" : REPO_URL;
    a.target = "_blank";
    a.rel = "noopener";
  });

  if (!window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
    document.querySelectorAll(".dockyrd-logo").forEach(function (logo) {
      logo.addEventListener("pointerenter", function () {
        logo.classList.add("dockyrd-logo--hover");
      });
      logo.addEventListener("pointerleave", function () {
        logo.classList.remove("dockyrd-logo--hover");
      });
    });
  }

  function fmt(n) {
    return n >= 1000
      ? (n / 1000).toFixed(n >= 10000 ? 0 : 1).replace(/\.0$/, "") + "k"
      : String(n);
  }
  function showStars(n) {
    document.querySelectorAll("[data-stars]").forEach(function (el) {
      el.textContent = fmt(n);
      el.setAttribute("aria-label", n + " stars");
    });
  }
  try {
    var c = JSON.parse(localStorage.getItem("dy-stars") || "null");
    if (c && typeof c.n === "number") showStars(c.n);
    if (!c || Date.now() - c.t > 3600000) throw 0;
  } catch (_) {
    fetch("https://api.github.com/repos/" + REPO, {
      headers: { Accept: "application/vnd.github+json" },
    })
      .then(function (r) {
        if (!r.ok) throw r.status;
        return r.json();
      })
      .then(function (d) {
        if (typeof d.stargazers_count !== "number") return;
        showStars(d.stargazers_count);
        try {
          localStorage.setItem(
            "dy-stars",
            JSON.stringify({ n: d.stargazers_count, t: Date.now() }),
          );
        } catch (_) {}
      })
      .catch(function () {});
  }

  function copy(text, btn) {
    var label = btn.querySelector("span");
    var done = function () {
      if (label) {
        var o = label.textContent;
        label.textContent = "Copied";
        setTimeout(function () {
          label.textContent = o;
        }, 1400);
      }
    };
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(done, function () {
        selectNear(btn);
      });
    } else selectNear(btn);
  }
  function selectNear(btn) {
    var el = btn.closest(".codeblock, .install").querySelector("pre, code");
    if (!el) return;
    var r = document.createRange();
    r.selectNodeContents(el);
    var s = getSelection();
    s.removeAllRanges();
    s.addRange(r);
  }
  document.addEventListener("click", function (e) {
    var b = e.target.closest("[data-copy]");
    if (!b) return;
    var block = b.closest(".codeblock");
    copy(
      b.dataset.copy || (block && block.querySelector("pre").innerText) || "",
      b,
    );
  });

  document.querySelectorAll(".doc > pre").forEach(function (pre) {
    if (pre.closest(".codeblock")) return;
    var box = document.createElement("div");
    box.className = "codeblock";
    var head = document.createElement("div");
    head.className = "cb-head";
    var lang = document.createElement("span");
    lang.textContent = pre.dataset.lang || "code";
    var btn = document.createElement("button");
    btn.type = "button";
    btn.className = "copy";
    btn.setAttribute("data-copy", "");
    btn.innerHTML =
      '<svg width="13" height="13" aria-hidden="true"><use href="#i-copy"/></svg><span>Copy</span>';
    head.appendChild(lang);
    head.appendChild(btn);
    pre.parentNode.insertBefore(box, pre);
    box.appendChild(head);
    box.appendChild(pre);
  });

  document.querySelectorAll(".doc h2[id]").forEach(function (h) {
    if (h.querySelector(".h-anchor")) return;
    var a = document.createElement("a");
    a.className = "h-anchor";
    a.href = "#" + h.id;
    a.textContent = "#";
    a.setAttribute("aria-label", "Link to " + h.textContent);
    h.appendChild(a);
  });

  var menuBtn = document.getElementById("menuBtn");
  var navBackdrop = document.getElementById("navBackdrop");
  var mobileNav = document.getElementById("mobileNav");
  var sidebar = document.getElementById("sidebar");

  function navDrawer() {
    return sidebar || mobileNav;
  }

  function closeNav() {
    document.body.classList.remove("nav-open");
    if (menuBtn) menuBtn.setAttribute("aria-expanded", "false");
  }

  if (menuBtn) {
    var drawer = navDrawer();
    if (drawer) menuBtn.setAttribute("aria-controls", drawer.id);
    menuBtn.addEventListener("click", function () {
      var open = !document.body.classList.contains("nav-open");
      document.body.classList.toggle("nav-open", open);
      menuBtn.setAttribute("aria-expanded", open ? "true" : "false");
    });
  }
  if (navBackdrop) navBackdrop.addEventListener("click", closeNav);
  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape" && document.body.classList.contains("nav-open")) {
      closeNav();
    }
  });
  document.querySelectorAll("#sidebar a, #mobileNav a").forEach(function (a) {
    a.addEventListener("click", closeNav);
  });

  var dlg = document.getElementById("searchDlg");
  var input = document.getElementById("searchInput");
  var results = document.getElementById("results");
  var searchBtn = document.getElementById("searchBtn");
  if (!dlg || !input || !results || !searchBtn) return;

  var index = Array.isArray(cfg.searchIndex) ? cfg.searchIndex.slice() : [];
  var sel = 0;

  function render() {
    var q = input.value.trim().toLowerCase();
    var list = index
      .filter(function (it) {
        if (!q) return true;
        return (it.title + " " + (it.text || "")).toLowerCase().indexOf(q) > -1;
      })
      .slice(0, 12);
    results.innerHTML = "";
    if (!list.length) {
      var li = document.createElement("li");
      li.className = "empty";
      li.textContent = 'No results for "' + input.value + '"';
      results.appendChild(li);
      return;
    }
    sel = Math.min(sel, list.length - 1);
    list.forEach(function (it, i) {
      var li = document.createElement("li");
      li.setAttribute("role", "option");
      var a = document.createElement("a");
      a.href = resolveSearchHref(it.href);
      if (i === sel) a.className = "sel";
      a.innerHTML =
        '<span class="sec"></span><span class="t"></span><span class="x"></span>';
      a.children[0].textContent = it.sec || "";
      a.children[1].textContent = it.title || "";
      a.children[2].textContent = (it.text || "").trim();
      a.addEventListener("click", function () {
        dlg.close();
      });
      li.appendChild(a);
      results.appendChild(li);
    });
  }

  function openSearch() {
    input.value = "";
    sel = 0;
    render();
    if (dlg.showModal) dlg.showModal();
    else dlg.setAttribute("open", "");
    input.focus();
  }

  searchBtn.addEventListener("click", openSearch);
  function resolveSearchHref(href) {
    if (!href || href.indexOf("://") > -1 || href.charAt(0) === "/") return href;
    var path = location.pathname || "";
    var inDocs =
      path.indexOf("/docs/") !== -1 || /\/docs\/?$/.test(path);
    if (inDocs && href.indexOf("docs/") === 0) return href.slice(5);
    return href;
  }
  input.addEventListener("input", function () {
    sel = 0;
    render();
  });
  input.addEventListener("keydown", function (e) {
    var items = results.querySelectorAll("a");
    if (e.key === "ArrowDown") {
      e.preventDefault();
      sel = Math.min(sel + 1, items.length - 1);
      render();
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      sel = Math.max(sel - 1, 0);
      render();
    } else if (e.key === "Enter" && items[sel]) {
      e.preventDefault();
      location.href = resolveSearchHref(items[sel].getAttribute("href"));
      dlg.close();
    }
  });
  dlg.addEventListener("click", function (e) {
    if (e.target === dlg) dlg.close();
  });
  document.addEventListener("keydown", function (e) {
    if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
      e.preventDefault();
      openSearch();
    } else if (
      e.key === "/" &&
      !/input|textarea/i.test(document.activeElement.tagName) &&
      !dlg.open
    ) {
      e.preventDefault();
      openSearch();
    }
  });
})();
