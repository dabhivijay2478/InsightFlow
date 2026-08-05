# Oria Release 5 enterprise test prompts

Hand-crafted test corpus for **Release 5 enterprise governance specialists** (12 agents).
Use with `AGENT_RUNTIME_ENABLED=true`, enterprise shadow mode as configured, and signed-in federation context.

## Agents covered

| Agent | Focus |
| --- | --- |
| `enterprise_mission_control` | Enterprise governance dashboard and cross-domain posture |
| `federated_workspace_manager` | Multi-workspace federation inventory and policy |
| `executive_intelligence` | C-level summaries, KPIs, and decision briefs |
| `strategic_goal_planner` | OKRs, strategic objectives, initiative mapping |
| `data_product_governor` | Data product registry, contracts, lifecycle |
| `change_approval_board` | CAB workflows, production change approvals |
| `enterprise_risk_manager` | Risk register, scoring, mitigation tracking |
| `compliance_evidence_manager` | Audit evidence packs, SOC 2, control mapping |
| `ai_model_governance` | LLM assignments, model policy, Oria agent oversight |
| `extension_governance` | Connector and extension risk review |
| `enterprise_finops_planner` | Enterprise spend, chargeback, budget guardrails |
| `continuity_command_center` | BCP, DR, severe incident command |

## Response rules

- Public identity is **Oria** — never internal capability or tool names.
- Release 5 **shadow mode** blocks writes and delegation without human confirmation.
- Governance artifacts are read-only previews unless explicit approval flow completes.
- No credentials, tokens, or raw secrets in responses.
- Executive summaries use leadership language with evidence caveats.

## Prompt index

| # | Agent | Category | Prompt | What to verify |
| ---: | --- | --- | --- | --- |
| 1 | enterprise_mission_control | Governance dashboard | Give me an enterprise mission control overview for our data platform | Executive-style summary with governance posture, open items, and no internal agent names |
| 2 | enterprise_mission_control | Governance dashboard | What is the current enterprise governance health score across all workspaces? | Quantified health indicators with evidence gaps called out when unavailable |
| 3 | enterprise_mission_control | Governance dashboard | Show me the mission control dashboard for pipeline operations this quarter | Quarter-scoped operational governance view with KPIs and status bands |
| 4 | enterprise_mission_control | Governance dashboard | Summarize enterprise-wide ELT governance posture for the board deck | Board-ready language; read-only; no credential leakage |
| 5 | enterprise_mission_control | Compliance posture | Which enterprise controls are currently failing or degraded? | Control status list with severity and remediation owner placeholders |
| 6 | enterprise_mission_control | Compliance posture | Are we in shadow mode for Release 5 enterprise governance actions? | Explicit shadow-mode acknowledgment; no silent writes |
| 7 | enterprise_mission_control | Compliance posture | List open governance exceptions requiring executive attention | Exception register format with expiry and approver state |
| 8 | enterprise_mission_control | Compliance posture | What policy violations were detected in the last 30 days? | Time-bounded violation summary from evidence or 'not available' |
| 9 | enterprise_mission_control | Cross-domain view | Correlate pipeline failures, approval backlog, and compliance gaps in one view | Integrated enterprise snapshot without overclaiming causality |
| 10 | enterprise_mission_control | Cross-domain view | Show enterprise mission control status during active production incident | Incident-aware governance summary; continuity cross-links allowed |
| 11 | enterprise_mission_control | Cross-domain view | Which business units have the weakest data governance maturity? | Maturity comparison with assumptions stated |
| 12 | enterprise_mission_control | Cross-domain view | Provide a red-yellow-green status for all enterprise governance domains | RAG status table with criteria explained |
| 13 | enterprise_mission_control | Executive reporting | Draft a weekly enterprise governance briefing for the COO | Executive summary tone; actionable bullets; read-only |
| 14 | enterprise_mission_control | Executive reporting | What changed in enterprise governance since last Monday? | Delta-oriented update with dated evidence |
| 15 | enterprise_mission_control | Executive reporting | Prepare a one-page mission control summary for investor diligence | High-level governance narrative; no secrets |
| 16 | enterprise_mission_control | Executive reporting | Highlight top three enterprise risks visible from mission control | Risk prioritization aligned to enterprise risk register themes |
| 17 | enterprise_mission_control | Workspace rollup | Roll up governance metrics from all federated workspaces | Federation-aware aggregation or scope limitation stated |
| 18 | enterprise_mission_control | Workspace rollup | Compare governance maturity between NA and EU workspace groups | Regional comparison with data availability caveats |
| 19 | enterprise_mission_control | Workspace rollup | Which workspace has the most overdue governance actions? | Ranked list with counts and aging |
| 20 | enterprise_mission_control | Workspace rollup | Show enterprise mission control for the finance analytics division only | Scoped filter respected; division boundary clear |
| 21 | enterprise_mission_control | Approval tracking | How many production changes are awaiting enterprise approval? | Approval queue count with status breakdown |
| 22 | enterprise_mission_control | Approval tracking | List critical-path approvals blocking go-live this week | Time-sensitive approval blockers with CAB references |
| 23 | enterprise_mission_control | Approval tracking | Show approval SLA compliance for enterprise change requests | SLA metrics or explicit missing-data message |
| 24 | enterprise_mission_control | Approval tracking | Which approvers are bottlenecks in the enterprise workflow? | Bottleneck analysis without naming private credentials |
| 25 | enterprise_mission_control | Audit readiness | Are we audit-ready from a mission control perspective? | Audit readiness assessment with evidence gaps |
| 26 | enterprise_mission_control | Audit readiness | What governance artifacts are missing for SOC 2 Type II? | Artifact gap list cross-referencing compliance themes |
| 27 | enterprise_mission_control | Audit readiness | Show last enterprise governance attestation date and owner | Attestation metadata or unavailable notice |
| 28 | enterprise_mission_control | Audit readiness | Summarize control testing results visible in mission control | Control test summary; no raw audit logs with secrets |
| 29 | enterprise_mission_control | Automation governance | Which automation policies bypass human confirmation? | Shadow mode and confirmation rules explained |
| 30 | enterprise_mission_control | Automation governance | Show enterprise automation compliance rate across workspaces | Compliance percentage with numerator/denominator if known |
| 31 | enterprise_mission_control | Automation governance | Are any self-healing actions running without CAB approval? | Automation governance boundary check |
| 32 | enterprise_mission_control | Automation governance | List automation incidents escalated to mission control this month | Monthly escalation log summary |
| 33 | enterprise_mission_control | Data platform | What is the enterprise status of critical Stripe-to-warehouse pipelines? | Critical pipeline governance lens; pipeline names from evidence |
| 34 | enterprise_mission_control | Data platform | Show mission control view for HubSpot and Salesforce federation pipelines | Multi-source pipeline governance summary |
| 35 | enterprise_mission_control | Data platform | Which pipelines lack assigned data product owners? | Ownership gap identification |
| 36 | enterprise_mission_control | Data platform | Enterprise overview of pipeline schedule change activity | Schedule change governance activity summary |
| 37 | enterprise_mission_control | FinOps linkage | Include FinOps spend signals in today's mission control briefing | FinOps cross-domain summary without billing credentials |
| 38 | enterprise_mission_control | FinOps linkage | Are any workspaces exceeding enterprise spend guardrails? | Guardrail breach summary with workspace scope |
| 39 | enterprise_mission_control | FinOps linkage | Show cost governance exceptions on the mission control dashboard | Exception list linked to FinOps themes |
| 40 | enterprise_mission_control | FinOps linkage | Correlate run volume spikes with governance alerts | Correlation narrative with confidence/assumptions |
| 41 | enterprise_mission_control | Risk linkage | Surface top enterprise risks on the mission control home view | Risk register integration preview |
| 42 | enterprise_mission_control | Risk linkage | Which mission control indicators triggered risk escalations? | Escalation trigger explanation |
| 43 | enterprise_mission_control | Risk linkage | Show risk trend for data platform operations over 90 days | Trend narrative with expiry on forecasts if any |
| 44 | enterprise_mission_control | Risk linkage | Map mission control KPIs to enterprise risk categories | KPI-to-risk mapping table |
| 45 | enterprise_mission_control | Continuity linkage | What is continuity command status reflected in mission control? | Continuity status cross-reference |
| 46 | enterprise_mission_control | Continuity linkage | Show active continuity playbooks linked to enterprise dashboard | Playbook linkage without executing writes |
| 47 | enterprise_mission_control | Continuity linkage | Is disaster recovery governance green across all regions? | DR governance RAG by region |
| 48 | enterprise_mission_control | Continuity linkage | Summarize last continuity exercise outcomes for leadership | Exercise summary in executive tone |
| 49 | enterprise_mission_control | Model governance | Include AI model governance status in enterprise overview | Model governance rollup without API keys |
| 50 | enterprise_mission_control | Model governance | Which LLM assignments lack enterprise approval? | Unapproved model usage gaps |
| 51 | enterprise_mission_control | Model governance | Show Oria agent governance posture on mission control | Agent governance summary; public identity Oria only |
| 52 | enterprise_mission_control | Model governance | Are shadow-mode restrictions visible on the governance dashboard? | Shadow mode UI/policy alignment stated |
| 53 | enterprise_mission_control | Extension governance | Summarize third-party extension risk on enterprise dashboard | Extension risk rollup |
| 54 | enterprise_mission_control | Extension governance | Which connectors are pending enterprise extension review? | Pending review queue |
| 55 | enterprise_mission_control | Extension governance | Show certified vs uncertified integrations at enterprise level | Certification status comparison |
| 56 | enterprise_mission_control | Extension governance | Highlight high-risk marketplace extensions for executive review | Executive risk highlight list |
| 57 | enterprise_mission_control | Strategic alignment | How do current governance metrics align to Q3 strategic objectives? | OKR alignment narrative |
| 58 | enterprise_mission_control | Strategic alignment | Which strategic goals are off-track based on governance data? | Goal drift identification |
| 59 | enterprise_mission_control | Strategic alignment | Show mission control KPIs tied to data freshness OKRs | OKR linkage with measurable targets |
| 60 | enterprise_mission_control | Strategic alignment | Executive summary: are we meeting enterprise data reliability goals? | Reliability goal assessment |
| 61 | enterprise_mission_control | Edge case | Enterprise mission control overview | Minimal prompt still routes correctly; governance artifact returned |
| 62 | enterprise_mission_control | Edge case | Mission control — nothing else, just status | Handles terse phrasing; no internal routing names |
| 63 | enterprise_mission_control | Edge case | Give me the governance dashboard for my organization | Org-scoped overview |
| 64 | enterprise_mission_control | Edge case | What should I look at first on enterprise mission control? | Prioritized starting points for operators |
| 65 | enterprise_mission_control | Edge case | Export-ready summary of enterprise governance for quarterly review | Quarterly review structure |
| 66 | enterprise_mission_control | Edge case | Compare this week vs last week enterprise governance posture | Week-over-week delta |
| 67 | enterprise_mission_control | Edge case | Who owns unresolved mission control action items? | Ownership assignment visibility |
| 68 | enterprise_mission_control | Edge case | Show enterprise governance for pipelines touching PCI data | Sensitive data scope handling; no PCI values returned |
| 69 | enterprise_mission_control | Edge case | Is enterprise mission control available in read-only mode? | Read-only/shadow mode confirmation |
| 70 | enterprise_mission_control | Edge case | Summarize enterprise governance in three bullets for the CEO | Ultra-concise executive summary |
| 71 | enterprise_mission_control | Edge case | What enterprise approvals expire in the next 7 days? | Time-bound expiry list |
| 72 | enterprise_mission_control | Edge case | Show governance dashboard filtered to production environments only | Environment filter respected |
| 73 | enterprise_mission_control | Edge case | Enterprise overview after major schema drift incident | Post-incident governance context |
| 74 | enterprise_mission_control | Edge case | List all governance domains and their current status labels | Domain inventory with status |
| 75 | federated_workspace_manager | Federation inventory | List all workspaces in my enterprise federation | Workspace inventory with names/ids; no cross-tenant credential leakage |
| 76 | federated_workspace_manager | Federation inventory | How many workspaces are federated under our enterprise tenant? | Count with federation boundary stated |
| 77 | federated_workspace_manager | Federation inventory | Show workspace hierarchy for our global data organization | Hierarchy tree or flat list with parent-child if available |
| 78 | federated_workspace_manager | Federation inventory | Which workspaces belong to the EMEA federation group? | Regional group filter applied |
| 79 | federated_workspace_manager | Federation inventory | Map federated workspaces to business units | BU mapping table |
| 80 | federated_workspace_manager | Cross-workspace ops | Compare pipeline counts across federated workspaces | Cross-workspace metric comparison |
| 81 | federated_workspace_manager | Cross-workspace ops | Which federated workspace has the highest failure rate this month? | Ranked failure rates with time scope |
| 82 | federated_workspace_manager | Cross-workspace ops | Show run volume totals aggregated across all federation members | Aggregated run statistics |
| 83 | federated_workspace_manager | Cross-workspace ops | List workspaces with no pipelines configured | Empty workspace identification |
| 84 | federated_workspace_manager | Cross-workspace ops | Compare Stripe pipeline configurations between prod and sandbox workspaces | Config comparison without secrets |
| 85 | federated_workspace_manager | Access governance | Who has admin access across federated workspaces? | Role summary; no raw tokens or passwords |
| 86 | federated_workspace_manager | Access governance | Show cross-workspace access policy violations | Policy violation list |
| 87 | federated_workspace_manager | Access governance | Which users span more than three federated workspaces? | Cross-workspace user breadth |
| 88 | federated_workspace_manager | Access governance | Summarize SSO and workspace membership for enterprise federation | Identity federation summary |
| 89 | federated_workspace_manager | Access governance | Are there orphaned workspaces without assigned owners? | Orphan workspace detection |
| 90 | federated_workspace_manager | Policy propagation | Which federation policies are not enforced in workspace APAC-2? | Policy drift per workspace |
| 91 | federated_workspace_manager | Policy propagation | Show enterprise policy inheritance across federated workspaces | Inheritance model explanation |
| 92 | federated_workspace_manager | Policy propagation | List workspaces exempt from standard federation schedule policies | Exemption register |
| 93 | federated_workspace_manager | Policy propagation | Compare automation policy versions across federation members | Version drift comparison |
| 94 | federated_workspace_manager | Policy propagation | Which workspaces lag behind enterprise baseline configuration? | Baseline compliance gaps |
| 95 | federated_workspace_manager | Data residency | Which federated workspaces store data in EU regions only? | Residency classification |
| 96 | federated_workspace_manager | Data residency | Show cross-border data flow paths between federation workspaces | Flow diagram narrative; compliance aware |
| 97 | federated_workspace_manager | Data residency | List workspaces processing customer PII under GDPR scope | PII scope identification without exposing PII |
| 98 | federated_workspace_manager | Data residency | Are any federation members violating data residency policy? | Residency violation flags |
| 99 | federated_workspace_manager | Data residency | Summarize regional workspace distribution for compliance review | Regional distribution summary |
| 100 | federated_workspace_manager | Onboarding | What is the checklist to onboard a new workspace into federation? | Onboarding steps; read-only unless confirmed write |
| 101 | federated_workspace_manager | Onboarding | Show pending workspace federation requests awaiting approval | Pending request queue |
| 102 | federated_workspace_manager | Onboarding | Which new workspaces joined the federation this quarter? | Quarterly join list |
| 103 | federated_workspace_manager | Onboarding | Draft federation onboarding plan for acquired subsidiary data team | Planning artifact with approval note |
| 104 | federated_workspace_manager | Onboarding | Validate prerequisites before linking workspace Finance-Analytics to federation | Prerequisite checklist |
| 105 | federated_workspace_manager | Offboarding | What happens when we de-federate a retired workspace? | Offboarding impact summary |
| 106 | federated_workspace_manager | Offboarding | List workspaces marked for decommission in federation | Decommission queue |
| 107 | federated_workspace_manager | Offboarding | Show data retention obligations for offboarding workspace EU-Sales | Retention requirements |
| 108 | federated_workspace_manager | Offboarding | Plan safe federation removal for sandbox workspace without prod impact | Blast radius analysis |
| 109 | federated_workspace_manager | Offboarding | Which pipelines must migrate before workspace defederation? | Migration dependency list |
| 110 | federated_workspace_manager | Cost allocation | Break down ELT spend by federated workspace | Spend allocation without payment credentials |
| 111 | federated_workspace_manager | Cost allocation | Which workspace drives the most AI token usage in federation? | Token usage ranking |
| 112 | federated_workspace_manager | Cost allocation | Show chargeback report for federation members last month | Chargeback summary |
| 113 | federated_workspace_manager | Cost allocation | Identify workspaces exceeding allocated compute budget | Budget overrun list |
| 114 | federated_workspace_manager | Cost allocation | Compare FinOps metrics between US-East and US-West workspace groups | Regional FinOps comparison |
| 115 | federated_workspace_manager | Compliance federation | Which workspaces lack SOC 2 evidence collection enabled? | Evidence gap per workspace |
| 116 | federated_workspace_manager | Compliance federation | Show federation-wide compliance score rollup | Compliance score aggregation |
| 117 | federated_workspace_manager | Compliance federation | List workspaces with open audit findings | Audit finding rollup |
| 118 | federated_workspace_manager | Compliance federation | Compare control maturity across federation business units | Maturity comparison |
| 119 | federated_workspace_manager | Compliance federation | Summarize HIPAA-scoped workspaces in the federation | Regulatory scope summary |
| 120 | federated_workspace_manager | Executive summary | Executive briefing: federated workspace landscape | C-level summary of federation state |
| 121 | federated_workspace_manager | Executive summary | One-page federation health report for leadership | Concise leadership report |
| 122 | federated_workspace_manager | Executive summary | What are the top federation risks this week? | Weekly risk highlights |
| 123 | federated_workspace_manager | Executive summary | Summarize cross-workspace pipeline reliability for the board | Board-level reliability narrative |
| 124 | federated_workspace_manager | Executive summary | How many critical pipelines span multiple federated workspaces? | Cross-workspace criticality count |
| 125 | federated_workspace_manager | Operational | Show federation sync status between enterprise hub and member workspaces | Sync status indicators |
| 126 | federated_workspace_manager | Operational | Which federation metadata sync jobs failed recently? | Sync failure log summary |
| 127 | federated_workspace_manager | Operational | List workspaces with stale federation catalog metadata | Staleness detection |
| 128 | federated_workspace_manager | Operational | When was federation membership last reconciled? | Reconciliation timestamp |
| 129 | federated_workspace_manager | Operational | Show cross-workspace search results for pipeline named Revenue Sync | Cross-workspace search results |
| 130 | federated_workspace_manager | Operational | Which federation groups include the Stripe All Streams pipeline? | Pipeline presence across workspaces |
| 131 | federated_workspace_manager | Operational | Compare HubSpot connector usage across all federation members | Connector adoption comparison |
| 132 | federated_workspace_manager | Operational | Identify duplicate pipelines across federated workspaces | Duplication detection |
| 133 | federated_workspace_manager | Operational | Show workspace quota utilization across federation | Quota utilization table |
| 134 | federated_workspace_manager | Operational | Which workspaces are near pipeline or connection limits? | Capacity warning list |
| 135 | federated_workspace_manager | Edge case | Federated workspaces | Terse prompt routes correctly |
| 136 | federated_workspace_manager | Edge case | Do I have more than one workspace? | Simple federation membership answer |
| 137 | federated_workspace_manager | Edge case | Switch context to enterprise federation view | Federation context acknowledgment |
| 138 | federated_workspace_manager | Edge case | List workspaces — federation scope only | Scope filter honored |
| 139 | federated_workspace_manager | Edge case | Cross workspace pipeline list for stripe | Stripe pipelines across federation |
| 140 | federated_workspace_manager | Edge case | Show me everything federated under Acme Corp | Org-named federation scope |
| 141 | federated_workspace_manager | Edge case | Federation manager summary please | Summary without internal agent names |
| 142 | federated_workspace_manager | Edge case | Which workspace should own the canonical Customer 360 pipeline? | Ownership recommendation with governance tone |
| 143 | federated_workspace_manager | Edge case | Can federated workspaces share destination connections safely? | Safety/governance guidance |
| 144 | federated_workspace_manager | Edge case | Read-only federation overview for external auditor | Auditor-safe read-only summary |
| 145 | federated_workspace_manager | Edge case | Compare federation size year over year | YoY growth narrative |
| 146 | federated_workspace_manager | Edge case | Show federation members with disabled automation | Automation status per workspace |
| 147 | federated_workspace_manager | Edge case | Workspace federation status during region outage | Outage-aware federation status |
| 148 | executive_intelligence | Executive summary | Give me an executive summary of pipeline operations this week | Weekly ops summary in leadership language |
| 149 | executive_intelligence | Executive summary | Prepare a C-level briefing on data platform reliability | C-level tone; KPIs; no jargon overload |
| 150 | executive_intelligence | Executive summary | Summarize ELT platform performance for the board meeting | Board-ready narrative |
| 151 | executive_intelligence | Executive summary | What should the CEO know about our data pipelines today? | CEO-focused top insights |
| 152 | executive_intelligence | Executive summary | Draft executive summary of last month's pipeline incidents | Incident summary without blame; factual |
| 153 | executive_intelligence | KPI dashboard | Show executive KPIs for data freshness and SLA adherence | KPI table with targets vs actuals |
| 154 | executive_intelligence | KPI dashboard | What is our enterprise data availability score? | Availability metric with methodology |
| 155 | executive_intelligence | KPI dashboard | Summarize rows processed and delivery success rate for leadership | Volume and success metrics |
| 156 | executive_intelligence | KPI dashboard | Compare this quarter's pipeline KPIs to last quarter | QoQ comparison |
| 157 | executive_intelligence | KPI dashboard | Highlight KPIs that missed target this month | Miss list with variance |
| 158 | executive_intelligence | Risk briefing | Top five operational risks for the data platform — executive view | Risk brief aligned to enterprise risk themes |
| 159 | executive_intelligence | Risk briefing | Summarize compliance posture for executive leadership | Compliance executive summary |
| 160 | executive_intelligence | Risk briefing | What regulatory exposure exists in our current pipeline footprint? | Regulatory exposure narrative; cautious language |
| 161 | executive_intelligence | Risk briefing | Executive risk brief after three consecutive Stripe sync failures | Incident-contextualized risk brief |
| 162 | executive_intelligence | Risk briefing | Are we meeting enterprise risk appetite for data operations? | Risk appetite assessment |
| 163 | executive_intelligence | Financial lens | Executive summary of data platform spend vs budget | Spend vs budget without payment details |
| 164 | executive_intelligence | Financial lens | What drove the increase in AI inference costs this month? | Cost driver analysis for executives |
| 165 | executive_intelligence | Financial lens | Summarize FinOps implications of doubling pipeline run frequency | FinOps impact narrative with assumptions |
| 166 | executive_intelligence | Financial lens | Show unit economics for pipeline processing at executive level | Unit economics summary |
| 167 | executive_intelligence | Financial lens | Is our data platform investment delivering expected ROI? | ROI discussion with stated assumptions |
| 168 | executive_intelligence | Strategic narrative | How does pipeline modernization support our digital transformation goals? | Strategy linkage narrative |
| 169 | executive_intelligence | Strategic narrative | Executive view of progress toward real-time analytics objective | Strategic objective progress |
| 170 | executive_intelligence | Strategic narrative | Summarize competitive implications of our data latency | Competitive framing; evidence-based |
| 171 | executive_intelligence | Strategic narrative | What strategic bets depend on HubSpot and Salesforce pipeline health? | Dependency on strategic initiatives |
| 172 | executive_intelligence | Strategic narrative | Brief the leadership team on data mesh adoption status | Data mesh progress summary |
| 173 | executive_intelligence | Customer impact | Did any pipeline issues affect customer-facing reports this week? | Customer impact assessment |
| 174 | executive_intelligence | Customer impact | Executive summary of data quality incidents affecting revenue teams | Business impact focus |
| 175 | executive_intelligence | Customer impact | Quantify business impact of yesterday's Postgres sync delay | Impact quantification with confidence bounds |
| 176 | executive_intelligence | Customer impact | Which downstream dashboards were stale during last outage? | Downstream impact list |
| 177 | executive_intelligence | Customer impact | Summarize customer SLA breaches for executive review | SLA breach executive summary |
| 178 | executive_intelligence | Trend analysis | Show 90-day trend in pipeline failure rate for executives | Trend chart narrative |
| 179 | executive_intelligence | Trend analysis | Is our mean time to recovery improving? | MTTR trend for leadership |
| 180 | executive_intelligence | Trend analysis | Executive view of automation adoption over six months | Automation adoption trend |
| 181 | executive_intelligence | Trend analysis | Summarize growth in federated workspaces for leadership | Federation growth narrative |
| 182 | executive_intelligence | Trend analysis | What trends should worry the executive team about data governance? | Governance trend warnings |
| 183 | executive_intelligence | Decision support | Should we approve additional OpenRouter budget for Oria agents? | Decision framing; requires human approval for writes |
| 184 | executive_intelligence | Decision support | Executive recommendation on consolidating duplicate Stripe pipelines | Recommendation with tradeoffs |
| 185 | executive_intelligence | Decision support | Summarize pros and cons of migrating to hourly sync schedules | Balanced decision brief |
| 186 | executive_intelligence | Decision support | What does leadership need to decide about EU data residency? | Decision agenda for residency |
| 187 | executive_intelligence | Decision support | Brief on whether to expand self-healing automation to production | Automation expansion brief; shadow mode noted |
| 188 | executive_intelligence | Stakeholder comms | Draft talking points for CFO on data platform costs | CFO-oriented talking points |
| 189 | executive_intelligence | Stakeholder comms | Prepare one-slide narrative on pipeline security for CISO | Security narrative for CISO |
| 190 | executive_intelligence | Stakeholder comms | Executive email draft: quarterly data platform achievements | Achievement summary draft |
| 191 | executive_intelligence | Stakeholder comms | How to explain last week's incident to non-technical board members | Plain-language incident explanation |
| 192 | executive_intelligence | Stakeholder comms | Summarize vendor dependency risk for OpenRouter and cloud providers | Vendor risk executive summary |
| 193 | executive_intelligence | Benchmarking | How do our pipeline SLAs compare to industry benchmarks? | Benchmark comparison with caveats |
| 194 | executive_intelligence | Benchmarking | Executive benchmark of automation maturity vs peers | Maturity benchmark narrative |
| 195 | executive_intelligence | Benchmarking | Are we leading or lagging on data governance maturity? | Governance maturity positioning |
| 196 | executive_intelligence | Benchmarking | Summarize external audit feedback for leadership | Audit feedback rollup |
| 197 | executive_intelligence | Benchmarking | Compare internal teams on pipeline reliability scores | Team comparison for executives |
| 198 | executive_intelligence | Portfolio view | Executive portfolio view of all critical data pipelines | Portfolio summary with criticality tiers |
| 199 | executive_intelligence | Portfolio view | Which pipelines are strategic vs operational in executive terms? | Strategic classification |
| 200 | executive_intelligence | Portfolio view | Summarize pipeline portfolio concentration risk | Concentration risk narrative |
| 201 | executive_intelligence | Portfolio view | Show executive heat map of pipeline health by business domain | Domain heat map description |
| 202 | executive_intelligence | Portfolio view | What percentage of revenue-critical pipelines are healthy? | Revenue-critical health percentage |
| 203 | executive_intelligence | Governance lens | Executive summary of change approval backlog | CAB backlog in exec terms |
| 204 | executive_intelligence | Governance lens | Summarize AI model governance for board risk committee | Model governance board brief |
| 205 | executive_intelligence | Governance lens | How effective is our enterprise governance program? | Program effectiveness assessment |
| 206 | executive_intelligence | Governance lens | Executive view of compliance evidence completeness | Evidence completeness summary |
| 207 | executive_intelligence | Governance lens | Summarize extension and connector governance for leadership | Extension governance exec summary |
| 208 | executive_intelligence | Continuity lens | Executive briefing on business continuity for data platform | BCP executive brief |
| 209 | executive_intelligence | Continuity lens | Summarize last disaster recovery exercise for leadership | DR exercise exec summary |
| 210 | executive_intelligence | Continuity lens | What is leadership visibility into active continuity incidents? | Continuity visibility statement |
| 211 | executive_intelligence | Continuity lens | Executive summary: can we survive region loss today? | Region loss resilience assessment |
| 212 | executive_intelligence | Continuity lens | RTO and RPO status for critical pipelines — executive view | RTO/RPO exec summary |
| 213 | executive_intelligence | Edge case | Executive summary | Minimal prompt produces exec-format output |
| 214 | executive_intelligence | Edge case | Brief me like I'm the CEO | CEO-appropriate tone and brevity |
| 215 | executive_intelligence | Edge case | Leadership update — pipelines | Terse leadership update handled |
| 216 | executive_intelligence | Edge case | What keeps the CTO up at night about our ELT stack? | Risk-focused CTO narrative |
| 217 | executive_intelligence | Edge case | Three bullets for the board on data platform status | Exactly concise board bullets |
| 218 | executive_intelligence | Edge case | Executive summary with no technical jargon | Plain language enforced |
| 219 | executive_intelligence | Edge case | Compare executive metrics before and after automation rollout | Before/after exec comparison |
| 220 | executive_intelligence | Edge case | Summarize Oria agent usage for executive oversight | Agent usage exec summary; Oria public identity |
| 221 | executive_intelligence | Edge case | Is shadow mode affecting what executives can see? | Shadow mode transparency for execs |
| 222 | executive_intelligence | Edge case | Executive intelligence report for Q2 close | Quarter-close themed summary |
| 223 | strategic_goal_planner | OKR planning | Define a strategic objective to improve data freshness across all pipelines | Objective statement with measurable key results draft |
| 224 | strategic_goal_planner | OKR planning | Draft OKRs for reducing pipeline failure rate by 50% this year | OKR structure with targets and owners placeholders |
| 225 | strategic_goal_planner | OKR planning | What strategic goals should we set for federated workspace governance? | Goal recommendations aligned to federation |
| 226 | strategic_goal_planner | OKR planning | Create a strategic plan for achieving sub-hour analytics latency | Strategic plan artifact with milestones |
| 227 | strategic_goal_planner | OKR planning | Align pipeline automation goals with enterprise digital strategy | Strategy alignment narrative |
| 228 | strategic_goal_planner | Objective tracking | Show progress on our data reliability strategic objective | Progress tracking with metrics |
| 229 | strategic_goal_planner | Objective tracking | Which key results are off-track this quarter? | Off-track KR identification |
| 230 | strategic_goal_planner | Objective tracking | Update me on OKR status for compliance evidence collection | Compliance OKR status |
| 231 | strategic_goal_planner | Objective tracking | How are we performing against the FinOps optimization goal? | FinOps goal performance |
| 232 | strategic_goal_planner | Objective tracking | Summarize strategic goal completion rate for leadership review | Completion rate summary |
| 233 | strategic_goal_planner | Initiative mapping | Map Stripe and HubSpot pipelines to strategic customer 360 initiative | Initiative-to-pipeline mapping |
| 234 | strategic_goal_planner | Initiative mapping | Which pipelines support the revenue forecasting strategic goal? | Supporting pipeline identification |
| 235 | strategic_goal_planner | Initiative mapping | List initiatives blocked by missing data product governance | Blocked initiative list |
| 236 | strategic_goal_planner | Initiative mapping | Show dependency between strategic goals and automation adoption | Goal dependency graph narrative |
| 237 | strategic_goal_planner | Initiative mapping | Prioritize strategic initiatives by pipeline readiness | Prioritized initiative ranking |
| 238 | strategic_goal_planner | Resource planning | What resources are needed to achieve real-time sync strategic goal? | Resource plan draft |
| 239 | strategic_goal_planner | Resource planning | Estimate timeline to meet enterprise data mesh objective | Timeline with assumptions |
| 240 | strategic_goal_planner | Resource planning | Identify skill gaps blocking strategic AI governance goals | Skill gap analysis |
| 241 | strategic_goal_planner | Resource planning | Budget implications of strategic goal to double pipeline throughput | Budget implication narrative |
| 242 | strategic_goal_planner | Resource planning | Recommend sequencing for three competing strategic objectives | Sequencing recommendation |
| 243 | strategic_goal_planner | Governance alignment | Ensure strategic goals comply with enterprise change approval policy | Policy alignment check |
| 244 | strategic_goal_planner | Governance alignment | Which strategic goals require CAB approval to proceed? | CAB requirement identification |
| 245 | strategic_goal_planner | Governance alignment | Align OKRs with SOC 2 compliance program objectives | Compliance alignment |
| 246 | strategic_goal_planner | Governance alignment | Show strategic goals touching regulated healthcare data | Regulatory scope on goals |
| 247 | strategic_goal_planner | Governance alignment | Draft governance guardrails for strategic automation expansion | Guardrail recommendations |
| 248 | strategic_goal_planner | Measurement | Define success metrics for enterprise extension governance goal | Success metric definition |
| 249 | strategic_goal_planner | Measurement | How will we measure progress on continuity preparedness objective? | Measurement framework |
| 250 | strategic_goal_planner | Measurement | Propose baseline and target for pipeline SLA strategic goal | Baseline/target proposal |
| 251 | strategic_goal_planner | Measurement | What leading indicators predict strategic goal achievement? | Leading indicator list |
| 252 | strategic_goal_planner | Measurement | Design a scorecard for strategic data platform goals | Scorecard structure |
| 253 | strategic_goal_planner | Scenario planning | What-if: delay strategic goal by one quarter — impact analysis | Scenario impact with assumptions |
| 254 | strategic_goal_planner | Scenario planning | Simulate achieving OKRs with 20% budget cut | Budget cut scenario |
| 255 | strategic_goal_planner | Scenario planning | Strategic plan if OpenRouter costs double next year | Cost shock scenario |
| 256 | strategic_goal_planner | Scenario planning | Plan strategic response to new EU data regulation | Regulatory response planning |
| 257 | strategic_goal_planner | Scenario planning | Evaluate strategic tradeoff: speed vs governance rigor | Tradeoff analysis |
| 258 | strategic_goal_planner | Executive alignment | Prepare strategic goal summary for quarterly business review | QBR strategic summary |
| 259 | strategic_goal_planner | Executive alignment | Draft strategic narrative linking pipelines to company mission | Mission linkage narrative |
| 260 | strategic_goal_planner | Executive alignment | Summarize strategic risks to OKR achievement for executives | Strategic risk summary |
| 261 | strategic_goal_planner | Executive alignment | Recommend strategic goal revisions based on H1 performance | Goal revision recommendations |
| 262 | strategic_goal_planner | Executive alignment | Show board-level strategic roadmap for data platform | Board roadmap artifact |
| 263 | strategic_goal_planner | Cross-functional | Align data engineering OKRs with finance reporting deadlines | Cross-functional alignment |
| 264 | strategic_goal_planner | Cross-functional | Strategic goals for partnership with security team on pipeline hardening | Security partnership goals |
| 265 | strategic_goal_planner | Cross-functional | Define objectives for legal review of cross-border data flows | Legal collaboration objectives |
| 266 | strategic_goal_planner | Cross-functional | Plan strategic onboarding goals for new federation workspaces | Federation onboarding goals |
| 267 | strategic_goal_planner | Cross-functional | Objectives for reducing mean time to approve production changes | Approval efficiency goals |
| 268 | strategic_goal_planner | Long-range | Three-year strategic plan for enterprise data platform maturity | 3-year plan outline |
| 269 | strategic_goal_planner | Long-range | Long-range goal for zero unapproved production pipeline changes | Long-range governance goal |
| 270 | strategic_goal_planner | Long-range | Strategic vision for fully governed AI agent operations | AI governance vision |
| 271 | strategic_goal_planner | Long-range | Plan strategic evolution from batch to streaming analytics | Streaming evolution plan |
| 272 | strategic_goal_planner | Long-range | Define 2027 strategic target for compliance evidence automation | Long-range compliance target |
| 273 | strategic_goal_planner | Review cadence | Schedule strategic goal review cadence for enterprise program | Review cadence recommendation |
| 274 | strategic_goal_planner | Review cadence | What strategic goals need mid-quarter correction? | Mid-quarter correction list |
| 275 | strategic_goal_planner | Review cadence | Summarize strategic goal changes since last planning cycle | Goal change log |
| 276 | strategic_goal_planner | Review cadence | Facilitate agenda for strategic goal planning workshop | Workshop agenda draft |
| 277 | strategic_goal_planner | Review cadence | Retrospective on last quarter's strategic goal outcomes | Retrospective summary |
| 278 | strategic_goal_planner | Edge case | Strategic objective for data freshness | Minimal objective prompt handled |
| 279 | strategic_goal_planner | Edge case | Set a business goal for pipeline ops | Business goal framing |
| 280 | strategic_goal_planner | Edge case | OKR help for enterprise governance | OKR assistance in governance context |
| 281 | strategic_goal_planner | Edge case | What strategic goals are we missing? | Gap analysis on goals |
| 282 | strategic_goal_planner | Edge case | Prioritize one strategic goal for this month | Single-goal prioritization |
| 283 | strategic_goal_planner | Edge case | Strategic plan read-only preview | Read-only plan; no silent writes |
| 284 | strategic_goal_planner | Edge case | Link strategic goals to Stripe All Streams pipeline | Pipeline-specific goal linkage |
| 285 | strategic_goal_planner | Edge case | Goals for shadow mode graduation to full governance | Shadow mode transition goals |
| 286 | strategic_goal_planner | Edge case | Strategic goal planner summary | Agent summary without internal names |
| 287 | strategic_goal_planner | Edge case | Define objective with executive approval required | Approval requirement noted |
| 288 | strategic_goal_planner | Edge case | Annual strategic planning for data products | Annual planning scope |
| 289 | strategic_goal_planner | Edge case | Compare planned vs actual strategic outcomes YTD | YTD planned vs actual |
| 290 | strategic_goal_planner | Review cadence | Quarterly strategic goal retrospective template for data platform program | Retrospective template artifact |
| 291 | strategic_goal_planner | Cross-functional | Strategic alignment workshop agenda for federation leaders | Workshop agenda draft |
| 292 | strategic_goal_planner | Measurement | Define north-star metric for enterprise data trust objective | North-star metric proposal |
| 293 | strategic_goal_planner | Governance alignment | Escalation path when strategic goals conflict with compliance controls | Escalation path explanation |
| 294 | data_product_governor | Product registry | Register a data product for Stripe revenue analytics | Registration preview; confirmation before write; product metadata shape |
| 295 | data_product_governor | Product registry | List all registered data products in the enterprise | Product registry listing |
| 296 | data_product_governor | Product registry | Show data product catalog with owners and SLAs | Catalog with ownership and SLA fields |
| 297 | data_product_governor | Product registry | Which data products lack assigned stewards? | Stewardship gap list |
| 298 | data_product_governor | Product registry | Search data products related to customer churn analytics | Search results by domain |
| 299 | data_product_governor | Product lifecycle | What is the lifecycle stage of the Customer 360 data product? | Lifecycle stage with criteria |
| 300 | data_product_governor | Product lifecycle | Show deprecated data products still receiving pipeline updates | Deprecation inconsistency flags |
| 301 | data_product_governor | Product lifecycle | Plan retirement for legacy HubSpot export data product | Retirement plan artifact |
| 302 | data_product_governor | Product lifecycle | Which data products are in draft vs published state? | State breakdown |
| 303 | data_product_governor | Product lifecycle | Summarize data product version history for Revenue Forecast | Version history summary |
| 304 | data_product_governor | Ownership | Who owns the Stripe Charges curated data product? | Owner identification |
| 305 | data_product_governor | Ownership | Show RACI matrix for enterprise data products | RACI summary |
| 306 | data_product_governor | Ownership | List data products with expired ownership assignments | Expired ownership list |
| 307 | data_product_governor | Ownership | Recommend owner for orphaned Salesforce pipeline data product | Owner recommendation |
| 308 | data_product_governor | Ownership | Which business units consume each registered data product? | Consumer mapping |
| 309 | data_product_governor | Quality standards | Do registered data products meet enterprise quality thresholds? | Quality compliance assessment |
| 310 | data_product_governor | Quality standards | Show data product quality scores from recent pipeline runs | Quality score linkage |
| 311 | data_product_governor | Quality standards | Which data products failed freshness SLA this week? | Freshness SLA breach list |
| 312 | data_product_governor | Quality standards | Define quality criteria for new Finance Ledger data product | Quality criteria draft |
| 313 | data_product_governor | Quality standards | Compare quality metrics across similar data products | Comparative quality analysis |
| 314 | data_product_governor | Contracts | Show data contract attached to Stripe analytics product | Contract summary without secrets |
| 315 | data_product_governor | Contracts | Which data products lack published data contracts? | Missing contract list |
| 316 | data_product_governor | Contracts | Draft contract schema for Postgres incremental sync output | Contract draft artifact |
| 317 | data_product_governor | Contracts | Validate data product output against its contract | Validation result summary |
| 318 | data_product_governor | Contracts | List contract violations for registered products this month | Violation log |
| 319 | data_product_governor | Lineage | Show lineage for Customer Lifetime Value data product | Lineage narrative |
| 320 | data_product_governor | Lineage | Which pipelines feed the Executive Dashboard data product? | Upstream pipeline list |
| 321 | data_product_governor | Lineage | Map downstream consumers of Stripe revenue data product | Downstream consumer map |
| 322 | data_product_governor | Lineage | Blast radius if we decommission HubSpot marketing data product | Impact analysis |
| 323 | data_product_governor | Lineage | Cross-workspace lineage for federated data product | Federation-aware lineage |
| 324 | data_product_governor | Access control | Who has access to sensitive payroll data product? | Access summary; no credential exposure |
| 325 | data_product_governor | Access control | Review access requests pending for Customer PII product | Pending access review queue |
| 326 | data_product_governor | Access control | Which data products require additional approval for consumption? | Approval-gated products |
| 327 | data_product_governor | Access control | Summarize RBAC policies on enterprise data products | RBAC policy summary |
| 328 | data_product_governor | Access control | Flag over-privileged access on financial data products | Over-privilege flags |
| 329 | data_product_governor | Compliance | Map data products to SOC 2 control requirements | Control mapping |
| 330 | data_product_governor | Compliance | Which products handle GDPR-subject data? | GDPR scope product list |
| 331 | data_product_governor | Compliance | Show compliance tags on all registered data products | Compliance tag inventory |
| 332 | data_product_governor | Compliance | Evidence pack for data product governance audit | Audit evidence orientation |
| 333 | data_product_governor | Compliance | Identify products missing privacy impact assessment | PIA gap list |
| 334 | data_product_governor | FinOps | Show compute cost allocation per data product | Cost allocation by product |
| 335 | data_product_governor | FinOps | Which data products are most expensive to maintain? | Cost ranking |
| 336 | data_product_governor | FinOps | Recommend consolidation of duplicate analytics products | Consolidation recommendation |
| 337 | data_product_governor | FinOps | Unit cost per row for top five data products | Unit economics by product |
| 338 | data_product_governor | FinOps | FinOps summary for data product portfolio | Portfolio FinOps summary |
| 339 | data_product_governor | Discovery | Help data analysts discover approved data products | Discovery guidance |
| 340 | data_product_governor | Discovery | Show metadata catalog for enterprise data products | Metadata catalog summary |
| 341 | data_product_governor | Discovery | Which products are certified for executive reporting? | Certification status for exec use |
| 342 | data_product_governor | Discovery | Recommend data product for marketing attribution use case | Use-case recommendation |
| 343 | data_product_governor | Discovery | Search products by domain: finance | Domain-filtered search |
| 344 | data_product_governor | Governance review | Schedule quarterly review for all tier-1 data products | Review schedule proposal |
| 345 | data_product_governor | Governance review | Summarize open governance issues on data products | Open issues rollup |
| 346 | data_product_governor | Governance review | Which products need re-certification this quarter? | Recertification queue |
| 347 | data_product_governor | Governance review | Data product governor executive summary | Executive summary of product governance |
| 348 | data_product_governor | Governance review | Compare product governance maturity across business units | BU maturity comparison |
| 349 | data_product_governor | Integration | Link Stripe All Streams pipeline to Revenue Analytics data product | Linkage preview; confirmation for writes |
| 350 | data_product_governor | Integration | Show pipelines not yet mapped to any data product | Unmapped pipeline list |
| 351 | data_product_governor | Integration | Validate product-pipeline mapping completeness | Completeness assessment |
| 352 | data_product_governor | Integration | Register data product from existing Postgres sync outputs | Registration from existing assets |
| 353 | data_product_governor | Integration | Propose data product boundaries for HubSpot federation | Boundary proposal |
| 354 | data_product_governor | Edge case | Data product for stripe analytics | Minimal registration intent |
| 355 | data_product_governor | Edge case | List data products | Simple list request |
| 356 | data_product_governor | Edge case | Who owns our data products? | Ownership overview |
| 357 | data_product_governor | Edge case | Data product governance status | Governance status summary |
| 358 | data_product_governor | Edge case | Tier-1 data products only | Tier filter applied |
| 359 | data_product_governor | Edge case | Product registry read-only view | Read-only confirmation |
| 360 | data_product_governor | Edge case | Data product without owner — what happens? | Policy explanation for missing owner |
| 361 | data_product_governor | Edge case | Governance for PCI-scoped data product | PCI scope handling |
| 362 | data_product_governor | Edge case | Export data product catalog for audit | Audit export orientation |
| 363 | data_product_governor | Edge case | Shadow mode: can I register products automatically? | Shadow mode blocks unconfirmed writes |
| 364 | data_product_governor | Edge case | Summary of data product governor responsibilities | Responsibility summary |
| 365 | change_approval_board | CAB intake | Submit production schedule change for Stripe All Streams to the change approval board | Submission preview; requires confirmation; CAB ticket shape |
| 366 | change_approval_board | CAB intake | Open a CAB request to add HubSpot streams to production federation | Intake form fields; risk section included |
| 367 | change_approval_board | CAB intake | Which production changes are pending CAB review? | Pending queue with status |
| 368 | change_approval_board | CAB intake | Show my submitted change requests awaiting approval | User-scoped pending list |
| 369 | change_approval_board | CAB intake | Draft CAB submission for migrating Postgres destination schema | Draft artifact with impact fields |
| 370 | change_approval_board | Approval workflow | Who must approve a tier-1 pipeline configuration change? | Approver chain without private contact secrets |
| 371 | change_approval_board | Approval workflow | What is the SLA for change approval board decisions? | SLA policy summary |
| 372 | change_approval_board | Approval workflow | Escalate overdue CAB approval for critical Stripe sync change | Escalation preview; human confirmation |
| 373 | change_approval_board | Approval workflow | Show approval history for last production run controller change | History log summary |
| 374 | change_approval_board | Approval workflow | Can this change be emergency-approved outside normal CAB? | Emergency process explanation |
| 375 | change_approval_board | Risk assessment | Include risk assessment in CAB packet for hourly schedule change | Risk section in CAB packet |
| 376 | change_approval_board | Risk assessment | Score change risk for enabling self-healing on production pipelines | Risk score with factors |
| 377 | change_approval_board | Risk assessment | What rollback plan is required for CAB approval? | Rollback requirement guidance |
| 378 | change_approval_board | Risk assessment | Assess blast radius of decommissioning Salesforce pipeline | Blast radius for CAB |
| 379 | change_approval_board | Risk assessment | Compare risk of incremental vs full-table sync change request | Comparative risk narrative |
| 380 | change_approval_board | Compliance gate | Does this change require compliance sign-off before CAB? | Compliance gate checklist |
| 381 | change_approval_board | Compliance gate | Show SOC 2 control impacts for proposed automation policy change | Control impact mapping |
| 382 | change_approval_board | Compliance gate | List changes blocked by missing compliance evidence | Blocked change list |
| 383 | change_approval_board | Compliance gate | Verify segregation of duties on approver list for finance pipeline change | SoD check result |
| 384 | change_approval_board | Compliance gate | CAB compliance checklist for cross-border data replication change | Checklist artifact |
| 385 | change_approval_board | Change types | Approve schema mapping change for Stripe charge_id column | Schema change CAB path |
| 386 | change_approval_board | Change types | CAB review for new OpenRouter model assignment in production | Model change approval path |
| 387 | change_approval_board | Change types | Submit connector extension upgrade for enterprise certification | Extension change CAB flow |
| 388 | change_approval_board | Change types | Request approval to promote staging pipeline to production | Promotion approval workflow |
| 389 | change_approval_board | Change types | Change approval for increasing MAX_CONCURRENT_RUNS enterprise-wide | Platform config change CAB |
| 390 | change_approval_board | Scheduling | When is the next CAB meeting and what is on the agenda? | Meeting schedule and agenda |
| 391 | change_approval_board | Scheduling | Fast-track approval for P1 incident remediation change | Fast-track process with guardrails |
| 392 | change_approval_board | Scheduling | Freeze window: can we approve changes during change freeze? | Freeze policy answer |
| 393 | change_approval_board | Scheduling | Show changes approved in last CAB session | Last session approvals |
| 394 | change_approval_board | Scheduling | Calendar of planned production changes this month | Change calendar summary |
| 395 | change_approval_board | Documentation | What documentation must attach to a CAB submission? | Required docs list |
| 396 | change_approval_board | Documentation | Generate CAB packet template for pipeline schedule change | Template artifact |
| 397 | change_approval_board | Documentation | Review completeness of CAB submission #1847 | Completeness review |
| 398 | change_approval_board | Documentation | Attach test evidence to change approval request | Evidence attachment guidance |
| 399 | change_approval_board | Documentation | Summarize post-implementation review requirements for CAB | PIR requirements |
| 400 | change_approval_board | Stakeholders | Notify stakeholders of approved Stripe pipeline change | Notification plan; no auto-send without confirm |
| 401 | change_approval_board | Stakeholders | Who are mandatory reviewers for tier-2 data product changes? | Reviewer roster |
| 402 | change_approval_board | Stakeholders | Show dissenting opinions on last rejected CAB request | Dissent summary if available |
| 403 | change_approval_board | Stakeholders | Assign business owner sign-off for revenue pipeline modification | Owner sign-off step |
| 404 | change_approval_board | Stakeholders | List CAB members and their approval domains | CAB membership by domain |
| 405 | change_approval_board | Audit trail | Export audit trail for all approved changes this quarter | Audit trail summary orientation |
| 406 | change_approval_board | Audit trail | Show who approved production destination change on July 12 | Approver attribution |
| 407 | change_approval_board | Audit trail | Were any CAB approvals bypassed in shadow mode? | Shadow mode bypass check |
| 408 | change_approval_board | Audit trail | Immutable log of rejected change requests last 90 days | Rejection log |
| 409 | change_approval_board | Audit trail | Verify CAB approval timestamp precedes production deployment | Temporal compliance check |
| 410 | change_approval_board | Integration | Link CAB request to enterprise risk register entry | Risk linkage preview |
| 411 | change_approval_board | Integration | Show FinOps impact section required for CAB cost-increasing changes | FinOps section guidance |
| 412 | change_approval_board | Integration | Connect change approval to continuity command pre-checks | Continuity pre-check linkage |
| 413 | change_approval_board | Integration | Reference strategic goal impacted by this CAB submission | Strategic goal reference |
| 414 | change_approval_board | Integration | Cross-reference compliance evidence for CAB approval | Evidence cross-reference |
| 415 | change_approval_board | Status queries | Status of CAB-2026-0142 pipeline schedule change | Ticket status lookup |
| 416 | change_approval_board | Status queries | Was my emergency change approved? | Approval status answer |
| 417 | change_approval_board | Status queries | Why was change request CAB-2026-0098 rejected? | Rejection reason summary |
| 418 | change_approval_board | Status queries | How many changes await second approver? | Multi-approver queue count |
| 419 | change_approval_board | Status queries | Show approval bottlenecks by change category | Category bottleneck analysis |
| 420 | change_approval_board | Executive | Executive summary of CAB backlog for leadership | Exec CAB backlog brief |
| 421 | change_approval_board | Executive | Board reporting: production change governance effectiveness | Board-level effectiveness narrative |
| 422 | change_approval_board | Executive | Summarize change failure rate post-CAB approval | Post-approval failure metrics |
| 423 | change_approval_board | Executive | CAB metrics: approval time, rejection rate, emergency changes | CAB KPI summary |
| 424 | change_approval_board | Executive | Recommend CAB process improvements for enterprise scale | Process improvement recommendations |
| 425 | change_approval_board | Edge case | Change approval board | Minimal routing prompt |
| 426 | change_approval_board | Edge case | Approve production change | Triggers approval flow preview not silent approve |
| 427 | change_approval_board | Edge case | CAB status | Status summary |
| 428 | change_approval_board | Edge case | Submit change for approval — stripe hourly sync | Specific change submission |
| 429 | change_approval_board | Edge case | Who approved this? | Approver lookup context-dependent |
| 430 | change_approval_board | Edge case | Reject change request with reason template | Rejection template; confirm before write |
| 431 | change_approval_board | Edge case | Read-only CAB queue view | Read-only queue |
| 432 | change_approval_board | Edge case | Shadow mode CAB behavior | Shadow mode explained |
| 433 | change_approval_board | Edge case | Dual control approval for tier-1 changes | Dual control policy |
| 434 | change_approval_board | Edge case | CAB summary for auditor | Auditor-safe summary |
| 435 | change_approval_board | Edge case | Change approval during active incident | Incident-period policy |
| 436 | enterprise_risk_manager | Risk register | Show operational risks for the pipeline platform | Risk register summary with scores |
| 437 | enterprise_risk_manager | Risk register | List top ten enterprise risks ranked by severity | Top-10 ranked list |
| 438 | enterprise_risk_manager | Risk register | Add risk entry for single-point-of-failure on Stripe connection | Risk entry preview; confirmation for write |
| 439 | enterprise_risk_manager | Risk register | Which risks lack assigned mitigation owners? | Unowned risk list |
| 440 | enterprise_risk_manager | Risk register | Export enterprise risk register for quarterly review | Export-oriented summary |
| 441 | enterprise_risk_manager | Risk scoring | Calculate risk score for unapproved production automation | Scored risk with methodology |
| 442 | enterprise_risk_manager | Risk scoring | Compare inherent vs residual risk for data breach scenario | Inherent/residual comparison |
| 443 | enterprise_risk_manager | Risk scoring | Show risk heat map across governance domains | Heat map narrative |
| 444 | enterprise_risk_manager | Risk scoring | Update risk likelihood after three pipeline failures this week | Likelihood update guidance |
| 445 | enterprise_risk_manager | Risk scoring | Enterprise risk appetite vs current exposure dashboard | Appetite vs exposure |
| 446 | enterprise_risk_manager | Mitigation | What mitigations exist for OpenRouter vendor concentration risk? | Mitigation list |
| 447 | enterprise_risk_manager | Mitigation | Track mitigation progress for R-1042 schema drift risk | Mitigation status tracking |
| 448 | enterprise_risk_manager | Mitigation | Recommend controls for cross-workspace credential leakage risk | Control recommendations |
| 449 | enterprise_risk_manager | Mitigation | Which mitigations are overdue? | Overdue mitigation list |
| 450 | enterprise_risk_manager | Mitigation | Cost estimate to mitigate tier-1 pipeline availability risks | Mitigation cost narrative |
| 451 | enterprise_risk_manager | Threat landscape | Summarize threat landscape for ELT data platform | Threat summary |
| 452 | enterprise_risk_manager | Threat landscape | Risk of insider threat on federated workspace admin roles | Insider threat assessment |
| 453 | enterprise_risk_manager | Threat landscape | Assess supply chain risk for third-party connectors | Supply chain risk |
| 454 | enterprise_risk_manager | Threat landscape | Cyber risk brief for AI agent automation expansion | Cyber risk brief |
| 455 | enterprise_risk_manager | Threat landscape | Geopolitical risk impact on EU data residency pipelines | Geopolitical risk narrative |
| 456 | enterprise_risk_manager | Operational risk | Operational risk from increasing concurrent pipeline runs | Ops risk analysis |
| 457 | enterprise_risk_manager | Operational risk | Risk of schedule overlap causing resource exhaustion | Resource exhaustion risk |
| 458 | enterprise_risk_manager | Operational risk | Human error risk in manual replication key configuration | Human error risk |
| 459 | enterprise_risk_manager | Operational risk | Risk register entry for missing primary keys on upsert delivery | PK-related risk |
| 460 | enterprise_risk_manager | Operational risk | Process risk when CAB backlog exceeds SLA | Process risk narrative |
| 461 | enterprise_risk_manager | Compliance risk | Compliance risks from incomplete SOC 2 evidence | Compliance risk linkage |
| 462 | enterprise_risk_manager | Compliance risk | GDPR risk assessment for Customer PII pipelines | GDPR risk without PII values |
| 463 | enterprise_risk_manager | Compliance risk | Regulatory risk if shadow mode disabled prematurely | Shadow mode transition risk |
| 464 | enterprise_risk_manager | Compliance risk | Audit finding risk recurrence tracker | Recurrence tracker summary |
| 465 | enterprise_risk_manager | Compliance risk | Map risks to SOC 2 trust service criteria | TSC mapping |
| 466 | enterprise_risk_manager | Financial risk | Financial exposure from runaway AI token usage | Financial risk quantification |
| 467 | enterprise_risk_manager | Financial risk | Budget overrun risk by federated workspace | Workspace budget risk |
| 468 | enterprise_risk_manager | Financial risk | Risk of unplanned infrastructure scaling costs | Scaling cost risk |
| 469 | enterprise_risk_manager | Financial risk | Insurance-relevant risks for data platform outage | Business interruption framing |
| 470 | enterprise_risk_manager | Financial risk | FinOps risk summary for executive committee | Exec FinOps risk brief |
| 471 | enterprise_risk_manager | Technology risk | Technology obsolescence risk for batch-only pipelines | Obsolescence risk |
| 472 | enterprise_risk_manager | Technology risk | Risk from DuckDB staging disk exhaustion at scale | Disk exhaustion risk |
| 473 | enterprise_risk_manager | Technology risk | Dependency risk on single destination warehouse | Dependency risk |
| 474 | enterprise_risk_manager | Technology risk | Model risk from ungoverned LLM version changes | Model version risk |
| 475 | enterprise_risk_manager | Technology risk | Extension risk from uncertified marketplace connectors | Extension tech risk |
| 476 | enterprise_risk_manager | Incident linkage | Update risk register after severity-2 pipeline incident | Post-incident risk update path |
| 477 | enterprise_risk_manager | Incident linkage | Which risks materialized in last month's incidents? | Materialized risk list |
| 478 | enterprise_risk_manager | Incident linkage | Near-miss events that should enter risk register | Near-miss identification |
| 479 | enterprise_risk_manager | Incident linkage | Correlate incident frequency with risk trend | Correlation narrative |
| 480 | enterprise_risk_manager | Incident linkage | Root cause patterns contributing to enterprise risk score | RCA pattern summary |
| 481 | enterprise_risk_manager | Reporting | Monthly enterprise risk report for leadership | Monthly report artifact |
| 482 | enterprise_risk_manager | Reporting | Risk committee deck talking points | Committee talking points |
| 483 | enterprise_risk_manager | Reporting | KRI dashboard for enterprise data platform | KRI dashboard summary |
| 484 | enterprise_risk_manager | Reporting | Trend analysis: enterprise risk score over 12 months | 12-month trend |
| 485 | enterprise_risk_manager | Reporting | Compare risk profile before and after automation rollout | Before/after risk profile |
| 486 | enterprise_risk_manager | Integration | Link risk R-2201 to change approval board item CAB-2026-0142 | Cross-system linkage |
| 487 | enterprise_risk_manager | Integration | Show risks blocked pending compliance evidence | Evidence-blocked risks |
| 488 | enterprise_risk_manager | Integration | Align risk register with strategic goal off-track items | Strategy-risk alignment |
| 489 | enterprise_risk_manager | Integration | Continuity risks reflected in enterprise risk register | Continuity risk integration |
| 490 | enterprise_risk_manager | Integration | Data product risks in centralized register | Product risk rollup |
| 491 | enterprise_risk_manager | Edge case | Enterprise risk manager summary | Summary without internal names |
| 492 | enterprise_risk_manager | Edge case | Risk register | Minimal prompt handled |
| 493 | enterprise_risk_manager | Edge case | What is our biggest pipeline risk? | Top risk identification |
| 494 | enterprise_risk_manager | Edge case | Show critical risks only | Critical filter applied |
| 495 | enterprise_risk_manager | Edge case | Risk score for stripe pipeline | Pipeline-scoped risk |
| 496 | enterprise_risk_manager | Edge case | Read-only risk register view | Read-only confirmation |
| 497 | enterprise_risk_manager | Edge case | Shadow mode risk acceptance workflow | Shadow mode workflow explained |
| 498 | enterprise_risk_manager | Edge case | Risk review for external auditor | Auditor-safe risk summary |
| 499 | enterprise_risk_manager | Edge case | Accept risk with executive sign-off template | Acceptance template; confirm before write |
| 500 | enterprise_risk_manager | Edge case | Close mitigated risk R-1055 | Close workflow preview |
| 501 | enterprise_risk_manager | Edge case | Emerging risks this week | Emerging risk list |
| 502 | enterprise_risk_manager | Edge case | Risk tolerance statement for data platform | Tolerance statement summary |
| 503 | enterprise_risk_manager | Reporting | Risk dashboard filtered to technology category only | Category filter applied |
| 504 | enterprise_risk_manager | Integration | Export risk heat map for enterprise mission control | Mission control integration |
| 505 | enterprise_risk_manager | Mitigation | Validate mitigation effectiveness with pipeline run metrics | Effectiveness validation narrative |
| 506 | enterprise_risk_manager | Edge case | Risk trend sparkline for executive dashboard | Trend visualization description |
| 507 | compliance_evidence_manager | Evidence collection | Collect compliance evidence for SOC 2 Type II audit | Evidence pack orientation; no secrets in output |
| 508 | compliance_evidence_manager | Evidence collection | Gather audit evidence for pipeline access controls | Access control evidence list |
| 509 | compliance_evidence_manager | Evidence collection | What evidence is missing for CC6.1 logical access? | Gap list for control CC6.1 |
| 510 | compliance_evidence_manager | Evidence collection | Automated evidence collection status for this workspace | Collection job status |
| 511 | compliance_evidence_manager | Evidence collection | Export evidence bundle for external auditor review | Export guidance; sanitized artifacts |
| 512 | compliance_evidence_manager | Control mapping | Map MantrixFlow controls to SOC 2 trust criteria | Control mapping table |
| 513 | compliance_evidence_manager | Control mapping | Show control owners and last test date for each control | Owner and test metadata |
| 514 | compliance_evidence_manager | Control mapping | Which controls failed last automated test? | Failed control list |
| 515 | compliance_evidence_manager | Control mapping | Link pipeline RLS policies to compliance controls | RLS-control linkage |
| 516 | compliance_evidence_manager | Control mapping | GDPR Article 30 processing records for pipeline data | Processing records summary |
| 517 | compliance_evidence_manager | Audit preparation | Audit readiness score for enterprise data platform | Readiness score with gaps |
| 518 | compliance_evidence_manager | Audit preparation | 30-day audit preparation checklist | Checklist artifact |
| 519 | compliance_evidence_manager | Audit preparation | Sample evidence for change management control | Sample evidence description |
| 520 | compliance_evidence_manager | Audit preparation | Walkthrough script for auditor: pipeline run lifecycle | Walkthrough script draft |
| 521 | compliance_evidence_manager | Audit preparation | Identify stale evidence older than 90 days | Stale evidence list |
| 522 | compliance_evidence_manager | Policy evidence | Evidence that credentials are never logged in API responses | Sanitization evidence reference |
| 523 | compliance_evidence_manager | Policy evidence | Prove JWT and RLS enforce cross-tenant isolation | Isolation evidence summary |
| 524 | compliance_evidence_manager | Policy evidence | Document internal auth for ELT routes X-ETL-Token | Internal auth evidence without token values |
| 525 | compliance_evidence_manager | Policy evidence | Evidence of encryption at rest for connection credentials | Encryption evidence summary |
| 526 | compliance_evidence_manager | Policy evidence | Show audit trail for Oria agent tool invocations | Agent audit trail summary |
| 527 | compliance_evidence_manager | Operational evidence | Pipeline run logs as evidence for processing integrity | Processing integrity evidence |
| 528 | compliance_evidence_manager | Operational evidence | Evidence of disk space pre-check before sync dispatch | Disk guard evidence |
| 529 | compliance_evidence_manager | Operational evidence | Callback payload audit fields delivery_outputs evidence | Callback audit field evidence |
| 530 | compliance_evidence_manager | Operational evidence | Prove destination tables are never auto-created by runner | Invariant evidence reference |
| 531 | compliance_evidence_manager | Operational evidence | Evidence pack for incremental cursor checkpoint extraction | Checkpoint evidence |
| 532 | compliance_evidence_manager | Regulatory | HIPAA compliance evidence for healthcare data pipelines | HIPAA evidence scope |
| 533 | compliance_evidence_manager | Regulatory | PCI DSS evidence scope for payment pipeline data | PCI scope without card data |
| 534 | compliance_evidence_manager | Regulatory | ISO 27001 annex control evidence mapping | ISO mapping summary |
| 535 | compliance_evidence_manager | Regulatory | Evidence for EU Schrems II cross-border transfers | Transfer evidence narrative |
| 536 | compliance_evidence_manager | Regulatory | FedRAMP-oriented evidence gap analysis if applicable | Gap analysis with applicability note |
| 537 | compliance_evidence_manager | Continuous monitoring | Continuous control monitoring dashboard for pipelines | CCM dashboard summary |
| 538 | compliance_evidence_manager | Continuous monitoring | Alerts when evidence collection jobs fail | Alert configuration summary |
| 539 | compliance_evidence_manager | Continuous monitoring | Drift detection: policy vs actual pipeline configuration | Drift evidence |
| 540 | compliance_evidence_manager | Continuous monitoring | Weekly compliance evidence freshness report | Freshness report |
| 541 | compliance_evidence_manager | Continuous monitoring | Integrate compliance monitor with automation policies | Integration summary |
| 542 | compliance_evidence_manager | Attestation | Manager attestation template for quarterly controls | Attestation template |
| 543 | compliance_evidence_manager | Attestation | Show pending attestations by control owner | Pending attestation queue |
| 544 | compliance_evidence_manager | Attestation | Evidence of board oversight for data platform risks | Board oversight evidence |
| 545 | compliance_evidence_manager | Attestation | Vendor SOC report evidence for OpenRouter dependency | Vendor SOC reference |
| 546 | compliance_evidence_manager | Attestation | Subservice organization evidence for cloud hosting | Subservice evidence list |
| 547 | compliance_evidence_manager | Remediation | Remediation plan for failed access review control | Remediation plan artifact |
| 548 | compliance_evidence_manager | Remediation | Track closure of audit findings F-2025-17 | Finding closure tracking |
| 549 | compliance_evidence_manager | Remediation | Evidence of corrective action after incident IR-442 | Corrective action evidence |
| 550 | compliance_evidence_manager | Remediation | POA&M summary for open compliance gaps | POA&M summary |
| 551 | compliance_evidence_manager | Remediation | Validate remediation evidence before auditor submission | Validation checklist |
| 552 | compliance_evidence_manager | Executive | Executive summary of compliance posture for board | Board compliance brief |
| 553 | compliance_evidence_manager | Executive | Compliance evidence completeness for CFO review | CFO-oriented completeness summary |
| 554 | compliance_evidence_manager | Executive | Regulatory exam readiness briefing | Exam readiness brief |
| 555 | compliance_evidence_manager | Executive | Summarize audit observations and management responses | Observation/response summary |
| 556 | compliance_evidence_manager | Executive | Compliance trend: open findings over four quarters | Finding trend narrative |
| 557 | compliance_evidence_manager | Cross-system | Pull evidence from change approval board for change management | CAB evidence linkage |
| 558 | compliance_evidence_manager | Cross-system | Risk register entries supporting compliance assessments | Risk-compliance crosswalk |
| 559 | compliance_evidence_manager | Cross-system | Data product governance evidence for audit | Product governance evidence |
| 560 | compliance_evidence_manager | Cross-system | AI model governance evidence for emerging regulator questions | AI governance evidence pack |
| 561 | compliance_evidence_manager | Cross-system | Extension certification evidence for integration controls | Extension cert evidence |
| 562 | compliance_evidence_manager | Edge case | SOC2 evidence pack | Minimal SOC2 prompt |
| 563 | compliance_evidence_manager | Edge case | Compliance evidence | General evidence orientation |
| 564 | compliance_evidence_manager | Edge case | Audit evidence for stripe pipeline | Pipeline-scoped evidence |
| 565 | compliance_evidence_manager | Edge case | What evidence do we have? | Inventory summary |
| 566 | compliance_evidence_manager | Edge case | Missing controls list | Missing control identification |
| 567 | compliance_evidence_manager | Edge case | Read-only evidence viewer | Read-only mode |
| 568 | compliance_evidence_manager | Edge case | Shadow mode evidence collection | Shadow mode behavior |
| 569 | compliance_evidence_manager | Edge case | Evidence redaction rules | Redaction policy explanation |
| 570 | compliance_evidence_manager | Edge case | Compliance manager summary | Agent summary |
| 571 | compliance_evidence_manager | Edge case | Evidence for last 30 days only | Date filter applied |
| 572 | compliance_evidence_manager | Edge case | Prepare evidence for ISO audit next month | Time-bound audit prep |
| 573 | compliance_evidence_manager | Continuous monitoring | Evidence gap alert when RLS policy changes without approval | Change-triggered evidence alert |
| 574 | compliance_evidence_manager | Audit preparation | Auditor request: sample of pipeline change approvals last 90 days | Sample selection methodology |
| 575 | compliance_evidence_manager | Regulatory | Evidence mapping for state privacy laws on customer data pipelines | State privacy mapping |
| 576 | compliance_evidence_manager | Cross-system | Synchronize evidence freshness with CAB approval timestamps | Timestamp sync validation |
| 577 | compliance_evidence_manager | Edge case | Compliance evidence retention schedule by control type | Retention schedule summary |
| 578 | ai_model_governance | Model inventory | Show OpenRouter model assignments for Oria agents | Model assignment table; no API keys |
| 579 | ai_model_governance | Model inventory | List all LLM models approved for enterprise use | Approved model list |
| 580 | ai_model_governance | Model inventory | Which models are blocked by governance policy? | Blocked model list |
| 581 | ai_model_governance | Model inventory | Model inventory for Release 5 enterprise specialists | Specialist model mapping |
| 582 | ai_model_governance | Model inventory | Compare model usage across workspace groups | Usage comparison |
| 583 | ai_model_governance | Policy | What is the enterprise policy on using frontier models? | Policy summary |
| 584 | ai_model_governance | Policy | Require human approval before promoting model to production tier | Approval workflow preview |
| 585 | ai_model_governance | Policy | Define allowed model tiers for PCI-scoped workspaces | Tier policy for PCI scope |
| 586 | ai_model_governance | Policy | Data residency constraints on LLM provider selection | Residency policy narrative |
| 587 | ai_model_governance | Policy | Shadow mode restrictions on autonomous model switching | Shadow mode policy |
| 588 | ai_model_governance | Evaluation | Show model evaluation results for routing accuracy | Evaluation metrics summary |
| 589 | ai_model_governance | Evaluation | Benchmark latency and cost for assigned OpenRouter models | Benchmark comparison |
| 590 | ai_model_governance | Evaluation | Quality review of Oria synthesis outputs last week | Quality review summary |
| 591 | ai_model_governance | Evaluation | A/B test plan for alternative model on failure investigation agent | A/B plan artifact |
| 592 | ai_model_governance | Evaluation | Red-team findings for prompt injection on enterprise agents | Red-team summary |
| 593 | ai_model_governance | Cost governance | Token spend by model for enterprise Oria usage | Spend breakdown without secrets |
| 594 | ai_model_governance | Cost governance | Alert thresholds for model cost overruns | Threshold configuration summary |
| 595 | ai_model_governance | Cost governance | Recommend model downgrade for non-critical specialists | Downgrade recommendation |
| 596 | ai_model_governance | Cost governance | FinOps approval required for premium model assignment | FinOps gate explanation |
| 597 | ai_model_governance | Cost governance | Unit cost per agent conversation by model | Unit cost metrics |
| 598 | ai_model_governance | Safety | Safety review for models processing customer support transcripts | Safety review artifact |
| 599 | ai_model_governance | Safety | PII leakage risk assessment for LLM synthesis path | PII risk assessment |
| 600 | ai_model_governance | Safety | Content filtering policy for model outputs in executive summaries | Filtering policy |
| 601 | ai_model_governance | Safety | Incident: model produced non-compliant recommendation — governance response | Incident response guidance |
| 602 | ai_model_governance | Safety | Verify models cannot exfiltrate connection credentials | Credential exfiltration guard summary |
| 603 | ai_model_governance | Lifecycle | Model deprecation timeline for legacy OpenRouter IDs | Deprecation timeline |
| 604 | ai_model_governance | Lifecycle | Migration plan when primary model reaches end of support | Migration plan draft |
| 605 | ai_model_governance | Lifecycle | Version pinning policy for production agent runtime | Pinning policy |
| 606 | ai_model_governance | Lifecycle | Rollback procedure after bad model deployment | Rollback procedure |
| 607 | ai_model_governance | Lifecycle | Certification checklist for new model before enterprise approval | Certification checklist |
| 608 | ai_model_governance | Oversight | Executive summary of AI model governance program | Executive program summary |
| 609 | ai_model_governance | Oversight | Board briefing on generative AI usage in data platform | Board AI brief |
| 610 | ai_model_governance | Oversight | Audit evidence for model change management | Change management evidence orientation |
| 611 | ai_model_governance | Oversight | Third-party model vendor risk assessment OpenRouter | Vendor risk without keys |
| 612 | ai_model_governance | Oversight | Alignment with EU AI Act requirements for internal copilots | Regulatory alignment narrative |
| 613 | ai_model_governance | Agent-specific | Governance review for enterprise_mission_control model routing | Agent routing governance |
| 614 | ai_model_governance | Agent-specific | Which model handles executive_intelligence summaries? | Model routing answer |
| 615 | ai_model_governance | Agent-specific | Ensure compliance_evidence_manager uses read-only model tier | Tier enforcement check |
| 616 | ai_model_governance | Agent-specific | Model assignment audit for change_approval_board specialist | Specialist audit entry |
| 617 | ai_model_governance | Agent-specific | Fallback model chain when OpenRouter primary unavailable | Fallback chain description |
| 618 | ai_model_governance | Approval | Submit model change request for CAB review | CAB-linked model change preview |
| 619 | ai_model_governance | Approval | Pending model approvals in governance queue | Pending queue |
| 620 | ai_model_governance | Approval | Who approved current OPENROUTER_MODEL_ORIA assignment? | Approver attribution |
| 621 | ai_model_governance | Approval | Reject unapproved model swap attempted in workspace | Rejection/block explanation |
| 622 | ai_model_governance | Approval | Emergency model switch procedure during provider outage | Emergency procedure |
| 623 | ai_model_governance | Monitoring | Monitor hallucination rate in governance artifact generation | Hallucination monitoring summary |
| 624 | ai_model_governance | Monitoring | Real-time alerts on policy-violating model requests | Alert summary |
| 625 | ai_model_governance | Monitoring | Dashboard: model errors, timeouts, and fallbacks | Ops dashboard narrative |
| 626 | ai_model_governance | Monitoring | Weekly model governance health report | Weekly health report |
| 627 | ai_model_governance | Monitoring | Track AGENT_LLM_SYNTHESIS impact on model usage | Synthesis flag impact |
| 628 | ai_model_governance | Edge case | Model governance | Minimal prompt |
| 629 | ai_model_governance | Edge case | OpenRouter models | Provider model list orientation |
| 630 | ai_model_governance | Edge case | Which model is Oria using? | Public Oria identity; model info if policy allows |
| 631 | ai_model_governance | Edge case | Approve gpt-4 for production | Approval preview not auto-approve |
| 632 | ai_model_governance | Edge case | Block uncertified models | Block policy explanation |
| 633 | ai_model_governance | Edge case | Read-only model policy | Read-only view |
| 634 | ai_model_governance | Edge case | Shadow mode model changes | Shadow mode for model writes |
| 635 | ai_model_governance | Edge case | Model governance for auditor | Auditor-safe summary |
| 636 | ai_model_governance | Edge case | Local fallback model governance | Fallback governance |
| 637 | ai_model_governance | Edge case | Model risk score summary | Risk score rollup |
| 638 | ai_model_governance | Edge case | AI governance executive one-pager | One-page exec summary |
| 639 | ai_model_governance | Evaluation | Compare hallucination rates between assigned production models | Comparative hallucination metrics |
| 640 | ai_model_governance | Safety | Review system prompt changes for enterprise agent specialists | Prompt change review process |
| 641 | ai_model_governance | Monitoring | Detect unauthorized model endpoint overrides in workspace config | Override detection summary |
| 642 | ai_model_governance | Lifecycle | Sunset plan for models failing enterprise quality threshold | Sunset plan artifact |
| 643 | ai_model_governance | Approval | Dual approval for tier-1 model changes affecting executive summaries | Dual approval workflow |
| 644 | ai_model_governance | Agent-specific | Model governance audit trail for federated workspace overrides | Override audit trail |
| 645 | ai_model_governance | Cost governance | Cap daily token spend per enterprise agent specialist | Spend cap configuration preview |
| 646 | ai_model_governance | Oversight | Quarterly AI governance committee pack | Committee pack structure |
| 647 | ai_model_governance | Policy | Restrict experimental models to sandbox workspaces only | Sandbox restriction policy |
| 648 | ai_model_governance | Edge case | Model drift detection after provider silent upgrade | Drift detection guidance |
| 649 | extension_governance | Extension review | Review Stripe connector extension risk for enterprise deployment | Risk review artifact with severity |
| 650 | extension_governance | Extension review | List all installed extensions pending governance review | Pending review queue |
| 651 | extension_governance | Extension review | Certification status for HubSpot marketplace connector | Certification status |
| 652 | extension_governance | Extension review | Compare security posture of Postgres vs MySQL connectors | Comparative security summary |
| 653 | extension_governance | Extension review | Extension risk score for custom Shopify integration | Risk score with factors |
| 654 | extension_governance | Marketplace | Which marketplace connectors are approved for production? | Approved marketplace list |
| 655 | extension_governance | Marketplace | Block uncertified connector from federated workspace APAC-1 | Block action preview; confirmation |
| 656 | extension_governance | Marketplace | Marketplace listing audit for Salesforce extension | Listing audit summary |
| 657 | extension_governance | Marketplace | Vendor security questionnaire status for Snowflake connector | Questionnaire status |
| 658 | extension_governance | Marketplace | Track extension version drift across workspaces | Version drift table |
| 659 | extension_governance | Security assessment | OAuth scope review for Stripe connector extension | Scope review without secrets |
| 660 | extension_governance | Security assessment | Pen test findings remediation for Kafka connector | Pen test remediation tracker |
| 661 | extension_governance | Security assessment | SBOM availability for enterprise connector bundle | SBOM status |
| 662 | extension_governance | Security assessment | CVE scan results for installed connector versions | CVE summary without exploit details overload |
| 663 | extension_governance | Security assessment | Secrets handling review for webhook integration extension | Secrets review; no raw secrets |
| 664 | extension_governance | Data handling | Does extension transmit PII outside approved regions? | Data flow assessment |
| 665 | extension_governance | Data handling | Extension data retention policy compliance check | Retention compliance |
| 666 | extension_governance | Data handling | Review logging behavior of diagnostic extensions | Logging review |
| 667 | extension_governance | Data handling | Minimization assessment for analytics tracking extension | Minimization assessment |
| 668 | extension_governance | Data handling | Cross-border transfer impact of EU CRM connector | Transfer impact narrative |
| 669 | extension_governance | Lifecycle | Extension onboarding checklist for new connector | Onboarding checklist |
| 670 | extension_governance | Lifecycle | Deprecate legacy Stripe API version extension | Deprecation plan |
| 671 | extension_governance | Lifecycle | Upgrade path for connector major version migration | Upgrade path artifact |
| 672 | extension_governance | Lifecycle | End-of-life timeline for unsupported extensions | EOL timeline |
| 673 | extension_governance | Lifecycle | Rollback plan after failed extension upgrade | Rollback plan |
| 674 | extension_governance | Policy | Enterprise policy for side-loaded vs marketplace extensions | Side-load policy |
| 675 | extension_governance | Policy | Require CAB approval for tier-1 connector changes | CAB requirement for connectors |
| 676 | extension_governance | Policy | Allowed extension permissions in production workspaces | Permission allowlist summary |
| 677 | extension_governance | Policy | Shadow mode: block unreviewed extension activation | Shadow mode block explanation |
| 678 | extension_governance | Policy | Segregation of duties on extension approval roles | SoD on extension approvals |
| 679 | extension_governance | Monitoring | Monitor extension API call anomalies | Anomaly monitoring summary |
| 680 | extension_governance | Monitoring | Alert on extension accessing undeclared scopes | Scope violation alerts |
| 681 | extension_governance | Monitoring | Weekly extension governance compliance report | Weekly compliance report |
| 682 | extension_governance | Monitoring | Failed extension health checks across federation | Health check failures |
| 683 | extension_governance | Monitoring | Usage metrics: which extensions consume most API quota | Quota usage ranking |
| 684 | extension_governance | Integration | Link extension review to enterprise risk register | Risk linkage |
| 685 | extension_governance | Integration | Compliance evidence for connector certification controls | Compliance evidence cross-ref |
| 686 | extension_governance | Integration | FinOps cost of licensed marketplace connectors | Licensed cost summary |
| 687 | extension_governance | Integration | Data product dependencies on specific extensions | Dependency mapping |
| 688 | extension_governance | Integration | Continuity impact if Stripe extension unavailable | Continuity impact analysis |
| 689 | extension_governance | Executive | Executive summary of integration risk landscape | Exec integration risk brief |
| 690 | extension_governance | Executive | Board update on third-party connector concentration | Concentration risk board brief |
| 691 | extension_governance | Executive | Recommend enterprise standard connector catalog | Standard catalog recommendation |
| 692 | extension_governance | Executive | Extension governance KPIs for leadership | KPI summary |
| 693 | extension_governance | Executive | Summary of rejected high-risk extensions this year | Rejection summary |
| 694 | extension_governance | Operational | How to request expedited review for critical connector | Expedited review process |
| 695 | extension_governance | Operational | Extension review SLA and current backlog | SLA and backlog |
| 696 | extension_governance | Operational | Assign reviewer for pending Postgres extension update | Reviewer assignment preview |
| 697 | extension_governance | Operational | Document exception for uncertified dev-only connector | Exception documentation path |
| 698 | extension_governance | Operational | Validate extension compatibility with strict ELT invariants | Invariant compatibility check |
| 699 | extension_governance | Edge case | Extension governance | Minimal prompt |
| 700 | extension_governance | Edge case | Review stripe connector | Connector-specific review |
| 701 | extension_governance | Edge case | Is hubspot connector certified? | Certification yes/no with detail |
| 702 | extension_governance | Edge case | Block extension | Block preview with confirmation |
| 703 | extension_governance | Edge case | Approved connectors list | Approved list |
| 704 | extension_governance | Edge case | Extension risk | Risk summary |
| 705 | extension_governance | Edge case | Read-only extension catalog | Read-only catalog |
| 706 | extension_governance | Edge case | Shadow mode extension installs | Shadow mode behavior |
| 707 | extension_governance | Edge case | Extension governance for audit | Audit-oriented summary |
| 708 | extension_governance | Edge case | Custom connector governance path | Custom connector process |
| 709 | extension_governance | Edge case | Extension summary for CISO | CISO-focused summary |
| 710 | extension_governance | Security assessment | Review extension network egress allowlist compliance | Egress allowlist review |
| 711 | extension_governance | Marketplace | Quarterly recertification calendar for production connectors | Recertification calendar |
| 712 | extension_governance | Monitoring | Detect shadow IT connectors not in governance registry | Shadow connector detection |
| 713 | extension_governance | Policy | Mandatory security review for extensions accessing billing APIs | Billing API review policy |
| 714 | extension_governance | Integration | Map extension permissions to least-privilege baseline | Least-privilege mapping |
| 715 | extension_governance | Operational | Expedite review queue for revenue-critical Stripe extension patch | Expedited patch review |
| 716 | extension_governance | Data handling | Validate extension logs exclude connection credential fields | Log scrubbing validation |
| 717 | extension_governance | Executive | Quarterly third-party integration risk committee summary | Committee summary artifact |
| 718 | extension_governance | Lifecycle | Freeze extension upgrades during enterprise change freeze | Freeze policy for extensions |
| 719 | extension_governance | Edge case | Extension waiver documentation for acquisition integration | Waiver documentation path |
| 720 | enterprise_finops_planner | Spend summary | FinOps summary for AI and ELT spend this month | Monthly spend summary without payment credentials |
| 721 | enterprise_finops_planner | Spend summary | Break down enterprise data platform costs by category | Category breakdown |
| 722 | enterprise_finops_planner | Spend summary | Compare actual vs budgeted spend for Q2 | Budget variance analysis |
| 723 | enterprise_finops_planner | Spend summary | Top five cost drivers in pipeline operations | Cost driver ranking |
| 724 | enterprise_finops_planner | Spend summary | Executive FinOps briefing for CFO | CFO briefing artifact |
| 725 | enterprise_finops_planner | Chargeback | Chargeback report by federated workspace | Chargeback table |
| 726 | enterprise_finops_planner | Chargeback | Show cost allocation tags on Stripe All Streams pipeline | Tag visibility |
| 727 | enterprise_finops_planner | Chargeback | Recommend chargeback model for shared destination warehouse | Chargeback model recommendation |
| 728 | enterprise_finops_planner | Chargeback | Which business unit exceeded allocated budget? | Budget exceedance list |
| 729 | enterprise_finops_planner | Chargeback | Unit economics: cost per million rows processed | Unit economics metrics |
| 730 | enterprise_finops_planner | Optimization | Optimize pipeline run costs in enterprise federation | Optimization recommendations |
| 731 | enterprise_finops_planner | Optimization | Identify idle pipelines consuming scheduled run budget | Idle pipeline list |
| 732 | enterprise_finops_planner | Optimization | Rightsize concurrency settings to reduce compute spend | Rightsizing recommendation |
| 733 | enterprise_finops_planner | Optimization | Recommend off-peak scheduling for cost reduction | Off-peak schedule plan |
| 734 | enterprise_finops_planner | Optimization | AI token usage optimization for Oria enterprise agents | Token optimization plan |
| 735 | enterprise_finops_planner | Forecasting | Forecast ELT spend for next quarter | Forecast with assumptions and confidence |
| 736 | enterprise_finops_planner | Forecasting | What-if cost impact of doubling HubSpot sync frequency | What-if narrative |
| 737 | enterprise_finops_planner | Forecasting | Predict budget overrun risk by September | Overrun risk forecast |
| 738 | enterprise_finops_planner | Forecasting | Scenario: 30% pipeline volume growth — FinOps impact | Growth scenario analysis |
| 739 | enterprise_finops_planner | Forecasting | Long-range FinOps plan for data mesh expansion | Long-range plan outline |
| 740 | enterprise_finops_planner | Guardrails | Configure spend guardrails for AI inference | Guardrail configuration preview; confirm for write |
| 741 | enterprise_finops_planner | Guardrails | Which workspaces breached FinOps guardrails this week? | Breach list |
| 742 | enterprise_finops_planner | Guardrails | Enterprise policy for premium OpenRouter model usage | Premium model cost policy |
| 743 | enterprise_finops_planner | Guardrails | Alert when monthly spend exceeds 80% of budget | Alert threshold guidance |
| 744 | enterprise_finops_planner | Guardrails | Hard stop vs soft warning on budget limits | Limit behavior explanation |
| 745 | enterprise_finops_planner | Showback | Showback dashboard for data engineering leadership | Showback dashboard summary |
| 746 | enterprise_finops_planner | Showback | Cost transparency report for product owners | Product owner report |
| 747 | enterprise_finops_planner | Showback | Trend: FinOps maturity across business units | Maturity trend narrative |
| 748 | enterprise_finops_planner | Showback | Communicate cost spikes after backfill operations | Spike communication template |
| 749 | enterprise_finops_planner | Showback | Compare showback vs chargeback adoption | Adoption comparison |
| 750 | enterprise_finops_planner | Vendor | OpenRouter invoice analysis and anomaly detection | Invoice analysis without account secrets |
| 751 | enterprise_finops_planner | Vendor | Cloud warehouse cost attribution for Postgres destination | Warehouse cost attribution |
| 752 | enterprise_finops_planner | Vendor | Vendor consolidation opportunities for connectors | Consolidation opportunities |
| 753 | enterprise_finops_planner | Vendor | Negotiation brief: enterprise discount eligibility | Negotiation brief |
| 754 | enterprise_finops_planner | Vendor | Multi-cloud cost comparison for staging storage | Multi-cloud comparison |
| 755 | enterprise_finops_planner | Governance | FinOps approval required for tier-1 cost-increasing changes | CAB-FinOps gate linkage |
| 756 | enterprise_finops_planner | Governance | Align FinOps objectives with strategic goals | Goal alignment summary |
| 757 | enterprise_finops_planner | Governance | FinOps controls mapped to SOC 2 evidence | Control-evidence mapping |
| 758 | enterprise_finops_planner | Governance | Risk register entries for uncontrolled spend | Spend risk linkage |
| 759 | enterprise_finops_planner | Governance | Executive approval for budget exception request | Exception approval workflow |
| 760 | enterprise_finops_planner | Reporting | Monthly FinOps committee pack | Committee pack structure |
| 761 | enterprise_finops_planner | Reporting | Dashboard KPIs: cost, efficiency, forecast accuracy | KPI dashboard narrative |
| 762 | enterprise_finops_planner | Reporting | Year-over-year FinOps performance summary | YoY summary |
| 763 | enterprise_finops_planner | Reporting | Savings realized from last optimization initiative | Savings tracking |
| 764 | enterprise_finops_planner | Reporting | FinOps metrics for board risk committee | Board metrics summary |
| 765 | enterprise_finops_planner | Pipeline-specific | Cost profile for Stripe All Streams incremental runs | Pipeline cost profile |
| 766 | enterprise_finops_planner | Pipeline-specific | Compare cost of full-table vs incremental for Salesforce | Sync mode cost comparison |
| 767 | enterprise_finops_planner | Pipeline-specific | FinOps impact of adding three new streams to HubSpot pipeline | Stream addition cost estimate |
| 768 | enterprise_finops_planner | Pipeline-specific | Most expensive pipeline runs last 7 days | Expensive run ranking |
| 769 | enterprise_finops_planner | Pipeline-specific | Cost of failed runs and retries this month | Failure/retry cost |
| 770 | enterprise_finops_planner | Edge case | FinOps summary | Minimal prompt |
| 771 | enterprise_finops_planner | Edge case | What did we spend on AI? | AI spend answer |
| 772 | enterprise_finops_planner | Edge case | Budget status | Budget status summary |
| 773 | enterprise_finops_planner | Edge case | Cost optimization help | Optimization guidance |
| 774 | enterprise_finops_planner | Edge case | Enterprise finops planner overview | Overview without internal names |
| 775 | enterprise_finops_planner | Edge case | Read-only spend view | Read-only confirmation |
| 776 | enterprise_finops_planner | Edge case | Shadow mode budget changes | Shadow mode for budget writes |
| 777 | enterprise_finops_planner | Edge case | FinOps for auditor | Auditor-safe spend summary |
| 778 | enterprise_finops_planner | Edge case | Chargeback for workspace EU-1 | Workspace-scoped chargeback |
| 779 | enterprise_finops_planner | Edge case | Forecast confidence intervals | Forecast includes confidence |
| 780 | enterprise_finops_planner | Edge case | FinOps one-pager for CEO | CEO one-pager |
| 781 | enterprise_finops_planner | Optimization | Identify duplicate pipeline runs inflating spend across federation | Duplicate run cost analysis |
| 782 | enterprise_finops_planner | Forecasting | Sensitivity analysis: OpenRouter price increase scenarios | Price sensitivity analysis |
| 783 | enterprise_finops_planner | Guardrails | Workspace-level hard budget caps for sandbox environments | Sandbox cap configuration preview |
| 784 | enterprise_finops_planner | Showback | Departmental showback for AI agent usage by specialist type | Specialist-type showback |
| 785 | enterprise_finops_planner | Vendor | Compare OpenRouter vs alternative provider total cost of ownership | TCO comparison with assumptions |
| 786 | enterprise_finops_planner | Governance | FinOps review gate before approving premium model assignments | Model-FinOps gate linkage |
| 787 | enterprise_finops_planner | Reporting | Anomaly detection on weekly ELT spend spikes | Spend anomaly summary |
| 788 | enterprise_finops_planner | Pipeline-specific | FinOps recommendation for consolidating HubSpot federation pipelines | Consolidation FinOps case |
| 789 | enterprise_finops_planner | Edge case | Cost allocation for shared pgmq worker infrastructure | Shared infra allocation method |
| 790 | enterprise_finops_planner | Edge case | FinOps impact statement template for CAB submissions | CAB FinOps template |
| 791 | continuity_command_center | Continuity status | Continuity status during severe pipeline platform incident | Continuity status with command structure |
| 792 | continuity_command_center | Continuity status | Show business continuity dashboard for data operations | BCP dashboard summary |
| 793 | continuity_command_center | Continuity status | Are we in continuity activation mode right now? | Activation mode status |
| 794 | continuity_command_center | Continuity status | Summarize active continuity incidents and severity | Active incident summary |
| 795 | continuity_command_center | Continuity status | Continuity command center weekly readiness report | Readiness report |
| 796 | continuity_command_center | DR planning | Disaster recovery plan for critical Stripe-to-warehouse pipelines | DR plan artifact |
| 797 | continuity_command_center | DR planning | RTO and RPO targets for tier-1 data products | RTO/RPO target table |
| 798 | continuity_command_center | DR planning | Failover procedure for primary Postgres destination loss | Failover procedure outline |
| 799 | continuity_command_center | DR planning | Last successful disaster recovery test results | DR test results summary |
| 800 | continuity_command_center | DR planning | Gap analysis: can we meet 4-hour RTO today? | Gap analysis with honest assessment |
| 801 | continuity_command_center | Incident command | Activate continuity command for region-wide cloud outage | Activation preview; human confirmation |
| 802 | continuity_command_center | Incident command | Incident commander checklist for P1 pipeline failure | IC checklist |
| 803 | continuity_command_center | Incident command | Assign roles: IC, comms lead, technical lead for active incident | Role assignment guidance |
| 804 | continuity_command_center | Incident command | Situation report template for executive stakeholders | Situation report template |
| 805 | continuity_command_center | Incident command | When to escalate from workspace incident to enterprise continuity | Escalation criteria |
| 806 | continuity_command_center | Recovery | Recovery playbook for DuckDB staging disk exhaustion event | Recovery playbook reference |
| 807 | continuity_command_center | Recovery | Steps to restore pipeline checkpoints after database corruption | Checkpoint recovery steps |
| 808 | continuity_command_center | Recovery | Validate data integrity post-failover to secondary warehouse | Integrity validation plan |
| 809 | continuity_command_center | Recovery | Estimated time to restore all critical pipelines | Recovery ETA with assumptions |
| 810 | continuity_command_center | Recovery | Post-incident recovery priority order for federated workspaces | Priority order |
| 811 | continuity_command_center | Communication | Executive communication draft during prolonged outage | Exec comms draft |
| 812 | continuity_command_center | Communication | Customer impact notification template for SLA breach | Customer notification template |
| 813 | continuity_command_center | Communication | Internal status page update for continuity incident | Status update draft |
| 814 | continuity_command_center | Communication | Regulatory notification requirements after data availability incident | Regulatory notification guidance |
| 815 | continuity_command_center | Communication | War room bridge details and cadence for leadership calls | War room logistics summary |
| 816 | continuity_command_center | Testing | Schedule next tabletop exercise for pipeline DR | Exercise scheduling preview |
| 817 | continuity_command_center | Testing | Tabletop scenario: OpenRouter complete outage during month-end close | Tabletop scenario artifact |
| 818 | continuity_command_center | Testing | Game day results for federated workspace failover | Game day results summary |
| 819 | continuity_command_center | Testing | Remediation items from last continuity exercise | Remediation tracker |
| 820 | continuity_command_center | Testing | Validate runbooks are current for all tier-1 pipelines | Runbook currency check |
| 821 | continuity_command_center | Dependencies | Map critical dependencies for continuity planning | Dependency map narrative |
| 822 | continuity_command_center | Dependencies | Single points of failure in current architecture | SPOF list |
| 823 | continuity_command_center | Dependencies | Third-party dependency: OpenRouter continuity plan | Vendor continuity summary |
| 824 | continuity_command_center | Dependencies | Cross-region dependency graph for ELT stack | Regional dependency graph |
| 825 | continuity_command_center | Dependencies | Impact if pgmq queue worker unavailable for 2 hours | Queue outage impact |
| 826 | continuity_command_center | Resilience | Resilience score for enterprise pipeline portfolio | Resilience score methodology |
| 827 | continuity_command_center | Resilience | Recommend redundancy for HubSpot federation pipelines | Redundancy recommendations |
| 828 | continuity_command_center | Resilience | Backup frequency and retention for pipeline configuration | Backup policy summary |
| 829 | continuity_command_center | Resilience | Immutable audit log availability during disaster | Audit log resilience |
| 830 | continuity_command_center | Resilience | Continuity implications of strict ELT invariant enforcement | Invariant-aware continuity note |
| 831 | continuity_command_center | Integration | Link continuity plan to enterprise risk register | Risk linkage |
| 832 | continuity_command_center | Integration | CAB approval for continuity-related production changes | CAB integration |
| 833 | continuity_command_center | Integration | Compliance evidence for business continuity controls | BCP control evidence |
| 834 | continuity_command_center | Integration | FinOps cost of hot standby vs cold standby | Standby cost comparison |
| 835 | continuity_command_center | Integration | Strategic goal alignment for continuity investments | Investment alignment |
| 836 | continuity_command_center | Monitoring | Early warning indicators before continuity event | Early warning KPIs |
| 837 | continuity_command_center | Monitoring | Monitor RPO breach risk for critical pipelines | RPO breach monitoring |
| 838 | continuity_command_center | Monitoring | Automated continuity health checks status | Health check status |
| 839 | continuity_command_center | Monitoring | Alert routing to continuity command center | Alert routing summary |
| 840 | continuity_command_center | Monitoring | Synthetic transaction monitoring for pipeline availability | Synthetic monitoring summary |
| 841 | continuity_command_center | Executive | CEO briefing: are we prepared for catastrophic data platform failure? | CEO preparedness brief |
| 842 | continuity_command_center | Executive | Board summary of continuity program maturity | Board maturity summary |
| 843 | continuity_command_center | Executive | Investment case for improving RTO from 8h to 2h | Investment case narrative |
| 844 | continuity_command_center | Executive | Continuity KPIs for quarterly business review | QBR continuity KPIs |
| 845 | continuity_command_center | Executive | Compare continuity posture to industry peers | Peer comparison with caveats |
| 846 | continuity_command_center | Edge case | Continuity command center | Minimal prompt |
| 847 | continuity_command_center | Edge case | DR status | DR status summary |
| 848 | continuity_command_center | Edge case | Are we in disaster mode? | Mode check |
| 849 | continuity_command_center | Edge case | Continuity during stripe outage | Outage-specific continuity |
| 850 | continuity_command_center | Edge case | Read-only continuity dashboard | Read-only view |
| 851 | continuity_command_center | Edge case | Shadow mode continuity activation | Shadow mode blocks auto-activation |
| 852 | continuity_command_center | Edge case | RTO for postgres pipeline | Pipeline-specific RTO |
| 853 | continuity_command_center | Edge case | Incident commander who? | Role assignment info |
| 854 | continuity_command_center | Edge case | Continuity summary for auditor | Auditor-safe summary |
| 855 | continuity_command_center | Edge case | Failover test plan draft | Test plan artifact |
| 856 | continuity_command_center | Edge case | Severe incident command overview | Command overview |
| 857 | continuity_command_center | DR planning | Multi-region failover runbook for Go API and ELT services | Multi-service runbook outline |
| 858 | continuity_command_center | Incident command | Handoff checklist between incident commander shifts | Shift handoff checklist |
| 859 | continuity_command_center | Recovery | Priority restore order when only partial destination capacity available | Partial capacity priority plan |
| 860 | continuity_command_center | Testing | Chaos engineering plan for queue worker failure injection | Chaos plan artifact |
| 861 | continuity_command_center | Executive | Post-incident executive summary template for continuity events | Post-incident exec template |

## Summary

| Metric | Value |
| --- | --- |
| Total prompts | 861 |
| Agents | 12 |
| Release | 5 — enterprise |
