
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { BarChart3, UploadCloud, HeartPulse, Info, Target, CheckSquare as NewVerificationIcon } from "lucide-react"; // Replaced Search with NewVerificationIcon for clarity

export default function HomePage() {
  return (
    <div className="container mx-auto py-12 px-4 space-y-10">
      {/* About This System Section */}
      <Card className="shadow-xl border-border bg-card">
        <CardHeader className="p-6">
          <CardTitle className="text-2xl font-headline text-primary-foreground flex items-center">
            <Target className="w-7 h-7 mr-3 text-gradient-primary" /> {/* Using Target icon */}
            About This System
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 px-4 pt-0 pb-6 text-base text-secondary-foreground">
          <p>
            The Vending Machine Verification System is an advanced AI-powered platform designed to automate the verification and analysis of vending machine layouts and inventory. Using cutting-edge computer vision and machine learning technologies, our system provides accurate, reliable verification results to help maintain optimal vending machine operations.
          </p>
          <div>
            <h4 className="font-semibold text-primary-foreground mb-2">Key Benefits:</h4>
            <ul className="list-disc list-inside space-y-1 pl-4">
              <li>Automated Analysis: Reduce manual inspection time and human error</li>
              <li>Real-time Results: Get instant verification feedback with detailed analysis</li>
              <li>Comprehensive Reporting: Access detailed metrics and performance data</li>
              <li>Easy Integration: Simple interface for seamless workflow integration</li>
            </ul>
          </div>
        </CardContent>
      </Card>

      {/* Key Features Section */}
      <div>
        <h2 className="text-3xl font-bold mb-6 text-primary-foreground">Key Features</h2>
        <div className="grid md:grid-cols-2 lg:grid-cols-2 gap-6"> {/* Adjusted to 2 columns for lg screens */}
          <FeatureCard
            icon={<NewVerificationIcon className="w-6 h-6 text-primary-foreground" />}
            title="New Verification"
            description="Initiate layout or previous vs. current checks."
            link="/verification/new"
            linkText="Go to New Verification"
          />
          <FeatureCard
            icon={<BarChart3 className="w-6 h-6 text-primary-foreground" />}
            title="View Results"
            description="Browse and analyze past verifications."
            link="/verifications"
            linkText="Go to View Results"
          />
          <FeatureCard
            icon={<UploadCloud className="w-6 h-6 text-primary-foreground" />}
            title="Upload Images"
            description="Manage reference and checking images."
            link="/tools/image-upload"
            linkText="Go to Upload Images"
          />
          <FeatureCard
            icon={<HeartPulse className="w-6 h-6 text-primary-foreground" />}
            title="System Health"
            description="Check API and system status."
            link="/tools/health-check"
            linkText="Go to System Health"
          />
        </div>
      </div>

      {/* Getting Started Section */}
      <Card className="shadow-xl border-border bg-card">
        <CardHeader className="p-6">
          <CardTitle className="text-2xl font-headline text-primary-foreground">Getting Started</CardTitle>
        </CardHeader>
        <CardContent className="p-6 pt-0 text-base">
          <div className="space-y-6">
            <div>
              <h4 className="text-lg font-semibold text-primary-foreground mb-2">Step 1: Upload Images</h4>
              <p className="text-secondary-foreground mb-3">
                Before starting verification, ensure your images are properly uploaded using the <Link href="/tools/image-upload" className="text-gradient-primary hover:underline">Upload Images</Link> tool:
              </p>
              <ul className="list-disc list-inside space-y-1 pl-4 text-secondary-foreground">
                <li>Upload reference images to establish baseline layouts</li>
                <li>Upload checking images for comparison analysis</li>
                <li>Organize images in appropriate S3 bucket folders</li>
                <li>Verify image quality and format compatibility</li>
              </ul>
            </div>
            <div>
              <h4 className="text-lg font-semibold text-primary-foreground mb-2">Step 2: Start a Verification</h4>
              <p className="text-secondary-foreground mb-3">
                Navigate to the <Link href="/verification/new" className="text-gradient-primary hover:underline">Verification System</Link> to begin a new verification process:
              </p>
              <ul className="list-disc list-inside space-y-1 pl-4 text-secondary-foreground">
                <li>Select your verification type (Layout vs Checking or Previous vs Current)</li>
                <li>Choose reference and checking images from S3 buckets</li>
                <li>Submit for analysis</li>
              </ul>
            </div>
            <div>
              <h4 className="text-lg font-semibold text-primary-foreground mb-2">Step 3: Monitor Progress</h4>
              <p className="text-secondary-foreground mb-3">
                Track your verification progress in real-time through the <Link href="/verifications" className="text-gradient-primary hover:underline">Verification Results</Link> page:
              </p>
              <ul className="list-disc list-inside space-y-1 pl-4 text-secondary-foreground">
                <li>View processing status and estimated completion time</li>
                <li>Access preliminary results as they become available</li>
              </ul>
            </div>
            <div>
              <h4 className="text-lg font-semibold text-primary-foreground mb-2">Step 4: Review Results</h4>
              <p className="text-secondary-foreground mb-3">
                Analyze detailed verification results in the <Link href="/verifications" className="text-gradient-primary hover:underline">Verification Results section</Link>:
              </p>
              <ul className="list-disc list-inside space-y-1 pl-4 text-secondary-foreground">
                <li>Browse all completed verifications with advanced filtering</li>
                <li>View detailed report and verification summary</li>
                <li>Access comprehensive LLM analysis</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Frequently Asked Questions Section */}
      <Card className="shadow-xl border-border bg-card">
        <CardHeader className="p-6">
          <CardTitle className="text-2xl font-headline text-primary-foreground">Frequently Asked Questions</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6 p-6 pt-0 text-base">
          <div>
            <h4 className="text-lg font-semibold text-primary-foreground mb-2">Q: What image formats are supported?</h4>
            <p className="text-secondary-foreground">The system supports common image formats including JPEG, PNG, and WEBP. For best results, use high-resolution images with clear visibility of vending machine contents.</p>
          </div>
          <div>
            <h4 className="text-lg font-semibold text-primary-foreground mb-2">Q: How long does verification take?</h4>
            <p className="text-secondary-foreground">Verification processing typically takes 2-5 minutes depending on image complexity and system load. You can monitor progress in real-time on the Results page.</p>
          </div>
          <div>
            <h4 className="text-lg font-semibold text-primary-foreground mb-2">Q: What's the difference between verification types?</h4>
            <p className="text-secondary-foreground">
              <strong>Layout vs Checking:</strong> Compares current machine layout against a reference layout.<br/>
              <strong>Previous vs Current:</strong> Compares two time-based snapshots to detect changes over time.
            </p>
          </div>
          <div>
            <h4 className="text-lg font-semibold text-primary-foreground mb-2">Q: How do I upload images to S3?</h4>
            <p className="text-secondary-foreground">Use the <Link href="/tools/image-upload" className="text-gradient-primary hover:underline">Upload Images</Link> tool to upload files directly to your configured S3 buckets. The system will organize them automatically.</p>
          </div>
          <div>
            <h4 className="text-lg font-semibold text-primary-foreground mb-2">Q: What should I do if verification fails?</h4>
            <p className="text-secondary-foreground">Check the <Link href="/tools/health-check" className="text-gradient-primary hover:underline">System Health</Link> page for API status, ensure your images are properly uploaded, and verify network connectivity. Contact support for persistent issues.</p>
          </div>
          <div>
            <h4 className="text-lg font-semibold text-primary-foreground mb-2">Q: Can I delete or modify past verification results?</h4>
            <p className="text-secondary-foreground">Verification results are stored permanently for audit purposes. You can filter and search through results on the <Link href="/verifications" className="text-gradient-primary hover:underline">Results</Link> page, but cannot modify historical data.</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

interface FeatureCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  link: string;
  linkText: string;
}

function FeatureCard({ icon, title, description, link, linkText }: FeatureCardProps) {
  return (
    <Card className="bg-card/80 hover:shadow-2xl transition-all duration-300 ease-in-out border-border transform hover:-translate-y-1 flex flex-col">
      <CardHeader className="p-5">
        <div className="flex items-center space-x-3">
          {icon}
          <CardTitle className="font-headline text-xl text-primary-foreground">{title}</CardTitle>
        </div>
      </CardHeader>
      <CardContent className="p-5 pt-0 flex-grow">
        <p className="text-secondary-foreground text-sm mb-4">{description}</p>
      </CardContent>
      <CardFooter className="p-5 pt-0">
        <Button asChild className="btn-gradient text-sm w-full py-2.5">
          <Link href={link}>{linkText}</Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
