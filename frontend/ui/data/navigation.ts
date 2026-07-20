import {
  Activity,
  Search,
  Boxes,
  FileClock,
  Settings,
  Users,
  ExternalLink,
} from "lucide-react";

export type NavigationItem = {
  id: string;
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
};

export type NavigationGroup = {
  id: string;
  title: string;
  items: NavigationItem[];
};

export const navigationGroups: NavigationGroup[] = [
  {
    id: "monitoring",
    title: "Monitoring",
    items: [
      {
        id: "logs-explorer",
        label: "Logs Explorer",
        href: "/logs-explorer",
        icon: Search,
      },
      {
        id: "live-tail",
        label: "Live Tail",
        href: "/logs/live",
        icon: Activity,
        badge: "Live",
      },
      {
        id: "audit-logs",
        label: "Audit Logs",
        href: "/audit-logs",
        icon: FileClock,
      },
    ],
  },
  {
    id: "management",
    title: "Management",
    items: [
      {
        id: "products",
        label: "Products",
        href: "/products",
        icon: Boxes,
      },
      {
        id: "users",
        label: "Users",
        href: "/users",
        icon: Users,
      },
      {
        id: "settings",
        label: "Settings",
        href: "/settings",
        icon: Settings,
      },
    ],
  },
  {
    id: "demo",
    title: "Demo",
    items: [
      {
        id: "demo-portal",
        label: "Demo Portal",
        href: "http://localhost:3001",
        icon: ExternalLink,
        badge: "Demo",
      },
    ],
  },
];