# Agent Model Providers

Mantrixflow AI routes are not locked to Anthropic. The Next.js app chooses a model at runtime through `apps/arcyria-platform/lib/ai/model-provider.ts`.

This affects:

- Standalone Agent Platform chat.
- Public embedded agent chat.
- AI Pipeline Builder.
- Existing pipeline builder AI assistant.

## Global Switch

Use one provider for every AI route:

```bash
AI_MODEL_PROVIDER=ollama
AI_MODEL=qwen3.5:4b
OLLAMA_BASE_URL=http://localhost:11434/v1
```

Supported values:

```text
anthropic
gateway
openrouter
ollama
openai-compatible
```

Aliases:

- `claude` -> `anthropic`
- `local` -> `ollama`
- `vercel-gateway` or `ai-gateway` -> `gateway`
- `compatible` -> `openai-compatible`

## Per-Flow Overrides

Use these when demos should run agents locally, but pipeline creation should still use a hosted model.

Custom agents:

```bash
AGENT_AI_MODEL_PROVIDER=ollama
AGENT_AI_MODEL=qwen3.5:4b
```

AI Pipeline Builder:

```bash
AI_PIPELINE_BUILDER_MODEL_PROVIDER=openrouter
AI_PIPELINE_BUILDER_MODEL=meta-llama/llama-3.1-8b-instruct
```

Existing pipeline builder assistant:

```bash
PIPELINE_ASSISTANT_AI_MODEL_PROVIDER=anthropic
PIPELINE_ASSISTANT_AI_MODEL=claude-sonnet-4-20250514
```

The lookup order is:

1. Flow-specific provider and model.
2. Global `AI_MODEL_PROVIDER` and `AI_MODEL`.
3. Provider-specific defaults.

## Anthropic

```bash
AI_MODEL_PROVIDER=anthropic
ANTHROPIC_API_KEY=...
ANTHROPIC_MODEL=claude-sonnet-4-20250514
```

`ANTHROPIC_MODEL` is optional. If unset, the default is `claude-sonnet-4-20250514`.

## OpenRouter

```bash
AI_MODEL_PROVIDER=openrouter
OPENROUTER_API_KEY=...
OPENROUTER_MODEL=meta-llama/llama-3.1-8b-instruct
OPENROUTER_SITE_URL=http://localhost:3000
OPENROUTER_APP_NAME=Mantrixflow
```

You can also use `AI_MODEL` instead of `OPENROUTER_MODEL`.

OpenRouter model ids are provider/model strings. Pick a model that supports tool calls if the agent needs to query data.

## Ollama / Local Demo

Install and run Ollama locally, then pull a tool-call-capable chat model:

```bash
ollama list
ollama pull qwen3.5:4b
ollama serve
```

If `ollama list` already shows `qwen3.5:4b`, you can skip the pull step.

Configure the app:

```bash
AI_MODEL_PROVIDER=ollama
AI_MODEL=qwen3.5:4b
OLLAMA_BASE_URL=http://localhost:11434/v1
```

For agent-only local demo:

```bash
AGENT_AI_MODEL_PROVIDER=ollama
AGENT_AI_MODEL=qwen3.5:4b
OLLAMA_BASE_URL=http://localhost:11434/v1
```

Notes:

- Ollama's OpenAI-compatible endpoint is `/v1`, so keep the base URL as `http://localhost:11434/v1`.
- No API key is required by default.
- `qwen3.5:4b` is the recommended local default for an M3 Pro demo: still light, but more reliable than the tiny 0.8B model for tool calls and SQL generation.
- If the AI Pipeline Builder misses tool calls, move only that heavier flow to `qwen3.5:9b`, OpenRouter, or Anthropic.

## Vercel AI Gateway

```bash
AI_MODEL_PROVIDER=gateway
AI_GATEWAY_API_KEY=...
AI_GATEWAY_MODEL=anthropic/claude-sonnet-4
```

You can also use `AI_MODEL` instead of `AI_GATEWAY_MODEL`.

## Generic OpenAI-Compatible Endpoint

Use this for LM Studio, vLLM, LiteLLM, or another proxy that exposes OpenAI-compatible chat completions.

```bash
AI_MODEL_PROVIDER=openai-compatible
OPENAI_COMPATIBLE_PROVIDER_NAME=local-proxy
OPENAI_COMPATIBLE_BASE_URL=http://localhost:8000/v1
OPENAI_COMPATIBLE_MODEL=qwen2.5-coder-7b-instruct
OPENAI_COMPATIBLE_API_KEY=
```

`OPENAI_COMPATIBLE_API_KEY` is optional.

## Recommended Demo Settings

Lowest-cost full demo:

```bash
AI_MODEL_PROVIDER=ollama
AI_MODEL=qwen3.5:4b
OLLAMA_BASE_URL=http://localhost:11434/v1
```

Safer hybrid demo:

```bash
AGENT_AI_MODEL_PROVIDER=ollama
AGENT_AI_MODEL=qwen3.5:4b
AI_PIPELINE_BUILDER_MODEL_PROVIDER=anthropic
ANTHROPIC_MODEL=claude-sonnet-4-20250514
OLLAMA_BASE_URL=http://localhost:11434/v1
```

The hybrid setup keeps pipeline creation on a stronger hosted model while making repeated agent Q&A cheap.

## M3 Pro Recommendation

For an Apple M3 Pro with 18 GB unified memory:

```bash
AGENT_AI_MODEL_PROVIDER=ollama
AGENT_AI_MODEL=qwen3.5:4b
PIPELINE_ASSISTANT_AI_MODEL_PROVIDER=ollama
PIPELINE_ASSISTANT_AI_MODEL=qwen3.5:4b
AI_PIPELINE_BUILDER_MODEL_PROVIDER=ollama
AI_PIPELINE_BUILDER_MODEL=qwen3.5:4b
OLLAMA_BASE_URL=http://localhost:11434/v1
```

Use `qwen3.5:4b` for the full local demo. Move the AI Pipeline Builder to `qwen3.5:9b` or a hosted model only if you need stronger multi-step schema/table/tool reasoning.
