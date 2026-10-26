package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// /docs — starting an assessment from CI/CD, and the HTTP API reference.
//
// The CI examples are data, rendered by one template, so adding a system is one
// entry and they cannot drift in layout. Every example sends the same request:
// POST /api/v1/pentests with an X-API-Token header and a domain.

type ciExample struct {
	Title string
	Code  string
	Note  template.HTML
}

var ciExamples = []ciExample{
	{"GitHub Actions", `name: pentest-on-deploy
on:
  push:
    branches: [ main, production ]
jobs:
  nsk-pentest:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Netsekurity assessment
        run: |
          curl -s -X POST https://netsekurity.com/api/v1/pentests \
            -H "X-API-Token: ${{ secrets.NETSEKURITY_API_TOKEN }}" \
            -d "domain=app.example.com&mode=standard"`,
		`Add <code>NETSEKURITY_API_TOKEN</code> under Settings → Secrets and variables → Actions.`},
	{"GitLab CI", `# .gitlab-ci.yml
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
  environment: production`,
		`Add <code>NETSEKURITY_API_TOKEN</code> under Settings → CI/CD → Variables (masked).`},
	{"Jenkins", `pipeline {
  agent any
  environment {
    NSK_TOKEN = credentials('nsk-api-token')   // Secret text credential
  }
  stages {
    stage('deploy') { steps { sh './deploy.sh' } }
    stage('assessment') {
      steps {
        sh "curl -s -X POST https://netsekurity.com/api/v1/pentests " +
           "-H 'X-API-Token: $NSK_TOKEN' " +
           "-d 'domain=app.example.com&mode=standard'"
      }
    }
  }
}`,
		`Create a <em>Secret text</em> credential named <code>nsk-api-token</code>.`},
	{"CircleCI", `# .circleci/config.yml
version: 2.1
jobs:
  pentest:
    docker: [{ image: cimg/base:stable }]
    steps:
      - run:
          name: Netsekurity assessment
          command: |
            curl -s -X POST https://netsekurity.com/api/v1/pentests \
              -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
              -d "domain=app.example.com&mode=standard"
workflows:
  version: 2
  deploy-and-assess:
    jobs:
      - pentest:
          filters: { branches: { only: [production] } }`,
		`Add <code>NETSEKURITY_API_TOKEN</code> under Project → Settings → Environment Variables.`},
	{"Azure DevOps", `# azure-pipelines.yml
trigger:
  branches:
    include: [ main ]
pool: { vmImage: ubuntu-latest }
steps:
  - script: |
      curl -s -X POST https://netsekurity.com/api/v1/pentests \
        -H "X-API-Token: $(NETSEKURITY_API_TOKEN)" \
        -d "domain=app.example.com&mode=standard"
    displayName: "Netsekurity assessment"`,
		`Add <code>NETSEKURITY_API_TOKEN</code> under Pipelines → Library → Variable group (keep it secret).`},
	{"Bitbucket Pipelines", `# bitbucket-pipelines.yml
pipelines:
  branches:
    main:
      - step:
          name: Netsekurity assessment
          script:
            - curl -s -X POST https://netsekurity.com/api/v1/pentests \
                -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
                -d "domain=app.example.com&mode=standard"`,
		`Add <code>NETSEKURITY_API_TOKEN</code> under Repository → Settings → Pipelines → Repository variables (secured).`},
	{"Travis CI", `# .travis.yml
language: generic
script:
  - ./deploy.sh
after_deploy:
  - curl -s -X POST https://netsekurity.com/api/v1/pentests \
      -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
      -d "domain=app.example.com&mode=standard"
branches: { only: [ master ] }`,
		`Add the encrypted secret with <code>travis encrypt NETSEKURITY_API_TOKEN=xxx</code>.`},
	{"TeamCity", `# Build step (Command Line): "Netsekurity assessment"
curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: %env.NSK_API_TOKEN%" \
  -d "domain=app.example.com&mode=standard"`,
		`Define <code>env.NSK_API_TOKEN</code> in the build configuration (Parameters, Password type).`},
	{"Buildkite", `# buildkite-agent pipeline
steps:
  - label: "Netsekurity assessment"
    command: |
      curl -s -X POST https://netsekurity.com/api/v1/pentests \
        -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
        -d "domain=app.example.com&mode=standard"
    agents: { queue: default }`,
		`Add <code>NETSEKURITY_API_TOKEN</code> under Pipeline → Settings → Environment variables (secret).`},
	{"Codefresh", `# codefresh.yml
version: "1.0"
steps:
  nsk_pentest:
    type: run
    image: curlimages/curl:latest
    commands:
      - curl -s -X POST https://netsekurity.com/api/v1/pentests \
          -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
          -d "domain=app.example.com&mode=standard"`,
		`Set <code>NETSEKURITY_API_TOKEN</code> in the pipeline environment as a secret.`},
	{"Semaphore", `# semaphore.yml
blocks:
  - name: "Netsekurity assessment"
    task:
      jobs:
        - name: pentest
          commands:
            - curl -s -X POST https://netsekurity.com/api/v1/pentests \
                -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
                -d "domain=app.example.com&mode=standard"`,
		`Set <code>NETSEKURITY_API_TOKEN</code> in the pipeline environment as a secret.`},
}

const docsBodyHTML = `
<section class="border-b-2 border-black bg-[#fffbeb]">
  <div class="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:py-16">
    <span class="nsk-tag">Docs</span>
    <h1 class="mt-4 max-w-4xl text-4xl font-extrabold leading-[1.1] tracking-tight sm:text-5xl">CI/CD integration and API reference</h1>
    <p class="mt-5 max-w-3xl text-lg text-gray-800">Start an external web security assessment from your pipeline on every deploy. One HTTP request, authenticated with an API token.</p>
    <ul class="mt-6 flex flex-wrap gap-2">
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#quickstart">Quickstart</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#api">API reference</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#ci">CI examples</a></li>
      <li><a class="nsk-btn nsk-btn-ghost nsk-btn-sm" href="#notes">Security notes</a></li>
    </ul>
  </div>
</section>

<div class="mx-auto max-w-7xl space-y-8 px-4 py-12 sm:px-6">

  <section id="quickstart" class="nsk-card" aria-labelledby="h-quick">
    <h2 id="h-quick" class="nsk-card-head">Quickstart</h2>
    <div class="p-5">
      <ol class="nsk-prose list-decimal space-y-2 pl-5 text-gray-800">
        <li><strong>Add and verify your domain.</strong> Dashboard → My websites → add the domain → add the TXT record → check. The domain must show as <em>Verified</em>.</li>
        <li><strong>Create an API token.</strong> Dashboard → Automation → choose how long it is valid (7, 14, 30, 60 or 90 days) → create. Copy it now: it is shown only once.</li>
        <li><strong>Store it as a secret</strong> in your CI, for example <code>NETSEKURITY_API_TOKEN</code>.</li>
        <li><strong>Call the endpoint</strong> on deploy. Netsekurity queues the assessment and runs it; the report appears in your dashboard.</li>
      </ol>
      <pre class="nsk-code mt-4 block overflow-x-auto whitespace-pre-wrap text-xs">curl -s -X POST https://netsekurity.com/api/v1/pentests \
  -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "domain=app.example.com&amp;mode=standard"
# 1 credit per assessment. External and read-only, so safe to point at production.</pre>
    </div>
  </section>

  <section id="api" class="nsk-card" aria-labelledby="h-api">
    <h2 id="h-api" class="nsk-card-head">HTTP API reference</h2>
    <div class="space-y-5 p-5">
      <div>
        <h3 class="nsk-mono text-base font-extrabold">POST /api/v1/pentests</h3>
        <p class="mt-1 text-gray-800">Start an assessment.</p>
        <div class="mt-3 overflow-x-auto border-2 border-black">
          <table class="w-full min-w-[520px] text-left text-sm">
            <thead><tr class="border-b-2 border-black bg-[#fffbeb]"><th scope="col" class="px-4 py-2">Parameter</th><th scope="col" class="px-4 py-2">Type</th><th scope="col" class="px-4 py-2">Description</th></tr></thead>
            <tbody class="[&>tr]:border-b-2 [&>tr]:border-gray-200">
              <tr><td class="nsk-mono px-4 py-2">X-API-Token</td><td class="px-4 py-2">header</td><td class="px-4 py-2">Required. Your API token.</td></tr>
              <tr><td class="nsk-mono px-4 py-2">domain</td><td class="px-4 py-2">form or JSON</td><td class="px-4 py-2">Required. A verified domain owned by the token's account.</td></tr>
              <tr><td class="nsk-mono px-4 py-2">mode</td><td class="px-4 py-2">form or JSON</td><td class="px-4 py-2">Optional. <code class="nsk-mono">standard</code> is the only value: 1 credit, read-only.</td></tr>
            </tbody>
          </table>
        </div>
        <pre class="nsk-code mt-3 block overflow-x-auto whitespace-pre-wrap text-xs"># 200 OK
{"pentest_id":"pt_1f2e…","domain":"app.example.com","mode":"standard","status":"queued"}
# 401 missing, invalid or expired token
# 400 domain not verified · 400 already queued or running (one per account at a time)
# 402 insufficient credits · 429 rate limited</pre>
      </div>
      <div>
        <h3 class="text-base font-extrabold">Reading the result</h3>
        <p class="nsk-prose mt-1 text-gray-800">When an assessment finishes its status becomes <code>completed</code> and the report is available in the dashboard, to the account that ordered it. The API starts assessments; results are read in the dashboard today.</p>
      </div>
    </div>
  </section>

  <section id="ci" aria-labelledby="h-ci">
    <h2 id="h-ci" class="text-2xl font-extrabold tracking-tight sm:text-3xl">CI examples</h2>
    <p class="mt-2 max-w-3xl text-gray-700">Any system that can send an HTTP request will work. These are ready to paste.</p>
    <div class="mt-6 space-y-6">
      {{range $i, $e := .Examples}}<article class="nsk-card" aria-labelledby="ci-{{$i}}">
        <h3 id="ci-{{$i}}" class="nsk-card-head">{{$e.Title}}</h3>
        <div class="p-5">
          <pre class="nsk-code block overflow-x-auto whitespace-pre-wrap text-xs">{{$e.Code}}</pre>
          <p class="nsk-prose mt-3 text-sm text-gray-700">{{$e.Note}}</p>
        </div>
      </article>
      {{end}}
    </div>
  </section>

  <section id="notes" class="nsk-card" aria-labelledby="h-notes">
    <h2 id="h-notes" class="nsk-card-head">Configuration and security notes</h2>
    <ul class="nsk-prose list-disc space-y-2 p-5 pl-10 text-gray-800">
      <li><strong>Tokens expire.</strong> Choose an expiry that matches your deploy cadence, and rotate by creating a new token and deleting the old one.</li>
      <li><strong>Store tokens as CI secrets.</strong> Never commit them to source control. A token is stored only as a SHA-256 hash on our side.</li>
      <li><strong>One token per pipeline</strong> is fine. Delete any token from the dashboard at any time.</li>
      <li><strong>One assessment at a time.</strong> Only one queued or running assessment per account; the API returns an error if one is already active.</li>
      <li><strong>Credits.</strong> One assessment is one credit. An insufficient balance returns HTTP 402.</li>
      <li><strong>Scope.</strong> External, unauthenticated and read-only. See the <a href="/product-scope">product scope</a> for exactly what is and is not checked.</li>
    </ul>
  </section>
</div>
`

var docsBodyTemplate = template.Must(template.New("docs").Parse(docsBodyHTML))

// handleDocs renders /docs inside the shared layout.
func handleDocs(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if err := docsBodyTemplate.Execute(&buf, struct{ Examples []ciExample }{ciExamples}); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
		return
	}
	serveRawPage(w, r, rawPage{
		Key: "docs", Path: "/docs", Crumb: "Docs", Body: template.HTML(buf.String()),
		Title:       "CI/CD Integration and API Reference | Netsekurity",
		Description: fmt.Sprintf("Start a Netsekurity external web security assessment from your pipeline on every deploy: quickstart, HTTP API reference and ready-to-paste examples for %d CI systems.", len(ciExamples)),
	})
}
