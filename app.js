// Video Saver v2 Logic
const BACKEND_URL = 'https://save-media-backend-production.up.railway.app';

let allFormats = [];
let selectedExt = 'mp4';
let selectedFormatID = '';

// Initialize Lucide icons
lucide.createIcons();

// GSAP Animations
document.addEventListener('DOMContentLoaded', () => {
    gsap.from('[data-gsap="fade-down"]', {
        y: -30,
        opacity: 0,
        duration: 0.8,
        ease: 'power3.out'
    });

    gsap.from('[data-gsap="fade-up"]', {
        y: 40,
        opacity: 0,
        duration: 1,
        delay: 0.2,
        stagger: 0.2,
        ease: 'power4.out'
    });
});

// Event Listeners
const urlInput = document.getElementById('urlInput');
const downloadBtn = document.getElementById('downloadBtn');
const clearBtn = document.getElementById('clearBtn');
const resultsSection = document.getElementById('resultsSection');
const resultStatus = document.getElementById('resultStatus');
const videoDetails = document.getElementById('videoDetails');

urlInput.addEventListener('input', () => {
    clearBtn.style.display = urlInput.value ? 'block' : 'none';
});

clearBtn.addEventListener('click', () => {
    urlInput.value = '';
    clearBtn.style.display = 'none';
    urlInput.focus();
});

urlInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') handleFetchMetadata();
});

downloadBtn.addEventListener('click', handleFetchMetadata);

async function handleFetchMetadata() {
    const url = urlInput.value.trim();
    if (!url) return showToast('Please paste a URL first');

    // Reset UI
    resultsSection.style.display = 'block';
    resultStatus.style.display = 'flex';
    videoDetails.style.display = 'none';
    
    gsap.fromTo(resultsSection, 
        { opacity: 0, y: 30 },
        { opacity: 1, y: 0, duration: 0.6, ease: 'back.out(1.7)' }
    );

    try {
        const response = await fetch(`${BACKEND_URL}/api/metadata?url=${encodeURIComponent(url)}`);
        const data = await response.json();

        if (data.error) throw new Error(data.error);

        populateVideoInfo(data);
        
        // Fetch specific formats
        await fetchFormats(url);

    } catch (err) {
        resultStatus.innerHTML = `<span style="color: #ef4444">⚠ Error: ${err.message}</span>`;
    }
}

function populateVideoInfo(data) {
    resultStatus.style.display = 'none';
    videoDetails.style.display = 'grid';
    
    document.getElementById('videoTitle').textContent = data.title || 'Untitled Video';
    document.getElementById('platformBadge').textContent = data.platform || 'Video';
    document.getElementById('videoDuration').innerHTML = `<i data-lucide="clock"></i> Duration: ${data.duration || 'N/A'}`;
    
    const thumb = document.getElementById('thumbnail');
    thumb.src = `${BACKEND_URL}/api/thumbnail?url=${encodeURIComponent(data.thumbnail)}`;
    
    lucide.createIcons();
}

async function fetchFormats(url) {
    try {
        const res = await fetch(`${BACKEND_URL}/api/formats?url=${encodeURIComponent(url)}`);
        const data = await res.json();
        
        if (data.formats && data.formats.length > 0) {
            allFormats = data.formats;
            setupCustomSelects();
        } else {
            showToast('No high-quality formats found. Using fallback.');
        }
    } catch (err) {
        console.error('Formats error:', err);
    }
}

function setupCustomSelects() {
    const formatOptions = document.getElementById('formatOptions');
    const qualityOptions = document.getElementById('qualityOptions');
    
    // Unique extensions
    const exts = [...new Set(allFormats.map(f => f.ext))];
    formatOptions.innerHTML = '';
    exts.forEach(ext => {
        const div = document.createElement('div');
        div.textContent = ext.toUpperCase();
        div.onclick = () => selectExt(ext);
        formatOptions.appendChild(div);
    });

    // Default to first ext and best quality
    selectExt(exts[0] || 'mp4');
}

function selectExt(ext) {
    selectedExt = ext;
    document.querySelector('#formatSelect .selected-value').textContent = `Format: ${ext.toUpperCase()}`;
    document.getElementById('formatOptions').classList.remove('show-options');

    const filtered = allFormats.filter(f => f.ext === ext);
    const qualityOptions = document.getElementById('qualityOptions');
    qualityOptions.innerHTML = '';
    
    filtered.forEach(f => {
        const div = document.createElement('div');
        div.textContent = f.quality;
        div.onclick = () => {
            selectedFormatID = f.format_id;
            document.querySelector('#qualitySelect .selected-value').textContent = f.quality;
            document.getElementById('fileSize').textContent = `Size: ${formatFileSize(f.filesize)}`;
            qualityOptions.classList.remove('show-options');
        };
        qualityOptions.appendChild(div);
    });

    // Pick first quality automatically
    if (filtered[0]) {
        selectedFormatID = filtered[0].format_id;
        document.querySelector('#qualitySelect .selected-value').textContent = filtered[0].quality;
        document.getElementById('fileSize').textContent = `Size: ${formatFileSize(filtered[0].filesize)}`;
    }
}

// Custom Select Toggles
document.getElementById('formatSelect').onclick = (e) => {
    document.getElementById('formatOptions').classList.toggle('show-options');
    document.getElementById('qualityOptions').classList.remove('show-options');
    e.stopPropagation();
};

document.getElementById('qualitySelect').onclick = (e) => {
    document.getElementById('qualityOptions').classList.toggle('show-options');
    document.getElementById('formatOptions').classList.remove('show-options');
    e.stopPropagation();
};

window.onclick = () => {
    document.getElementById('formatOptions').classList.remove('show-options');
    document.getElementById('qualityOptions').classList.remove('show-options');
};

// Download Trigger
document.getElementById('mainDownloadBtn').onclick = () => {
    const title = document.getElementById('videoTitle').textContent;
    const downloadUrl = `${BACKEND_URL}/api/download?url=${encodeURIComponent(urlInput.value)}&format_id=${encodeURIComponent(selectedFormatID)}&ext=${encodeURIComponent(selectedExt)}&title=${encodeURIComponent(title)}`;
    
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.click();
    showToast('Download started!');
};

function formatFileSize(bytes) {
    if (!bytes || bytes === 0) return 'Unknown';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function showToast(msg) {
    const toast = document.getElementById('toast');
    toast.textContent = msg;
    toast.classList.add('show');
    setTimeout(() => toast.classList.remove('show'), 3000);
}
