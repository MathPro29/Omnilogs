import {
  Home,
  Search,
  Boxes,
  FileClock,
  Users,
  DatabasePlus,
  Clock,
} from "lucide-react";

export type NavigationItem = {
  id: string;
  label: string;
  href: string;
  icon: React.ComponentType<{ className?: string }>;
  badge?: string;
  /** Platform roles that must not see this item. */
  hiddenForRoles?: string[];
};

export type NavigationGroup = {
  id: string;
  title: string;
  items: NavigationItem[];
};

export const navigationGroups: NavigationGroup[] = [
  {
    id: 'overview',
    title: 'Overview',
    items: [{ id: 'dashboard', label: 'Dashboard', href: '/dashboard', icon: Home }],
  },
  {
    id: "monitoring",
    title: "Monitoring",
    items: [
      {
        id: "explore-logs",
        label: "Logs Explorer",
        href: "/logs-explorer",
        icon: Search,
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
        id: "retention-policies",
        label: "Retention Policy",
        href: "/retention-policies",
        icon: Clock,
      },
      {
        id: "products",
        label: "Products",
        href: "/products",
        icon: Boxes,
        hiddenForRoles: ["user"],
      },
      {
        id: "users",
        label: "Users",
        href: "/users",
        icon: Users,
        hiddenForRoles: ["user"],
      }
    ],
  },
  {
    id: "connections",
    title: "Connections",
    items: [
      {
        id: "connections",
        label: "Connections",
        href: "/connections",
        icon: DatabasePlus,
        hiddenForRoles: ["user"],
      },
    ],
  },
];
export function isNavigationPathActive(pathname: string, href: string) {
  return href === '/'
    ? pathname === href
    : pathname === href || pathname.startsWith(`${href}/`);
}