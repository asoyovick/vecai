import { useEffect, useRef, useState } from "react"
import { RotateCcw, Send, Sparkles } from "lucide-react"
import type { NavigateFn } from "../App"
import { api, ApiRequestError } from "../lib/api"
import type { AssistantAnswer } from "../lib/api"

interface Props {
  onNavigate: NavigateFn
}

interface Message {
  id: string
  role: "user" | "assistant"
  text?: string
  answer?: AssistantAnswer
  error?: string
}

const SUGGESTED_PROMPTS = [
  {
    label: "Summarize latest site reports",
    prompt: "Summarize the latest site reports",
  },
  {
    label: "Materials running low",
    prompt: "Which materials are likely to run low this week?",
  },
  {
    label: "Tasks behind schedule",
    prompt: "Show me all tasks that are behind schedule or blocked.",
  },
  {
    label: "Budget variance",
    prompt:
      "Compare the planned and actual budget spend and explain any variances.",
  },
  {
    label: "Weekly status report",
    prompt: "Generate a weekly project status report for all active projects.",
  },
  {
    label: "Why is a project behind?",
    prompt:
      "Explain why Lakeview Residence might be at risk of missing its completion date.",
  },
]

const TONE_STYLE: Record<string, string> = {
  danger: "border-l-2 border-red-500 pl-2.5",
  warn: "border-l-2 border-orange-400 pl-2.5",
  success: "border-l-2 border-green-500 pl-2.5",
  default: "",
}

function AnswerCard({ answer }: { answer: AssistantAnswer }) {
  return (
    <div>
      <h3 className="text-sm font-bold text-slate-900">{answer.title}</h3>
      <p className="mt-0.5 text-xs leading-relaxed text-slate-500">
        {answer.intro}
      </p>
      <div className="mt-3 space-y-3.5">
        {answer.sections.map((section, i) => (
          <div key={i}>
            {section.heading && (
              <p className="mb-1.5 text-xs font-semibold uppercase tracking-wide text-slate-500">
                {section.heading}
              </p>
            )}
            <ul className="space-y-1.5">
              {section.lines.map((line, j) => (
                <li
                  key={j}
                  className={`text-sm leading-relaxed text-slate-700 ${TONE_STYLE[line.tone ?? "default"]}`}
                >
                  {line.text}
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
      <p className="mt-3 border-t border-slate-100 pt-2 text-xs text-slate-400">
        {answer.footnote}
      </p>
    </div>
  )
}

export default function AIAssistantPage(_props: Props) {
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState("")
  const [loading, setLoading] = useState(false)
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth", block: "nearest" })
  }, [messages, loading])

  const send = async (text: string) => {
    const question = text.trim()
    if (!question || loading) return
    const userMsg: Message = {
      id: `u-${Date.now()}`,
      role: "user",
      text: question,
    }
    setMessages((m) => [...m, userMsg])
    setInput("")
    setLoading(true)
    try {
      const data = await api.post<{ answer: AssistantAnswer }>(
        "/api/assistant",
        { question },
      )
      setMessages((m) => [
        ...m,
        { id: `a-${Date.now()}`, role: "assistant", answer: data.answer },
      ])
    } catch (err) {
      setMessages((m) => [
        ...m,
        {
          id: `e-${Date.now()}`,
          role: "assistant",
          error:
            err instanceof ApiRequestError
              ? err.message
              : "The assistant is unavailable right now.",
        },
      ])
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      void send(input)
    }
  }

  return (
    <div className="flex h-[calc(100vh-56px)] flex-col lg:h-screen">
      {/* Header */}
      <div className="shrink-0 border-b border-slate-200 bg-white px-4 py-4 sm:px-8">
        <div className="flex items-center gap-3">
          <span
            className="flex h-9 w-9 items-center justify-center rounded bg-slate-900"
            aria-hidden="true"
          >
            <Sparkles size={15} className="text-orange-500" />
          </span>
          <div>
            <h1 className="text-lg font-bold tracking-tight text-slate-900">
              Project assistant
            </h1>
            <p className="text-xs text-slate-500">
              Answers computed from your live project data
            </p>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div
        className="flex-1 overflow-y-auto px-4 py-6 sm:px-8"
        aria-live="polite"
      >
        <div className="mx-auto max-w-3xl space-y-5">
          {messages.length === 0 && !loading && (
            <div className="rounded border border-slate-200 bg-white p-5">
              <div className="flex items-start gap-3">
                <span
                  className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-900"
                  aria-hidden="true"
                >
                  <Sparkles size={13} className="text-orange-500" />
                </span>
                <div>
                  <p className="text-sm leading-relaxed text-slate-700">
                    Ask about your projects in plain language. I can summarize
                    site reports, forecast material shortages, list blocked or
                    overdue tasks, analyse budget variance, or compile a weekly
                    status report, all from the data your team has recorded.
                  </p>
                  <p className="mt-2 text-xs text-slate-400">
                    Start with one of the suggested questions below.
                  </p>
                </div>
              </div>
            </div>
          )}

          {messages.map((msg) =>
            msg.role === "user" ? (
              <div key={msg.id} className="flex justify-end">
                <div className="max-w-[80%] rounded bg-slate-900 px-4 py-2.5 text-sm text-white">
                  {msg.text}
                </div>
              </div>
            ) : (
              <div key={msg.id} className="flex items-start gap-3">
                <span
                  className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-900"
                  aria-hidden="true"
                >
                  <Sparkles size={13} className="text-orange-500" />
                </span>
                <div className="min-w-0 max-w-[85%] flex-1 rounded border border-slate-200 bg-white p-4">
                  {msg.error ? (
                    <p className="text-sm text-red-700" role="alert">
                      {msg.error}
                    </p>
                  ) : (
                    msg.answer && <AnswerCard answer={msg.answer} />
                  )}
                </div>
              </div>
            ),
          )}

          {loading && (
            <div className="flex items-start gap-3">
              <span
                className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-900"
                aria-hidden="true"
              >
                <Sparkles size={13} className="text-orange-500" />
              </span>
              <div
                className="flex items-center gap-2 rounded border border-slate-200 bg-white px-4 py-3 text-sm text-slate-500"
                role="status"
              >
                <span
                  className="h-1.5 w-1.5 animate-pulse rounded-full bg-slate-400"
                  aria-hidden="true"
                />
                Analysing project data…
              </div>
            </div>
          )}
          <div ref={bottomRef} />
        </div>
      </div>

      {/* Suggested prompts */}
      <div className="shrink-0 border-t border-slate-200 bg-white px-4 pt-3 sm:px-8">
        <div className="mx-auto flex max-w-3xl gap-2 overflow-x-auto pb-1">
          {SUGGESTED_PROMPTS.map((p) => (
            <button
              key={p.label}
              onClick={() => void send(p.prompt)}
              disabled={loading}
              className="shrink-0 rounded border border-slate-200 bg-white px-2.5 py-1 text-xs text-slate-600 transition-colors hover:border-slate-400 hover:text-slate-900 disabled:opacity-50"
            >
              {p.label}
            </button>
          ))}
        </div>
      </div>

      {/* Input */}
      <div className="shrink-0 bg-white px-4 pb-4 pt-2 sm:px-8">
        <div className="mx-auto max-w-3xl">
          <div className="flex items-end gap-2">
            <label htmlFor="ai-input" className="sr-only">
              Ask the assistant
            </label>
            <textarea
              id="ai-input"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Ask about materials, tasks, budget, site reports, or request a status report…"
              rows={2}
              className="flex-1 resize-none rounded border border-slate-300 bg-white px-3.5 py-2.5 text-sm text-slate-900 outline-none transition-colors placeholder:text-slate-400 focus:border-slate-800 focus:ring-2 focus:ring-slate-900/10"
            />
            <button
              onClick={() => void send(input)}
              disabled={!input.trim() || loading}
              aria-label="Send question"
              className="flex shrink-0 items-center justify-center rounded bg-slate-900 p-3 text-white transition-colors hover:bg-slate-700 disabled:bg-slate-200 disabled:text-slate-400"
            >
              <Send size={15} aria-hidden="true" />
            </button>
          </div>
          <div className="mt-2 flex items-center justify-between">
            <p className="text-xs text-slate-400">
              Answers are computed from your recorded project data. Verify
              critical decisions with your team.
            </p>
            {messages.length > 0 && (
              <button
                onClick={() => setMessages([])}
                className="flex items-center gap-1 text-xs text-slate-400 hover:text-slate-600"
              >
                <RotateCcw size={11} aria-hidden="true" /> Clear
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
