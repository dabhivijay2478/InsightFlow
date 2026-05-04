import { useEffect, useRef, useState } from "react";
import {
	Bar,
	BarChart,
	CartesianGrid,
	Line,
	LineChart,
	ResponsiveContainer,
	XAxis,
	YAxis,
} from "recharts";

export interface MantrixAgentProps {
	agentId: string;
	apiBaseUrl?: string;
	title?: string;
	description?: string;
}

interface Message {
	role: "user" | "assistant";
	text: string;
}

interface ChartData {
	type: "bar" | "line";
	labels: string[];
	datasets: Array<{ label: string; data: number[] }>;
}

function sessionIdForAgent(agentId: string) {
	if (typeof window === "undefined") return "";
	const key = `mantrix-agent-session:${agentId}`;
	const existing = localStorage.getItem(key);
	if (existing) return existing;
	const next = `sess_${Math.random().toString(36).slice(2)}_${Date.now().toString(36)}`;
	localStorage.setItem(key, next);
	return next;
}

function parseChart(text: string): ChartData | null {
	const match = text.match(/<chart_data>([\s\S]*?)<\/chart_data>/);
	if (!match) return null;
	try {
		return JSON.parse(match[1]) as ChartData;
	} catch {
		return null;
	}
}

function stripChart(text: string) {
	return text.replace(/<chart_data>[\s\S]*?<\/chart_data>/g, "").trim();
}

function textFromStreamChunk(value: unknown): string {
	if (!value || typeof value !== "object") return "";
	const chunk = value as Record<string, unknown>;
	const type = typeof chunk.type === "string" ? chunk.type : "";
	if (type.includes("text")) {
		if (typeof chunk.text === "string") return chunk.text;
		if (typeof chunk.delta === "string") return chunk.delta;
	}
	if (Array.isArray(chunk.parts)) {
		return chunk.parts
			.map((part) => textFromStreamChunk(part))
			.filter(Boolean)
			.join("");
	}
	if (typeof chunk.content === "string" && type.includes("delta")) {
		return chunk.content;
	}
	return "";
}

function assistantTextFromStream(raw: string): string {
	const lines = raw.split(/\r?\n/);
	const parts: string[] = [];
	for (const line of lines) {
		const trimmed = line.trim();
		if (!trimmed || trimmed === "data: [DONE]") continue;
		if (/^\d+:/.test(trimmed)) {
			const payload = trimmed.slice(trimmed.indexOf(":") + 1);
			try {
				const parsed = JSON.parse(payload) as unknown;
				if (typeof parsed === "string") parts.push(parsed);
				else parts.push(textFromStreamChunk(parsed));
			} catch {
				// Ignore non-text protocol frames.
			}
			continue;
		}
		if (trimmed.startsWith("data:")) {
			const payload = trimmed.slice(5).trim();
			try {
				parts.push(textFromStreamChunk(JSON.parse(payload) as unknown));
			} catch {
				if (payload && payload !== "[DONE]") parts.push(payload);
			}
		}
	}
	return parts.filter(Boolean).join("") || raw.trim();
}

function InlineChart({ chart }: { chart: ChartData }) {
	const dataset = chart.datasets[0];
	const rows = chart.labels.map((label, index) => ({
		label,
		value: dataset?.data[index] ?? 0,
	}));
	const Chart = chart.type === "line" ? LineChart : BarChart;
	return (
		<div style={{ height: 180, marginTop: 12 }}>
			<ResponsiveContainer width="100%" height="100%">
				<Chart data={rows}>
					<CartesianGrid strokeDasharray="3 3" />
					<XAxis dataKey="label" tick={{ fontSize: 10 }} />
					<YAxis tick={{ fontSize: 10 }} />
					{chart.type === "line" ? (
						<Line
							dataKey="value"
							stroke="#0f766e"
							strokeWidth={2}
							dot={false}
						/>
					) : (
						<Bar dataKey="value" fill="#0f766e" radius={[4, 4, 0, 0]} />
					)}
				</Chart>
			</ResponsiveContainer>
		</div>
	);
}

export function useMantrixAgent(agentId: string, apiBaseUrl = "") {
	const [sessionId, setSessionId] = useState("");
	const [messages, setMessages] = useState<Message[]>([]);
	const [isLoading, setIsLoading] = useState(false);
	useEffect(() => {
		setSessionId(sessionIdForAgent(agentId));
	}, [agentId]);
	const sendMessage = async (text: string) => {
		const trimmed = text.trim();
		if (!trimmed || isLoading || !sessionId) return;
		const nextMessages = [
			...messages,
			{ role: "user" as const, text: trimmed },
		];
		setMessages(nextMessages);
		setIsLoading(true);
		try {
			const response = await fetch(`${apiBaseUrl}/api/agents/${agentId}/chat`, {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					sessionId,
					messages: nextMessages.map((message, index) => ({
						id: `${message.role}-${index}`,
						role: message.role,
						parts: [{ type: "text", text: message.text }],
					})),
				}),
			});
			if (!response.ok) {
				throw new Error(await response.text());
			}
			const raw = await response.text();
			setMessages([
				...nextMessages,
				{ role: "assistant", text: assistantTextFromStream(raw) },
			]);
		} finally {
			setIsLoading(false);
		}
	};
	return { messages, sendMessage, isLoading };
}

export function MantrixAgent({
	agentId,
	apiBaseUrl = "",
	title = "Pipeline Agent",
	description = "Ask about published pipeline data",
}: MantrixAgentProps) {
	const [open, setOpen] = useState(false);
	const [input, setInput] = useState("");
	const { messages, sendMessage, isLoading } = useMantrixAgent(
		agentId,
		apiBaseUrl,
	);
	const endRef = useRef<HTMLDivElement | null>(null);
	useEffect(() => {
		endRef.current?.scrollIntoView({ behavior: "smooth" });
	}, [messages]);

	return (
		<div
			style={{
				position: "fixed",
				right: 20,
				bottom: 20,
				zIndex: 2147483000,
				fontFamily: "Inter, system-ui, sans-serif",
			}}
		>
			{open && (
				<div
					style={{
						width: 380,
						height: 560,
						background: "#fff",
						border: "1px solid #d4d4d8",
						borderRadius: 8,
						boxShadow: "0 18px 70px rgba(15,23,42,.22)",
						display: "flex",
						flexDirection: "column",
						overflow: "hidden",
						marginBottom: 12,
					}}
				>
					<div
						style={{
							padding: 14,
							borderBottom: "1px solid #eee",
							background: "#fafafa",
						}}
					>
						<div style={{ fontWeight: 700, color: "#111827" }}>{title}</div>
						<div style={{ marginTop: 2, fontSize: 12, color: "#6b7280" }}>
							{description}
						</div>
					</div>
					<div style={{ flex: 1, overflowY: "auto", padding: 12 }}>
						{messages.map((message, index) => {
							const chart = parseChart(message.text);
							return (
								<div
									key={`${message.role}-${index}`}
									style={{
										marginBottom: 10,
										textAlign: message.role === "user" ? "right" : "left",
									}}
								>
									<div
										style={{
											display: "inline-block",
											maxWidth: "85%",
											borderRadius: 8,
											padding: "8px 10px",
											background:
												message.role === "user" ? "#111827" : "#f4f4f5",
											color: message.role === "user" ? "#fff" : "#111827",
											fontSize: 13,
											lineHeight: 1.45,
										}}
									>
										{stripChart(message.text)}
										{chart && <InlineChart chart={chart} />}
									</div>
								</div>
							);
						})}
						<div ref={endRef} />
					</div>
					<form
						style={{
							display: "flex",
							gap: 8,
							padding: 12,
							borderTop: "1px solid #eee",
						}}
						onSubmit={(event) => {
							event.preventDefault();
							void sendMessage(input);
							setInput("");
						}}
					>
						<input
							value={input}
							onChange={(event) => setInput(event.target.value)}
							placeholder="Ask a question..."
							style={{
								flex: 1,
								border: "1px solid #ddd",
								borderRadius: 6,
								padding: "9px 10px",
							}}
						/>
						<button
							type="submit"
							disabled={isLoading || !input.trim()}
							style={{
								border: 0,
								borderRadius: 6,
								background: "#0f766e",
								color: "#fff",
								padding: "0 12px",
							}}
						>
							Send
						</button>
					</form>
				</div>
			)}
			<button
				type="button"
				onClick={() => setOpen((value) => !value)}
				style={{
					height: 56,
					width: 56,
					borderRadius: "50%",
					border: 0,
					background: "#0f766e",
					color: "#fff",
					boxShadow: "0 10px 30px rgba(0,0,0,.22)",
					cursor: "pointer",
					fontWeight: 700,
				}}
			>
				{open ? "X" : "M"}
			</button>
		</div>
	);
}
