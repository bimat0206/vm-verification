
// This file is no longer used for the main app shell structure as TopNavBar
// is handled globally by src/app/layout.tsx.
// It can be safely removed or repurposed if a nested shell is needed later.
// For now, we'll keep it empty or with a placeholder to avoid breaking imports if any exist,
// though ideally it should be removed if truly unused.

// import React from 'react';

// export function AppShell({ children }: { children: React.ReactNode }) {
//   return (
//     <div className="min-h-screen flex flex-col">
//        {/* The TopNavBar is now in RootLayout */}
//       <main className="flex-1"> {/* Padding is handled in RootLayout's main tag */}
//         {children}
//       </main>
//     </div>
//   );
// }

// To prevent breaking existing imports immediately, let's just pass children through.
// This component should be removed once all references are updated.
import React from 'react';

export function AppShell({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}
