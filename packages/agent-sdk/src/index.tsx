import React, { useCallback, useEffect, useRef, useState } from "react";
import {
	Bar,
	BarChart,
	CartesianGrid,
	Line,
	LineChart,
	Pie,
	PieChart,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts";

export interface MantrixAgentProps {
	agentId: string;
	apiBaseUrl?: string;
	title?: string;
	description?: string;
	theme?: "light" | "dark";
	placeholder?: string;
	initialMessage?: string;
	primaryColor?: string;
}

interface Message {
	role: "user" | "assistant";
	text: string;
	chart?: ChartResult | null;
	suggestions?: string[];
}

interface ChartResult {
	type: "chart";
	chart_type: "bar" | "line" | "area" | "pie" | "table";
	title: string;
	data: Array<{ x: unknown; y: unknown }>;
	x_label?: string;
	y_label?: string;
}

interface SuggestionsResult {
	type: "suggestions";
	last_query: string;
	available_tables: string[];
}

// Legacy XML chart format (backward compat)
interface LegacyChartData {
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

function parseLegacyChart(text: string): LegacyChartData | null {
	const match = text.match(/<chart_data>([\s\S]*?)<\/chart_data>/);
	if (!match) return null;
	try {
		return JSON.parse(match[1]) as LegacyChartData;
	} catch {
		return null;
	}
}

function stripChart(text: string) {
	return text.replace(/<chart_data>[\s\S]*?<\/chart_data>/g, "").trim();
}

// Parse AI SDK v6 data-stream protocol lines:
// 0:"text" → text delta
// a:{toolCallId,result} → tool result
function parseDataStream(raw: string): { text: string; toolResults: unknown[] } {
	const lines = raw.split(/\r?\n/);
	const parts: string[] = [];
	const toolResults: unknown[] = [];
	for (const line of lines) {
		const trimmed = line.trim();
		if (!trimmed) continue;
		const colonIdx = trimmed.indexOf(":");
		if (colonIdx === -1) continue;
		const prefix = trimmed.slice(0, colonIdx);
		const payload = trimmed.slice(colonIdx + 1);
		try {
			const parsed = JSON.parse(payload) as unknown;
			if (prefix === "0" && typeof parsed === "string") {
				parts.push(parsed);
			} else if (prefix === "a" || prefix === "10") {
				toolResults.push(parsed);
			}
		} catch {
			// ignore unparseable lines
		}
	}
	return { text: parts.join(""), toolResults };
}

function extractChartFromResults(results: unknown[]): ChartResult | null {
	for (const r of results) {
		const obj = r as Record<string, unknown>;
		const result = (obj.result ?? obj) as Record<string, unknown>;
		if (result?.type === "chart") return result as unknown as ChartResult;
	}
	return null;
}

function extractSuggestionsFromResults(results: unknown[]): string[] {
	for (const r of results) {
		const obj = r as Record<string, unknown>;
		const result = (obj.result ?? obj) as SuggestionsResult;
		if (result?.type === "suggestions") {
			const tables = result.available_tables ?? [];
			return tables.slice(0, 3).map((t) => `What's in ${t}?`);
		}
	}
	return [];
}

function InlineChart({ chart, color }: { chart: ChartResult; color: string }) {
	const rows = chart.data.map((d) => ({ x: d.x, y: d.y }));
	if (chart.chart_type === "pie") {
		const pieData = rows.map((r) => ({ name: String(r.x), value: Number(r.y) }));
		return (
			<div style={{ marginTop: 10 }}>
				<div style={{ fontSize: 11, fontWeight: 600, marginBottom: 4, opacity: 0.7 }}>{chart.title}</div>
				<div style={{ height: 180 }}>
					<ResponsiveContainer width="100%" height="100%">
						<PieChart>
							<Pie data={pieData} dataKey="value" nameKey="name" cx="50%" cy="50%" outerRadius={70} fill={color} />
							<Tooltip />
						</PieChart>
					</ResponsiveContainer>
				</div>
			</div>
		);
	}
	const isLine = chart.chart_type === "line" || chart.chart_type === "area";
	const Chart = isLine ? LineChart : BarChart;
	return (
		<div style={{ marginTop: 10 }}>
			<div style={{ fontSize: 11, fontWeight: 600, marginBottom: 4, opacity: 0.7 }}>{chart.title}</div>
			<div style={{ height: 180 }}>
				<ResponsiveContainer width="100%" height="100%">
					<Chart data={rows}>
						<CartesianGrid strokeDasharray="3 3" stroke="rgba(128,128,128,.2)" />
						<XAxis dataKey="x" tick={{ fontSize: 10 }} label={chart.x_label ? { value: chart.x_label, position: "insideBottom", offset: -4, fontSize: 10 } : undefined} />
						<YAxis tick={{ fontSize: 10 }} label={chart.y_label ? { value: chart.y_label, angle: -90, position: "insideLeft", fontSize: 10 } : undefined} />
						<Tooltip />
						{isLine ? (
							<Line dataKey="y" stroke={color} strokeWidth={2} dot={false} />
						) : (
							<Bar dataKey="y" fill={color} radius={[4, 4, 0, 0]} />
						)}
					</Chart>
				</ResponsiveContainer>
			</div>
		</div>
	);
}

function LegacyInlineChart({ chart, color }: { chart: LegacyChartData; color: string }) {
	const dataset = chart.datasets[0];
	const rows = chart.labels.map((label, index) => ({ x: label, y: dataset?.data[index] ?? 0 }));
	const isLine = chart.type === "line";
	const Chart = isLine ? LineChart : BarChart;
	return (
		<div style={{ height: 180, marginTop: 10 }}>
			<ResponsiveContainer width="100%" height="100%">
				<Chart data={rows}>
					<CartesianGrid strokeDasharray="3 3" />
					<XAxis dataKey="x" tick={{ fontSize: 10 }} />
					<YAxis tick={{ fontSize: 10 }} />
					{isLine ? <Line dataKey="y" stroke={color} strokeWidth={2} dot={false} /> : <Bar dataKey="y" fill={color} radius={[4, 4, 0, 0]} />}
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

	const sendMessage = useCallback(
		async (text: string) => {
			const trimmed = text.trim();
			if (!trimmed || isLoading || !sessionId) return;
			const userMsg: Message = { role: "user", text: trimmed };
			const history = [...messages, userMsg];
			setMessages([...history, { role: "assistant", text: "" }]);
			setIsLoading(true);
			try {
				const response = await fetch(`${apiBaseUrl}/api/agents/${agentId}/chat`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify({
						sessionId,
						messages: history.map((m, i) => ({
							id: `${m.role}-${i}`,
							role: m.role,
							parts: [{ type: "text", text: m.text }],
						})),
					}),
				});
				if (!response.ok || !response.body) {
					throw new Error(await response.text());
				}
				// Stream the response chunk-by-chunk
				const reader = response.body.getReader();
				const decoder = new TextDecoder();
				let rawAccum = "";
				while (true) {
					const { done, value } = await reader.read();
					if (done) break;
					rawAccum += decoder.decode(value, { stream: true });
					const { text: streamText } = parseDataStream(rawAccum);
					setMessages((prev: Message[]) => [
						...prev.slice(0, -1),
						{ role: "assistant", text: streamText },
					]);
				}
				// Final parse including tool results
				const { text: finalText, toolResults } = parseDataStream(rawAccum);
				const chart = extractChartFromResults(toolResults);
				const suggestions = extractSuggestionsFromResults(toolResults);
				setMessages((prev: Message[]) => [
					...prev.slice(0, -1),
					{ role: "assistant", text: finalText, chart, suggestions },
				]);
			} catch (err) {
				setMessages((prev: Message[]) => [
					...prev.slice(0, -1),
					{ role: "assistant", text: err instanceof Error ? `Error: ${err.message}` : "Something went wrong." },
				]);
			} finally {
				setIsLoading(false);
			}
		},
		[agentId, apiBaseUrl, isLoading, messages, sessionId],
	);

	return { messages, sendMessage, isLoading };
}

export function MantrixAgent({
	agentId,
	apiBaseUrl = "",
	title = "Pipeline Agent",
	description = "Ask about published pipeline data",
	theme = "light",
	placeholder = "Ask a question...",
	initialMessage,
	primaryColor = "#0f766e",
}: MantrixAgentProps) {
	const [open, setOpen] = useState(false);
	const [input, setInput] = useState("");
	const { messages, sendMessage, isLoading } = useMantrixAgent(agentId, apiBaseUrl);
	const endRef = useRef<HTMLDivElement | null>(null);
	const isDark = theme === "dark";
	const bg = isDark ? "#18181b" : "#ffffff";
	const border = isDark ? "#3f3f46" : "#d4d4d8";
	const textColor = isDark ? "#f4f4f5" : "#111827";
	const mutedColor = isDark ? "#a1a1aa" : "#6b7280";
	const bubbleBg = isDark ? "#27272a" : "#f4f4f5";

	useEffect(() => {
		endRef.current?.scrollIntoView({ behavior: "smooth" });
	}, [messages]);

	useEffect(() => {
		if (initialMessage && open && messages.length === 0) {
			void sendMessage(initialMessage);
		}
	// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [open]);

	return (
		<div style={{ position: "fixed", right: 20, bottom: 20, zIndex: 2147483000, fontFamily: "Inter, system-ui, sans-serif" }}>
			{open && (
				<div style={{ width: 390, height: 580, background: bg, border: `1px solid ${border}`, borderRadius: 12, boxShadow: "0 20px 80px rgba(15,23,42,.28)", display: "flex", flexDirection: "column", overflow: "hidden", marginBottom: 12 }}>
					{/* Header */}
					<div style={{ padding: "12px 14px", borderBottom: `1px solid ${border}`, display: "flex", alignItems: "center", justifyContent: "space-between" }}>
						<div>
							<div style={{ fontWeight: 700, fontSize: 14, color: textColor }}>{title}</div>
							<div style={{ marginTop: 1, fontSize: 11, color: mutedColor }}>{description}</div>
						</div>
						<button type="button" onClick={() => setOpen(false)} style={{ border: 0, background: "transparent", color: mutedColor, cursor: "pointer", fontSize: 16, lineHeight: 1 }}>✕</button>
					</div>

					{/* Messages */}
					<div style={{ flex: 1, overflowY: "auto", padding: 12, display: "flex", flexDirection: "column", gap: 8 }}>
						{messages.length === 0 && (
							<div style={{ margin: "auto", textAlign: "center", color: mutedColor, fontSize: 13, padding: 24 }}>
								Ask a question about your data
							</div>
						)}
						{messages.map((message, index) => {
							const legacyChart = parseLegacyChart(message.text);
							const isUser = message.role === "user";
							return (
								<div key={`${message.role}-${index}`} style={{ display: "flex", flexDirection: "column", alignItems: isUser ? "flex-end" : "flex-start" }}>
									<div style={{ maxWidth: "88%", borderRadius: 10, padding: "8px 11px", background: isUser ? primaryColor : bubbleBg, color: isUser ? "#fff" : textColor, fontSize: 13, lineHeight: 1.5, whiteSpace: "pre-wrap", wordBreak: "break-word" }}>
										{stripChart(message.text) || (isLoading && index === messages.length - 1 ? (
											<span style={{ display: "flex", gap: 3 }}>
												{[0,1,2].map(i => <span key={i} style={{ width: 5, height: 5, borderRadius: "50%", background: mutedColor, display: "inline-block", animation: `pulse 1s ease-in-out ${i * 0.2}s infinite` }} />)}
											</span>
										) : null)}
										{message.chart && <InlineChart chart={message.chart} color={primaryColor} />}
										{legacyChart && <LegacyInlineChart chart={legacyChart} color={primaryColor} />}
									</div>
									{message.suggestions && message.suggestions.length > 0 && (
										<div style={{ display: "flex", flexWrap: "wrap", gap: 4, marginTop: 6 }}>
											{message.suggestions.map((s: string) => (
												<button key={s} type="button" onClick={() => void sendMessage(s)} style={{ fontSize: 11, border: `1px solid ${border}`, borderRadius: 6, padding: "3px 8px", background: "transparent", color: textColor, cursor: "pointer" }}>
													{s}
												</button>
											))}
										</div>
									)}
								</div>
							);
						})}
						<div ref={endRef} />
					</div>

					{/* Input */}
					<form style={{ display: "flex", gap: 8, padding: 10, borderTop: `1px solid ${border}` }} onSubmit={(e: React.FormEvent) => { e.preventDefault(); void sendMessage(input); setInput(""); }}>
						<input
							value={input}
							onChange={(e: React.ChangeEvent<HTMLInputElement>) => setInput(e.target.value)}
							placeholder={placeholder}
							disabled={isLoading}
							style={{ flex: 1, border: `1px solid ${border}`, borderRadius: 8, padding: "9px 10px", background: bg, color: textColor, fontSize: 13, outline: "none" }}
						/>
						<button type="submit" disabled={isLoading || !input.trim()} style={{ border: 0, borderRadius: 8, background: primaryColor, color: "#fff", padding: "0 14px", fontSize: 13, fontWeight: 600, cursor: isLoading ? "wait" : "pointer", opacity: isLoading || !input.trim() ? 0.6 : 1 }}>
							{isLoading ? "…" : "↑"}
						</button>
					</form>
				</div>
			)}
			{/* FAB trigger */}
			<button type="button" onClick={() => setOpen((v: boolean) => !v)} style={{ height: 56, width: 56, borderRadius: "50%", border: 0, background: primaryColor, color: "#fff", boxShadow: "0 10px 30px rgba(0,0,0,.25)", cursor: "pointer", fontSize: 22, display: "flex", alignItems: "center", justifyContent: "center" }}>
				{open ? "✕" : "💬"}
			</button>
		</div>
	);
}
