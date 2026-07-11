import { StrictMode, useState } from "react";
import { createRoot } from "react-dom/client";

import { FloatingDockNavbar } from "./docking";
import type { NavigationItem } from "./data/navigation";
import "./style.css";

function App() {
  const [activePath, setActivePath] = useState("/overview");

  function handleNavigate(item: NavigationItem) {
    setActivePath(item.href);
  }

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

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
