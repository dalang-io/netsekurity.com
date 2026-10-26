package main

// The CI examples on /docs and /id/dokumentasi. The code is the same in both
// languages, so it is data; the one-line note under each (where to put the token)
// is language copy, docsCopy.CINotes, one per entry here in the same order.
//
// Every example starts an assessment with the same request:
// POST /api/v1/assessments with an X-API-Token header and a domain.

type ciExample struct {
	Title string
	Code  string
}

var ciExamples = []ciExample{
	{"GitHub Actions", `name: assessment-on-deploy
on:
  push:
    branches: [ main, production ]
jobs:
  nsk-assessment:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Netsekurity assessment
        run: |
          curl -s -X POST https://netsekurity.com/api/v1/assessments \
            -H "X-API-Token: ${{ secrets.NETSEKURITY_API_TOKEN }}" \
            -d "domain=app.example.com&authorised=true"`},
	{"GitLab CI", `# .gitlab-ci.yml
nsk-assessment:
  stage: .post
  image: alpine:latest
  script:
    - apk add --no-cache curl
    - curl -s -X POST https://netsekurity.com/api/v1/assessments \
        -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
        -d "domain=app.example.com&authorised=true"
  only:
    - production
  environment: production`},
	{"Jenkins", `pipeline {
  agent any
  environment {
    NSK_TOKEN = credentials('nsk-api-token')   // Secret text credential
  }
  stages {
    stage('deploy') { steps { sh './deploy.sh' } }
    stage('assessment') {
      steps {
        sh "curl -s -X POST https://netsekurity.com/api/v1/assessments " +
           "-H 'X-API-Token: $NSK_TOKEN' " +
           "-d 'domain=app.example.com&authorised=true'"
      }
    }
  }
}`},
	{"CircleCI", `# .circleci/config.yml
version: 2.1
jobs:
  assessment:
    docker: [{ image: cimg/base:stable }]
    steps:
      - run:
          name: Netsekurity assessment
          command: |
            curl -s -X POST https://netsekurity.com/api/v1/assessments \
              -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
              -d "domain=app.example.com&authorised=true"
workflows:
  version: 2
  deploy-and-assess:
    jobs:
      - assessment:
          filters: { branches: { only: [production] } }`},
	{"Azure DevOps", `# azure-pipelines.yml
trigger:
  branches:
    include: [ main ]
pool: { vmImage: ubuntu-latest }
steps:
  - script: |
      curl -s -X POST https://netsekurity.com/api/v1/assessments \
        -H "X-API-Token: $(NETSEKURITY_API_TOKEN)" \
        -d "domain=app.example.com&authorised=true"
    displayName: "Netsekurity assessment"`},
	{"Bitbucket Pipelines", `# bitbucket-pipelines.yml
pipelines:
  branches:
    main:
      - step:
          name: Netsekurity assessment
          script:
            - curl -s -X POST https://netsekurity.com/api/v1/assessments \
                -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
                -d "domain=app.example.com&authorised=true"`},
	{"Travis CI", `# .travis.yml
language: generic
script:
  - ./deploy.sh
after_deploy:
  - curl -s -X POST https://netsekurity.com/api/v1/assessments \
      -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
      -d "domain=app.example.com&authorised=true"
branches: { only: [ master ] }`},
	{"TeamCity", `# Build step (Command Line): "Netsekurity assessment"
curl -s -X POST https://netsekurity.com/api/v1/assessments \
  -H "X-API-Token: %env.NSK_API_TOKEN%" \
  -d "domain=app.example.com&authorised=true"`},
	{"Buildkite", `# buildkite-agent pipeline
steps:
  - label: "Netsekurity assessment"
    command: |
      curl -s -X POST https://netsekurity.com/api/v1/assessments \
        -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
        -d "domain=app.example.com&authorised=true"
    agents: { queue: default }`},
	{"Codefresh", `# codefresh.yml
version: "1.0"
steps:
  nsk_assessment:
    type: run
    image: curlimages/curl:latest
    commands:
      - curl -s -X POST https://netsekurity.com/api/v1/assessments \
          -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
          -d "domain=app.example.com&authorised=true"`},
	{"Semaphore", `# semaphore.yml
blocks:
  - name: "Netsekurity assessment"
    task:
      jobs:
        - name: assessment
          commands:
            - curl -s -X POST https://netsekurity.com/api/v1/assessments \
                -H "X-API-Token: $NETSEKURITY_API_TOKEN" \
                -d "domain=app.example.com&authorised=true"`},
}
