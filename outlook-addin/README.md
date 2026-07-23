# ComplyMail Outlook Add-in

Office Web Add-in (TypeScript + HTML) that provides compose-time email checks.

## Features (planned)
- "Check Email" button in the compose pane task panel.
- Displays style recommendations and security warnings from the backend API.
- Users can accept or ignore suggestions.

## Development

```bash
npm install
# Generate & trust localhost HTTPS certs (once). Office requires HTTPS.
npx office-addin-dev-certs install
npm run dev            # serves https://localhost:3000/taskpane.html
```

The dev server proxies `/api/*` to the backend at `http://localhost:8080`
(see `vite.config.ts`), so the HTTPS task pane can call the HTTP API without
mixed-content or CORS problems. Start the backend separately:

```bash
cd ../backend && go run ./cmd/api -config config.yaml
```

## Sideload into local Outlook (macOS)

1. Ensure `npm run dev` and the backend are running.
2. Copy the manifest into Outlook's sideload folder and restart Outlook:
   ```bash
   mkdir -p ~/Library/Containers/com.microsoft.Outlook/Data/Documents/wef
   cp manifest.xml ~/Library/Containers/com.microsoft.Outlook/Data/Documents/wef/
   ```
   Or, in classic Outlook for Mac: **Tools > Get Add-ins > My add-ins >
   Add a custom add-in > Add from file…** and pick `manifest.xml`.
   For new Outlook / web, sideload via https://aka.ms/olksideload.
3. Compose an email, open the **ComplyMail** task pane, and click **Check Email**.

Requires a Microsoft 365 / Exchange Online mailbox (sideloading is not
supported for IMAP/Gmail/POP accounts).

## Build & deployment

```bash
npm run build          # outputs to ../dist
```

Update `manifest.xml` (`SourceLocation` and `<Id>`) with the production URL and
a unique GUID before publishing.
