import { Navbar, Footer } from "@/components/layout";
import { HeroSection } from "@/components/home/herosection";
import { WhoWeAre } from "@/components/home/who-we-are";
import { WhatWeOffer } from "@/components/home/what-we-offer";
import { HowItWorks } from "@/components/home/how-it-works";

export default function HomePage() {
  return (
    <>
      <Navbar />

      <main className="flex-grow">
        <HeroSection />
        <WhoWeAre />
        <WhatWeOffer />
        <HowItWorks />
      </main>

      <Footer />
    </>
  );
}