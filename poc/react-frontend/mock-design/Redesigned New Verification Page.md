import React, { useState, useEffect } from 'react';

// --- MOCK DATA & ICONS ---
// In a real app, this would come from an API
const s3MockData = {
    '': { type: 'folder', items: ['processed/', 'raw/', 'system-prompt.json'] },
    'processed/': { type: 'folder', items: ['AACZ 1.png', 'AACZ 2.png', 'AACZ layout chuan.png'] },
    'raw/': { type: 'folder', items: ['AANP 1.png', 'AANP 2.png'] },
    'system-prompt.json': { type: 'file', content: { "prompt": "Analyze the provided image for discrepancies against the reference layout.", "model": "vision-pro-4.0", "parameters": { "temperature": 0.5, "max_tokens": 512 } } }
};
const FolderIcon = () => <svg className="w-5 h-5 mr-3 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>;
const FileIcon = () => <svg className="w-5 h-5 mr-3 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20"><path d="M9 2a2 2 0 00-2 2v8a2 2 0 002 2h2a2 2 0 002-2V4a2 2 0 00-2-2H9z"></path><path d="M4 3a2 2 0 100 4h12a2 2 0 100-4H4zM4 7a2 2 0 100 4h12a2 2 0 100-4H4zM4 11a2 2 0 100 4h12a2 2 0 100-4H4z"></path></svg>;
const UpIcon = () => <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 10l7-7m0 0l7 7m-7-7v18"></path></svg>;
const RootIcon = () => <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3"></path></svg>;

// --- S3 Image Selector Component ---
const S3ImageSelector = ({ title, selectedImageUrl, onSelectImage }) => {
    const [path, setPath] = useState('');
    const [items, setItems] = useState([]);
    const [previewItem, setPreviewItem] = useState(null);

    useEffect(() => {
        const data = s3MockData[path] || { items: [] };
        setItems(data.items);
        setPreviewItem(null); // Clear preview when path changes
    }, [path]);

    const handleItemClick = (item) => {
        const fullPath = path + item;
        const isFolder = item.endsWith('/');
        if (isFolder) {
            setPath(fullPath);
        } else {
            setPreviewItem({ name: item, fullPath: fullPath, type: item.endsWith('.json') ? 'json' : 'image' });
        }
    };

    const goUp = () => {
        if (path === '') return;
        const parts = path.split('/').filter(p => p);
        parts.pop();
        const newPath = parts.length > 0 ? parts.join('/') + '/' : '';
        setPath(newPath);
    };

    return (
        <div className="bg-[#111111] p-6 rounded-2xl border border-[#2F2F2F] h-[70vh] flex flex-col">
            <h3 className="font-bold text-lg text-white mb-4">{title}</h3>
            {/* Header */}
            <div className="flex items-center gap-2 mb-4 p-2 bg-[#1E1E1E] rounded-lg border border-[#2F2F2F] flex-shrink-0">
                <button onClick={goUp} className="p-2 hover:bg-white/10 rounded-md"><UpIcon /></button>
                <button onClick={() => setPath('')} className="p-2 hover:bg-white/10 rounded-md"><RootIcon /></button>
                <span className="px-3 py-2 bg-[#111111] rounded-md text-sm font-mono text-blue-400 flex-grow">s3://{path}</span>
            </div>

            <div className="flex-grow flex gap-6 overflow-hidden">
                {/* Item List (Left) */}
                <div className="w-1/2 flex-shrink-0 overflow-y-auto pr-2 border-r border-[#2F2F2F]">
                    {items.map(item => {
                        const isFolder = item.endsWith('/');
                        const isPreviewed = previewItem && previewItem.name === item;
                        return (
                            <div
                                key={item}
                                onClick={() => handleItemClick(item)}
                                className={`flex items-center p-2 rounded-md cursor-pointer transition-colors duration-200 ${isPreviewed ? 'bg-purple-600/30' : 'hover:bg-white/10'}`}
                            >
                                {isFolder ? <FolderIcon /> : <FileIcon />}
                                <span className="text-sm font-mono truncate">{item}</span>
                            </div>
                        );
                    })}
                </div>

                {/* Preview Pane (Right) */}
                <div className="w-1/2 flex flex-col">
                    <p className="text-sm text-[#A0A0A0] mb-2 flex-shrink-0">Preview</p>
                    <div className="flex-grow bg-[#1E1E1E] rounded-lg p-4 flex items-center justify-center border border-[#2F2F2F]">
                        {!previewItem ? (
                            <p className="text-[#A0A0A0]">Select a file to preview</p>
                        ) : previewItem.type === 'image' ? (
                            <img src={`https://placehold.co/400x300/1E1E1E/A0A0A0?text=${previewItem.name}`} alt="preview" className="max-w-full max-h-full object-contain rounded-md" />
                        ) : (
                            <pre className="text-xs text-green-400 bg-[#111111] p-3 rounded-md w-full h-full overflow-auto">
                                {JSON.stringify(s3MockData[previewItem.fullPath]?.content || { error: 'No content found' }, null, 2)}
                            </pre>
                        )}
                    </div>
                    {previewItem && previewItem.type === 'image' && (
                        <button onClick={() => onSelectImage(`s3://${previewItem.fullPath}`)} className="w-full mt-4 flex-shrink-0 text-white font-bold py-2 px-4 rounded-lg bg-gradient-to-r from-green-500 to-teal-500 hover:opacity-90">
                            ✅ Select this Image
                        </button>
                    )}
                </div>
            </div>

            {/* Selected Item Footer */}
            <div className="flex-shrink-0 mt-4 pt-4 border-t border-[#2F2F2F]">
                <p className="text-sm text-[#A0A0A0]">Final Selection:</p>
                {selectedImageUrl ? (
                    <div className="flex items-center gap-2 mt-2">
                        <img src={`https://placehold.co/40x40/EC4899/FFFFFF?text=✓`} alt="selected thumbnail" className="w-10 h-10 rounded-md" />
                        <span className="font-mono text-sm text-green-400">{selectedImageUrl}</span>
                    </div>
                ) : (
                    <p className="text-sm text-gray-500 mt-1">No image selected</p>
                )}
            </div>
        </div>
    );
};

// --- Wizard Component ---
export default function NewVerificationPage() {
    const [step, setStep] = useState(1);
    const [verificationType, setVerificationType] = useState('LAYOUT_VS_CHECKING');
    const [referenceImage, setReferenceImage] = useState(null);
    const [checkingImage, setCheckingImage] = useState(null);
    const [verificationResult, setVerificationResult] = useState(null);

    const handleSubmit = () => {
        setStep(5);
        setTimeout(() => {
            setVerificationResult({ id: `verif-${Date.now()}`, status: 'PROCESSING', llmAnalysis: 'Analysis is in progress...' });
        }, 1000);
        setTimeout(() => {
            setVerificationResult({ id: `verif-${Date.now()}`, status: 'COMPLETE', accuracy: 0.95, confidence: 0.99, llmAnalysis: 'LLM Analysis Complete: 1 discrepancy found in slot B4. Product is tilted.' });
        }, 5000);
    };
    
    const resetWizard = () => {
        setStep(1);
        setReferenceImage(null);
        setCheckingImage(null);
        setVerificationResult(null);
    }

    const Step = ({ number, title, active }) => (
        <div className="flex items-center">
            <div className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm transition-all duration-300 ${active ? 'bg-gradient-to-r from-blue-500 to-pink-500 text-white' : 'bg-[#2F2F2F] text-[#A0A0A0]'}`}>
                {number}
            </div>
            <p className={`ml-3 font-medium ${active ? 'text-white' : 'text-[#A0A0A0]'}`}>{title}</p>
        </div>
    );

    return (
        <div className="bg-[#111111] text-[#F5F5F5] min-h-screen font-sans p-8 flex items-center justify-center">
            <div className="w-full max-w-5xl">
                <h1 className="text-center text-4xl font-bold mb-4">
                    <span className="bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 bg-clip-text text-transparent">
                        New Verification
                    </span>
                </h1>
                <p className="text-center text-[#A0A0A0] mb-10">Follow the steps below to initiate a new visual verification.</p>

                {/* Step Indicator */}
                <div className="flex justify-between items-center mb-10 px-4">
                    <Step number={1} title="Type" active={step >= 1} />
                    <div className="flex-1 h-0.5 bg-[#2F2F2F] mx-4"></div>
                    <Step number={2} title="Reference" active={step >= 2} />
                    <div className="flex-1 h-0.5 bg-[#2F2F2F] mx-4"></div>
                    <Step number={3} title="Checking" active={step >= 3} />
                     <div className="flex-1 h-0.5 bg-[#2F2F2F] mx-4"></div>
                    <Step number={4} title="Review" active={step >= 4} />
                     <div className="flex-1 h-0.5 bg-[#2F2F2F] mx-4"></div>
                    <Step number={5} title="Result" active={step >= 5} />
                </div>
                
                {/* Wizard Content */}
                <div className="bg-[#1E1E1E] p-8 rounded-2xl border border-[#2F2F2F]">
                    {step === 1 && (
                        <div>
                            <h2 className="text-xl font-bold mb-4">Step 1: Choose Verification Type</h2>
                            <select value={verificationType} onChange={(e) => setVerificationType(e.target.value)} className="w-full bg-[#111111] border border-[#2F2F2F] rounded-lg p-3 focus:ring-2 focus:ring-purple-500 focus:outline-none">
                                <option>LAYOUT_VS_CHECKING</option>
                                <option>PREVIOUS_VS_CURRENT</option>
                            </select>
                            <button onClick={() => setStep(2)} className="w-full mt-6 text-white font-bold py-3 px-8 rounded-lg bg-gradient-to-r from-blue-500 to-pink-500 hover:opacity-90">Next Step</button>
                        </div>
                    )}
                    {step === 2 && <S3ImageSelector title="Step 2: Select Reference Image" selectedImageUrl={referenceImage} onSelectImage={setReferenceImage} />}
                    {step === 3 && <S3ImageSelector title="Step 3: Select Checking Image" selectedImageUrl={checkingImage} onSelectImage={setCheckingImage} />}
                    {step === 4 && (
                        <div>
                             <h2 className="text-xl font-bold mb-6">Step 4: Review and Submit</h2>
                             <div className="bg-[#111111] p-6 rounded-lg border border-[#2F2F2F] space-y-4">
                                <p><strong className="text-[#A0A0A0] w-32 inline-block">Type:</strong> <span className="font-mono">{verificationType}</span></p>
                                <p><strong className="text-[#A0A0A0] w-32 inline-block">Reference:</strong> <span className="font-mono text-green-400">{referenceImage || 'Not Selected'}</span></p>
                                <p><strong className="text-[#A0A0A0] w-32 inline-block">Checking:</strong> <span className="font-mono text-green-400">{checkingImage || 'Not Selected'}</span></p>
                             </div>
                             <button onClick={handleSubmit} disabled={!referenceImage || !checkingImage} className="w-full mt-6 text-white font-bold py-3 px-8 rounded-lg bg-gradient-to-r from-blue-500 to-pink-500 hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed">Submit Verification</button>
                        </div>
                    )}
                     {step === 5 && verificationResult && (
                        <div>
                             <h2 className="text-xl font-bold mb-6">Step 5: Verification Result</h2>
                             <div className="bg-[#111111] p-6 rounded-lg border border-[#2F2F2F] space-y-3">
                                <p><strong className="text-[#A0A0A0] w-28 inline-block">ID:</strong> <span className="font-mono text-blue-400">{verificationResult.id}</span></p>
                                <p><strong className="text-[#A0A0A0] w-28 inline-block">Status:</strong> <span className={`font-bold ${verificationResult.status === 'COMPLETE' ? 'text-green-400' : 'text-yellow-400'}`}>{verificationResult.status}</span></p>
                                <div className="pt-3 mt-3 border-t border-[#2F2F2F]">
                                    <p className="text-[#A0A0A0] mb-2">LLM Analysis:</p>
                                    <p className="font-mono text-sm whitespace-pre-wrap">{verificationResult.llmAnalysis}</p>
                                </div>
                             </div>
                              <button onClick={resetWizard} className="w-full mt-6 text-white font-bold py-3 px-8 rounded-lg bg-gradient-to-r from-blue-500 to-pink-500 hover:opacity-90">Start New Verification</button>
                        </div>
                    )}

                    {/* Navigation for steps 2 & 3 */}
                    {(step === 2 || step === 3) && (
                         <div className="flex justify-between mt-6">
                            <button onClick={() => setStep(step - 1)} className="text-[#A0A0A0] font-bold py-3 px-8 rounded-lg bg-[#2F2F2F] hover:bg-white/10">Back</button>
                            <button onClick={() => setStep(step + 1)} disabled={(step === 2 && !referenceImage) || (step === 3 && !checkingImage)} className="text-white font-bold py-3 px-8 rounded-lg bg-gradient-to-r from-blue-500 to-pink-500 hover:opacity-90 disabled:opacity-50 disabled:cursor-not-allowed">Next Step</button>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}