
// This component is no longer used as the navigation logic has been
// moved to src/components/layout/top-nav-bar.tsx.
// The navItems structure might be imported by top-nav-bar.tsx if needed.
// This file can be safely deleted or kept for reference.

// "use client";

// import Link from "next/link";
// import { usePathname } from "next/navigation";
// import { Home, CheckSquare, Search, UploadCloud, HeartPulse, Wrench, Bot } from "lucide-react";
// import { cn } from "@/lib/utils";
// import {
//   Accordion,
//   AccordionContent,
//   AccordionItem,
//   AccordionTrigger,
// } from "@/components/ui/accordion";

// interface NavItem {
//   title: string;
//   href: string;
//   icon: React.ElementType;
//   matchPaths?: string[];
//   subItems?: NavItem[];
// }

// const navItems: NavItem[] = [
//   { title: "Home", href: "/", icon: Home },
//   {
//     title: "Verification",
//     href: "/verification",
//     icon: CheckSquare,
//     matchPaths: ["/verification", "/verifications"],
//     subItems: [
//       { title: "New Verification", href: "/verification/new", icon: CheckSquare, matchPaths: ["/verification/new"] },
//       { title: "Results", href: "/verifications", icon: Search, matchPaths: ["/verifications"] },
//       { title: "Lookup", href: "/verification/lookup", icon: Search, matchPaths: ["/verification/lookup"] },
//     ],
//   },
//   {
//     title: "Tools",
//     href: "/tools",
//     icon: Wrench,
//     matchPaths: ["/tools"],
//     subItems: [
//       { title: "Image Upload", href: "/tools/image-upload", icon: UploadCloud, matchPaths: ["/tools/image-upload"] },
//       { title: "Health Check", href: "/tools/health-check", icon: HeartPulse, matchPaths: ["/tools/health-check"] },
//       { title: "Predictive Health", href: "/tools/predictive-health", icon: Bot, matchPaths: ["/tools/predictive-health"] },
//     ],
//   },
// ];

// export function SidebarNav() {
//   const pathname = usePathname();

//   const isActive = (item: NavItem) => {
//     if (item.matchPaths?.some(p => pathname.startsWith(p))) return true;
//     return pathname === item.href;
//   };
  
//   const getAccordionDefaultValues = () => {
//     const activeParent = navItems.find(item => item.subItems && item.subItems.some(subItem => isActive(subItem)));
//     return activeParent ? [activeParent.title] : [];
//   };

//   return (
//     <nav className="flex flex-col space-y-1 px-2">
//       <Accordion type="multiple" defaultValue={getAccordionDefaultValues()} className="w-full">
//         {navItems.map((item) =>
//           item.subItems ? (
//             <AccordionItem value={item.title} key={item.title} className="border-b-0">
//               <AccordionTrigger
//                 className={cn(
//                   "flex items-center justify-between rounded-md px-3 py-2 text-sm font-medium hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
//                   isActive(item) ? "bg-sidebar-primary text-sidebar-primary-foreground" : "text-sidebar-foreground",
//                   "data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
//                 )}
//               >
//                 <div className="flex items-center">
//                   <item.icon className="mr-2 h-5 w-5" />
//                   {item.title}
//                 </div>
//               </AccordionTrigger>
//               <AccordionContent className="pb-0 pl-4 pt-1">
//                 <div className="flex flex-col space-y-1">
//                 {item.subItems.map((subItem) => (
//                   <Link key={subItem.href} href={subItem.href} legacyBehavior passHref>
//                     <a
//                       className={cn(
//                         "flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
//                         isActive(subItem) ? "bg-sidebar-accent text-sidebar-accent-foreground" : "text-sidebar-foreground"
//                       )}
//                     >
//                       <subItem.icon className="mr-2 h-4 w-4" />
//                       {subItem.title}
//                     </a>
//                   </Link>
//                 ))}
//                 </div>
//               </AccordionContent>
//             </AccordionItem>
//           ) : (
//             <Link key={item.href} href={item.href} legacyBehavior passHref>
//               <a
//                 className={cn(
//                   "flex items-center rounded-md px-3 py-2 text-sm font-medium hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
//                   isActive(item) ? "bg-sidebar-primary text-sidebar-primary-foreground" : "text-sidebar-foreground"
//                 )}
//               >
//                 <item.icon className="mr-2 h-5 w-5" />
//                 {item.title}
//               </a>
//             </Link>
//           )
//         )}
//       </Accordion>
//     </nav>
//   );
// }

// Add an empty export to satisfy module requirements if this file is still imported elsewhere.
export {};
