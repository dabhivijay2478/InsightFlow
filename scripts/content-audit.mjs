#!/usr/bin/env node

import { createHash } from "node:crypto";
import { existsSync, mkdirSync, readFileSync, readdirSync, statSync, writeFileSync } from "node:fs";
import { dirname, extname, join, relative, resolve } from "node:path";

const root = resolve(import.meta.dirname, "..");
const outputDir = join(root, "md-docs", "content-audit", "generated");
const generatedAt = new Date().toISOString();
const isAfterSnapshot = process.argv.includes("--after");
const outputName = (name) => (isAfterSnapshot ? `after-${name}` : name);

const roots = {
  website: ["apps/arcyria-website/app", "apps/arcyria-website/components", "apps/arcyria-website/lib"],
  application: ["apps/arcyria-platform/app", "apps/arcyria-platform/components", "apps/arcyria-platform/lib"],
  documentation: ["apps/arcyria-docs"],
  internal_reference: ["md-docs"],
};

const ignored = [
  "/node_modules/",
  "/.git/",
  "/.next/",
  "/content-audit/generated/",
  "/components/ui/",
  "/apps/arcyria-platform/app/api/",
  "/__tests__/",
];

function walk(path, extensions) {
  const absolute = join(root, path);
  if (!existsSync(absolute)) return [];
  const files = [];
  const visit = (current) => {
    for (const entry of readdirSync(current, { withFileTypes: true })) {
      const full = join(current, entry.name);
      const normalized = `/${relative(root, full).replaceAll("\\", "/")}`;
      if (ignored.some((part) => normalized.includes(part))) continue;
      if (entry.isDirectory()) visit(full);
      else if (extensions.has(extname(entry.name))) files.push(full);
    }
  };
  visit(absolute);
  return files;
}

function cleanText(value) {
  return value
    .replace(/<[^>]+>/g, " ")
    .replace(/\{[^}]+\}/g, " ")
    .replace(/[`*_>#\[\]]/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function titleFor(file, source) {
  const metadataTitle = source.match(/^title\s*:\s*(?:["'`]([^"'`]+)["'`]|([^\n]+))$/m);
  const parsedMetadataTitle = metadataTitle?.[1] || metadataTitle?.[2]?.trim();
  if (parsedMetadataTitle) return parsedMetadataTitle;
  if (file.endsWith(".md") || file.endsWith(".mdx")) {
    return source.match(/^#\s+(.+)$/m)?.[1]?.trim() || "Untitled article";
  }
  const heading = source.match(/<h1[^>]*>([\s\S]*?)<\/h1>/i)?.[1];
  if (heading) {
    const cleaned = cleanText(heading);
    if (cleaned) return cleaned;
  }
  const basename = file.split("/").at(-1)?.replace(/\.(tsx?|md)$/, "") || "Untitled";
  return basename
    .replace(/[-_]/g, " ")
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function routeFor(file, surface) {
  if (!file.endsWith("/page.tsx") && !file.endsWith("/layout.tsx")) return "";
  const routeRoot = surface === "website" ? "apps/arcyria-website/app" : "apps/arcyria-platform/app";
  let route = relative(join(root, routeRoot), dirname(file)).replaceAll("\\", "/");
  route = route
    .split("/")
    .filter((segment) => !segment.startsWith("("))
    .join("/");
  return route ? `/${route}` : "/";
}

function purposeFor(location, kind, title) {
  if (kind === "article") {
    if (location.includes("/manual-testing/")) return "Internal manual QA procedure or acceptance evidence";
    if (location.includes("/test-logs/")) return "Generated test evidence";
    if (/setup|deployment|runbook|observability|template|email/i.test(location)) return "Internal setup or operations guidance";
    if (/plan|implementation|prompts/i.test(location)) return "Implementation plan or engineering reference";
    return `Technical guidance: ${title}`;
  }
  if (kind === "route") return `Primary content for ${title}`;
  if (kind === "layout") return "Shared navigation, metadata, or layout content";
  if (kind === "dialog") return "Contextual dialog, sheet, or confirmation content";
  if (kind === "tooltip") return "Contextual tooltip content";
  return `Reusable content for ${title}`;
}

function targetFor(surface, location) {
  if (surface === "website") return "Prospective and evaluating customers";
  if (surface === "application") {
    if (location.includes("/auth/") || location.includes("/onboarding/")) return "New or returning workspace users";
    if (location.includes("/billing/")) return "Workspace owners and billing administrators";
    if (location.includes("/team/") || location.includes("/settings/")) return "Workspace owners and administrators";
    return "Authenticated MantrixFlow users";
  }
  if (location.includes("/manual-testing/") || location.includes("/test-logs/") || /plan|implementation|deployment|runbook|setup/i.test(location)) {
    return "MantrixFlow engineering and operations";
  }
  return "Technical product users and administrators";
}

function documentationAction(location) {
  if (location.includes("/test-logs/")) return "ARCHIVE";
  if (location.includes("/manual-testing/")) return "MOVE TO ADVANCED SECTION";
  if (/e2e-report|e2e-summary|testcase|testing-local/i.test(location)) return "ARCHIVE";
  if (/plan|implementation|ai-prompts/i.test(location)) return "ARCHIVE";
  if (/deployment|runbook|setup|template|observability|billing-dodo|dodo-|posthog|betterstack|tigris|aws-ses|autosend/i.test(location)) {
    return "MOVE TO ADVANCED SECTION";
  }
  if (location.includes("/agents/")) return "REQUIRES PRODUCT CONFIRMATION";
  if (/slack-guide|salesforce/i.test(location)) return "REQUIRES PRODUCT CONFIRMATION";
  return "KEEP AND IMPROVE";
}

function websiteAction(location) {
  if (location.includes("/agents/")) return "REQUIRES PRODUCT CONFIRMATION";
  if (location.includes("/integrations/")) return "KEEP AND IMPROVE";
  if (location.endsWith("marketing-content.ts")) return "SHORTEN";
  if (location.includes("/components/ui/")) return "KEEP";
  return "KEEP AND IMPROVE";
}

function appAction(location, source) {
  if (/mock|placeholder|coming soon/i.test(source)) return "REQUIRES PRODUCT CONFIRMATION";
  if (/Dialog|Sheet|Drawer/.test(source)) return "SHORTEN";
  return "KEEP AND IMPROVE";
}

function kindFor(file, source) {
  if (file.endsWith(".md") || file.endsWith(".mdx")) return "article";
  if (file.endsWith("/page.tsx")) return "route";
  if (file.endsWith("/layout.tsx")) return "layout";
  if (/Dialog(Content|Title|Description)|AlertDialog(Content|Title|Description)|Sheet(Content|Title|Description)|Drawer(Content|Title|Description)/.test(source)) return "dialog";
  if (/Tooltip(Content|Trigger)|<Tooltip/.test(source)) return "tooltip";
  if (file.endsWith(".ts")) return "content-source";
  return "component";
}

function hasAuditableContent(file, source) {
  if (file.endsWith("docs.json")) return true;
  if (file.endsWith(".json")) return false;
  if (!file.endsWith(".ts")) return true;
  if (file.endsWith("marketing-content.ts")) return true;
  return /(?:throw new Error|message\s*:|description\s*:|title\s*:|label\s*:|toast\.|userMessage|errorMessage)/.test(source);
}

function detectFlags(source) {
  return {
    dialogs: (source.match(/<(?:Dialog|AlertDialog|Sheet|Drawer)(?:Content|Title|Description)\b/g) || []).length,
    tooltips: (source.match(/<TooltipContent\b/g) || []).length,
    notifications: (source.match(/(?:toast|sonnerToast)\.(?:success|error|info|warning|promise)\b/g) || []).length,
    emptyStates: (source.match(/(?:EmptyState|No [A-Za-z][^\n<{]{0,80}(?:yet|found|available))/g) || []).length,
    loadingStates: (source.match(/(?:LoadingState|Skeleton|isLoading)/g) || []).length,
    claims: (source.match(/\b(?:live|production-ready|beta|planned|roadmap|coming soon|unlimited|free forever|SLA|SSO|HIPAA|SOC 2|GDPR|\d+[+%])\b/gi) || []).length,
  };
}

function csvCell(value) {
  const text = String(value ?? "");
  return /[",\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
}

function toCsv(rows) {
  const headers = Object.keys(rows[0] || {});
  return [headers.join(","), ...rows.map((row) => headers.map((header) => csvCell(row[header])).join(","))].join("\n") + "\n";
}

const inventory = [];
const baseline = [];
const sourceByFile = new Map();

for (const [surface, paths] of Object.entries(roots)) {
  const extensions = surface === "documentation"
    ? new Set([".md", ".mdx", ".json"])
    : surface === "internal_reference"
      ? new Set([".md"])
      : new Set([".ts", ".tsx"]);
  const files = paths.flatMap((path) => walk(path, extensions));
  for (const file of [...new Set(files)].sort()) {
    const source = readFileSync(file, "utf8");
    const location = relative(root, file).replaceAll("\\", "/");
    if (surface === "internal_reference" && location.startsWith("md-docs/content-audit/")) continue;
    if (surface === "documentation" && file.endsWith("package-lock.json")) continue;
    if (!hasAuditableContent(file, source)) continue;
    const kind = kindFor(file, source);
    const title = titleFor(file, source);
    const flags = detectFlags(source);
    const words = source.trim() ? source.trim().split(/\s+/).length : 0;
    const lines = source.split(/\r?\n/).length;
    const action = surface === "documentation"
      ? (location.endsWith("docs.json") ? "KEEP AND IMPROVE" : documentationAction(location))
      : surface === "internal_reference"
        ? documentationAction(location)
      : surface === "website"
        ? websiteAction(location)
        : appAction(location, source);
    const uncertain = action === "REQUIRES PRODUCT CONFIRMATION";
    inventory.push({
      surface,
      kind,
      location,
      route: routeFor(file, surface),
      current_title: title,
      content_purpose: purposeFor(location, kind, title),
      target_user: targetFor(surface, location),
      accurate: uncertain ? "Unverified" : "Verify during page review",
      duplicated: "See duplicate-content.csv",
      outdated: uncertain ? "Possible" : "Verify during page review",
      too_detailed: words > 1200 ? "Likely" : "Review",
      unclear: "Review in context",
      needed_in_current_workflow: surface === "application" ? "Yes or contextual" : "Not workflow-bound",
      belongs_in: surface === "internal_reference"
        ? "Internal engineering documentation"
        : surface,
      recommended_action: action,
      source_words: words,
      source_lines: lines,
      dialog_markers: flags.dialogs,
      tooltip_markers: flags.tooltips,
      notification_markers: flags.notifications,
      empty_state_markers: flags.emptyStates,
      loading_state_markers: flags.loadingStates,
      claim_markers: flags.claims,
    });
    baseline.push({
      surface,
      location,
      bytes: Buffer.byteLength(source),
      words,
      lines,
      sha256: createHash("sha256").update(source).digest("hex"),
    });
    sourceByFile.set(location, source);
  }
}

const paragraphs = new Map();
for (const [location, source] of sourceByFile) {
  const candidates = [];
  if (location.endsWith(".md") || location.endsWith(".mdx")) {
    candidates.push(...source.split(/\n\s*\n/));
  } else {
    const quoted = source.matchAll(/(["'`])((?:(?!\1)[\s\S]){35,}?)\1/g);
    for (const match of quoted) candidates.push(match[2]);
    const jsx = source.matchAll(/>([^<>{}\n][^<>{}]{35,})</g);
    for (const match of jsx) candidates.push(match[1]);
  }
  for (const candidate of candidates) {
    const display = cleanText(candidate);
    if (display.split(/\s+/).length < 8) continue;
    if (!location.endsWith(".md") && !location.endsWith(".mdx") && /\b(?:className|const|return|undefined|state|aria|function|module|import|true|false|rgba|async|await)\b/i.test(display)) continue;
    if (/^(?:https?:|[\w-]+\/|[a-z-]+\s){1,}/i.test(display) && !display.includes(".")) continue;
    const normalized = display.toLowerCase().replace(/[^a-z0-9 ]/g, "").replace(/\s+/g, " ").trim();
    if (normalized.length < 45) continue;
    const locations = paragraphs.get(normalized) || new Set();
    locations.add(location);
    paragraphs.set(normalized, locations);
  }
}

const duplicates = [...paragraphs.entries()]
  .filter(([, locations]) => locations.size > 1)
  .map(([content, locations]) => ({
    occurrences: locations.size,
    content,
    locations: [...locations].sort().join(" | "),
    recommended_action: "MERGE WITH ANOTHER PAGE",
  }))
  .sort((a, b) => b.occurrences - a.occurrences || b.content.length - a.content.length);

const websiteRoutes = new Set(
  inventory
    .filter((item) => item.surface === "website" && item.kind === "route")
    .map((item) => item.route),
);
const applicationRoutes = new Set(
  inventory
    .filter((item) => item.surface === "application" && item.kind === "route")
    .map((item) => item.route),
);
const documentationRoutes = new Set(
  inventory
    .filter((item) => item.surface === "documentation" && item.kind === "article" && item.location.endsWith(".mdx"))
    .map((item) => {
      const value = item.location
        .replace(/^apps\/arcyria-docs\//, "")
        .replace(/\.mdx$/, "")
        .replace(/\/index$/, "");
      return value ? `/${value}` : "/";
    }),
);

function routeExists(route, routes) {
  if (routes.has(route)) return true;
  for (const candidate of routes) {
    const pattern = `^${candidate.replace(/[.*+?^${}()|[\]\\]/g, "\\$&").replace(/\\\[[^/]+\\\]/g, "[^/]+")}$`;
    if (new RegExp(pattern).test(route)) return true;
  }
  return false;
}

const links = [];
for (const [location, source] of sourceByFile) {
  const candidates = [];
  for (const match of source.matchAll(/(?:href\s*=\s*["']|\]\()([^"')#?]+(?:#[^"')]*)?)["')]/g)) {
    candidates.push(match[1]);
  }
  for (const href of candidates) {
    if (!href || href.startsWith("#") || href.includes("${") || href.includes("{")) continue;
    let status = "External — network check required";
    if (/^(?:mailto:|tel:)/.test(href)) status = "External contact link — syntax only";
    if (href.startsWith("/")) {
      const route = href.split(/[?#]/)[0].replace(/\/$/, "") || "/";
      if (location.startsWith("apps/arcyria-docs/") && href.startsWith("/images/")) {
        status = existsSync(join(root, "apps/arcyria-docs", route)) ? "OK" : "Missing local target";
      } else if (location.startsWith("apps/arcyria-docs/")) {
        status = routeExists(route, documentationRoutes) ? "OK" : "Missing documentation route";
      } else if (location.endsWith(".md") && existsSync(join(root, route))) {
        status = "OK";
      } else {
        const routes = location.startsWith("apps/arcyria-platform/") ? applicationRoutes : websiteRoutes;
        status = routeExists(route, routes) ? "OK" : "Missing route";
      }
    } else if (!/^(?:https?:|mailto:|tel:)/.test(href) && (location.endsWith(".md") || location.endsWith(".mdx"))) {
      const pathOnly = href.split("#")[0];
      const resolved = resolve(root, dirname(location), pathOnly);
      status = !pathOnly || existsSync(resolved) ? "OK" : "Missing local target";
    }
    links.push({ location, href, status });
  }
}

mkdirSync(outputDir, { recursive: true });
writeFileSync(join(outputDir, outputName("content-inventory.csv")), toCsv(inventory));
writeFileSync(join(outputDir, outputName("duplicate-content.csv")), toCsv(duplicates.length ? duplicates : [{ occurrences: 0, content: "", locations: "", recommended_action: "" }]));
writeFileSync(join(outputDir, outputName("link-inventory.csv")), toCsv(links.length ? links : [{ location: "", href: "", status: "" }]));
writeFileSync(join(outputDir, outputName("baseline-manifest.json")), `${JSON.stringify({ generatedAt, files: baseline }, null, 2)}\n`);

const summary = {
  generatedAt,
  files: inventory.length,
  bySurface: Object.fromEntries(Object.keys(roots).map((surface) => [surface, inventory.filter((item) => item.surface === surface).length])),
  routes: inventory.filter((item) => item.kind === "route").length,
  layouts: inventory.filter((item) => item.kind === "layout").length,
  contentComponents: inventory.filter((item) => ["component", "dialog", "tooltip"].includes(item.kind)).length,
  dialogs: inventory.filter((item) => item.dialog_markers > 0).length,
  tooltips: inventory.filter((item) => item.tooltip_markers > 0).length,
  articles: inventory.filter((item) => item.kind === "article").length,
  exactDuplicateBlocks: duplicates.length,
  links: links.length,
  questionableLinks: links.filter((link) => !["OK", "External — network check required", "External contact link — syntax only"].includes(link.status)).length,
  externalLinksRequiringNetworkCheck: links.filter((link) => link.status === "External — network check required").length,
  actions: Object.fromEntries(
    [...new Set(inventory.map((item) => item.recommended_action))].sort().map((action) => [action, inventory.filter((item) => item.recommended_action === action).length]),
  ),
};
writeFileSync(join(outputDir, outputName("summary.json")), `${JSON.stringify(summary, null, 2)}\n`);
console.log(JSON.stringify(summary, null, 2));
