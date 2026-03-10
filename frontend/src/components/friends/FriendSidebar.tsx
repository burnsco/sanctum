import { Settings, UserCheck, UserPlus, Users } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface FriendSidebarProps {
  activeView: "suggestions" | "requests" | "friends";
  onViewChange: (view: "suggestions" | "requests" | "friends") => void;
  requestCount?: number;
}

export function FriendSidebar({ activeView, onViewChange, requestCount = 0 }: FriendSidebarProps) {
  type NavItem = {
    id: "suggestions" | "requests" | "friends";
    label: string;
    icon: React.ElementType;
    badge?: number;
  };

  const navItems: NavItem[] = [
    {
      id: "suggestions",
      label: "Suggestions",
      icon: UserPlus,
    },
    {
      id: "requests",
      label: "Friend Requests",
      icon: UserCheck,
      badge: requestCount > 0 ? requestCount : undefined,
    },
    {
      id: "friends",
      label: "All Friends",
      icon: Users,
    },
  ];

  return (
    <div className="shrink-0 bg-background border-b md:border-b-0 md:border-r md:w-90 md:h-full md:flex md:flex-col">
      {/* Header: desktop only */}
      <div className="hidden md:flex p-4 pt-5 pb-2 items-center justify-between">
        <h2 className="text-2xl font-bold tracking-tight">Friends</h2>
        <Button
          variant="ghost"
          size="icon"
          className="rounded-full bg-muted/50 hover:bg-muted cursor-pointer"
          type="button"
        >
          <Settings className="w-5 h-5 text-foreground" />
        </Button>
      </div>

      {/* Nav: horizontal scroll on mobile, vertical list on desktop */}
      <nav className="flex md:block overflow-x-auto gap-1 px-2 py-2 md:space-y-0.5 md:px-2 md:pb-2 scrollbar-none">
        {navItems.map((item) => {
          const isActive = activeView === item.id;
          return (
            <button
              type="button"
              key={item.id}
              onClick={() => onViewChange(item.id)}
              className={cn(
                "relative shrink-0 flex items-center gap-2 rounded-lg transition-colors font-medium",
                // Mobile: compact pill
                "md:w-full md:gap-3 md:px-2 md:py-2",
                "px-3 py-2",
                isActive
                  ? "bg-muted/80 text-foreground"
                  : "hover:bg-muted/50 text-foreground/80 hover:text-foreground",
              )}
            >
              <div
                className={cn(
                  "flex items-center justify-center rounded-full shrink-0 transition-colors",
                  // Mobile: smaller icon circle
                  "w-7 h-7 md:w-9 md:h-9",
                  isActive ? "bg-primary text-primary-foreground" : "bg-muted text-foreground",
                )}
              >
                <item.icon className="w-4 h-4 md:w-5 md:h-5" />
              </div>
              <span className="text-[13px] md:text-[17px] whitespace-nowrap md:flex-1 md:text-left">
                {item.label}
              </span>
              {item.badge && (
                <div className="flex items-center justify-center min-w-5 h-5 px-1.5 text-xs font-bold text-white bg-red-500 rounded-full">
                  {item.badge}
                </div>
              )}
            </button>
          );
        })}
      </nav>
    </div>
  );
}
