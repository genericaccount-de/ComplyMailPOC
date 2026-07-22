## ComplyMail POC Design Document

### 1. Overview

**Product name (working):** ComplyMail POC  
**Goal:** Validate that an AI‑powered assistant can reliably enforce a basic company email style guide and flag obvious sensitive content in outbound emails, with EU‑native processing (Mistral).  
**Scope:** One pilot customer, Outlook (desktop + web) only, English‑language emails, limited set of style and security rules. [sc1.checkpoint](https://sc1.checkpoint.com/documents/Harmony_Email_and_Collaboration/oxy_ex-1/Topics/outlook-add-in/outlook-add-in-supported-outlook-types-and-platform.html)

***

### 2. Objectives and Success Criteria

**Objectives**

- Provide real‑time, inline suggestions in Outlook when composing emails:
  - Style issues (tone, greeting, closing, forbidden phrases).
  - Simple compliance checks (missing disclaimer, sending to external domain).
- Inspect outbound emails via SMTP proxy and flag/score possible sensitive content:
  - Keywords and patterns (e.g. “confidential”, project code names, internal links).
  - LLM‑based “sensitivity score” for the body text.
- Keep all processing within EU infrastructure using Mistral models. [weventure](https://weventure.de/en/blog/mistral)

**Success Criteria (for POC)**

- Latency: < 1s average response for compose‑time hints on typical email length.  
- Precision: Pilot users report ≥ 70% of suggestions as “useful / relevant” in feedback.  
- Adoption: At least 5 active users in the pilot sending ≥ 10 checked emails/day over 2 weeks.  
- Compliance: Data processed and stored only in EU regions, documented for customer review. [llmdeploy](https://llmdeploy.to/solutions/mistral)

***

### 3. Functional Requirements

**Outlook Add‑in**

- As a user composes an email, the add‑in can:
  - Send the draft (subject + body, recipients) to the backend API on button click (“Check email”).  
  - Display:
    - Style recommendations (e.g. “Greeting too informal”, “Add closing line”).  
    - Security warnings (e.g. “Contains keyword ‘internal only’ but recipient is external”).  
- Users can accept or ignore suggestions; each action is logged anonymously for later analysis.  

**SMTP Proxy (Outbound Gateway)**

- Accept outbound SMTP connections from the pilot customer’s mail server.  
- For each outbound email:
  - Run deterministic checks (regex/keyword rules, domain checks).  
  - Call LLM for a sensitivity classification: `LOW | MEDIUM | HIGH`.  
- If sensitivity is `HIGH` or a hard rule is violated:
  - Add a header and optionally redirect to a review mailbox (configurable for pilot).  
- Maintain basic metrics: total emails scanned, number flagged, latency.

**Admin Console (Minimal)**

- Simple web UI for pilot admin to:
  - Define style rules (required greeting/closing, forbidden phrases).  
  - Define security rules (keywords, domains considered “internal/external”).  
  - View aggregate stats (emails scanned, flagged).

***

### 4. Non‑Functional Requirements

**Security & Privacy**

- Emails and metadata must be processed only on EU‑hosted infrastructure (e.g. Mistral La Plateforme API, EU cloud region). [help.mistral](https://help.mistral.ai/en/articles/347629-where-do-you-store-my-data-or-my-organization-s-data)
- No training on customer data; logs contain only minimal pseudonymised text where possible. [opper](https://opper.ai/provider/mistral)
- HTTPS/TLS for all communication, API key‑based auth between add‑in/proxy and backend.

**Compliance**

- Provide a short data‑processing description and architecture diagram suitable for GDPR review. [donneespersonnelles](https://www.donneespersonnelles.fr/mistral-rgpd)
- Avoid storing full email bodies by default; keep only short samples for debugging via explicit “debug mode” toggle.

**Performance**

- Gateway must not introduce more than 1–2s additional latency per email on average.  
- System should handle at least 10 emails/minute without degradation for POC.

***

### 5. High‑Level Architecture

**Outlook Add‑in (client)**

- Office Web Add‑in using JavaScript and HTML panel; communicates with backend via REST API. [techdocs.broadcom](https://techdocs.broadcom.com/jp/ja/symantec-security-software/information-security/data-loss-prevention/16-1/about-discovering-and-preventing-data-loss-on-endpoints/adding-and-editing-agent-configurations/channel-settings/enable-monitoring-settings/monitoring-microsoft-outlook-using-the-on-send-web-add-in.html)

**Backend API (EU Cloud)**

- Auth, rule engine, LLM orchestrator.  
- Exposes endpoints:
  - `POST /check-style-email` for add‑in.  
  - `POST /scan-outbound-email` for SMTP proxy.

**LLM Layer**

- Calls Mistral API (EU endpoint) or self‑hosted Mistral model for: [innfactory](https://innfactory.ai/en/ai-models/mistral/)
  - Style compliance analysis (prompt with company style guide + email text).  
  - Sensitivity classification (prompt with security categories + email text).

**SMTP Proxy**

- Lightweight service in front of customer mail server; parses email, invokes backend, adds headers or reroutes as configured. [techdocs.broadcom](https://techdocs.broadcom.com/jp/ja/symantec-security-software/information-security/data-loss-prevention/16-1/about-discovering-and-preventing-data-loss-on-endpoints/adding-and-editing-agent-configurations/channel-settings/enable-monitoring-settings/monitoring-microsoft-outlook-using-the-on-send-web-add-in.html)

**Admin Console**

- Single‑page app + simple backend endpoints to manage rules.

***

### 6. Pilot Scope and Limitations

- Single customer, single mail domain.  
- English‑only emails.  
- No attachment content scanning (only text in body/subject).  
- Limited rule set: up to 20 style rules, 20 security rules.  
- No deep integration with Purview or other DLP tools in POC stage. [learn.microsoft](https://learn.microsoft.com/en-us/purview/ai-microsoft-purview)

***

If you imagine actually implementing this, which section of the Markdown doc would you extend next (e.g. detailed API spec, data model, or threat model) to make it actionable for you as a developer?