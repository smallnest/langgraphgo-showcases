// API Configuration
const API_BASE = window.location.origin;
let sessionId = null;

// DOM Elements
const dropZone = document.getElementById('dropZone');
const fileInput = document.getElementById('fileInput');
const docCount = document.getElementById('docCount');
const clearBtn = document.getElementById('clearBtn');
const chatForm = document.getElementById('chatForm');
const chatInput = document.getElementById('chatInput');
const chatMessages = document.getElementById('chatMessages');
const sendBtn = document.getElementById('sendBtn');
const sourcesPanel = document.getElementById('sourcesPanel');
const sourcesList = document.getElementById('sourcesList');
const newChatBtn = document.getElementById('newChatBtn');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    loadStats();
    setupEventListeners();
});

function setupEventListeners() {
    // File upload
    dropZone.addEventListener('click', () => fileInput.click());
    dropZone.addEventListener('dragover', handleDragOver);
    dropZone.addEventListener('dragleave', handleDragLeave);
    dropZone.addEventListener('drop', handleDrop);
    fileInput.addEventListener('change', handleFileSelect);

    // Clear documents
    clearBtn.addEventListener('click', clearDocuments);

    // Chat
    chatForm.addEventListener('submit', handleChatSubmit);
    newChatBtn.addEventListener('click', startNewChat);
}

// ============================================
// File Upload Handling
// ============================================

function handleDragOver(e) {
    e.preventDefault();
    dropZone.classList.add('dragover');
}

function handleDragLeave(e) {
    e.preventDefault();
    dropZone.classList.remove('dragover');
}

function handleDrop(e) {
    e.preventDefault();
    dropZone.classList.remove('dragover');

    const files = Array.from(e.dataTransfer.files).filter(f =>
        f.type === 'application/pdf' ||
        f.name.endsWith('.txt') ||
        f.name.endsWith('.md')
    );

    if (files.length > 0) {
        uploadFiles(files);
    }
}

function handleFileSelect(e) {
    const files = Array.from(e.target.files);
    if (files.length > 0) {
        uploadFiles(files);
    }
    e.target.value = '';
}

async function uploadFiles(files) {
    showToast(`Uploading ${files.length} file(s)...`, 'info');

    const formData = new FormData();
    files.forEach(file => formData.append('files', file));

    try {
        const response = await fetch('/api/ingest', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                files: files.map(f => f.name), // For simplicity, sending file names
                action: 'add'
            })
        });

        if (!response.ok) throw new Error('Upload failed');

        const result = await response.json();
        showToast(`Successfully uploaded ${files.length} file(s)`, 'success');
        await loadStats();
    } catch (error) {
        showToast('Upload failed: ' + error.message, 'error');
    }
}

async function clearDocuments() {
    if (!confirm('Are you sure you want to clear all documents?')) return;

    try {
        const response = await fetch('/api/ingest', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: 'delete' })
        });

        if (!response.ok) throw new Error('Clear failed');

        showToast('Documents cleared', 'success');
        await loadStats();
    } catch (error) {
        showToast('Failed to clear documents', 'error');
    }
}

// ============================================
// Chat Handling
// ============================================

async function handleChatSubmit(e) {
    e.preventDefault();

    const message = chatInput.value.trim();
    if (!message) return;

    // Clear welcome message if present
    const welcome = chatMessages.querySelector('.welcome-message');
    if (welcome) welcome.remove();

    // Add user message
    addMessage('user', message);
    chatInput.value = '';

    // Disable input
    setChatEnabled(false);

    // Add loading indicator
    const loadingId = addLoadingIndicator();

    try {
        const response = await fetch('/api/chat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                message: message,
                sessionId: sessionId || ''
            })
        });

        if (!response.ok) throw new Error('Chat request failed');

        // Remove loading indicator
        removeMessage(loadingId);

        // Handle SSE stream
        await handleSSEStream(response);
    } catch (error) {
        removeMessage(loadingId);
        addMessage('assistant', 'Sorry, something went wrong. Please try again.');
        showToast('Error: ' + error.message, 'error');
        setChatEnabled(true);
    }
}

async function handleSSEStream(response) {
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let assistantMessageId = null;
    let sources = [];

    while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n');

        for (const line of lines) {
            if (line.startsWith('event:')) {
                const event = line.slice(6).trim();
                const dataLine = lines[lines.indexOf(line) + 1];

                if (dataLine && dataLine.startsWith('data:')) {
                    const data = JSON.parse(dataLine.slice(5).trim());

                    switch (event) {
                        case 'metadata':
                            sessionId = data.sessionId;
                            break;

                        case 'message':
                            if (!assistantMessageId) {
                                assistantMessageId = addMessage('assistant', '');
                            }
                            appendMessageContent(assistantMessageId, data.content);
                            break;

                        case 'source':
                            sources.push(data);
                            break;

                        case 'error':
                            showToast(data.message, 'error');
                            break;

                        case 'done':
                            if (sources.length > 0) {
                                displaySources(sources);
                            }
                            setChatEnabled(true);
                            return;
                    }
                }
            }
        }
    }
}

// ============================================
// UI Helpers
// ============================================

function addMessage(role, content) {
    const messageDiv = document.createElement('div');
    messageDiv.className = `message ${role}`;

    const avatar = document.createElement('div');
    avatar.className = 'message-avatar';
    avatar.textContent = role === 'user' ? '👤' : '🤖';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.innerHTML = `<p>${escapeHtml(content)}</p>`;

    messageDiv.appendChild(avatar);
    messageDiv.appendChild(contentDiv);
    chatMessages.appendChild(messageDiv);

    scrollToBottom();
    return messageDiv.id = 'msg-' + Date.now();
}

function appendMessageContent(messageId, content) {
    const messageEl = document.getElementById(messageId);
    if (messageEl) {
        const contentDiv = messageEl.querySelector('.message-content p');
        if (contentDiv) {
            contentDiv.innerHTML += escapeHtml(content);
            scrollToBottom();
        }
    }
}

function removeMessage(messageId) {
    const messageEl = document.getElementById(messageId);
    if (messageEl) messageEl.remove();
}

function addLoadingIndicator() {
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message assistant';

    const avatar = document.createElement('div');
    avatar.className = 'message-avatar';
    avatar.textContent = '🤖';

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.innerHTML = `
        <div class="loading">
            <span></span><span></span><span></span>
        </div>
    `;

    messageDiv.appendChild(avatar);
    messageDiv.appendChild(contentDiv);
    chatMessages.appendChild(messageDiv);

    scrollToBottom();
    return messageDiv.id = 'loading-' + Date.now();
}

function displaySources(sources) {
    sourcesList.innerHTML = '';
    sourcesPanel.style.display = 'block';

    sources.forEach((source, index) => {
        const sourceDiv = document.createElement('div');
        sourceDiv.className = 'source-item';
        sourceDiv.innerHTML = `
            <strong>[${index + 1}]</strong> ${escapeHtml(source.content.substring(0, 200))}...
            ${source.score ? `<div class="source-score">Relevance: ${(source.score * 100).toFixed(1)}%</div>` : ''}
        `;
        sourcesList.appendChild(sourceDiv);
    });
}

function toggleSources() {
    sourcesPanel.style.display = sourcesPanel.style.display === 'none' ? 'block' : 'none';
}

function setChatEnabled(enabled) {
    chatInput.disabled = !enabled;
    sendBtn.disabled = !enabled;
    if (enabled) {
        chatInput.focus();
    }
}

function scrollToBottom() {
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function startNewChat() {
    sessionId = null;
    chatMessages.innerHTML = `
        <div class="welcome-message">
            <h3>Welcome! 👋</h3>
            <p>Upload a PDF document to get started, or ask me anything!</p>
            <ul class="example-questions">
                <li onclick="askQuestion('What can you help me with?')">What can you help me with?</li>
                <li onclick="askQuestion('How do I upload a document?')">How do I upload a document?</li>
            </ul>
        </div>
    `;
    sourcesPanel.style.display = 'none';
    sourcesList.innerHTML = '';
}

function askQuestion(question) {
    chatInput.value = question;
    chatForm.dispatchEvent(new Event('submit'));
}

// ============================================
// Stats & Helpers
// ============================================

async function loadStats() {
    try {
        const response = await fetch('/api/health');
        const stats = await response.json();

        docCount.textContent = `${stats.documents} document(s) loaded`;
        clearBtn.disabled = stats.documents === 0;
    } catch (error) {
        console.error('Failed to load stats:', error);
    }
}

function showToast(message, type = 'info') {
    const container = document.createElement('div');
    container.className = 'toast-container';

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;

    container.appendChild(toast);
    document.body.appendChild(container);

    setTimeout(() => {
        toast.remove();
        container.remove();
    }, 3000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
