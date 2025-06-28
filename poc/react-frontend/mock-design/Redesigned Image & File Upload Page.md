import React, { useState, useCallback, useEffect } from 'react';

// --- ICONS (Helper Components) ---
const UploadIcon = () => <svg className="w-10 h-10 mb-3 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-4-4V7a4 4 0 014-4h.5a3.5 3.5 0 017 0H15a4 4 0 014 4v5a4 4 0 01-4 4" /><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 11v9m0 0l-3-3m3 3l3-3" /></svg>;
const CheckCircleIcon = () => <svg className="w-6 h-6 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>;
const XCircleIcon = () => <svg className="w-6 h-6 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>;
const FolderIcon = () => <svg className="w-5 h-5 mr-3 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20"><path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z"></path></svg>;
const UpIcon = () => <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 10l7-7m0 0l7 7m-7-7v18"></path></svg>;
const RootIcon = () => <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3"></path></svg>;

// --- MOCK DATA ---
const s3MockData = {
    '': { type: 'folder', items: ['processed/', 'raw/', 'planograms/'] },
    'processed/': { type: 'folder', items: ['2025/', '2024/'] },
    'raw/': { type: 'folder', items: ['device_images/'] },
    'planograms/': { type: 'folder', items: ['store123/', 'store456/'] }
};

// --- S3 Path Browser Panel (for Reference Upload) ---
const S3PathBrowserPanel = ({ path, setPath }) => {
    const [items, setItems] = useState([]);

    useEffect(() => {
        const data = s3MockData[path] || { items: [] };
        setItems(data.items.filter(item => item.endsWith('/'))); // Only show folders
    }, [path]);

    const goUp = () => {
        if (path === '') return;
        const parts = path.split('/').filter(p => p);
        parts.pop();
        const newPath = parts.length > 0 ? parts.join('/') + '/' : '';
        setPath(newPath);
    };

    return (
        <div className="w-full lg:w-1/3 flex-shrink-0 flex flex-col h-[50vh] lg:h-auto">
            <h3 className="font-bold text-white mb-2">1. Select Destination</h3>
            <div className="bg-[#111111] border border-[#2F2F2F] rounded-xl p-4 flex-grow flex flex-col">
                <div className="flex items-center gap-2 mb-4 p-2 bg-[#1E1E1E] rounded-lg border border-[#2F2F2F] flex-shrink-0">
                    <button onClick={goUp} className="p-2 hover:bg-white/10 rounded-md"><UpIcon /></button>
                    <button onClick={() => setPath('')} className="p-2 hover:bg-white/10 rounded-md"><RootIcon /></button>
                </div>
                <p className="px-3 py-2 bg-[#1E1E1E] rounded-md text-sm font-mono text-blue-400 flex-grow mb-4 whitespace-nowrap overflow-x-auto">s3://reference-bucket/{path}</p>
                <div className="overflow-y-auto flex-grow">
                    {items.map(item => (
                        <div key={item} onClick={() => setPath(path + item)} className="flex items-center p-2 rounded-md cursor-pointer hover:bg-white/10 transition-colors">
                            <FolderIcon />
                            <span className="text-sm font-mono truncate">{item}</span>
                        </div>
                    ))}
                </div>
            </div>
        </div>
    );
};

// --- Main App Component ---
export default function FileUploadPage() {
    const [uploadTarget, setUploadTarget] = useState('checking');
    const [selectedFile, setSelectedFile] = useState(null);
    const [previewData, setPreviewData] = useState(null);
    const [uploadStatus, setUploadStatus] = useState({ status: 'idle', message: '' });
    const [history, setHistory] = useState([]);
    const [activeTab, setActiveTab] = useState('uploader');
    const [destinationPath, setDestinationPath] = useState('');

    const handleFileSelect = useCallback((file) => {
        if (!file) return;
        const isReference = uploadTarget === 'reference';
        if (isReference && file.type !== 'application/json') {
            setUploadStatus({ status: 'error', message: 'Invalid file type. Please select a JSON file.' });
            return;
        }
        if (!isReference && !file.type.startsWith('image/')) {
            setUploadStatus({ status: 'error', message: 'Invalid file type. Please select an image.' });
            return;
        }
        if (file.size > 10 * 1024 * 1024) { // 10MB
            setUploadStatus({ status: 'error', message: 'File is too large (Max 10MB).' });
            return;
        }
        setSelectedFile(file);
        setUploadStatus({ status: 'idle', message: '' });
        const reader = new FileReader();
        if (isReference) {
            reader.onload = (e) => setPreviewData(e.target.result);
            reader.readAsText(file);
        } else {
            reader.onload = (e) => setPreviewData(e.target.result);
            reader.readAsDataURL(file);
        }
    }, [uploadTarget]);

    const handleDrop = (e) => {
        e.preventDefault();
        if (e.dataTransfer.files && e.dataTransfer.files[0]) {
            handleFileSelect(e.dataTransfer.files[0]);
        }
    };

    const handleUpload = () => {
        if (!selectedFile) return;
        setUploadStatus({ status: 'uploading', message: 'Uploading file...' });
        setTimeout(() => {
            const finalPath = destinationPath || (uploadTarget === 'checking' ? `uploads/${new Date().toISOString().split('T')[0]}/` : '');
            const newHistoryItem = {
                id: Date.now(), name: selectedFile.name, bucket: uploadTarget,
                url: `s3://${uploadTarget}-bucket/${finalPath}${selectedFile.name}`,
                date: new Date().toLocaleString()
            };
            setHistory([newHistoryItem, ...history]);
            setUploadStatus({ status: 'success', message: `Successfully uploaded to ${newHistoryItem.url}` });
            resetSelection();
        }, 1500);
    };

    const resetSelection = () => {
        setSelectedFile(null);
        setPreviewData(null);
        setUploadStatus({ status: 'idle', message: '' });
        if (uploadTarget === 'checking') setDestinationPath('');
    };
    
    const handleClearHistory = () => setHistory([]);

    const renderUploader = () => {
        if (uploadTarget === 'reference') {
            return (
                <div className="flex flex-col lg:flex-row gap-8">
                    <S3PathBrowserPanel path={destinationPath} setPath={setDestinationPath} />
                    <div className="w-full lg:w-2/3">
                        <h3 className="font-bold text-white mb-2">2. Select & Upload File</h3>
                        {!selectedFile ? (
                             <div onDrop={handleDrop} onDragOver={(e) => e.preventDefault()} className="w-full p-10 h-full border-2 border-dashed border-[#2F2F2F] rounded-2xl text-center cursor-pointer hover:border-purple-500 transition-colors flex flex-col justify-center items-center" onClick={() => document.getElementById('fileInput').click()}>
                                <input type="file" id="fileInput" className="hidden" accept=".json" onChange={(e) => handleFileSelect(e.target.files[0])} />
                                <UploadIcon />
                                <p className="text-lg font-semibold">Drop JSON file here</p>
                            </div>
                        ) : (
                             <div className="bg-[#111111] border border-[#2F2F2F] rounded-xl p-4">
                                <h4 className="text-sm font-bold mb-2">Preview</h4>
                                <pre className="text-xs text-green-400 w-full h-48 overflow-auto whitespace-pre-wrap bg-[#1E1E1E] p-2 rounded-md">{previewData}</pre>
                                <div className="mt-4">
                                    <label className="text-sm text-[#A0A0A0]">Custom Filename (Optional)</label>
                                    <input type="text" placeholder={selectedFile.name} className="w-full mt-1 bg-[#1E1E1E] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none" />
                                </div>
                                <div className="flex gap-4 mt-4">
                                    <button onClick={handleUpload} className="w-full text-white font-bold py-3 px-8 rounded-lg shadow-md bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 hover:opacity-90">Upload File</button>
                                    <button onClick={resetSelection} className="w-full text-sm text-[#A0A0A0] hover:text-white bg-[#2F2F2F] rounded-lg">Cancel</button>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            )
        }
        
        // Render 'Checking' bucket uploader
        return (
             <>
                {!selectedFile && (
                    <div onDrop={handleDrop} onDragOver={(e) => e.preventDefault()} className="w-full p-10 border-2 border-dashed border-[#2F2F2F] rounded-2xl text-center cursor-pointer hover:border-purple-500 transition-colors" onClick={() => document.getElementById('fileInput').click()}>
                        <input type="file" id="fileInput" className="hidden" accept="image/*" onChange={(e) => handleFileSelect(e.target.files[0])} />
                        <div className="flex flex-col items-center">
                            <UploadIcon />
                            <p className="text-lg font-semibold">Drag & drop your image here</p>
                            <p className="text-sm text-[#A0A0A0]">or click to browse</p>
                        </div>
                    </div>
                )}
                {selectedFile && (
                    <div className="flex flex-col md:flex-row gap-8">
                        <div className="w-full md:w-1/2">
                            <h3 className="font-bold text-white mb-2">Preview</h3>
                            <div className="bg-[#111111] border border-[#2F2F2F] rounded-xl aspect-square flex items-center justify-center p-4">
                                <img src={previewData} alt="preview" className="max-w-full max-h-full object-contain rounded-md" />
                            </div>
                        </div>
                        <div className="w-full md:w-1/2">
                             <h3 className="font-bold text-white mb-2">Configuration</h3>
                             <div className="bg-[#111111] border border-[#2F2F2F] rounded-xl p-6 space-y-4">
                                <div>
                                    <label className="text-sm text-[#A0A0A0]">File Name</label>
                                    <p className="font-mono text-blue-400">{selectedFile.name}</p>
                                </div>
                                <div>
                                    <label className="text-sm text-[#A0A0A0]">Destination Path (Optional)</label>
                                    <input type="text" value={destinationPath} onChange={(e) => setDestinationPath(e.target.value)} placeholder="e.g., device_images/shelf1" className="w-full mt-1 bg-[#1E1E1E] border border-[#2F2F2F] rounded-lg p-2 focus:ring-2 focus:ring-purple-500 focus:outline-none" />
                                </div>
                             </div>
                             <button onClick={handleUpload} className="w-full mt-6 text-white font-bold py-3 px-8 rounded-lg shadow-md bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 hover:opacity-90">Upload File</button>
                             <button onClick={resetSelection} className="w-full mt-3 text-sm text-[#A0A0A0] hover:text-white">Cancel</button>
                        </div>
                    </div>
                )}
            </>
        )
    };
    
    const renderHistory = () => (
         <div className="space-y-3">
             {history.length === 0 ? <p className="text-[#A0A0A0] text-center py-4">No recent uploads.</p> : history.map(item => (
                 <div key={item.id} className="bg-[#111111] p-3 rounded-lg border border-[#2F2F2F] text-sm flex justify-between items-center">
                     <div>
                        <p className="font-mono text-purple-400">{item.name}</p>
                        <p className="text-xs text-[#A0A0A0]">{item.date}</p>
                     </div>
                     <span className={`px-2 py-0.5 rounded-full text-xs font-bold ${item.bucket === 'reference' ? 'bg-blue-900 text-blue-300' : 'bg-pink-900 text-pink-300'}`}>{item.bucket}</span>
                 </div>
             ))}
        </div>
    );

    return (
        <div className="bg-[#111111] text-[#F5F5F5] min-h-screen font-sans p-8 flex items-center justify-center">
            <div className="w-full max-w-5xl">
                <h1 className="text-center text-4xl font-bold mb-4">
                    <span className="bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 bg-clip-text text-transparent">Image & File Upload</span>
                </h1>
                <p className="text-center text-[#A0A0A0] mb-10">Upload JSON layouts or checking images directly to S3.</p>

                <div className="bg-[#1E1E1E] p-8 rounded-2xl border border-[#2F2F2F]">
                    <div className="flex justify-center mb-8">
                        <div className="relative flex p-1 bg-[#111111] rounded-full border border-[#2F2F2F]">
                            <button onClick={() => { setUploadTarget('checking'); resetSelection(); }} className={`relative z-10 w-48 text-center py-2 rounded-full text-sm font-bold transition-colors ${uploadTarget === 'checking' ? 'text-white' : 'text-[#A0A0A0]'}`}>Checking Bucket (Images)</button>
                            <button onClick={() => { setUploadTarget('reference'); resetSelection(); }} className={`relative z-10 w-48 text-center py-2 rounded-full text-sm font-bold transition-colors ${uploadTarget === 'reference' ? 'text-white' : 'text-[#A0A0A0]'}`}>Reference Bucket (JSON)</button>
                            <span className="absolute top-1 h-10 w-48 rounded-full bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 transition-transform duration-300 ease-out" style={{ transform: `translateX(${uploadTarget === 'reference' ? '100%' : '0%'})` }} />
                        </div>
                    </div>
                    
                    <div className="flex justify-between items-center border-b border-[#2F2F2F] mb-6">
                        <div className="flex">
                            <button onClick={() => setActiveTab('uploader')} className={`py-2 px-4 text-sm font-medium ${activeTab === 'uploader' ? 'text-white border-b-2 border-pink-500' : 'text-[#A0A0A0]'}`}>Uploader</button>
                            <button onClick={() => setActiveTab('history')} className={`py-2 px-4 text-sm font-medium ${activeTab === 'history' ? 'text-white border-b-2 border-pink-500' : 'text-[#A0A0A0]'}`}>Recent Uploads</button>
                        </div>
                         {activeTab === 'history' && history.length > 0 && (
                            <button onClick={handleClearHistory} className="text-xs text-red-400 hover:text-red-300 px-3 py-1 rounded-md bg-red-900/50">Clear History</button>
                         )}
                    </div>

                    <div>
                        {activeTab === 'uploader' ? renderUploader() : renderHistory()}
                    </div>

                    {uploadStatus.status !== 'idle' && uploadStatus.status !== 'uploading' && (
                        <div className={`mt-6 flex items-center gap-3 p-3 rounded-lg text-sm ${uploadStatus.status === 'success' ? 'bg-green-900/50' : 'bg-red-900/50'}`}>
                            {uploadStatus.status === 'success' ? <CheckCircleIcon /> : <XCircleIcon />}
                            <span>{uploadStatus.message}</span>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}