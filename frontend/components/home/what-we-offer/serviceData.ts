import { Building2, Ruler, Calculator, BrainCircuit, PackageSearch, Users } from "lucide-react";
import type { OfferCardData } from "./types";

export const OFFER_SERVICES_DATA: OfferCardData[] = [
  {
    id: 1,
    title: "Construction Management",
    description:
      "Plan, organize and monitor your construction projects from one intelligent platform.",
    icon: Building2,
    href: "/services/construction-management",
  },
  {
    id: 2,
    title: "Design & Planning",
    description:
      "Turn your construction idea into structured plans and actionable project information.",
    icon: Ruler,
    href: "/services/design-planning",
  },
  {
    id: 3,
    title: "Cost Estimation",
    description:
      "Generate reliable project estimates and understand your construction costs before you build.",
    icon: Calculator,
    href: "/services/cost-estimation",
  },
  {
    id: 4,
    title: "AI-Powered Intelligence",
    description:
      "Use artificial intelligence to analyze project information and support smarter construction decisions.",
    icon: BrainCircuit,
    href: "/services/ai-intelligence",
  },
  {
    id: 5,
    title: "Materials & Suppliers",
    description:
      "Find materials, compare supplier options and make better purchasing decisions for your project.",
    icon: PackageSearch,
    href: "/services/materials-suppliers",
  },
  {
    id: 6,
    title: "Stakeholder Collaboration",
    description:
      "Connect clients, architects, quantity surveyors, contractors and suppliers in one ecosystem.",
    icon: Users,
    href: "/services/stakeholder-collaboration",
  },
];

export const SERVICE_PAGES: Record<string, {
  title: string;
  eyebrow: string;
  summary: string;
  highlights: string[];
  outcomes: string[];
  cta: { label: string; href: string }[];
}> = {
  "construction-management": {
    title: "Construction Management",
    eyebrow: "Project Operations",
    summary:
      "Run each project from one workspace: scope, schedule, team, budget, and documents stay connected instead of scattered across spreadsheets and messages.",
    highlights: [
      "Keep plans, documents, and status updates attached to the same project record",
      "Invite the right team members and see who is responsible for what",
      "Track progress against milestones instead of relying on memory",
      "Surface blockers early so small delays do not become expensive ones",
    ],
    outcomes: [
      "Less time spent chasing information",
      "Clearer ownership across the team",
      "Faster decisions when plans change",
      "A shared source of truth for clients and contractors",
    ],
    cta: [
      { label: "Start a project", href: "/register" },
      { label: "See how it works", href: "/how-it-works" },
    ],
  },
  "design-planning": {
    title: "Design & Planning",
    eyebrow: "Plan Before You Build",
    summary:
      "Turn a rough idea into a structured plan: room layouts, finishes, scope, and requirements that the rest of the platform can build on.",
    highlights: [
      "Capture project goals, size, style, and constraints in one place",
      "Translate ideas into structured plans that estimators and teams can use",
      "Make changes once and propagate them across the project",
      "Keep design conversations tied to the actual scope, not separate notes",
    ],
    outcomes: [
      "Clearer briefs before work starts",
      "Fewer misunderstandings with contractors and suppliers",
      "Faster iteration on layout and scope",
      "A plan that cost and procurement can read directly",
    ],
    cta: [
      { label: "Start a project", href: "/register" },
      { label: "Explore solutions", href: "/solutions" },
    ],
  },
  "cost-estimation": {
    title: "Cost Estimation",
    eyebrow: "Know Your Numbers Early",
    summary:
      "Build reliable cost estimates before you commit: materials, labor, and contingencies broken down so you can see where the money is going.",
    highlights: [
      "Generate itemized estimates from the project scope and plan",
      "Compare estimated vs actual costs as the project moves forward",
      "Spot high-risk line items before they turn into overruns",
      "Edit any line directly instead of rerunning the whole estimate",
    ],
    outcomes: [
      "Better budgeting before construction starts",
      "Earlier visibility into likely cost drift",
      "More confident conversations with clients and suppliers",
      "Estimates that stay connected to the live project",
    ],
    cta: [
      { label: "Start a project", href: "/register" },
      { label: "Read estimate insights", href: "/blog" },
    ],
  },
  "ai-intelligence": {
    title: "AI-Powered Intelligence",
    eyebrow: "Smarter Input, Faster Decisions",
    summary:
      "Use AI to turn rough project descriptions into structured plans, cost signals, and decision support, so the routine work gets done fast and the hard calls stay human.",
    highlights: [
      "Turn plain-language project briefs into organized project information",
      "Highlight risks, gaps, and likely cost pressure points early",
      "Support design and estimating decisions without replacing professional judgment",
      "Keep recommendations tied to the project, not generic advice",
    ],
    outcomes: [
      "Faster first drafts of plans and estimates",
      "Less time spent starting from zero",
      "Better-structured information for every team member",
      "A clearer path from idea to action",
    ],
    cta: [
      { label: "Try the platform", href: "/register" },
      { label: "See AI in context", href: "/solutions" },
    ],
  },
  "materials-suppliers": {
    title: "Materials & Suppliers",
    eyebrow: "Source With Confidence",
    summary:
      "Find materials and compare supplier options in one place, so purchasing decisions are based on price, availability, and fit for the project rather than whoever replies first.",
    highlights: [
      "Compare material options against the project's actual requirements",
      "Review supplier options side by side instead of across separate messages",
      "Keep procurement tied to the project scope and estimate",
      "Flag when quotes or availability look off before you commit",
    ],
    outcomes: [
      "More informed purchasing decisions",
      "Less back-and-forth with suppliers",
      "Better alignment between materials and budget",
      "Faster sourcing when timelines matter",
    ],
    cta: [
      { label: "Start a project", href: "/register" },
      { label: "Explore sourcing", href: "/solutions" },
    ],
  },
  "stakeholder-collaboration": {
    title: "Stakeholder Collaboration",
    eyebrow: "Everyone on the Same Plan",
    summary:
      "Bring clients, architects, quantity surveyors, contractors, and suppliers into the same project context, so communication happens around the work instead of around email threads.",
    highlights: [
      "Give each stakeholder the access and view they actually need",
      "Keep discussions attached to the relevant project detail",
      "Reduce duplicated updates across separate inboxes",
      "Make handoffs between roles cleaner and more auditable",
    ],
    outcomes: [
      "Fewer lost messages and stale versions",
      "Clearer responsibilities across the team",
      "Faster responses to design and budget changes",
      "A project that is easier to manage as it scales",
    ],
    cta: [
      { label: "Start a project", href: "/register" },
      { label: "See who it's for", href: "/solutions" },
    ],
  },
};

export function serviceSlug(title: string): string {
  return title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

export function getServicePage(slug: string) {
  return SERVICE_PAGES[slug] ?? null;
}
