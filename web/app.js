// Global variables
let currentProjectId = null;
let statusCheckInterval = null;

// DOM elements
const scrapeForm = document.getElementById('scrapeForm');
const progressSection = document.getElementById('progress');
const resultsSection = document.getElementById('results');
const submitBtn = document.getElementById('submitBtn');
const downloadZipBtn = document.getElementById('downloadZip');
const generatePdfBtn = document.getElementById('generatePdf');

// Status elements
const statusText = document.getElementById('statusText');
const progressPercent = document.getElementById('progressPercent');
const pagesCount = document.getElementById('pagesCount');
const currentUrl = document.getElementById('currentUrl');
const progressBar = document.getElementById('progressBar');
const errorsDiv = document.getElementById('errors');
const errorList = document.getElementById('errorList');

// Form submission
scrapeForm.addEventListener('submit', async (e) => {
    e.preventDefault();

    // Get form data
    const url = document.getElementById('url').value;
    const depth = parseInt(document.getElementById('depth').value);
    const filtersText = document.getElementById('filters').value;

    // Parse filters
    const filters = [];
    if (filtersText.trim()) {
        const lines = filtersText.split('\n');
        for (const line of lines) {
            const trimmed = line.trim();
            if (trimmed && trimmed.includes('|||')) {
                const [start, end] = trimmed.split('|||');
                filters.push({ start: start.trim(), end: end.trim() });
            }
        }
    }

    // Prepare request
    const requestBody = {
        url,
        depth,
        filters
    };

    // Disable form
    submitBtn.disabled = true;
    submitBtn.textContent = 'Rozpoczynanie...';

    // Show progress section
    progressSection.style.display = 'block';
    resultsSection.style.display = 'none';
    resetProgress();

    try {
        // Start scraping
        const response = await fetch('/api/scrape', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestBody)
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(`Server error: ${errorText}`);
        }

        const data = await response.json();
        currentProjectId = data.project_id;

        // Start polling for status
        startStatusPolling();

    } catch (error) {
        alert(`Błąd: ${error.message}`);
        submitBtn.disabled = false;
        submitBtn.textContent = 'Start Scraping';
        progressSection.style.display = 'none';
    }
});

// Start polling for status updates
function startStatusPolling() {
    if (statusCheckInterval) {
        clearInterval(statusCheckInterval);
    }

    statusCheckInterval = setInterval(checkStatus, 1000);
    checkStatus(); // Check immediately
}

// Check scraping status
async function checkStatus() {
    if (!currentProjectId) return;

    try {
        const response = await fetch(`/api/project/${currentProjectId}/status`);
        
        if (!response.ok) {
            throw new Error('Failed to fetch status');
        }

        const status = await response.json();
        updateProgress(status);

        // Check if completed or failed
        if (status.status === 'completed') {
            clearInterval(statusCheckInterval);
            showResults();
        } else if (status.status === 'failed') {
            clearInterval(statusCheckInterval);
            showError('Scraping failed');
        }

    } catch (error) {
        console.error('Status check error:', error);
    }
}

// Update progress UI
function updateProgress(status) {
    // Update status text
    const statusMap = {
        'in_progress': 'W trakcie...',
        'completed': 'Zakończone',
        'failed': 'Błąd'
    };
    statusText.textContent = statusMap[status.status] || status.status;

    // Update progress percentage
    const progress = status.progress || 0;
    progressPercent.textContent = `${progress}%`;
    progressBar.style.width = `${progress}%`;

    // Update pages count
    pagesCount.textContent = `${status.pages_downloaded} / ${status.total_pages || '?'}`;

    // Update current URL
    if (status.current_url) {
        currentUrl.textContent = status.current_url;
    }

    // Show errors if any
    if (status.errors && status.errors.length > 0) {
        errorsDiv.style.display = 'block';
        errorList.innerHTML = '';
        status.errors.forEach(error => {
            const li = document.createElement('li');
            li.textContent = error;
            errorList.appendChild(li);
        });
    }
}

// Show results section
function showResults() {
    progressSection.style.display = 'none';
    resultsSection.style.display = 'block';
    submitBtn.disabled = false;
    submitBtn.textContent = 'Start Scraping';

    // Setup export buttons
    downloadZipBtn.onclick = () => downloadZip();
    generatePdfBtn.onclick = () => generatePdf();
}

// Show error
function showError(message) {
    alert(`Błąd: ${message}`);
    submitBtn.disabled = false;
    submitBtn.textContent = 'Start Scraping';
}

// Reset progress UI
function resetProgress() {
    statusText.textContent = 'Inicjalizacja...';
    progressPercent.textContent = '0%';
    pagesCount.textContent = '0';
    currentUrl.textContent = '-';
    progressBar.style.width = '0%';
    errorsDiv.style.display = 'none';
    errorList.innerHTML = '';
}

// Download ZIP
function downloadZip() {
    if (!currentProjectId) return;
    
    downloadZipBtn.disabled = true;
    downloadZipBtn.textContent = 'Pobieranie...';

    const url = `/api/project/${currentProjectId}/export/zip`;
    
    // Create temporary link and trigger download
    const a = document.createElement('a');
    a.href = url;
    a.download = `${currentProjectId}.zip`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);

    setTimeout(() => {
        downloadZipBtn.disabled = false;
        downloadZipBtn.textContent = '📦 Pobierz ZIP';
    }, 1000);
}

// Generate and download PDF
async function generatePdf() {
    if (!currentProjectId) return;

    generatePdfBtn.disabled = true;
    generatePdfBtn.textContent = 'Generowanie...';

    try {
        const response = await fetch(`/api/project/${currentProjectId}/export/pdf`, {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('Failed to generate PDF');
        }

        // Download PDF
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${currentProjectId}.pdf`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);

    } catch (error) {
        alert(`Błąd generowania PDF: ${error.message}`);
    } finally {
        generatePdfBtn.disabled = false;
        generatePdfBtn.textContent = '📄 Generuj PDF';
    }
}
