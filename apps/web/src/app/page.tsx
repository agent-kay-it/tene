import { Hero } from "@/components/hero";
import { Terminal } from "@/components/terminal";
import { Features } from "@/components/features";
import { Comparison } from "@/components/comparison";
import { HowItWorks } from "@/components/how-it-works";
import { Security } from "@/components/security";
import { CTA } from "@/components/cta";
import { Footer } from "@/components/footer";
import { Nav } from "@/components/nav";

export default function Home() {
  return (
    <>
      <Nav />
      <main>
        <Hero />
        <Terminal />
        <Features />
        <HowItWorks />
        <Security />
        <Comparison />
        <CTA />
      </main>
      <Footer />
    </>
  );
}
