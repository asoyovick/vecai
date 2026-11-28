import ProjectDetailClient from "./ProjectDetailClient";

export function generateStaticParams() {
  return [{ id: "placeholder" }];
}

export default function Page() {
  return <ProjectDetailClient />;
}
