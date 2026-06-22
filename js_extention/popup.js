const $ = (id) => document.getElementById(id);

function setStatus(msg, isError = false) {
  const el = $("status");
  el.textContent = msg || "";
  el.className = "status" + (isError ? " error" : "");
}

function send(msg) {
  return new Promise((resolve) => chrome.runtime.sendMessage(msg, resolve));
}

async function refreshState() {
  const state = await send({ type: "getAuthState" });
  if (!state || !state.ok) {
    setStatus("Не удалось связаться с расширением", true);
    return;
  }
  $("apiBase").value = state.apiBase;
  $("logged-in").hidden = !state.loggedIn;
  $("logged-out").hidden = state.loggedIn;
}

async function saveApiBase() {
  await send({ type: "setApiBase", apiBase: $("apiBase").value });
}

$("login-btn").addEventListener("click", async () => {
  setStatus("Вход…");
  await saveApiBase();
  const resp = await send({
    type: "login",
    username: $("username").value.trim(),
    password: $("password").value,
  });
  if (resp && resp.ok) {
    setStatus("");
    $("password").value = "";
    await refreshState();
  } else {
    setStatus((resp && resp.error) || "Ошибка входа", true);
  }
});

$("logout-btn").addEventListener("click", async () => {
  await send({ type: "logout" });
  await refreshState();
  setStatus("");
});

$("apiBase").addEventListener("change", saveApiBase);

refreshState();
