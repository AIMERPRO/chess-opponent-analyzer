const $ = (id) => document.getElementById(id);

let mode = "login"; // "login" | "register"

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

// Switch between login and register: toggles tab styling, the lichess field
// (register-only) and the submit button label.
function setMode(next) {
  mode = next;
  const isRegister = mode === "register";
  $("tab-login").classList.toggle("active", !isRegister);
  $("tab-register").classList.toggle("active", isRegister);
  $("lichess-field").hidden = !isRegister;
  $("submit-btn").textContent = isRegister ? "Зарегистрироваться" : "Войти";
  setStatus("");
}

$("tab-login").addEventListener("click", () => setMode("login"));
$("tab-register").addEventListener("click", () => setMode("register"));

// Returns an error message if the form is invalid, otherwise null.
function validateForm(username, password) {
  if (!username) return "Введите логин";
  if (!password) return "Введите пароль";
  if (mode === "register" && !$("lichessUsername").value.trim()) {
    return "Введите Lichess username";
  }
  return null;
}

$("submit-btn").addEventListener("click", async () => {
  const username = $("username").value.trim();
  const password = $("password").value;

  const validationError = validateForm(username, password);
  if (validationError) {
    setStatus(validationError, true);
    return;
  }

  // lock the button so a slow request can't be submitted twice
  const btn = $("submit-btn");
  btn.disabled = true;
  try {
    await saveApiBase();

    let resp;
    if (mode === "register") {
      setStatus("Регистрация…");
      resp = await send({
        type: "register",
        username,
        password,
        lichessUsername: $("lichessUsername").value.trim(),
      });
    } else {
      setStatus("Вход…");
      resp = await send({ type: "login", username, password });
    }

    if (resp && resp.ok) {
      setStatus("");
      $("password").value = "";
      await refreshState();
    } else {
      setStatus((resp && resp.error) || "Ошибка", true);
    }
  } finally {
    btn.disabled = false;
  }
});

$("logout-btn").addEventListener("click", async () => {
  await send({ type: "logout" });
  await refreshState();
  setStatus("");
});

$("apiBase").addEventListener("change", saveApiBase);

refreshState();
