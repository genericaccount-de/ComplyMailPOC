// taskpane.ts — entry point for the ComplyMail task pane.
// Reads the current compose draft via Office.js and sends it to the backend
// /check-style-email endpoint (proxied under /api by the Vite dev server) to
// render style & security suggestions.

// API base is proxied by Vite (see vite.config.ts) to avoid mixed-content and
// CORS issues in local development.
const API_BASE = "/api";

interface StyleSuggestion {
  type: string;
  severity: "info" | "warning" | "error";
  message: string;
}

interface CheckStyleResponse {
  suggestions: StyleSuggestion[];
}

Office.onReady((info) => {
  const button = document.getElementById("check-btn") as HTMLButtonElement | null;
  const status = document.getElementById("status");

  if (info.host !== Office.HostType.Outlook) {
    if (status) status.textContent = "This add-in only runs in Outlook.";
    return;
  }

  if (button) {
    button.disabled = false;
    button.addEventListener("click", () => void runCheck());
  }
  if (status) status.textContent = "Ready. Compose an email and click Check Email.";
});

async function runCheck(): Promise<void> {
  const button = document.getElementById("check-btn") as HTMLButtonElement | null;
  const status = document.getElementById("status");
  const results = document.getElementById("results");

  if (results) results.innerHTML = "";
  if (button) button.disabled = true;
  if (status) status.textContent = "Checking…";

  try {
    const item = Office.context.mailbox.item;
    if (!item) throw new Error("No active email item.");

    const [subject, body, recipients] = await Promise.all([
      getSubject(item),
      getBody(item),
      getRecipients(item),
    ]);

    const res = await fetch(`${API_BASE}/check-style-email`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ subject, body, recipients }),
    });

    if (!res.ok) {
      throw new Error(`Backend returned ${res.status} ${res.statusText}`);
    }

    const data = (await res.json()) as CheckStyleResponse;
    renderSuggestions(data.suggestions ?? []);
    if (status) {
      const n = data.suggestions?.length ?? 0;
      status.textContent = n === 0 ? "No issues found." : `${n} suggestion(s):`;
    }
  } catch (err) {
    if (status) status.textContent = `Error: ${err instanceof Error ? err.message : String(err)}`;
  } finally {
    if (button) button.disabled = false;
  }
}

function renderSuggestions(suggestions: StyleSuggestion[]): void {
  const results = document.getElementById("results");
  if (!results) return;
  results.innerHTML = "";
  for (const s of suggestions) {
    const li = document.createElement("li");
    li.className = s.severity;
    const sev = document.createElement("span");
    sev.className = "sev";
    sev.textContent = s.severity;
    li.appendChild(sev);
    li.appendChild(document.createTextNode(` ${s.message}`));
    results.appendChild(li);
  }
}

// --- Office.js async helpers wrapped as promises ---

function getSubject(item: Office.MessageCompose): Promise<string> {
  return new Promise((resolve, reject) => {
    item.subject.getAsync((result) => {
      if (result.status === Office.AsyncResultStatus.Succeeded) resolve(result.value ?? "");
      else reject(result.error);
    });
  });
}

function getBody(item: Office.MessageCompose): Promise<string> {
  return new Promise((resolve, reject) => {
    item.body.getAsync(Office.CoercionType.Text, (result) => {
      if (result.status === Office.AsyncResultStatus.Succeeded) resolve(result.value ?? "");
      else reject(result.error);
    });
  });
}

function getRecipients(item: Office.MessageCompose): Promise<string[]> {
  return new Promise((resolve, reject) => {
    item.to.getAsync((result) => {
      if (result.status === Office.AsyncResultStatus.Succeeded) {
        resolve(result.value.map((r) => r.emailAddress));
      } else {
        reject(result.error);
      }
    });
  });
}
