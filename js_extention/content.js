// content.js — внедряется на lichess.org/*.
// Рисует плавающую кнопку, по клику забирает username оппонента + speed
// и показывает панель с результатом анализа.

(() => {
  if (window.__chessAnalyzerInjected) return;
  window.__chessAnalyzerInjected = true;

  const SPEEDS = ["bullet", "blitz", "rapid", "classical"];

  // ---------- извлечение данных со страницы ----------

  // Верхний игрок на доске = оппонент во время своей партии.
  function getOpponentUsername() {
    // Сначала пробуем верхнего игрока (.ruser-top), затем — любой профиль на странице.
    const selectors = [
      ".ruser-top a[href*='/@/']",
      ".game__meta .player a[href*='/@/']",
      "a.user-link[href*='/@/']",
    ];
    for (const sel of selectors) {
      const el = document.querySelector(sel);
      const name = usernameFromHref(el && el.getAttribute("href"));
      if (name) return name;
    }
    return null;
  }

  function usernameFromHref(href) {
    if (!href) return null;
    const m = href.match(/\/@\/([^/?#]+)/);
    return m ? decodeURIComponent(m[1]) : null;
  }

  // Best-effort определение speed. Надёжного DOM-якоря у lichess нет,
  // поэтому это лишь предзаполнение дропдауна — пользователь может поправить.
  function detectSpeed() {
    const text = (document.querySelector(".game__meta, .header, .round__meta") || document.body)
      .textContent.toLowerCase();
    for (const s of SPEEDS) {
      if (text.includes(s)) return s;
    }
    // По тайм-контролу вида "5+3" → суммарная оценка (initial + 40*inc).
    const tc = text.match(/(\d+)\s*\+\s*(\d+)/);
    if (tc) {
      const est = parseInt(tc[1], 10) * 60 + 40 * parseInt(tc[2], 10);
      if (est < 179) return "bullet";
      if (est < 479) return "blitz";
      if (est < 1499) return "rapid";
      return "classical";
    }
    return "blitz";
  }

  // ---------- UI ----------

  function createButton() {
    const btn = document.createElement("button");
    btn.id = "coa-button";
    btn.type = "button";
    btn.textContent = "♟ Анализ оппонента";
    btn.addEventListener("click", onAnalyzeClick);
    document.body.appendChild(btn);
  }

  function ensurePanel() {
    let panel = document.getElementById("coa-panel");
    if (panel) return panel;

    panel = document.createElement("div");
    panel.id = "coa-panel";
    panel.innerHTML = `
      <div class="coa-panel__head">
        <span class="coa-panel__title">Анализ оппонента</span>
        <button type="button" class="coa-panel__close" title="Закрыть">×</button>
      </div>
      <div class="coa-panel__controls">
        <label>Оппонент: <input type="text" class="coa-username" placeholder="username"></label>
        <label>Speed:
          <select class="coa-speed">
            ${SPEEDS.map((s) => `<option value="${s}">${s}</option>`).join("")}
          </select>
        </label>
        <button type="button" class="coa-run">Анализировать</button>
      </div>
      <div class="coa-panel__body"></div>
    `;
    document.body.appendChild(panel);

    panel.querySelector(".coa-panel__close").addEventListener("click", () => {
      panel.style.display = "none";
    });
    panel.querySelector(".coa-run").addEventListener("click", () => {
      const username = panel.querySelector(".coa-username").value.trim();
      const speed = panel.querySelector(".coa-speed").value;
      if (username) runAnalysis(username, speed);
    });
    return panel;
  }

  function setBody(html) {
    ensurePanel().querySelector(".coa-panel__body").innerHTML = html;
  }

  // ---------- обработчики ----------

  function onAnalyzeClick() {
    const panel = ensurePanel();
    panel.style.display = "block";

    const username = getOpponentUsername();
    if (username) panel.querySelector(".coa-username").value = username;
    panel.querySelector(".coa-speed").value = detectSpeed();

    if (username) {
      runAnalysis(username, panel.querySelector(".coa-speed").value);
    } else {
      setBody(`<p class="coa-hint">Не удалось найти оппонента автоматически. Введите username вручную и нажмите «Анализировать».</p>`);
    }
  }

  function runAnalysis(username, speed) {
    setBody(`<p class="coa-loading">Анализирую <b>${escapeHtml(username)}</b> (${speed})…</p>`);

    chrome.runtime.sendMessage({ type: "analyze", username, speed }, (resp) => {
      if (chrome.runtime.lastError) {
        setBody(`<p class="coa-error">Ошибка связи с расширением: ${escapeHtml(chrome.runtime.lastError.message)}</p>`);
        return;
      }
      if (!resp || !resp.ok) {
        setBody(`<p class="coa-error">${escapeHtml((resp && resp.error) || "Неизвестная ошибка")}</p>`);
        return;
      }
      renderResult(resp.data);
    });
  }

  // ---------- рендер результата ----------

  function renderResult(d) {
    const rows = [
      ["Speed", d.speed],
      ["Винрейт", pct(d.winrate)],
      ["Винрейт (10 дней)", pct(d.winrate_last10_days)],
      ["Ср. точность", pct(d.avg_accuracy)],
      ["Ср. точность (10 дней)", pct(d.avg_accuracy_last10_days)],
      ["Тильт-фактор", pct(d.tilt_factor)],
      ["Частый дебют (белые)", d.most_popular_debut_white || "—"],
      ["Лучший дебют (белые)", d.most_winrate_debut_white || "—"],
      ["Частый дебют (чёрные)", d.most_popular_debut_black || "—"],
      ["Лучший дебют (чёрные)", d.most_winrate_debut_black || "—"],
    ];
    const html = `<table class="coa-table">${rows
      .map(([k, v]) => `<tr><td class="coa-k">${k}</td><td class="coa-v">${escapeHtml(String(v))}</td></tr>`)
      .join("")}</table>`;
    setBody(html);
  }

  function pct(v) {
    if (v === null || v === undefined || Number.isNaN(v)) return "—";
    return `${Number(v).toFixed(1)}%`;
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    }[c]));
  }

  // ---------- init ----------
  createButton();
})();
