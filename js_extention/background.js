// background.js — service worker.
// Все обращения к Go API идут отсюда: в MV3 фоновый воркер имеет
// кросс-доменный доступ к хостам из host_permissions в обход CORS,
// поэтому бэкенд не обязан отдавать CORS-заголовки.

const DEFAULT_API_BASE = "http://localhost:8080";

// ---------- storage helpers ----------

async function getStorage(keys) {
  return chrome.storage.local.get(keys);
}

async function setStorage(obj) {
  return chrome.storage.local.set(obj);
}

async function getApiBase() {
  const { apiBase } = await getStorage("apiBase");
  return (apiBase || DEFAULT_API_BASE).replace(/\/+$/, "");
}

// Стабильный идентификатор устройства для логина (бэкенд хранит refresh per-device).
async function getDeviceId() {
  const { deviceId } = await getStorage("deviceId");
  if (deviceId) return deviceId;
  const newId = crypto.randomUUID();
  await setStorage({ deviceId: newId });
  return newId;
}

// ---------- auth ----------

async function login(username, password) {
  const apiBase = await getApiBase();
  const deviceId = await getDeviceId();

  const res = await fetch(`${apiBase}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ username, password, device_id: deviceId }),
  });

  if (!res.ok) {
    const msg = await readError(res);
    throw new Error(msg || `Ошибка входа (HTTP ${res.status})`);
  }

  const data = await res.json();
  await setStorage({
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  });
  return true;
}

async function refreshTokens() {
  const apiBase = await getApiBase();
  const { refreshToken } = await getStorage("refreshToken");
  if (!refreshToken) return false;

  const res = await fetch(`${apiBase}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: refreshToken }),
  });

  if (!res.ok) {
    // refresh протух — чистим, потребуется повторный логин
    await setStorage({ accessToken: null, refreshToken: null });
    return false;
  }

  const data = await res.json();
  await setStorage({
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
  });
  return true;
}

async function logout() {
  await setStorage({ accessToken: null, refreshToken: null });
}

async function isLoggedIn() {
  const { accessToken } = await getStorage("accessToken");
  return Boolean(accessToken);
}

// ---------- analyze ----------

async function analyze(username, speed) {
  const apiBase = await getApiBase();
  const url = `${apiBase}/analyze/${encodeURIComponent(username)}?speed=${encodeURIComponent(speed)}`;

  let res = await authedFetch(url);

  // access протух — пробуем один раз обновить и повторить
  if (res.status === 401) {
    const refreshed = await refreshTokens();
    if (!refreshed) {
      throw new Error("Сессия истекла. Войдите снова через иконку расширения.");
    }
    res = await authedFetch(url);
  }

  if (res.status === 401) {
    throw new Error("Не авторизовано. Войдите через иконку расширения.");
  }
  if (!res.ok) {
    const msg = await readError(res);
    throw new Error(msg || `Ошибка анализа (HTTP ${res.status})`);
  }

  return res.json();
}

async function authedFetch(url) {
  const { accessToken } = await getStorage("accessToken");
  return fetch(url, {
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : {},
  });
}

async function readError(res) {
  try {
    const data = await res.json();
    return data.error || null;
  } catch {
    return null;
  }
}

// ---------- message router ----------

chrome.runtime.onMessage.addListener((msg, _sender, sendResponse) => {
  (async () => {
    try {
      switch (msg.type) {
        case "login":
          await login(msg.username, msg.password);
          sendResponse({ ok: true });
          break;
        case "logout":
          await logout();
          sendResponse({ ok: true });
          break;
        case "getAuthState":
          sendResponse({
            ok: true,
            loggedIn: await isLoggedIn(),
            apiBase: await getApiBase(),
          });
          break;
        case "setApiBase":
          await setStorage({ apiBase: (msg.apiBase || "").trim() || DEFAULT_API_BASE });
          sendResponse({ ok: true });
          break;
        case "analyze": {
          const data = await analyze(msg.username, msg.speed);
          sendResponse({ ok: true, data });
          break;
        }
        default:
          sendResponse({ ok: false, error: `Неизвестный запрос: ${msg.type}` });
      }
    } catch (err) {
      sendResponse({ ok: false, error: err.message || String(err) });
    }
  })();

  // async-ответ
  return true;
});
