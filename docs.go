package main

import (
	"html/template"
	"net/http"
	"strings"
)

// docsHTMLRaw is the /docs page body (header is injected server-side by the shared
// renderHeader component).
var docsHTMLRaw = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>docs — netsekurity.com / CI/CD integration</title>
<meta name="robots" content="index,follow"/>
<meta name="description" content="Documentation &amp; tutorials to integrate Netsekurity automated pentests into your CI/CD pipeline, plus the HTTP API reference."/>
<link rel="icon" type="image/svg+xml" href="/favicon.svg"/>
<link rel="apple-touch-icon" href="/favicon-512.png"/>
<link rel="stylesheet" href="/css/styles.css?v={{cssHash}}"/>
<style>pre{white-space:pre-wrap;word-break:break-word}code{font-family:ui-monospace,monospace}</style>
</head>
<body class="scanlines bg-ink text-gray-300 min-h-screen overflow-x-hidden">
  __DOCS_HEADER__

<main class="mx-auto max-w-6xl overflow-x-hidden px-4 py-8 sm:px-6">
  <h1 class="font-mono text-2xl font-bold text-white"><span class="text-cyan-400">$</span> man netsekurity --cicd-integration</h1>
  <p class="mt-2 font-mono text-sm text-gray-500"># pentest on every deploy — automated API integration for your CI/CD</p>

  <div class="mt-6 space-y-6">
    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-emerald-300">01 · Quickstart</h2>
      <ol class="mt-3 list-inside list-decimal space-y-2 font-mono text-xs leading-relaxed text-gray-300">
        <li><b>Add &amp; verify your domain</b> — dashboard → add domain → add the TXT record → verify. The domain must be <span class="text-emerald-300">[verified]</span>.</li>
        <li><b>Generate an API token</b> — dashboard → <span class="text-cyan-300">api tokens (CI/CD)</span> → choose expiry (7/14/30/60/90 days) → generate. Copy it now (shown only once).</li>
        <li><b>Store as a secret</b> in your CI — e.g. GitHub Actions <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code>.</li>
        <li><b>Call the endpoint</b> on deploy. Done — Netsekurity queues and runs the pentest, then uploads the report.</li>
      </ol>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300">curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "domain=app.example.com&amp;mode=standard"
# standard = 1 credit · destructive = 2 credits (exploit/RCE/webshell — use a dev server)</pre>
    </section>

    <section class="rounded border border-white/10 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">02 · HTTP API Reference</h2>
      <div class="mt-3 space-y-4 font-mono text-xs">
        <div>
          <div class="text-cyan-300">POST /api/v1/pentests — start a pentest from CI/CD</div>
          <table class="mt-2 w-full border-collapse text-left">
            <thead><tr class="border-b border-white/15 text-gray-500"><th class="py-1 pr-3">param</th><th class="py-1 pr-3">type</th><th class="py-1">desc</th></tr></thead>
            <tbody class="text-gray-300">
              <tr class="border-b border-white/5"><td class="py-1 pr-3">X-API-Token</td><td class="py-1 pr-3">header</td><td class="py-1">required — your API token</td></tr>
              <tr class="border-b border-white/5"><td class="py-1 pr-3">domain</td><td class="py-1 pr-3">form/JSON</td><td class="py-1">required — verified domain owned by the token user</td></tr>
              <tr class="border-b border-white/5"><td class="py-1 pr-3">mode</td><td class="py-1 pr-3">form/JSON</td><td class="py-1">standard (default, 1 credit) or destructive (2 credits)</td></tr>
            </tbody>
          </table>
          <pre class="mt-2 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># 200 OK
{"pentest_id":"pt_1f2e…","domain":"app.example.com","mode":"standard","status":"queued"}
# 401 missing/invalid/expired token · 400 domain not verified · 402 insufficient credits
# 400 already queued/running (in-flight cap: 1 per user)</pre>
        </div>
        <div>
          <div class="text-cyan-300">Report download</div>
          <p class="mt-1 text-gray-400">When a pentest finishes it sets <code class="text-gray-200">status=completed</code> and the report is
          available on the dashboard (<code class="text-gray-200">/reports/&lt;name&gt;.pdf</code>, owner/admin only). Poll
          <code class="text-gray-200">/api/pentests/list</code> or check the dashboard.</p>
        </div>
      </div>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">03 · GitHub Actions</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300">name: pentest-on-deploy
on:
  push:
    branches: [ main, production ]
jobs:
  nsk-pentest:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Netsekurity pentest
        run: |
          curl -s -X POST https://netsekurity.com/api/v1/pentests \
            -H "X-API-Token: $&#123;&#123; secrets.NETSEKURITY_API_TOKEN &#125;&#125;" \
            -d "domain=app.example.com&amp;mode=standard"
        env:
          NETSEKURITY_API_TOKEN: $&#123;&#123; secrets.NETSEKURITY_API_TOKEN &#125;&#125;</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> under Settings → Secrets and variables → Actions.</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">04 · GitLab CI</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># .gitlab-ci.yml
nsk-pentest:
  stage: .post
  image: alpine:latest
  script:
    - apk add --no-cache curl
    - curl -s -X POST https://netsekurity.com/api/v1/pentests \
        -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
        -d "domain=app.example.com&mode=standard"
  only:
    - production
  environment: production</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> under Settings → CI/CD → Variables (Masked).</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">05 · Jenkins</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300">pipeline {
  agent any
  environment {
    NSK_TOKEN = credentials('nsk-api-token')   // Secret text credential
  }
  stages {
    stage('deploy') { steps { sh './deploy.sh' } }
    stage('pentest') {
      steps {
        sh "curl -s -X POST https://netsekurity.com/api/v1/pentests " +
           "-H 'X-API-Token: $NSK_TOKEN' " +
           "-d 'domain=app.example.com&mode=standard'"
      }
    }
  }
}</pre>
      <p class="mt-2 text-[11px] text-gray-500">Create a <em>Secret text</em> credential named <code class="text-cyan-300">nsk-api-token</code>.</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">06 · CircleCI</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># .circleci/config.yml
version: 2.1
jobs:
  pentest:
    docker: [{ image: cimg/base:stable }]
    steps:
      - run:
          name: Netsekurity pentest
          command: |
            curl -s -X POST https://netsekurity.com/api/v1/pentests \
              -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
              -d "domain=app.example.com&mode=standard"
workflows:
  version: 2
  deploy-and-pentest:
    jobs:
      - pentest:
          filters: { branches: { only: [production] } }</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> under Project → Settings → Environment Variables.</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">07 · Azure DevOps</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># azure-pipelines.yml
trigger:
  branches:
    include: [ main ]
pool: { vmImage: ubuntu-latest }
variables:
  NSK_TOKEN: $[variables.NETSEKURITY_API_TOKEN]
steps:
  - script: |
      curl -s -X POST https://netsekurity.com/api/v1/pentests \
        -H "X-API-Token: $(NETSEKURITY_API_TOKEN)" \
        -d "domain=app.example.com&mode=standard"
    displayName: "Netsekurity pentest"</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> under Pipelines → Library → Variable group (keep secret).</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">08 · Bitbucket Pipelines</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># bitbucket-pipelines.yml
pipelines:
  branches:
    main:
      - step:
          name: Netsekurity pentest
          script:
            - curl -s -X POST https://netsekurity.com/api/v1/pentests \
                -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
                -d "domain=app.example.com&mode=standard"</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> under Repository → Settings → Pipelines → Repository variables (Secured).</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">09 · Travis CI</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># .travis.yml
language: generic
script:
  - ./deploy.sh
after_deploy:
  - curl -s -X POST https://netsekurity.com/api/v1/pentests \
      -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
      -d "domain=app.example.com&mode=standard"
branches: { only: [ master ] }</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add the encrypted secret via <code class="text-cyan-300">travis encrypt NETSEKURITY_API_TOKEN=xxx</code> in <code class="text-cyan-300">.travis.yml</code>.</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">10 · TeamCity</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># Build step (Command Line): "Netsekurity pentest"
curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: %env.NSK_API_TOKEN%" \
  -d "domain=app.example.com&mode=standard"</pre>
      <p class="mt-2 text-[11px] text-gray-500">Define <code class="text-cyan-300">env.NSK_API_TOKEN</code> in the build configuration (Parameters, Password type).</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">11 · Buildkite</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># buildkite-agent pipeline
steps:
  - label: "Netsekurity pentest"
    command: |
      curl -s -X POST https://netsekurity.com/api/v1/pentests \
        -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
        -d "domain=app.example.com&mode=standard"
    agents: { queue: default }</pre>
      <p class="mt-2 text-[11px] text-gray-500">Add <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> under Pipeline → Settings → Environment variables (secret).</p>
    </section>

    <section class="rounded border border-emerald-500/25 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">12 · Codefresh &amp; Semaphore</h2>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-emerald-300"># codefresh (codefresh.yml)
version: "1.0"
steps:
  nsk_pentest:
    type: run
    image: curlimages/curl:latest
    commands:
      - curl -s -X POST https://netsekurity.com/api/v1/pentests \
          -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
          -d "domain=app.example.com&mode=standard"

# semaphore (semaphore.yml)
blocks:
  - name: "Netsekurity pentest"
    task:
      jobs:
        - name: pentest
          commands:
            - curl -s -X POST https://netsekurity.com/api/v1/pentests \
                -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
                -d "domain=app.example.com&mode=standard"</pre>
      <p class="mt-2 text-[11px] text-gray-500">Set <code class="text-cyan-300">NETSEKURITY_API_TOKEN</code> in the pipeline environment (secret) in each tool.</p>
    </section>

    <section class="rounded border border-white/10 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">13 · Destructive mode</h2>
      <p class="mt-2 font-mono text-xs leading-relaxed text-gray-300"><span class="text-red-300 font-bold">DANGER.</span>
      Set <code class="text-cyan-300">mode=destructive</code> (2 credits) to run active exploitation — RCE, webshell upload,
      malware/exploit injection, takeover attempts. It may damage the target. Always point at a
      <span class="text-yellow-300">development / staging server</span>, never production, unless explicitly authorized.</p>
      <pre class="mt-3 rounded bg-black/60 p-3 text-[11px] text-red-300">curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
  -d "domain=staging.example.com&mode=destructive"</pre>
    </section>

    <section class="rounded border border-white/10 bg-[#04060c] p-5">
      <h2 class="font-mono text-lg font-bold text-white">14 · Configuration &amp; security notes</h2>
      <ul class="mt-3 list-inside list-disc space-y-1.5 font-mono text-xs text-gray-300">
        <li><b>Tokens expire</b> — set an expiry that matches your deploy cadence; rotate regularly by generating a new token and revoking the old one.</li>
        <li><b>Store tokens as CI secrets</b> — never commit them to source control.</li>
        <li><b>One token per user/pipeline</b> is fine; revoke from the dashboard any time.</li>
        <li><b>In-flight cap</b> — only 1 queued/running pentest per user at a time; the API returns an error if a scan is already active.</li>
        <li><b>Credits</b> — standard pentest = 1 credit, destructive = 2 credits. Insufficient balance returns HTTP 402.</li>
      </ul>
    </section>
  </div>
</main>

<footer class="mt-10 border-t border-white/10 py-6 text-center font-mono text-[11px] text-gray-600">
  netsekurity.com — automated pentest on every deploy · <a href="/" class="text-cyan-400 hover:underline">back home</a>
</footer>
</body>
</html>`

// handleDocs renders /docs. Uses the shared renderHeader component so the
// header is identical to the landing page.
func handleDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	docsNav := []hdrLink{
		{Href: "/", Text: "man home"},
		{Href: "/docs", Text: "man docs"},
		{Href: "/#cicd", Text: "$ pip install cicd"},
		{Href: "/#pricing", Text: "cat pricing"},
		{Href: "/#faq", Text: "man faq"},
	}
	authNav := `<a href="/login" class="rounded border border-emerald-400 bg-emerald-500/10 px-3 py-1.5 text-[13px] font-bold text-emerald-300 hover:bg-emerald-500/20 glow">login</a>`
	hdr := string(renderHeader(docsNav, template.HTML(authNav), "/")) + headerMobileJS
	out := strings.ReplaceAll(docsHTMLRaw, "__DOCS_HEADER__", hdr)
	out = strings.ReplaceAll(out, "__CSS_HASH__", cssHash)
	w.Write([]byte(out))
}