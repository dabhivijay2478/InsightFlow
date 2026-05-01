import { createRoot, type Root } from "react-dom/client";
import { MantrixAgent } from "./index";

type InitOptions = {
	apiBaseUrl?: string;
	title?: string;
};

type MantrixCommand = {
	(command: "init", agentId: string, options?: InitOptions): void;
	q?: unknown[][];
};

declare global {
	interface Window {
		MantrixAgent?: string;
		mantrix?: MantrixCommand;
		[key: string]: unknown;
	}
}

let root: Root | null = null;

function currentScriptOrigin() {
	if (typeof document === "undefined") return "";
	const script = document.currentScript as HTMLScriptElement | null;
	if (!script?.src) return "";
	try {
		return new URL(script.src).origin;
	} catch {
		return "";
	}
}

function mount(agentId: string, options: InitOptions = {}) {
	if (typeof document === "undefined" || !agentId) return;
	const hostId = "mantrix-agent-shadow-root";
	let host = document.getElementById(hostId);
	if (!host) {
		host = document.createElement("div");
		host.id = hostId;
		document.body.appendChild(host);
	}
	const shadow = host.shadowRoot ?? host.attachShadow({ mode: "open" });
	let mountNode = shadow.getElementById("mantrix-agent-mount");
	if (!mountNode) {
		mountNode = document.createElement("div");
		mountNode.id = "mantrix-agent-mount";
		shadow.appendChild(mountNode);
	}
	root?.unmount();
	root = createRoot(mountNode);
	root.render(
		<MantrixAgent
			agentId={agentId}
			apiBaseUrl={options.apiBaseUrl ?? currentScriptOrigin()}
			title={options.title}
		/>,
	);
}

function install() {
	if (typeof window === "undefined") return;
	const commandName =
		typeof window.MantrixAgent === "string" ? window.MantrixAgent : "mantrix";
	const queued = ((window[commandName] as MantrixCommand | undefined)?.q ??
		[]) as unknown[][];
	const command: MantrixCommand = (name, agentId, options) => {
		if (name === "init") {
			mount(agentId, options);
		}
	};
	window[commandName] = command;
	window.mantrix = command;
	for (const args of queued) {
		command(args[0] as "init", args[1] as string, args[2] as InitOptions);
	}
}

install();
