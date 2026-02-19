// Global state
let currentProjectId = null;
let pollingInterval = null;

// DOM Elements
const form = document.getElementById('form');
const scrapeCard = document.getElementById('scrapeForm');
const progressCard = document.getElementById('progressCard');
const exportCard = document.getElementById('exportCard');

const progressFill = document.getElementById('progressFill');
const progressText = document.getElementById('progressText');
const statusText = document.getElementById('statusText');
const downloadedText = document.getElementById('downloadedText');
const totalText = document.getElementById('totalText');
const currentUrlText = document.getElementById('currentUrlText');
const errorsDiv = document.getElementById('errors');
const errorsList = document.getElementById('errorsList');

const startBtn = document.getElementById('startBtn');
const cancelBtn = document.getElementById('cancelBtn');
const exportZipBtn = document.getElementById('exportZipBtn');
const exportPdfBtn = document.getElementById('exportPdfBtn');
const newScrapeBtn = document.getElementById('newScrapeBtn');

// Event Listeners
form.addEventListener('submit', handleFormSubmit);
cancelBtn.addEventListener('click', handleCancel);
exportZipBtn.addEventListener('click', handleExportZip);
exportPdfBtn.addEventListener('click', handleExportPdf);
newScrapeBtn.addEventListener('click', handleNewScrape);

// Handle form submission
async function handleFormSubmit(e) {
    e.preventDefault();

    // Get form data
    const url = document.getElementById('url').value;
    const depth = parseInt(document.getElementById('depth').value);
    const filtersText = document.getElementById('filters').value;

    // Parse filters
    const filters = parseFilters(filtersText);

    // Prepare request
    const requestData = {
        url: url,
        depth: depth,
        filters: filters
    };

    try {
        // Start scraping
        const response = await fetch('/api/scrape', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(requestData)
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to start scraping');
        }

        const data = await response.json();
        currentProjectId = data.project_id;

        // Show progress card
        showProgress();

        // Start polling
        startPolling();

    } catch (error) {
        alert('Error: ' + error.message);
    }
}

// Parse filters from textarea
function parseFilters(text) {
    if (!text.trim()) return [];

    const lines = text.split('\n');
    const filters = [];

    for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed) continue;

        const parts = trimmed.split('|||');
        if (parts.length === 2) {
            filters.push({
                start: parts[0].trim(),
                end: parts[1].trim()
            });
        }
    }

    return filters;
}

// Show progress card
function showProgress() {
    scrapeCard.classList.add('hidden');
    progressCard.classList.remove('hidden');
    exportCard.classList.add('hidden');
    cancelBtn.classList.remove('hidden');
}

// Show export card
function showExport() {
    scrapeCard.classList.add('hidden');
    progressCard.classList.add('hidden');
    exportCard.classList.remove('hidden');
}

// Start status polling
function startPolling() {
    // Initial poll
    pollStatus();

    // Poll every 2 seconds
    pollingInterval = setInterval(pollStatus, 2000);
}

// Stop status polling
function stopPolling() {
    if (pollingInterval) {
        clearInterval(pollingInterval);
        pollingInterval = null;
    }
}

// Poll status endpoint
async function pollStatus() {
    if (!currentProjectId) return;

    try {
        const response = await fetch(`/api/project/${currentProjectId}/status`);
        
        if (!response.ok) {
            throw new Error('Failed to get status');
        }

        const data = await response.json();
        updateProgress(data);

        // Check if completed
        if (data.status === 'completed') {
            stopPolling();
            showExport();
        } else if (data.status === 'failed') {
            stopPolling();
            progressText.textContent = '❌ Scraping nie powiódł się';
            cancelBtn.textContent = 'Powrót';
        }

    } catch (error) {
        console.error('Polling error:', error);
    }
}

// Update progress UI
function updateProgress(data) {
    // Progress bar
    const progress = data.progress || 0;
    progressFill.style.width = progress + '%';
    progressFill.textContent = progress + '%';

    // Status
    statusText.textContent = formatStatus(data.status);
    downloadedText.textContent = data.pages_downloaded || 0;
    totalText.textContent = data.total_pages || 0;
    currentUrlText.textContent = data.current_url || '-';

    // Progress text
    if (data.status === 'in_progress') {
        progressText.textContent = `Pobieranie w toku... (${data.pages_downloaded}/${data.total_pages})`;
    } else if (data.status === 'completed') {
        progressText.textContent = '✅ Scraping zakończony!';
    } else if (data.status === 'failed') {
        progressText.textContent = '❌ Scraping nie powiódł się';
    }

    // Errors
    if (data.errors && data.errors.length > 0) {
        errorsDiv.classList.remove('hidden');
        errorsList.innerHTML = '';
        data.errors.forEach(error => {
            const li = document.createElement('li');
            li.textContent = error;
            errorsList.appendChild(li);
        });
    } else {
        errorsDiv.classList.add('hidden');
    }
}

// Format status for display
function formatStatus(status) {
    const statusMap = {
        'started': '🔄 Rozpoczęty',
        'in_progress': '⏳ W toku',
        'completed': '✅ Zakończony',
        'failed': '❌ Błąd'
    };
    return statusMap[status] || status;
}

// Handle cancel
function handleCancel() {
    stopPolling();
    handleNewScrape();
}

// Handle new scrape
function handleNewScrape() {
    currentProjectId = null;
    stopPolling();

    // Reset form
    form.reset();

    // Reset UI
    scrapeCard.classList.remove('hidden');
    progressCard.classList.add('hidden');
    exportCard.classList.add('hidden');

    progressFill.style.width = '0%';
    errorsDiv.classList.add('hidden');
}

// Handle ZIP export
async function handleExportZip() {
    if (!currentProjectId) return;

    try {
        window.location.href = `/api/project/${currentProjectId}/export/zip`;
    } catch (error) {
        alert('Export error: ' + error.message);
    }
}

// Handle PDF export
async function handleExportPdf() {
    if (!currentProjectId) return;

    try {
        // Show loading state
        exportPdfBtn.disabled = true;
        exportPdfBtn.textContent = '⏳ Generowanie PDF...';

        const response = await fetch(`/api/project/${currentProjectId}/export/pdf`, {
            method: 'POST'
        });

        if (!response.ok) {
            throw new Error('PDF generation failed');
        }

        // Download PDF
        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${currentProjectId}.pdf`;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(a);

    } catch (error) {
        alert('PDF export error: ' + error.message);
    } finally {
        // Reset button state
        exportPdfBtn.disabled = false;
        exportPdfBtn.textContent = '📄 Generuj PDF';
    }
}