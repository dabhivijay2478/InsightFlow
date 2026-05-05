import "dotenv/config";
import { App, type MessageEvent } from "@slack/bolt";

// ---------------------------------------------------------------------------
// Env
// ---------------------------------------------------------------------------
const SLACK_BOT_TOKEN = process.env.SLACK_BOT_TOKEN ?? "";
const SLACK_APP_TOKEN = process.env.SLACK_APP_TOKEN ?? "";
const SLACK_SIGNING_SECRET = process.env.SLACK_SIGNING_SECRET ?? "";
const AGENT_API_BASE_URL = (process.env.AGENT_API_BASE_URL ?? "").replace(/\/$/, "");

if (!SLACK_BOT_TOKEN || !SLACK_APP_TOKEN || !SLACK_SIGNING_SECRET) {
  console.error(
    "Missing required env vars: SLACK_BOT_TOKEN, SLACK_APP_TOKEN, SLACK_SIGNING_SECRET",
  );
  process.exit(1);
}

// ---------------------------------------------------------------------------
// AI SDK data-stream parser — same logic as the browser SDK
// ---------------------------------------------------------------------------
function parseDataStream(raw: string): string {
  const lines = raw.split(/\r?\n/);
  const parts: string[] = [];
  for (const line of lines) {
    const trimmed = line.trim();
    if (!trimmed) continue;
    const colonIdx = trimmed.indexOf(":");
    if (colonIdx === -1) continue;
    const prefix = trimmed.slice(0, colonIdx);
    const payload = trimmed.slice(colonIdx + 1);
    if (prefix !== "0") continue;
    try {
      const parsed: unknown = JSON.parse(payload);
      if (typeof parsed === "string") parts.push(parsed);
    } catch {
      // ignore
    }
  }
  return parts.join("").trim();
}

// ---------------------------------------------------------------------------
// Session store  (in-memory per channel thread)
// ---------------------------------------------------------------------------
interface Session {
  agentKey: string;
  sessionId: string;
  history: Array<{ role: "user" | "assistant"; text: string }>;
}

const sessions = new Map<string, Session>();

function threadKey(channelId: string, threadTs: string) {
  return `${channelId}:${threadTs}`;
}

function getOrCreateSession(channelId: string, threadTs: string, agentKey: string): Session {
  const key = threadKey(channelId, threadTs);
  if (!sessions.has(key)) {
    sessions.set(key, {
      agentKey,
      sessionId: `slack_${Math.random().toString(36).slice(2)}_${Date.now().toString(36)}`,
      history: [],
    });
  }
  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
  return sessions.get(key)!;
}

// ---------------------------------------------------------------------------
// Query the agent
// ---------------------------------------------------------------------------
async function queryAgent(session: Session, userText: string): Promise<string> {
  const userMsg = { role: "user" as const, text: userText };
  const allMessages = [...session.history, userMsg];

  const response = await fetch(`${AGENT_API_BASE_URL}/api/agents/${session.agentKey}/chat`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      sessionId: session.sessionId,
      messages: allMessages.map((m, i) => ({
        id: `${m.role}-${i}`,
        role: m.role,
        parts: [{ type: "text", text: m.text }],
      })),
    }),
  });

  if (!response.ok) {
    const err = await response.text();
    throw new Error(`Agent API error ${response.status}: ${err.slice(0, 200)}`);
  }

  const raw = await response.text();
  const assistantText = parseDataStream(raw) || "*(no response)*";

  session.history.push(userMsg, { role: "assistant", text: assistantText });
  if (session.history.length > 40) {
    session.history = session.history.slice(-40);
  }

  return assistantText;
}

// ---------------------------------------------------------------------------
// Parse agent key from message: "@bot agentKey: some_key  question..."
// or from env SLACK_DEFAULT_AGENT_KEY
// ---------------------------------------------------------------------------
const DEFAULT_AGENT_KEY = process.env.SLACK_DEFAULT_AGENT_KEY ?? "";

function extractAgentKey(text: string): { agentKey: string; query: string } {
  const match = text.match(/agent[_\s]key[:\s]+([a-zA-Z0-9_-]+)\s*(.*)/is);
  if (match) {
    return { agentKey: match[1].trim(), query: (match[2] ?? "").trim() };
  }
  return { agentKey: DEFAULT_AGENT_KEY, query: text };
}

// ---------------------------------------------------------------------------
// Bolt App — Socket Mode (no public URL needed)
// ---------------------------------------------------------------------------
const app = new App({
  token: SLACK_BOT_TOKEN,
  appToken: SLACK_APP_TOKEN,
  signingSecret: SLACK_SIGNING_SECRET,
  socketMode: true,
});

// Strip Slack user mentions like <@U12345>
function stripMentions(text: string) {
  return text.replace(/<@[A-Z0-9]+>/g, "").trim();
}

// Respond to app_mention events
app.event("app_mention", async ({ event, say, client }) => {
  const msgEvent = event as MessageEvent & { thread_ts?: string };
  const channelId = msgEvent.channel;
  const threadTs = msgEvent.thread_ts ?? msgEvent.ts;
  const rawText = stripMentions(msgEvent.text ?? "");

  if (!rawText) {
    await say({ text: "Hi! Ask me anything about your pipeline data.", thread_ts: threadTs });
    return;
  }

  const { agentKey, query } = extractAgentKey(rawText);
  if (!agentKey) {
    await say({
      text: "No agent key configured. Set `SLACK_DEFAULT_AGENT_KEY` or include `agent_key: <key>` in your message.",
      thread_ts: threadTs,
    });
    return;
  }

  const session = getOrCreateSession(channelId, threadTs, agentKey);

  // Show a "thinking" reaction
  await client.reactions.add({ channel: channelId, timestamp: msgEvent.ts, name: "hourglass_flowing_sand" }).catch(() => undefined);

  let answer: string;
  try {
    answer = await queryAgent(session, query);
  } catch (err) {
    answer = `Error: ${err instanceof Error ? err.message : String(err)}`;
  }

  // Remove thinking reaction
  await client.reactions.remove({ channel: channelId, timestamp: msgEvent.ts, name: "hourglass_flowing_sand" }).catch(() => undefined);

  await say({ text: answer, thread_ts: threadTs });
});

// Also respond to DMs (message.im)
app.message(async ({ message, say }) => {
  const msg = message as MessageEvent & { thread_ts?: string; bot_id?: string };
  if (msg.bot_id) return; // Ignore bot messages
  const rawText = stripMentions(msg.text ?? "");
  if (!rawText) return;

  const channelId = msg.channel;
  const threadTs = msg.thread_ts ?? msg.ts;
  const { agentKey, query } = extractAgentKey(rawText);
  if (!agentKey) {
    await say({ text: "Set `SLACK_DEFAULT_AGENT_KEY` env var to use this bot without specifying a key." });
    return;
  }

  const session = getOrCreateSession(channelId, threadTs, agentKey);
  let answer: string;
  try {
    answer = await queryAgent(session, query);
  } catch (err) {
    answer = `Error: ${err instanceof Error ? err.message : String(err)}`;
  }
  await say({ text: answer, thread_ts: threadTs });
});

// ---------------------------------------------------------------------------
// Start
// ---------------------------------------------------------------------------
(async () => {
  const port = Number(process.env.PORT ?? 3010);
  await app.start(port);
  console.log(`⚡ Mantrixflow Slack bot running on port ${port}`);
  console.log(`   Default agent key: ${DEFAULT_AGENT_KEY || "(none — pass agent_key: in message)"}`);
  console.log(`   Agent API base:    ${AGENT_API_BASE_URL || "(none — set AGENT_API_BASE_URL)"}`);
})();
