#!/usr/bin/env node

import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { dirname, extname, relative, resolve } from "node:path";

const repositoryRoot = resolve(import.meta.dirname, "..");
const ignoredDirectories = new Set([
  ".git",
  ".agents",
  ".codex",
  ".cursor",
  ".next",
  ".venv",
  "node_modules",
  "graphify-out",
]);
const markdownExtensions = new Set([".md", ".mdx"]);

function walk(directory) {
  const files = [];
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    if (entry.isDirectory() && ignoredDirectories.has(entry.name)) continue;
    const path = resolve(directory, entry.name);
    if (entry.isDirectory()) files.push(...walk(path));
    else if (markdownExtensions.has(extname(entry.name))) files.push(path);
  }
  return files;
}

function localTarget(rawTarget) {
  let target = rawTarget.trim();
  if (target.startsWith("<") && target.endsWith(">")) {
    target = target.slice(1, -1);
  } else {
    target = target.split(/\s+["']/)[0];
  }
  if (
    !target ||
    target.startsWith("#") ||
    target.startsWith("/") ||
    /^[a-z][a-z\d+.-]*:/i.test(target) ||
    target.includes("{") ||
    target.includes("}")
  ) {
    return null;
  }
  const withoutAnchor = target.split("#", 1)[0].split("?", 1)[0];
  if (!withoutAnchor) return null;
  try {
    return decodeURIComponent(withoutAnchor.replaceAll("\\ ", " "));
  } catch {
    return withoutAnchor;
  }
}

function lineNumber(source, index) {
  return source.slice(0, index).split("\n").length;
}

const failures = [];
let checkedLinks = 0;
const files = walk(repositoryRoot);
for (const file of files) {
  const relativeFile = relative(repositoryRoot, file).replaceAll("\\", "/");
  if (relativeFile.startsWith("md-docs/content-audit/generated/")) continue;
  const source = readFileSync(file, "utf8");
  const linkPattern = /!?\[[^\]]*\]\(([^)]+)\)/g;
  for (const match of source.matchAll(linkPattern)) {
    const target = localTarget(match[1]);
    if (!target) continue;
    checkedLinks += 1;
    const resolvedTarget = resolve(dirname(file), target);
    if (!existsSync(resolvedTarget)) {
      failures.push(
        `${relativeFile}:${lineNumber(source, match.index ?? 0)} -> ${target}`,
      );
      continue;
    }
    if (statSync(resolvedTarget).isDirectory()) {
      const hasIndex = ["README.md", "index.md", "index.mdx"].some((name) =>
        existsSync(resolve(resolvedTarget, name)),
      );
      if (!hasIndex) {
        failures.push(
          `${relativeFile}:${lineNumber(source, match.index ?? 0)} -> ${target} (directory has no README/index)`,
        );
      }
    }
  }
}

if (failures.length > 0) {
  console.error(`Broken local Markdown links (${failures.length}):`);
  for (const failure of failures) console.error(`- ${failure}`);
  process.exitCode = 1;
} else {
  console.log(
    `Markdown links valid: ${checkedLinks} local targets across ${files.length} files.`,
  );
}
