import type { Metadata } from "next";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, ArrowRight } from "lucide-react";
import { Navbar } from "@/components/layout/Navbar";
import { Footer } from "@/components/layout/Footer";
import {
  OFFER_SERVICES_DATA,
  getServicePage,
  serviceSlug,
} from "@/components/home/what-we-offer/serviceData";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/Card";
import { ButtonLink } from "@/components/ui/Button";
import { Badge } from "@/components/ui/Badge";

export function generateStaticParams() {
  return OFFER_SERVICES_DATA.map((service) => ({
    slug: serviceSlug(service.title),
  }));
}

export function generateMetadata({
  params,
}: {
  params: { slug: string };
}): Metadata {
  const page = getServicePage(params.slug);
  if (!page) {
    return {};
  }

  return {
    title: `${page.title} — VECAI`,
    description: page.summary,
  };
}

export default function ServicePage({
  params,
}: {
  params: { slug: string };
}) {
  const service = OFFER_SERVICES_DATA.find(
    (item) => serviceSlug(item.title) === params.slug,
  );
  const page = getServicePage(params.slug);

  if (!service || !page) {
    notFound();
  }

  return (
    <>
      <Navbar />

      <main className="bg-slate-50">
        {/* Hero */}
        <section className="relative overflow-hidden bg-[#000080] text-white">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20 lg:py-28">
            <Link
              href="/services"
              className="inline-flex items-center gap-2 text-sm font-semibold text-[#F28500] hover:text-white transition-colors"
            >
              <ArrowLeft className="h-4 w-4" />
              All services
            </Link>

            <div className="mt-6 max-w-3xl">
              <Badge tone="accent" className="border-0 bg-[#F28500]/20 text-[#F28500]">
                {page.eyebrow}
              </Badge>
              <h1 className="mt-4 text-3xl sm:text-4xl lg:text-5xl font-extrabold tracking-tight">
                {service.title}
              </h1>
              <p className="mt-5 text-base sm:text-lg text-slate-200/90 leading-relaxed">
                {page.summary}
              </p>
            </div>
          </div>
        </section>

        {/* Body */}
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16 lg:py-20">
          <div className="grid gap-12 lg:grid-cols-12 lg:items-start">
            {/* Main content */}
            <div className="lg:col-span-7">
              <Card className="border-0 shadow-sm">
                <CardHeader className="border-b bg-white">
                  <CardTitle className="text-2xl font-bold text-[#000080]">
                    What this service does
                  </CardTitle>
                </CardHeader>
                <CardContent className="bg-white">
                  <p className="text-slate-600 leading-relaxed">
                    {service.description}
                  </p>
                </CardContent>
              </Card>

              <Card className="border-0 shadow-sm mt-8">
                <CardHeader className="border-b bg-white">
                  <CardTitle className="text-xl font-bold text-[#000080]">
                    Key benefits
                  </CardTitle>
                </CardHeader>
                <CardContent className="bg-white">
                  <ul className="flex flex-col gap-4">
                    {page.highlights.map((item) => (
                      <li
                        key={item}
                        className="flex gap-3 text-slate-700"
                      >
                        <span className="mt-1 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-[#F28500]/10 text-[#F28500] text-xs font-bold">
                          +
                        </span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>

              <Card className="border-0 shadow-sm mt-8">
                <CardHeader className="border-b bg-white">
                  <CardTitle className="text-xl font-bold text-[#000080]">
                    Why teams use it
                  </CardTitle>
                </CardHeader>
                <CardContent className="bg-white">
                  <ul className="flex flex-col gap-4">
                    {page.outcomes.map((item) => (
                      <li key={item} className="flex gap-3 text-slate-700">
                        <span className="mt-1 flex h-5 w-5 shrink-0 items-center justify-center rounded-full bg-[#000080]/10 text-[#000080] text-xs font-bold">
                          ✓
                        </span>
                        {item}
                      </li>
                    ))}
                  </ul>
                </CardContent>
              </Card>
            </div>

            {/* Sidebar */}
            <div className="lg:col-span-5">
              <Card className="border-0 shadow-sm h-fit">
                <CardHeader className="bg-white border-b">
                  <CardTitle className="text-xl font-bold text-[#000080]">
                    Next steps
                  </CardTitle>
                </CardHeader>
                <CardContent className="bg-white">
                  <p className="text-sm text-slate-600 mb-5">
                    Explore this capability and then move to the next step that
                    fits your project.
                  </p>
                  <div className="flex flex-col gap-3">
                    {page.cta.map((action) => (
                      <ButtonLink
                        key={action.href}
                        href={action.href}
                        variant={action.href === "/register" ? "primary" : "outline"}
                        className="w-full justify-start"
                        icon={<ArrowRight className="h-4 w-4" />}
                      >
                        {action.label}
                      </ButtonLink>
                    ))}
                  </div>

                  <Link
                    href="/services"
                    className="mt-6 block text-center rounded-xl border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-50"
                  >
                    Back to all services
                  </Link>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </main>

      <Footer />
    </>
  );
}
