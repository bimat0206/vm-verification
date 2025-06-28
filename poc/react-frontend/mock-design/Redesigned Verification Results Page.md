import React, { useState, useEffect, useRef } from 'react';

// --- ICONS (Helper Components) ---
const FilterIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-2" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M3 3a1 1 0 011-1h12a1 1 0 011 1v3a1 1 0 01-.293.707L12 11.414V15a1 1 0 01-.293.707l-2 2A1 1 0 018 17v-5.586L3.293 6.707A1 1 0 013 6V3z" clipRule="evenodd" /></svg>;
const SearchIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" /></svg>;
const CloseIcon = () => <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" /></svg>;
const PanelLeftIcon = () => <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M9 3v18"/></svg>;
const PanelRightIcon = () => <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect width="18" height="18" x="3" y="3" rx="2"/><path d="M15 3v18"/></svg>;

// --- MOCK DATA ---
const mockVerifications = Array.from({ length: 57 }, (_, i) => {
    const statusMap = ['CORRECT', 'INCORRECT', 'PENDING'];
    const status = statusMap[i % 3];
    return { 
        id: `verif-20250607${String(1000 + i).padStart(4, '0')}`, 
        status: status, 
        machineId: `VM-${100 + (i % 10)}`, 
        date: new Date(Date.now() - i * 3600000).toISOString(), 
        accuracy: status === 'PENDING' ? null : Math.random(), 
        confidence: status === 'PENDING' ? null : Math.random(), 
        summary: 'Verification summary details go here...', 
        details: 'Detailed LLM analysis text goes here. This can be a very long text to demonstrate the scrolling capability of the modal window. It will wrap and scroll as needed to ensure the UI remains clean and functional even with extensive data output.',
        referenceImage: `https://placehold.co/800x1200/111111/444444?text=Ref-${i}`,
        checkingImage: `https://placehold.co/800x1200/111111/555555?text=Check-${i}`,
    };
});


// --- Image Viewer with Loading State ---
const ImageContainer = ({ src, alt, className }) => {
    const [isLoading, setIsLoading] = useState(true);
    return (
        <div className={`relative w-full h-full bg-[#111111] rounded-lg ${className}`}>
            {isLoading && (
                <div className="absolute inset-0 flex items-center justify-center">
                    <div className="w-8 h-8 border-4 border-dashed rounded-full animate-spin border-purple-500"></div>
                </div>
            )}
            <img 
                src={src} 
                alt={alt} 
                onLoad={() => setIsLoading(false)} 
                className={`w-full h-full object-contain rounded-lg transition-opacity duration-300 ${isLoading ? 'opacity-0' : 'opacity-100'}`}
            />
        </div>
    );
};

// --- Full Screen Zoom & Pan Viewer ---
const FullScreenViewer = ({ referenceSrc, checkingSrc, onClose }) => {
    const [transform, setTransform] = useState({ scale: 1, x: 0, y: 0 });
    const viewerRef = useRef(null);
    const isPanning = useRef(false);
    const startPos = useRef({ x: 0, y: 0 });

    const handleWheel = (e) => {
        e.preventDefault();
        const scaleFactor = 0.1;
        const newScale = transform.scale - e.deltaY * scaleFactor;
        setTransform(prev => ({...prev, scale: Math.max(0.5, Math.min(newScale, 5)) }));
    };

    const handleMouseDown = (e) => {
        e.preventDefault();
        isPanning.current = true;
        startPos.current = { x: e.clientX - transform.x, y: e.clientY - transform.y };
        viewerRef.current.style.cursor = 'grabbing';
    };

    const handleMouseMove = (e) => {
        if (!isPanning.current) return;
        e.preventDefault();
        setTransform(prev => ({
            ...prev,
            x: e.clientX - startPos.current.x,
            y: e.clientY - startPos.current.y
        }));
    };

    const handleMouseUp = () => {
        isPanning.current = false;
        viewerRef.current.style.cursor = 'grab';
    };

    useEffect(() => {
        const viewer = viewerRef.current;
        if (viewer) {
            viewer.addEventListener('wheel', handleWheel, { passive: false });
            return () => viewer.removeEventListener('wheel', handleWheel);
        }
    }, [transform.scale]);

    return (
        <div className="fixed inset-0 bg-black/90 backdrop-blur-sm z-50 flex flex-col p-4" ref={viewerRef} onMouseMove={handleMouseMove} onMouseUp={handleMouseUp} onMouseLeave={handleMouseUp}>
            <div className="flex justify-between items-center mb-4 flex-shrink-0">
                <p className="text-[#A0A0A0] text-sm">Use mouse wheel to zoom, click and drag to pan. Press ESC to close.</p>
                <button onClick={onClose} className="p-2 rounded-full bg-white/10 hover:bg-white/20 text-white"><CloseIcon /></button>
            </div>
            <div className="flex-grow flex gap-4 overflow-hidden">
                <div 
                    className="flex gap-4 w-full h-full transition-transform duration-100 ease-out"
                    style={{ transform: `translate(${transform.x}px, ${transform.y}px) scale(${transform.scale})`}}
                    onMouseDown={handleMouseDown}
                >
                    <div className="w-1/2 h-full"><ImageContainer src={referenceSrc} alt="Reference Full Screen" /></div>
                    <div className="w-1/2 h-full"><ImageContainer src={checkingSrc} alt="Checking Full Screen" /></div>
                </div>
            </div>
        </div>
    );
};


// --- MAIN APP COMPONENT ---
export default function VerificationResultsPage() {
    const [results, setResults] = useState([]);
    const [loading, setLoading] = useState(true);
    const [selectedVerification, setSelectedVerification] = useState(null);
    const [isPanelVisible, setPanelVisible] = useState(true);
    const [resultsPerPage, setResultsPerPage] = useState(5);
    const [currentPage, setCurrentPage] = useState(1);
    const [isFullScreen, setFullScreen] = useState(false);

    useEffect(() => {
        setLoading(true);
        setTimeout(() => {
            setResults(mockVerifications);
            setLoading(false);
        }, 1000);
    }, []);
    
     useEffect(() => {
        const handleKeyDown = (e) => {
            if (e.key === 'Escape') {
                setFullScreen(false);
                setSelectedVerification(null);
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, []);

    const StatusPill = ({ status }) => {
        const baseClasses = "px-3 py-1 text-xs font-bold rounded-full text-white";
        const statusConfig = {
            CORRECT: { text: "Correct", bg: "bg-green-500" },
            INCORRECT: { text: "Incorrect", bg: "bg-red-500" },
            PENDING: { text: "Pending", bg: "bg-yellow-500" },
        };
        const config = statusConfig[status] || { text: "Unknown", bg: "bg-gray-500" };
        return <span className={`${baseClasses} ${config.bg}`}>{config.text}</span>;
    };

    const GradientText = ({ children, className }) => (
        <span className={`bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 bg-clip-text text-transparent ${className}`}>
            {children}
        </span>
    );
    
    const FilterPanel = ({ setResultsPerPage, setCurrentPage }) => {
        return (
            <div className="w-full lg:w-80 lg:flex-shrink-0 transition-all duration-300">
                <div className="sticky top-8 bg-[#1E1E1E]/60 backdrop-blur-lg rounded-2xl border border-[#2F2F2F] p-6 space-y-6">
                    <div>
                        <h2 className="text-lg font-bold text-white mb-4">Quick Lookup</h2>
                         <div className="relative">
                            <input type="text" placeholder="Enter Verification ID..." className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg py-2 pl-4 pr-10 text-white focus:ring-2 focus:ring-purple-500 focus:outline-none"/>
                            <div className="absolute inset-y-0 right-0 flex items-center pr-3 pointer-events-none">
                                <SearchIcon />
                            </div>
                        </div>
                    </div>
                    <div>
                        <h2 className="flex items-center text-lg font-bold text-white mb-4"><FilterIcon /> Filters & Sorting</h2>
                        <div className="space-y-4 text-sm">
                            <div>
                                <label className="block mb-1 text-[#A0A0A0]">Status</label>
                                <select className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none"><option>All</option><option>Correct</option><option>Incorrect</option></select>
                            </div>
                            <div>
                                <label className="block mb-1 text-[#A0A0A0]">Machine ID</label>
                                <input type="text" placeholder="e.g., VM-102A" className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none" />
                            </div>
                            <div>
                                <label className="block mb-1 text-[#A0A0A0]">Date Range</label>
                                <input type="date" className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none" />
                            </div>
                            <div>
                                <label className="block mb-1 text-[#A0A0A0]">Sort By</label>
                                <select className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none"><option>Newest First</option><option>Oldest First</option></select>
                            </div>
                            <div>
                                <label className="block mb-1 text-[#A0A0A0]">Results per page</label>
                                <select onChange={(e) => { setResultsPerPage(Number(e.target.value)); setCurrentPage(1); }} defaultValue={5} className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none">
                                    <option>5</option>
                                    <option>10</option>
                                    <option>15</option>
                                    <option>20</option>
                                </select>
                            </div>
                            <button className="w-full mt-6 text-white font-bold py-3 px-8 rounded-lg shadow-md transition-all duration-300 transform hover:scale-105 bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500">
                                Apply Filters & Search
                            </button>
                             <button className="w-full mt-2 text-sm text-[#A0A0A0] hover:text-white">
                                Reset All Filters
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        );
    }
    
    const ResultsList = ({ results, resultsPerPage, currentPage, setCurrentPage }) => {
        const totalResults = results.length;
        const totalPages = Math.ceil(totalResults / resultsPerPage);
        const startIndex = (currentPage - 1) * resultsPerPage;
        const endIndex = startIndex + resultsPerPage;
        const paginatedResults = results.slice(startIndex, endIndex);

        return (
            <div className="bg-[#1E1E1E] rounded-2xl border border-[#2F2F2F] p-2">
                <div className="flex items-center justify-between p-4 text-sm text-[#A0A0A0] border-b border-[#2F2F2F]">
                    <span>Showing {startIndex + 1}-{Math.min(endIndex, totalResults)} of {totalResults} results</span>
                    <div className="flex space-x-2">
                        <button onClick={() => setCurrentPage(p => Math.max(1, p - 1))} disabled={currentPage === 1} className="px-3 py-1 border border-[#2F2F2F] rounded-md hover:bg-[#111111] disabled:opacity-50">Previous</button>
                        <button onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))} disabled={currentPage === totalPages} className="px-3 py-1 border border-[#2F2F2F] rounded-md hover:bg-[#111111] disabled:opacity-50">Next</button>
                    </div>
                </div>
                <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-[#2F2F2F]">
                        <thead className="bg-[#111111]">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-[#A0A0A0] uppercase tracking-wider">Verification ID</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-[#A0A0A0] uppercase tracking-wider">Status</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-[#A0A0A0] uppercase tracking-wider">Machine</th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-[#A0A0A0] uppercase tracking-wider">Date</th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-[#A0A0A0] uppercase tracking-wider">Action</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-[#2F2F2F]">
                            {paginatedResults.map((r) => (
                                <tr key={r.id} className="hover:bg-[#111111] transition-colors">
                                    <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-blue-400">{r.id}</td>
                                    <td className="px-6 py-4 whitespace-nowrap"><StatusPill status={r.status} /></td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-[#F5F5F5]">{r.machineId}</td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-[#A0A0A0]">{new Date(r.date).toLocaleString()}</td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <button onClick={() => setSelectedVerification(r)} className="text-purple-400 hover:text-pink-500">View Details</button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        );
    }
    
    const VerificationModal = ({ verification, onClose }) => {
        if (!verification) return null;
        
        return (
            <div className="fixed inset-0 bg-black/80 backdrop-blur-sm z-40 flex items-center justify-center p-4" onClick={onClose}>
                <div className="bg-[#1E1E1E] w-full max-w-6xl max-h-[90vh] rounded-2xl border border-[#2F2F2F] shadow-2xl flex flex-col" onClick={e => e.stopPropagation()}>
                    <div className="flex items-center justify-between p-4 border-b border-[#2F2F2F] flex-shrink-0">
                        <h2 className="text-xl font-bold"><GradientText>Verification Details</GradientText></h2>
                        <button onClick={onClose} className="text-[#A0A0A0] hover:text-white"><CloseIcon /></button>
                    </div>
                    <div className="p-6 overflow-y-auto">
                        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
                            <div>
                                <h3 className="font-bold text-white mb-3">Summary</h3>
                                <div className="space-y-2 text-sm bg-[#111111] p-4 rounded-lg border border-[#2F2F2F]">
                                    <p><strong className="text-[#A0A0A0] w-28 inline-block">ID:</strong> <span className="font-mono text-blue-400">{verification.id}</span></p>
                                    <p><strong className="text-[#A0A0A0] w-28 inline-block">Status:</strong> <StatusPill status={verification.status} /></p>
                                    <p><strong className="text-[#A0A0A0] w-28 inline-block">Accuracy:</strong> {verification.accuracy !== null ? `${(verification.accuracy * 100).toFixed(0)}%` : 'N/A'}</p>
                                    <p><strong className="text-[#A0A0A0] w-28 inline-block">Confidence:</strong> {verification.confidence !== null ? `${(verification.confidence * 100).toFixed(0)}%` : 'N/A'}</p>
                                    <p><strong className="text-[#A0A0A0] w-28 inline-block">Summary:</strong> {verification.summary}</p>
                                </div>
                                <h3 className="font-bold text-white mb-3 mt-6">LLM Analysis</h3>
                                <div className="text-sm bg-[#111111] p-4 rounded-lg border border-[#2F2F2F] font-mono text-green-400 max-h-64 overflow-y-auto whitespace-pre-wrap">{verification.details}</div>
                            </div>
                            <div>
                                 <h3 className="font-bold text-white mb-3">Image Comparison</h3>
                                 <div className="grid grid-cols-2 gap-4">
                                     <div>
                                         <p className="text-center text-[#A0A0A0] text-sm mb-2">Reference Image</p>
                                         <div className="aspect-[3/4]"><ImageContainer src={verification.referenceImage} alt="Reference Layout" /></div>
                                     </div>
                                      <div>
                                         <p className="text-center text-[#A0A0A0] text-sm mb-2">Checking Image</p>
                                         <div className="aspect-[3/4]"><ImageContainer src={verification.checkingImage} alt="Checking Image" /></div>
                                     </div>
                                 </div>
                                 <button onClick={() => setFullScreen(true)} className="w-full mt-4 text-white font-bold py-2 px-4 rounded-lg bg-white/10 hover:bg-white/20">View Full Screen</button>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        );
    }

    return (
        <div className="bg-[#111111] text-[#F5F5F5] min-h-screen font-sans">
            <main className="container mx-auto px-4 sm:px-6 lg:px-8 py-8">
                <h1 className="text-center text-4xl font-bold mb-4">
                    <GradientText>Verification Results</GradientText>
                </h1>
                <p className="text-center text-[#A0A0A0] mb-10">Review, filter, and analyze all verification outcomes.</p>

                <div className="flex flex-row gap-8 items-start">
                    {isPanelVisible && <FilterPanel setResultsPerPage={setResultsPerPage} setCurrentPage={setCurrentPage} />}
                    <div className={`transition-all duration-300 ${isPanelVisible ? 'flex-1' : 'w-full'}`}>
                        <div className="mb-4">
                           <button onClick={() => setPanelVisible(!isPanelVisible)} className="p-2 rounded-md bg-[#1E1E1E] text-[#A0A0A0] hover:bg-white/10" title={isPanelVisible ? 'Collapse Panel' : 'Expand Panel'}>
                                {isPanelVisible ? <PanelLeftIcon/> : <PanelRightIcon/>}
                            </button>
                        </div>
                         {loading ? <div className="flex-1 text-center py-20">Loading results...</div> : <ResultsList results={results} resultsPerPage={resultsPerPage} currentPage={currentPage} setCurrentPage={setCurrentPage} />}
                    </div>
                </div>
            </main>
            {selectedVerification && <VerificationModal verification={selectedVerification} onClose={() => setSelectedVerification(null)} />}
            {isFullScreen && selectedVerification && <FullScreenViewer referenceSrc={selectedVerification.referenceImage} checkingSrc={selectedVerification.checkingImage} onClose={() => setFullScreen(false)} />}
        </div>
    );
}