import { defineRailway, github, project, service } from "railway/iac";

export default defineRailway(() => {
  const edj = github("solutiono1/pagor");

  const edjApiV2 = service("edj-api-v2", {
    source: edj,
    build: { buildEnvironment: "V3", builder: "RAILPACK" },
    replicas: { "iad": 1 },
  });

  const edjApiV3 = service("edj-api-v3", {
    source: edj,
    build: {
      buildCommand: "go build -o bin/pagor ./cmd/edj",
      buildEnvironment: "V3",
      builder: "NIXPACKS",
    },
    start: "./bin/pagor",
    replicas: { "iad": 1 },
  });

  const edjApi = service("edj-api", {
    build: { buildEnvironment: "V3", builder: "RAILPACK" },
    replicas: { "iad": 1 },
  });

  return project("edj", {
    variables: { managed: false },
    resources: [edjApiV2, edjApiV3, edjApi],
  });
});
