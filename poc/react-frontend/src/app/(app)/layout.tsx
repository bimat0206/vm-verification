
// This layout is now simplified as TopNavBar is handled by the root layout.
// It can be used for specific (app) group styling if needed in the future.
export default function AppPagesLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  // The AppShell is removed as TopNavBar is now global via RootLayout
  // The main content area will have padding from RootLayout's main tag
  return <>{children}</>;
}
