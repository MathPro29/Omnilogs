import { useState } from "react";
import { FloatingDockNavbar } from "./docking";
import type { NavigationItem } from "./data/navigation";

export function DemoApp() {
  const [activePath, setActivePath] = useState("/overview");

  const handleNavigate = (item: NavigationItem) => {
    setActivePath(item.href);
  };

  return (
    <main className="app-shell">
      <FloatingDockNavbar activePath={activePath} onNavigate={handleNavigate} />
      <section className="page-panel" aria-live="polite">
        <p>Current page</p>
        <h1>{activePath}</h1>
      </section>
    </main>
  );
}
