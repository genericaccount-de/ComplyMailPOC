// taskpane.ts — entry point for the ComplyMail task pane.
// Will use Office.js to read the current compose draft and
// send it to the backend API for style & security checks.

document.getElementById("check-btn")?.addEventListener("click", async () => {
  const resultsDiv = document.getElementById("results");
  if (resultsDiv) {
    resultsDiv.textContent = "Checking…";
  }
  // TODO: read draft via Office.js, POST to backend, render suggestions.
});
