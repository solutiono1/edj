import { defineRailway, github, project, service } from "railway/iac";

export default defineRailway(() => {
  const repo = github("solutiono1/pagor");

  const pagorApiV2 = service("pagor-api-v2", {
    source: repo,
    build: { buildEnvironment: "V3", builder: "RAILPACK" },
    replicas: { "iad": 1 },
  });

  const pagorApi = service("pagor-api", {
    source: repo,
    build: {
      buildCommand: "go build -o bin/pagor ./cmd/edj",
      buildEnvironment: "V3",
      builder: "NIXPACKS",
    },
    start: "./bin/pagor",
    replicas: { "iad": 1 },
  });

  const pagorApiLegacy = service("pagor-api-legacy", {
    build: { buildEnvironment: "V3", builder: "RAILPACK" },
    replicas: { "iad": 1 },
  });

  return project("pagor", {
    variables: { managed: false },
    resources: [pagorApiV2, pagorApi, pagorApiLegacy],
  });
});
