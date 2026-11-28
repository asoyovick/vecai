import type { Metadata } from "next";
import { Navbar } from "@/components/layout/Navbar";
import { Footer } from "@/components/layout/Footer";
import { WhatWeOffer } from "@/components/home/what-we-offer";
import { CTA } from "@/components/home/CTA";

export const metadata: Metadata = {
  title: "What We Offer — VECAI",
  description:
    "Construction management, design and planning, cost estimation, AI-powered intelligence, materials and suppliers, and stakeholder collaboration — one connected platform for every stage of a build.",
};

export default function WhatWeOfferPage() {
  return (
    <>
      <Navbar />

      <main>
        <WhatWeOffer />
        <CTA />
      </main>

      <Footer />
    </>
  );
}
