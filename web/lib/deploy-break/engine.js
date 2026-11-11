// @ts-nocheck — ported breakout engine
export function mountDeployBreak(canvas, root, options) {
  const W = 640,
    H = 480;
  const cv = canvas;
  const ctx = cv.getContext("2d");
  if (!ctx) throw new Error("no ctx");
  const dpr = Math.min(2, window.devicePixelRatio || 1);
  cv.width = W * dpr;
  cv.height = H * dpr;
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0);

  const C = {
    ink: "#0B1F3A",
    brand: "#8EBEFF",
    mark: "#3B93F5",
    soft: "#E3F0FF",
    border: "#ECEAE6",
    strong: "#CBD0D8",
    ok: "#1A7F4B",
    warn: "#D98A00",
    err: "#B3261E",
    muted: "#5A6577",
  };
  const MONO = '"Geist Mono",ui-monospace,Menlo,monospace';

  // ---------- DOM ----------
  const $ = (id) => root.querySelector('[data-dy-id="' + id + '"]');
  const elScore = $("score"),
    elLives = $("lives"),
    elLevel = $("level"),
    elBest = $("best");
  const elPct = $("pct"),
    elBar = $("bar"),
    elFill = $("barfill"),
    elStage = $("stage"),
    elBadge = $("badge"),
    elBadgeText = $("badgeText");
  const overlay = $("overlay"),
    ovTitle = $("ovTitle"),
    ovText = $("ovText"),
    ovBtn = $("ovBtn");
  const logEl = $("log"),
    muteBtn = $("muteBtn"),
    pauseBtn = $("pauseBtn");

  // ---------- Audio ----------
  let ac = null,
    muted = false;
  function beep(f, d, type, v, slide) {
    if (muted) return;
    try {
      if (!ac) ac = new (window.AudioContext || window.webkitAudioContext)();
      if (ac.state === "suspended") ac.resume();
      const o = ac.createOscillator(),
        g = ac.createGain(),
        t = ac.currentTime;
      o.type = type || "square";
      o.frequency.setValueAtTime(f, t);
      if (slide)
        o.frequency.exponentialRampToValueAtTime(
          Math.max(30, f * slide),
          t + d,
        );
      g.gain.setValueAtTime(v || 0.035, t);
      g.gain.exponentialRampToValueAtTime(0.0001, t + d);
      o.connect(g);
      g.connect(ac.destination);
      o.start(t);
      o.stop(t + d);
    } catch (e) {}
  }
  const sfx = {
    paddle: () => beep(220, 0.07, "square", 0.04),
    brick: () => beep(520 + Math.random() * 120, 0.06, "square", 0.03),
    kill: () => {
      beep(660, 0.08, "square", 0.035);
      setTimeout(() => beep(880, 0.1, "square", 0.03), 50);
    },
    good: () => {
      beep(520, 0.08, "triangle", 0.05);
      setTimeout(() => beep(780, 0.14, "triangle", 0.05), 70);
    },
    bad: () => beep(200, 0.25, "sawtooth", 0.04, 0.5),
    lose: () => beep(300, 0.5, "sawtooth", 0.05, 0.25),
    clear: () => {
      for (const [i, f] of [523, 659, 784, 1046].entries()) {
        setTimeout(() => beep(f, 0.14, "square", 0.04), i * 90);
      }
    },
    live: () => {
      for (const [i, f] of [392, 523, 659, 784, 1046].entries()) {
        setTimeout(() => beep(f, 0.12, "triangle", 0.05), i * 80);
      }
    },
  };

  // ---------- Storage (optional) ----------
  let best = options.initialBest || 0;
  function saveBest() {
    if (options.onBestChange) options.onBestChange(best);
  }

  // ---------- Log ----------
  const logLines = [];
  function log(msg, cls) {
    const d = new Date(),
      ts = d.toTimeString().slice(0, 8);
    logLines.push(
      '<div class="' +
        (cls || "") +
        '"><span class="dim">' +
        ts +
        "</span> " +
        msg +
        "</div>",
    );
    while (logLines.length > 5) logLines.shift();
    logEl.innerHTML = logLines.join("");
  }
  const BRICK_MSGS = [
    "pulled node:22-alpine",
    "layer cached",
    "npm install · 1,204 packages",
    "tests passed (3 skipped)",
    "env vars injected",
    "TLS certificate issued",
    "health check OK",
    "container started",
    "migrations applied",
    "static assets uploaded",
    "image pushed to registry",
    "dependency audit: 0 vulnerabilities",
    "build cache hit",
    "webhook verified (HMAC)",
    "Caddy reloaded",
    "database linked to app",
  ];
  const KILL_MSGS = [
    "node_modules is 1.4 GB. Nobody is surprised.",
    "works on my machine",
    "1 flaky test retried, passed",
    "TODO removed. Just kidding",
    "lockfile conflict resolved",
    "yarn.lock vs package-lock: fight!",
    "curl localhost:3000 → 200",
    "sudo make me a sandwich",
  ];
  const pick = (a) => a[Math.floor(Math.random() * a.length)];

  // ---------- State ----------
  let state = "title"; // title | ready | play | paused | clear | over
  let score = 0,
    lives = 3,
    level = 1,
    deploy = 0,
    deployDone = false;
  const paddle = {
    x: W / 2 - 48,
    w: 96,
    h: 14,
    y: H - 36,
    target: W / 2,
    tw: 96,
  };
  let balls = [],
    bricks = [],
    items = [],
    parts = [];
  let eff = { wide: 0, narrow: 0, slow: 0, fast: 0 };
  let shake = 0,
    clearTimer = 0,
    keys = { left: false, right: false },
    flash = 0;

  const COLS = 10,
    ROWS = 5,
    BW = 56,
    BH = 22,
    GAP = 4,
    OX = (W - (COLS * BW + (COLS - 1) * GAP)) / 2,
    OY = 54;
  const HPCOL = { 1: C.brand, 2: C.mark, 3: C.ink };

  function layout(l) {
    const p = (l - 1) % 4,
      boost = Math.min(2, Math.floor((l - 1) / 4));
    const out = [];
    for (let r = 0; r < ROWS; r++)
      for (let c = 0; c < COLS; c++) {
        let on = true;
        if (p === 1) on = (r + c) % 2 === 0;
        else if (p === 2) on = c >= r && c <= COLS - 1 - r;
        else if (p === 3) on = c % 3 !== 2 || r === 0;
        if (!on) continue;
        let hp = 1 + (r < 2 ? 1 : 0) + (boost && r < 3 ? 1 : 0);
        if (l === 1 && r >= 2) hp = 1;
        hp = Math.min(3, hp);
        out.push({
          x: OX + c * (BW + GAP),
          y: OY + r * (BH + GAP),
          w: BW,
          h: BH,
          hp,
          max: hp,
          hit: 0,
        });
      }
    return out;
  }
  function baseSpeed() {
    return Math.min(8.2, 4.6 + (level - 1) * 0.35);
  }
  function speedMul() {
    return (eff.slow > 0 ? 0.72 : 1) * (eff.fast > 0 ? 1.35 : 1);
  }

  function newBall(stuck) {
    return {
      x: paddle.x + paddle.w / 2,
      y: paddle.y - 8,
      vx: 0,
      vy: 0,
      stuck: !!stuck,
      spin: 0,
    };
  }
  function resetBall() {
    balls = [newBall(true)];
    items = [];
  }

  function setupLevel() {
    bricks = layout(level);
    resetBall();
    eff = { wide: 0, narrow: 0, slow: 0, fast: 0 };
    paddle.tw = 96;
    elLevel.textContent = "v0.1." + level;
  }
  function startGame() {
    score = 0;
    lives = 3;
    level = 1;
    deploy = deployDone ? deploy : deploy;
    setupLevel();
    updateHud();
    state = "ready";
    overlay.classList.add("hidden");
    log("Dockyrd Deploy Break started. Good luck.", "good");
  }

  // ---------- HUD ----------
  const cubeSvg =
    '<svg width="22" height="22" viewBox="0 0 28 28"><rect x="2" y="2" width="24" height="24" rx="7" fill="#0B1F3A"/><path d="M5 18c2.3-2.6 4.6-2.6 6.9 0s4.6 2.6 6.9 0c1.4-1.6 2.8-2.2 4.2-2" fill="none" stroke="#8EBEFF" stroke-width="2.4" stroke-linecap="round"/></svg>';
  const cubeOff =
    '<svg width="22" height="22" viewBox="0 0 28 28"><rect x="3" y="3" width="22" height="22" rx="6.5" fill="none" stroke="#CBD0D8" stroke-width="2" stroke-dasharray="3 3"/></svg>';
  function updateHud() {
    elScore.textContent = score.toLocaleString("en");
    let h = "";
    for (let i = 0; i < 3; i++) h += i < lives ? cubeSvg : cubeOff;
    if (lives > 3)
      h +=
        '<span style="font:600 13px ' + MONO + '">+' + (lives - 3) + "</span>";
    elLives.innerHTML = h;
    if (score > best) {
      best = score;
      saveBest();
    }
    if (options.onScoreChange) options.onScoreChange(score);
    elBest.textContent = best.toLocaleString("en");
  }
  const STAGES = [
    [20, "Pulling image…"],
    [45, "Building with Railpack…"],
    [70, "Running migrations…"],
    [90, "Health checks…"],
    [100, "Switching traffic…"],
  ];
  function updateDeploy() {
    const p = Math.min(100, deploy);
    elFill.style.width = p + "%";
    elPct.textContent = Math.floor(p) + "%";
    if (p >= 100) {
      if (!deployDone) {
        deployDone = true;
        sfx.live();
        log(
          "✓ Deploy finished. web-api is live. Keep playing if you like.",
          "good",
        );
      }
      elBar.classList.add("live");
      elBadge.classList.add("live");
      elBadgeText.textContent = "Live";
      elStage.textContent = "web-api is live. This game is now just for fun.";
    } else {
      for (const s of STAGES)
        if (p < s[0]) {
          elStage.textContent = s[1];
          break;
        }
    }
  }

  // ---------- Helpers ----------
  function rr(x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.arcTo(x + w, y, x + w, y + h, r);
    ctx.arcTo(x + w, y + h, x, y + h, r);
    ctx.arcTo(x, y + h, x, y, r);
    ctx.arcTo(x, y, x + w, y, r);
    ctx.closePath();
  }
  function burst(x, y, col, n) {
    for (let i = 0; i < n; i++) {
      const a = Math.random() * 6.28,
        s = 1 + Math.random() * 3;
      parts.push({
        x,
        y,
        vx: Math.cos(a) * s,
        vy: Math.sin(a) * s - 1,
        life: 26 + Math.random() * 14,
        col,
        sz: 3 + Math.random() * 3,
      });
    }
  }
  const ITEMS = {
    replicas: { label: "REPLICAS", good: true, col: C.ok },
    scale: { label: "SCALE UP", good: true, col: C.mark },
    cache: { label: "CACHE HIT", good: true, col: C.mark },
    rollback: { label: "ROLLBACK", good: true, col: C.ok },
    oom: { label: "OOM KILLED", good: false, col: C.err },
    cpu: { label: "CPU SPIKE", good: false, col: C.warn },
  };
  const ITEM_POOL = [
    "replicas",
    "replicas",
    "scale",
    "scale",
    "cache",
    "cache",
    "rollback",
    "oom",
    "oom",
    "cpu",
    "cpu",
  ];
  function dropItem(x, y) {
    let t = pick(ITEM_POOL);
    if (t === "rollback" && lives >= 5) t = "scale";
    items.push({ x, y, type: t, vy: 1.9 });
  }
  function applyItem(t) {
    const it = ITEMS[t];
    if (t === "replicas") {
      const src = balls.find((b) => !b.stuck) || balls[0];
      if (src) {
        const sp = Math.hypot(src.vx, src.vy) || baseSpeed();
        for (const a of [-0.5, 0.5]) {
          const ang = Math.atan2(src.vy, src.vx) + a;
          balls.push({
            x: src.x,
            y: src.y,
            vx: Math.cos(ang) * sp,
            vy: Math.sin(ang) * sp,
            stuck: false,
            spin: 0,
          });
        }
      }
      log("replicas: 3 → ball ×3 scheduled", "good");
      sfx.good();
    } else if (t === "scale") {
      eff.wide = 600;
      eff.narrow = 0;
      log("scaled up dy-proxy (wider paddle)", "good");
      sfx.good();
    } else if (t === "cache") {
      eff.slow = 480;
      eff.fast = 0;
      log("cache hit! everything is slower", "good");
      sfx.good();
    } else if (t === "rollback") {
      lives++;
      log("rolled back to last good release (+1 uptime)", "good");
      sfx.live();
      updateHud();
    } else if (t === "oom") {
      eff.narrow = 600;
      eff.wide = 0;
      log("✗ OOM killed (exit 137). paddle shrank", "bad");
      sfx.bad();
    } else if (t === "cpu") {
      eff.fast = 480;
      eff.slow = 0;
      log("✗ CPU spike: ball is going fast", "bad");
      sfx.bad();
    }
  }

  // ---------- Launch / input ----------
  function launch() {
    if (state !== "ready") return;
    const b = balls[0];
    if (!b) return;
    const s = baseSpeed() * speedMul();
    const a = -Math.PI / 2 + (Math.random() * 0.6 - 0.3);
    b.vx = Math.cos(a) * s;
    b.vy = Math.sin(a) * s;
    b.stuck = false;
    state = "play";
    log("deploying release v0.1." + level + "…", "dim");
  }
  function togglePause() {
    if (state === "play" || (state === "ready" && false)) {
      state = "paused";
      showOverlay(
        "Paused",
        "Your deploy keeps running in the background. Promise.",
        "Resume",
        () => {
          hideOverlay();
          state = "play";
        },
      );
    } else if (state === "paused") {
      hideOverlay();
      state = "play";
    }
    pauseBtn.textContent = state === "paused" ? "Resume" : "Pause";
  }
  function showOverlay(title, text, btn, fn) {
    ovTitle.textContent = title;
    ovText.textContent = text;
    ovBtn.textContent = btn;
    ovBtn.onclick = fn;
    overlay.classList.remove("hidden");
    ovBtn.focus({ preventScroll: true });
  }
  function hideOverlay() {
    overlay.classList.add("hidden");
    pauseBtn.textContent = "Pause";
  }
  function setTargetFromEvent(e) {
    const r = cv.getBoundingClientRect();
    paddle.target = ((e.clientX - r.left) / r.width) * W;
  }
  canvas.addEventListener("pointermove", setTargetFromEvent);
  canvas.addEventListener("pointerdown", (e) => {
    setTargetFromEvent(e);
    launch();
  });
  window.addEventListener("keydown", (e) => {
    const k = e.key;
    if (k === "ArrowLeft" || k === "a" || k === "A") {
      keys.left = true;
      e.preventDefault();
    } else if (k === "ArrowRight" || k === "d" || k === "D") {
      keys.right = true;
      e.preventDefault();
    } else if (k === " ") {
      if (state === "title" || state === "over") {
        ovBtn.click();
      } else launch();
      e.preventDefault();
    } else if (k === "p" || k === "P" || k === "Escape") {
      togglePause();
    } else if (k === "m" || k === "M") {
      muteBtn.click();
    }
  });
  window.addEventListener("keyup", (e) => {
    const k = e.key;
    if (k === "ArrowLeft" || k === "a" || k === "A") keys.left = false;
    if (k === "ArrowRight" || k === "d" || k === "D") keys.right = false;
  });
  muteBtn.onclick = () => {
    muted = !muted;
    muteBtn.textContent = "Sound: " + (muted ? "off" : "on");
    muteBtn.setAttribute("aria-pressed", String(muted));
  };
  pauseBtn.onclick = togglePause;
  document.addEventListener("visibilitychange", () => {
    if (document.hidden && state === "play") togglePause();
  });

  // ---------- Update ----------
  function stepBall(b) {
    const r = 7,
      sub = 2;
    for (let s = 0; s < sub; s++) {
      b.x += b.vx / sub;
      b.y += b.vy / sub;
      if (b.x < r) {
        b.x = r;
        b.vx = Math.abs(b.vx);
        beep(300, 0.03, "square", 0.02);
      }
      if (b.x > W - r) {
        b.x = W - r;
        b.vx = -Math.abs(b.vx);
        beep(300, 0.03, "square", 0.02);
      }
      if (b.y < r) {
        b.y = r;
        b.vy = Math.abs(b.vy);
        beep(300, 0.03, "square", 0.02);
      }
      // paddle
      if (
        b.vy > 0 &&
        b.y + r >= paddle.y &&
        b.y + r <= paddle.y + paddle.h + 10 &&
        b.x >= paddle.x - r &&
        b.x <= paddle.x + paddle.w + r
      ) {
        const rel = (b.x - (paddle.x + paddle.w / 2)) / (paddle.w / 2);
        const ang = Math.max(-1, Math.min(1, rel)) * 1.0; // up to ~57deg
        const sp = baseSpeed() * speedMul();
        b.vx = Math.sin(ang) * sp;
        b.vy = -Math.cos(ang) * sp;
        b.y = paddle.y - r;
        sfx.paddle();
        paddle.squish = 6;
      }
      // bricks
      for (let i = 0; i < bricks.length; i++) {
        const k = bricks[i];
        const cx = Math.max(k.x, Math.min(b.x, k.x + k.w)),
          cy = Math.max(k.y, Math.min(b.y, k.y + k.h));
        const dx = b.x - cx,
          dy = b.y - cy;
        if (dx * dx + dy * dy < r * r) {
          if (dx === 0 && dy === 0) {
            b.vy = -b.vy;
          } else if (Math.abs(dx) > Math.abs(dy)) {
            b.vx = dx > 0 ? Math.abs(b.vx) : -Math.abs(b.vx);
            b.x = cx + (dx > 0 ? r : -r);
          } else {
            b.vy = dy > 0 ? Math.abs(b.vy) : -Math.abs(b.vy);
            b.y = cy + (dy > 0 ? r : -r);
          }
          k.hp--;
          k.hit = 6;
          if (k.hp <= 0) {
            bricks.splice(i, 1);
            score += 10 * k.max;
            deploy += 0.25;
            burst(k.x + k.w / 2, k.y + k.h / 2, HPCOL[k.max], 10);
            sfx.kill();
            log(
              Math.random() < 0.2 ? pick(KILL_MSGS) : "✓ " + pick(BRICK_MSGS),
              Math.random() < 0.2 ? "" : "good",
            );
            if (Math.random() < 0.14) dropItem(k.x + k.w / 2, k.y + k.h / 2);
          } else {
            score += 5;
            sfx.brick();
            burst(b.x, b.y, C.strong, 3);
          }
          updateHud();
          break;
        }
      }
    }
    b.spin += 0.12 * Math.sign(b.vx || 1);
  }
  function update() {
    // paddle movement
    const kspeed = 8;
    if (keys.left) paddle.target = paddle.x + paddle.w / 2 - kspeed * 2;
    if (keys.right) paddle.target = paddle.x + paddle.w / 2 + kspeed * 2;
    const wantW = 96 * (eff.wide > 0 ? 1.5 : 1) * (eff.narrow > 0 ? 0.62 : 1);
    paddle.w += (wantW - paddle.w) * 0.15;
    let tx = paddle.target - paddle.w / 2;
    tx = Math.max(0, Math.min(W - paddle.w, tx));
    paddle.x += (tx - paddle.x) * (keys.left || keys.right ? 1 : 0.5);
    if (paddle.squish > 0) paddle.squish--;
    for (const k in eff) if (eff[k] > 0) eff[k]--;

    for (const b of balls) {
      if (b.stuck) {
        b.x = paddle.x + paddle.w / 2;
        b.y = paddle.y - 8;
      }
    }
    if (state === "play") {
      deploy += 100 / (110 * 60);
      // keep speed in sync with effects
      const want = baseSpeed() * speedMul();
      for (const b of balls) {
        const sp = Math.hypot(b.vx, b.vy) || want;
        const ns = sp + (want - sp) * 0.05;
        b.vx *= ns / sp;
        b.vy *= ns / sp;
        // avoid near-horizontal stalls
        if (Math.abs(b.vy) < want * 0.25) {
          b.vy = (b.vy < 0 ? -1 : 1) * want * 0.25;
        }
        stepBall(b);
      }
      balls = balls.filter((b) => b.y - 7 <= H);
      if (balls.length === 0) {
        lives--;
        sfx.lose();
        shake = 14;
        flash = 10;
        updateHud();
        log(
          "✗ container crashed (exit 137). uptime " + Math.max(0, lives) + "/3",
          "bad",
        );
        if (lives <= 0) {
          state = "over";
          showOverlay(
            "Deploy failed",
            "Exit code 137. You scored " +
              score.toLocaleString("en") +
              " deploys" +
              (score >= best && score > 0 ? " (new best!)" : "") +
              ". Retries are free.",
            "Retry deploy",
            () => {
              startGame();
            },
          );
        } else {
          eff.wide = eff.narrow = eff.slow = eff.fast = 0;
          resetBall();
          state = "ready";
        }
      }
      if (bricks.length === 0) {
        state = "clear";
        clearTimer = 110;
        sfx.clear();
        score += 100 * level;
        deploy += 5;
        updateHud();
        log("✓ release v0.1." + level + " shipped to production", "good");
      }
    }
    if (state === "clear") {
      clearTimer--;
      if (clearTimer <= 0) {
        level++;
        setupLevel();
        state = "ready";
        updateHud();
      }
    }
    // items
    for (const it of items) {
      it.y += it.vy;
    }
    items = items.filter((it) => {
      if (it.y > H + 20) return false;
      if (
        it.y + 8 >= paddle.y &&
        it.y - 8 <= paddle.y + paddle.h &&
        it.x >= paddle.x - 30 &&
        it.x <= paddle.x + paddle.w + 30
      ) {
        applyItem(it.type);
        burst(it.x, it.y, ITEMS[it.type].col, 10);
        return false;
      }
      return true;
    });
    for (const p of parts) {
      p.x += p.vx;
      p.y += p.vy;
      p.vy += 0.12;
      p.life--;
    }
    parts = parts.filter((p) => p.life > 0);
    for (const k of bricks) if (k.hit > 0) k.hit--;
    if (shake > 0) shake--;
    if (flash > 0) flash--;
    updateDeploy();
  }

  // ---------- Draw ----------
  function drawBall(b) {
    ctx.save();
    ctx.translate(b.x, b.y);
    ctx.rotate(Math.sin(b.spin) * 0.35);
    rr(-8, -8, 16, 16, 4.5);
    ctx.fillStyle = C.ink;
    ctx.fill();
    ctx.beginPath();
    ctx.moveTo(-5.5, 2.2);
    ctx.bezierCurveTo(-3.5, -0.4, -1.5, -0.4, 0, 2.2);
    ctx.bezierCurveTo(1.5, 4.8, 3.5, 4.8, 5.5, 2.2);
    ctx.strokeStyle = C.brand;
    ctx.lineWidth = 2;
    ctx.lineCap = "round";
    ctx.stroke();
    ctx.restore();
  }
  function drawBrick(k) {
    const col = HPCOL[k.hp] || C.brand;
    const off = k.hit > 0 ? 1 : 0;
    rr(k.x, k.y + off, k.w, k.h, 5);
    ctx.fillStyle = col;
    ctx.fill();
    // retro bevel
    ctx.fillStyle = "rgba(255,255,255,.28)";
    rr(k.x + 2, k.y + off + 2, k.w - 4, 4, 2);
    ctx.fill();
    ctx.fillStyle = "rgba(11,31,58,.18)";
    rr(k.x + 2, k.y + off + k.h - 6, k.w - 4, 4, 2);
    ctx.fill();
    if (k.max > k.hp) {
      // cracks
      ctx.strokeStyle = "rgba(255,255,255,.7)";
      ctx.lineWidth = 1.5;
      ctx.beginPath();
      ctx.moveTo(k.x + k.w * 0.3, k.y + off + 2);
      ctx.lineTo(k.x + k.w * 0.42, k.y + off + k.h * 0.5);
      ctx.lineTo(k.x + k.w * 0.34, k.y + off + k.h - 2);
      ctx.stroke();
    }
  }
  function drawItem(it) {
    const d = ITEMS[it.type];
    ctx.font = "600 10px " + MONO;
    const tw = ctx.measureText(d.label).width + 14;
    rr(it.x - tw / 2, it.y - 9, tw, 18, 9);
    ctx.fillStyle = d.col;
    ctx.fill();
    ctx.fillStyle = "#fff";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(d.label, it.x, it.y + 0.5);
  }
  function draw() {
    ctx.save();
    if (shake > 0)
      ctx.translate(
        (Math.random() - 0.5) * shake * 0.6,
        (Math.random() - 0.5) * shake * 0.6,
      );
    ctx.clearRect(-20, -20, W + 40, H + 40);
    ctx.fillStyle = "#fff";
    ctx.fillRect(-20, -20, W + 40, H + 40);
    // dotted grid
    ctx.fillStyle = C.border;
    for (let x = 16; x < W; x += 32)
      for (let y = 16; y < H; y += 32) ctx.fillRect(x - 1, y - 1, 2, 2);
    // status header strip
    ctx.fillStyle = C.muted;
    ctx.font = "500 11px " + MONO;
    ctx.textAlign = "left";
    ctx.textBaseline = "middle";
    ctx.fillText("PIPELINE · release v0.1." + level, 16, 20);
    ctx.textAlign = "right";
    const efx = [];
    if (eff.wide > 0) efx.push("SCALE UP");
    if (eff.narrow > 0) efx.push("OOM");
    if (eff.slow > 0) efx.push("CACHE HIT");
    if (eff.fast > 0) efx.push("CPU SPIKE");
    ctx.fillText(efx.join(" · "), W - 16, 20);
    for (const k of bricks) drawBrick(k);
    for (const it of items) drawItem(it);
    // paddle
    const pw = paddle.w,
      ph = paddle.h,
      px = paddle.x,
      py = paddle.y + (paddle.squish > 0 ? 2 : 0);
    rr(px, py, pw, ph, 7);
    ctx.fillStyle = C.ink;
    ctx.fill();
    ctx.fillStyle = C.brand;
    rr(px + 4, py + 2, pw - 8, 3, 1.5);
    ctx.fill();
    if (pw > 70) {
      ctx.fillStyle = "#fff";
      ctx.font = "500 8px " + MONO;
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText("dy-proxy", px + pw / 2, py + ph / 2 + 2);
    }
    for (const b of balls) drawBall(b);
    for (const p of parts) {
      ctx.globalAlpha = Math.max(0, p.life / 40);
      ctx.fillStyle = p.col;
      ctx.fillRect(p.x, p.y, p.sz, p.sz);
    }
    ctx.globalAlpha = 1;
    if (state === "ready") {
      ctx.fillStyle = C.ink;
      ctx.font = "500 12px " + MONO;
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText("CLICK OR PRESS SPACE TO DEPLOY", W / 2, H / 2 + 60);
    }
    if (state === "clear") {
      ctx.fillStyle = "rgba(255,255,255,.85)";
      ctx.fillRect(0, H / 2 - 44, W, 88);
      ctx.fillStyle = C.ok;
      ctx.font = "600 26px " + MONO;
      ctx.textAlign = "center";
      ctx.textBaseline = "middle";
      ctx.fillText("✓ RELEASE v0.1." + level + " SHIPPED", W / 2, H / 2 - 8);
      ctx.fillStyle = C.muted;
      ctx.font = "400 12px " + MONO;
      ctx.fillText(
        "+" + 100 * level + " bonus · rolling out next release…",
        W / 2,
        H / 2 + 22,
      );
    }
    if (flash > 0) {
      ctx.fillStyle = "rgba(179,38,30," + flash / 40 + ")";
      ctx.fillRect(-20, -20, W + 40, H + 40);
    }
    ctx.restore();
  }

  // ---------- Loop (fixed 60Hz) ----------
  let last = performance.now(),
    acc = 0;
  function frame(now) {
    acc += Math.min(100, now - last);
    last = now;
    while (acc >= 1000 / 60) {
      if (state !== "paused" && state !== "title" && state !== "over") update();
      else if (state === "title" || state === "over") {
        for (const p of parts) {
          p.x += p.vx;
          p.y += p.vy;
          p.life--;
        }
      }
      acc -= 1000 / 60;
    }
    draw();

    function applyExternalDeploy() {
      if (typeof options.deployProgress === "number") {
        const ext = Math.max(deploy, options.deployProgress);
        deploy = ext;
        deployDone = !!options.deployDone || deploy >= 100;
        updateDeploy();
      }
      if (options.releaseLabel) elLevel.textContent = options.releaseLabel;
    }
    setInterval(applyExternalDeploy, 500);

    requestAnimationFrame(frame);
  }

  // ---------- Init ----------
  ovBtn.onclick = startGame;
  setupLevel();
  updateHud();
  updateDeploy();
  log("web-api: deployment queued", "dim");
  log("Dockyrd: waiting for build slot…", "dim");

  function applyExternalDeploy() {
    if (typeof options.deployProgress === "number") {
      const ext = Math.max(deploy, options.deployProgress);
      deploy = ext;
      deployDone = !!options.deployDone || deploy >= 100;
      updateDeploy();
    }
    if (options.releaseLabel) elLevel.textContent = options.releaseLabel;
  }
  setInterval(applyExternalDeploy, 500);

  requestAnimationFrame(frame);
  // demo hooks for testing
  window.__dy = {
    get state() {
      return state;
    },
    get balls() {
      return balls;
    },
    get bricks() {
      return bricks;
    },
    get lives() {
      return lives;
    },
    get score() {
      return score;
    },
    get deploy() {
      return deploy;
    },
    set deploy(v) {
      deploy = v;
    },
    paddle: paddle,
    launch: launch,
    clearBricks() {
      bricks.length = 0;
    },
  };
  return { getScore: () => score, getBest: () => best };
}
