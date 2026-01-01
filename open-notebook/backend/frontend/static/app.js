// ============================================
// OPEN NOTEBOOK - Application Logic
// ============================================

class OpenNotebook {
    constructor() {
        this.notebooks = [];
        this.currentNotebook = null;
        this.apiBase = '/api';
        this.currentChatSession = null;

        this.init();
    }

    async init() {
        this.bindEvents();
        await this.loadNotebooks();

        // Auto-select first notebook if available
        if (this.notebooks.length > 0) {
            this.selectNotebook(this.notebooks[0].id);
        }
    }

    bindEvents() {
        // Notebook actions
        document.getElementById('btnNewNotebook').addEventListener('click', () => this.showNewNotebookModal());
        document.getElementById('btnCreateFirst').addEventListener('click', () => this.showNewNotebookModal());
        document.getElementById('newNotebookForm').addEventListener('submit', (e) => this.handleCreateNotebook(e));
        document.getElementById('btnCloseNotebookModal').addEventListener('click', () => this.closeModals());
        document.getElementById('btnCancelNotebook').addEventListener('click', () => this.closeModals());

        // Source actions
        document.getElementById('btnAddSource').addEventListener('click', () => this.showAddSourceModal());
        document.getElementById('btnCloseSourceModal').addEventListener('click', () => this.closeModals());
        document.getElementById('dropZone').addEventListener('click', () => document.getElementById('fileInput').click());
        document.getElementById('fileInput').addEventListener('change', (e) => this.handleFileUpload(e));
        document.getElementById('textSourceForm').addEventListener('submit', (e) => this.handleTextSource(e));
        document.getElementById('urlSourceForm').addEventListener('submit', (e) => this.handleURLSource(e));
        document.getElementById('btnCancelText').addEventListener('click', () => this.closeModals());
        document.getElementById('btnCancelURL').addEventListener('click', () => this.closeModals());

        // Source tabs
        document.querySelectorAll('.source-tab').forEach(tab => {
            tab.addEventListener('click', () => {
                document.querySelectorAll('.source-tab').forEach(t => t.classList.remove('active'));
                document.querySelectorAll('.source-content').forEach(c => c.classList.remove('active'));
                tab.classList.add('active');
                document.getElementById(`source${tab.dataset.source.charAt(0).toUpperCase() + tab.dataset.source.slice(1)}`).classList.add('active');
            });
        });

        // Panel tabs
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', () => {
                document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
                document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
                btn.classList.add('active');
                document.getElementById(`tab-${btn.dataset.tab}`).classList.add('active');
            });
        });

        // Transform cards
        document.querySelectorAll('.transform-card').forEach(card => {
            card.addEventListener('click', (e) => {
                e.preventDefault();
                this.handleTransform(card.dataset.type, card);
            });
        });

        document.getElementById('btnCustomTransform').addEventListener('click', (e) => {
            this.handleTransform('custom', e.currentTarget);
        });

        // Chat
        document.getElementById('chatForm').addEventListener('submit', (e) => this.handleChat(e));

        // Modal overlay
        document.getElementById('modalOverlay').addEventListener('click', (e) => {
            if (e.target.id === 'modalOverlay') {
                this.closeModals();
            }
        });

        // Drag and drop
        const dropZone = document.getElementById('dropZone');
        dropZone.addEventListener('dragover', (e) => {
            e.preventDefault();
            dropZone.classList.add('drag-over');
        });
        dropZone.addEventListener('dragleave', () => {
            dropZone.classList.remove('drag-over');
        });
        dropZone.addEventListener('drop', (e) => this.handleDrop(e));
    }

    // API Methods
    async api(endpoint, options = {}) {
        const defaults = {
            headers: {
                'Content-Type': 'application/json',
            },
            cache: 'no-store'
        };

        // Add timestamp to prevent caching for GET requests
        let url = `${this.apiBase}${endpoint}`;
        if (!options.method || options.method === 'GET') {
            const separator = url.includes('?') ? '&' : '?';
            url += `${separator}_t=${Date.now()}`;
        }

        const response = await fetch(url, { ...defaults, ...options });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ error: 'Request failed' }));
            throw new Error(error.error || 'Request failed');
        }

        // Handle 204 No Content responses
        if (response.status === 204) {
            return null;
        }

        return response.json();
    }

    // Notebook Methods
    async loadNotebooks() {
        try {
            this.notebooks = await this.api('/notebooks');
            this.renderNotebooks();
            this.updateFooter();
        } catch (error) {
            this.showError('Failed to load notebooks');
        }
    }

    renderNotebooks() {
        const container = document.getElementById('notebookList');
        const template = document.getElementById('notebookTemplate');

        // Clear existing content
        container.innerHTML = '';

        // Add empty state if no notebooks
        if (this.notebooks.length === 0) {
            container.innerHTML = `
                <div class="empty-state" style="display: flex;">
                    <svg width="48" height="48" viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5">
                        <rect x="8" y="8" width="32" height="32" rx="2"/>
                        <line x1="16" y1="16" x2="32" y2="16"/>
                        <line x1="16" y1="22" x2="28" y2="22"/>
                    </svg>
                    <p>No notebooks yet</p>
                    <button id="btnCreateFirst" class="btn-primary">Create your first notebook</button>
                </div>
            `;
            // Re-bind the create button event
            document.getElementById('btnCreateFirst').addEventListener('click', () => this.showNewNotebookModal());
            return;
        }

        // Render notebook items
        this.notebooks.forEach(nb => {
            const clone = template.content.cloneNode(true);
            const item = clone.querySelector('.notebook-item');

            item.dataset.id = nb.id;
            if (this.currentNotebook?.id === nb.id) {
                item.classList.add('active');
            }

            item.querySelector('.notebook-name').textContent = nb.name;

            // Fetch counts
            this.loadNotebookCounts(nb.id, item);

            item.addEventListener('click', (e) => {
                if (!e.target.closest('.btn-delete-notebook')) {
                    this.selectNotebook(nb.id);
                }
            });

            item.querySelector('.btn-delete-notebook').addEventListener('click', (e) => {
                e.stopPropagation();
                if (confirm('Delete this notebook?')) {
                    this.deleteNotebook(nb.id);
                }
            });

            container.appendChild(clone);
        });

        document.getElementById('notebookCount').textContent = this.notebooks.length;
    }

    async loadNotebookCounts(notebookId, element) {
        try {
            const [sources, notes] = await Promise.all([
                this.api(`/notebooks/${notebookId}/sources`),
                this.api(`/notebooks/${notebookId}/notes`)
            ]);

            element.querySelector('.notebook-sources').textContent = `${sources.length} sources`;
            element.querySelector('.notebook-notes').textContent = `${notes.length} notes`;
        } catch (error) {
            // Ignore errors for counts
        }
    }

    async selectNotebook(id) {
        this.currentNotebook = this.notebooks.find(nb => nb.id === id);

        document.querySelectorAll('.notebook-item').forEach(item => {
            item.classList.toggle('active', item.dataset.id === id);
        });

        await Promise.all([
            this.loadSources(),
            this.loadNotes(),
            this.loadChatSessions()
        ]);

        this.setStatus(`Selected: ${this.currentNotebook.name}`);
    }

    showNewNotebookModal() {
        document.getElementById('newNotebookModal').classList.add('active');
        document.getElementById('modalOverlay').classList.add('active');
        document.querySelector('#newNotebookForm input[name="name"]').focus();
    }

    async handleCreateNotebook(e) {
        e.preventDefault();
        const form = e.target;
        const data = new FormData(form);

        this.showLoading('Creating notebook...');

        try {
            const notebook = await this.api('/notebooks', {
                method: 'POST',
                body: JSON.stringify({
                    name: data.get('name'),
                    description: data.get('description') || undefined,
                }),
            });

            this.notebooks.push(notebook);
            this.renderNotebooks();
            this.selectNotebook(notebook.id);
            this.closeModals();
            form.reset();
            this.hideLoading();
        } catch (error) {
            this.hideLoading();
            this.showError(error.message);
        }
    }

    async deleteNotebook(id) {
        try {
            console.log('Deleting notebook:', id);
            await this.api(`/notebooks/${id}`, { method: 'DELETE' });
            console.log('Notebook deleted successfully');

            this.notebooks = this.notebooks.filter(nb => nb.id !== id);

            if (this.currentNotebook?.id === id) {
                this.currentNotebook = null;
                // Clear content areas
                this.clearContentAreas();
            }

            this.renderNotebooks();
            this.updateFooter();
            console.log('Notebooks after delete:', this.notebooks.length);
        } catch (error) {
            console.error('Delete error:', error);
            this.showError('Failed to delete notebook: ' + error.message);
        }
    }

    clearContentAreas() {
        // Clear sources
        const sourcesContainer = document.getElementById('sourcesGrid');
        sourcesContainer.innerHTML = `
            <div class="empty-state">
                <svg width="64" height="64" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1">
                    <path d="M20 8 L44 8 L48 12 L48 56 L20 56 Z"/>
                    <polyline points="44,8 44,12 48,12"/>
                    <line x1="28" y1="24" x2="40" y2="24"/>
                    <line x1="28" y1="32" x2="40" y2="32"/>
                    <line x1="28" y1="40" x2="36" y2="40"/>
                </svg>
                <p>Add sources to begin</p>
                <p class="empty-hint">PDF, TXT, MD, DOCX, HTML supported</p>
            </div>
        `;

        // Clear notes
        const notesContainer = document.getElementById('notesList');
        notesContainer.innerHTML = `
            <div class="empty-state">
                <svg width="48" height="48" viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M12 4 L36 4 L40 8 L40 44 L12 44 Z"/>
                    <polyline points="36,4 36,8 40,8"/>
                </svg>
                <p>No notes generated</p>
                <p class="empty-hint">Use transformations to create notes from sources</p>
            </div>
        `;

        // Clear chat
        const chatContainer = document.getElementById('chatMessages');
        chatContainer.innerHTML = `
            <div class="chat-welcome">
                <svg width="40" height="40" viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5">
                    <circle cx="20" cy="12" r="6"/>
                    <path d="M8 38 C8 28 14 22 20 22 C26 22 32 28 32 38"/>
                </svg>
                <h3>Chat with your sources</h3>
                <p>Ask questions about the content in your notebook</p>
            </div>
        `;

        this.currentChatSession = null;
    }

    // Source Methods
    async loadSources() {
        if (!this.currentNotebook) return;

        const container = document.getElementById('sourcesGrid');
        const template = document.getElementById('sourceTemplate');

        try {
            const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);

            if (sources.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <svg width="64" height="64" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1">
                            <path d="M20 8 L44 8 L48 12 L48 56 L20 56 Z"/>
                            <polyline points="44,8 44,12 48,12"/>
                            <line x1="28" y1="24" x2="40" y2="24"/>
                            <line x1="28" y1="32" x2="40" y2="32"/>
                            <line x1="28" y1="40" x2="36" y2="40"/>
                        </svg>
                        <p>Add sources to begin</p>
                        <p class="empty-hint">PDF, TXT, MD, DOCX, HTML supported</p>
                    </div>
                `;
                return;
            }

            container.innerHTML = '';

            sources.forEach(source => {
                const clone = template.content.cloneNode(true);
                const card = clone.querySelector('.source-card');

                card.dataset.id = source.id;
                card.querySelector('.source-type-badge').textContent = source.type;
                card.querySelector('.source-name').textContent = source.name;
                card.querySelector('.source-meta').textContent = this.formatFileSize(source.file_size) || 'Text source';
                card.querySelector('.chunk-count').textContent = source.chunk_count || 0;

                // Icon based on type
                const icon = this.getSourceIcon(source.type);
                card.querySelector('.source-icon').innerHTML = icon;

                card.querySelector('.btn-remove-source').addEventListener('click', () => {
                    this.removeSource(source.id);
                });

                container.appendChild(clone);
            });

            this.updateFooter();
        } catch (error) {
            console.error('Failed to load sources:', error);
        }
    }

    getSourceIcon(type) {
        const icons = {
            file: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 4 L24 4 L30 10 L30 36 L10 36 Z"/><polyline points="24,4 24,10 30,10"/></svg>',
            text: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M8 6 L32 6"/><path d="M8 12 L32 12"/><path d="M8 18 L28 18"/><path d="M8 24 L32 24"/><path d="M8 30 L24 30"/></svg>',
            url: '<svg viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 20 C12 14 16 10 22 10 C28 10 32 14 32 20 C32 26 28 30 22 30"/><path d="M28 20 C28 26 24 30 18 30 C12 30 8 26 8 20 C8 14 12 10 18 10"/></svg>',
        };
        return icons[type] || icons.file;
    }

    formatFileSize(bytes) {
        if (!bytes) return null;
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    }

    showAddSourceModal() {
        if (!this.currentNotebook) {
            this.showError('Please select a notebook first');
            return;
        }
        document.getElementById('addSourceModal').classList.add('active');
        document.getElementById('modalOverlay').classList.add('active');
    }

    async handleFileUpload(e) {
        const files = e.target.files;
        if (!files.length) return;

        this.showLoading('Uploading and processing...');

        for (const file of files) {
            const formData = new FormData();
            formData.append('file', file);
            formData.append('notebook_id', this.currentNotebook.id);

            try {
                await this.api('/upload', {
                    method: 'POST',
                    headers: {}, // Let browser set content-type for FormData
                    body: formData,
                });
            } catch (error) {
                this.showError(`Failed to upload ${file.name}`);
            }
        }

        this.hideLoading();
        this.closeModals();
        await this.loadSources();
        // Update notebook card counts in left panel
        await this.updateCurrentNotebookCounts();
        document.getElementById('fileInput').value = '';
    }

    async handleTextSource(e) {
        e.preventDefault();
        const form = e.target;
        const data = new FormData(form);

        this.showLoading('Adding source...');

        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/sources`, {
                method: 'POST',
                body: JSON.stringify({
                    name: data.get('name'),
                    type: 'text',
                    content: data.get('content'),
                }),
            });

            this.hideLoading();
            this.closeModals();
            form.reset();
            await this.loadSources();
            // Update notebook card counts in left panel
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.hideLoading();
            this.showError(error.message);
        }
    }

    async handleURLSource(e) {
        e.preventDefault();
        const form = e.target;
        const data = new FormData(form);

        this.showLoading('Fetching URL...');

        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/sources`, {
                method: 'POST',
                body: JSON.stringify({
                    name: data.get('name') || data.get('url'),
                    type: 'url',
                    url: data.get('url'),
                }),
            });

            this.hideLoading();
            this.closeModals();
            form.reset();
            await this.loadSources();
            // Update notebook card counts in left panel
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.hideLoading();
            this.showError(error.message);
        }
    }

    handleDrop(e) {
        e.preventDefault();
        document.getElementById('dropZone').classList.remove('drag-over');

        const files = e.dataTransfer.files;
        if (!files.length) return;

        document.getElementById('fileInput').files = files;
        this.handleFileUpload({ target: { files } });
    }

    async removeSource(id) {
        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/sources/${id}`, {
                method: 'DELETE',
            });
            await this.loadSources();
            // Update notebook card counts in left panel
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.showError('Failed to remove source');
        }
    }

    async updateCurrentNotebookCounts() {
        if (!this.currentNotebook) return;

        // Get fresh counts
        const [sources, notes] = await Promise.all([
            this.api(`/notebooks/${this.currentNotebook.id}/sources`),
            this.api(`/notebooks/${this.currentNotebook.id}/notes`)
        ]);

        // Find and update the notebook card in the left panel
        const notebookCard = document.querySelector(`.notebook-item[data-id="${this.currentNotebook.id}"]`);
        if (notebookCard) {
            notebookCard.querySelector('.notebook-sources').textContent = `${sources.length} sources`;
            notebookCard.querySelector('.notebook-notes').textContent = `${notes.length} notes`;
        }
    }

    // Note Methods
    async loadNotes() {
        if (!this.currentNotebook) return;

        const container = document.getElementById('notesList');
        const template = document.getElementById('noteTemplate');

        try {
            const notes = await this.api(`/notebooks/${this.currentNotebook.id}/notes`);

            if (notes.length === 0) {
                container.innerHTML = `
                    <div class="empty-state">
                        <svg width="48" height="48" viewBox="0 0 48 48" fill="none" stroke="currentColor" stroke-width="1.5">
                            <path d="M12 4 L36 4 L40 8 L40 44 L12 44 Z"/>
                            <polyline points="36,4 36,8 40,8"/>
                        </svg>
                        <p>No notes generated</p>
                        <p class="empty-hint">Use transformations to create notes from sources</p>
                    </div>
                `;
                return;
            }

            container.innerHTML = '';

            notes.forEach(note => {
                const clone = template.content.cloneNode(true);
                const item = clone.querySelector('.note-item');

                item.dataset.id = note.id;
                item.querySelector('.note-type-badge').textContent = note.type;
                item.querySelector('.note-title').textContent = note.title;

                // Strip markdown for preview - get plain text
                const plainText = note.content
                    .replace(/^#+\s+/gm, '') // Remove headers
                    .replace(/\*\*/g, '') // Remove bold
                    .replace(/\*/g, '') // Remove italic
                    .replace(/`/g, '') // Remove code
                    .replace(/\[([^\]]+)\]\([^)]+\)/g, '$1') // Replace links with text
                    .replace(/\n+/g, ' ') // Replace newlines with spaces
                    .trim();

                item.querySelector('.note-preview').textContent = plainText;
                item.querySelector('.note-date').textContent = this.formatDate(note.created_at);
                item.querySelector('.note-sources').textContent = `${note.source_ids?.length || 0} sources`;

                item.querySelector('.btn-delete-note').addEventListener('click', () => {
                    this.deleteNote(note.id);
                });

                // Expand on click
                item.addEventListener('click', (e) => {
                    if (!e.target.closest('.btn-delete-note')) {
                        this.viewNote(note);
                    }
                });

                container.appendChild(clone);
            });

            this.updateFooter();
        } catch (error) {
            console.error('Failed to load notes:', error);
        }
    }

    async viewNote(note) {
        // Render markdown content
        const renderedContent = marked.parse(note.content);

        const modal = document.createElement('div');
        modal.className = 'modal active';
        modal.style.cssText = 'max-width: 800px; max-height: 85vh;';
        modal.innerHTML = `
            <div class="modal-header">
                <h3>${note.title}</h3>
                <div class="modal-header-actions">
                    <button class="btn-copy-note" title="Copy Markdown">
                        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                            <rect x="3" y="3" width="10" height="10" rx="1"/>
                            <path d="M7 3 L7 1 C7 1 13 1 13 1 L13 13 L11 13"/>
                        </svg>
                    </button>
                    <button class="btn-close" onclick="this.closest('.modal').remove()">×</button>
                </div>
            </div>
            <div class="modal-body" style="max-height: calc(85vh - 120px); overflow-y: auto;">
                <div class="markdown-content">${renderedContent}</div>
            </div>
        `;

        // Add copy button functionality
        const copyBtn = modal.querySelector('.btn-copy-note');
        copyBtn.addEventListener('click', async () => {
            try {
                await navigator.clipboard.writeText(note.content);
                const originalHTML = copyBtn.innerHTML;
                copyBtn.innerHTML = `
                    <svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="2">
                        <polyline points="4,8 6,10 12,4"/>
                    </svg>
                `;
                copyBtn.classList.add('copied');
                setTimeout(() => {
                    copyBtn.innerHTML = originalHTML;
                    copyBtn.classList.remove('copied');
                }, 2000);
            } catch (err) {
                this.showError('Failed to copy');
            }
        });

        const overlay = document.getElementById('modalOverlay');
        overlay.classList.add('active');
        overlay.appendChild(modal);
    }

    async deleteNote(id) {
        try {
            await this.api(`/notebooks/${this.currentNotebook.id}/notes/${id}`, {
                method: 'DELETE',
            });
            await this.loadNotes();
            // Update notebook card counts in left panel
            await this.updateCurrentNotebookCounts();
        } catch (error) {
            this.showError('Failed to delete note');
        }
    }

    // Transform Methods
    async handleTransform(type, element) {
        if (!this.currentNotebook) {
            this.showError('Please select a notebook first');
            return;
        }

        const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);
        if (sources.length === 0) {
            this.showError('Please add sources first');
            return;
        }

        const customPrompt = document.getElementById('customPrompt').value;

        // Visual feedback on button
        let originalContent = '';
        let originalIcon = '';
        if (element) {
            if (element.classList.contains('transform-card')) {
                originalContent = element.innerHTML;
                originalIcon = element.querySelector('.transform-icon').innerHTML;
                element.classList.add('loading');
                element.disabled = true;
                element.querySelector('.transform-icon').innerHTML = `
                    <svg class="hourglass-icon" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M5 22h14"/>
                        <path d="M5 2h14"/>
                        <path d="M17 22v-4.172a2 2 0 0 0-.586-1.414L12 12l4.414-4.414A2 2 0 0 0 17 6.172V2"/>
                        <path d="M7 22v-4.172a2 2 0 0 1 .586-1.414L12 12l-4.414-4.414A2 2 0 0 1 7 6.172V2"/>
                    </svg>
                `;
                element.querySelector('.transform-name').textContent = 'Generating...';
            } else {
                element.textContent = 'Generating...';
            }
        }

        try {
            const result = await this.api(`/notebooks/${this.currentNotebook.id}/transform`, {
                method: 'POST',
                body: JSON.stringify({
                    type: type,
                    prompt: customPrompt || undefined,
                    length: 'medium',
                    format: 'markdown',
                    language: 'en',
                }),
            });

            if (element && element.classList.contains('transform-card')) {
                element.classList.remove('loading');
                element.disabled = false;
                element.querySelector('.transform-icon').innerHTML = originalIcon;
                element.querySelector('.transform-name').textContent = element.querySelector('.transform-name').textContent.replace('Generating...', '');
            }

            await this.loadNotes();
            // Update notebook card counts in left panel
            await this.updateCurrentNotebookCounts();
            this.switchTab('notes');
            document.getElementById('customPrompt').value = '';
            this.setStatus(`Generated ${type}`);
        } catch (error) {
            if (element && element.classList.contains('transform-card')) {
                element.classList.remove('loading');
                element.disabled = false;
                element.querySelector('.transform-icon').innerHTML = originalIcon;
                element.querySelector('.transform-name').textContent = element.querySelector('.transform-name').textContent.replace('Generating...', '');
            }
            this.showError(error.message);
        }
    }

    // Chat Methods
    async loadChatSessions() {
        if (!this.currentNotebook) return;

        try {
            const sessions = await this.api(`/notebooks/${this.currentNotebook.id}/chat/sessions`);

            // Reset chat view
            const container = document.getElementById('chatMessages');
            container.innerHTML = `
                <div class="chat-welcome">
                    <svg width="40" height="40" viewBox="0 0 40 40" fill="none" stroke="currentColor" stroke-width="1.5">
                        <circle cx="20" cy="12" r="6"/>
                        <path d="M8 38 C8 28 14 22 20 22 C26 22 32 28 32 38"/>
                    </svg>
                    <h3>Chat with your sources</h3>
                    <p>Ask questions about the content in your notebook</p>
                </div>
            `;

            this.currentChatSession = null;
        } catch (error) {
            console.error('Failed to load chat sessions:', error);
        }
    }

    async handleChat(e) {
        e.preventDefault();

        if (!this.currentNotebook) {
            this.showError('Please select a notebook first');
            return;
        }

        const input = document.getElementById('chatInput');
        const message = input.value.trim();

        if (!message) return;

        // Add user message
        this.addMessage('user', message);
        input.value = '';

        const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);
        if (sources.length === 0) {
            this.addMessage('assistant', 'Please add some sources to your notebook first.');
            return;
        }

        this.setStatus('Thinking...');

        try {
            const response = await this.api(`/notebooks/${this.currentNotebook.id}/chat`, {
                method: 'POST',
                body: JSON.stringify({
                    message: message,
                    session_id: this.currentChatSession || undefined,
                }),
            });

            this.addMessage('assistant', response.message, response.sources);
            this.currentChatSession = response.session_id;
            this.setStatus('Ready');
        } catch (error) {
            this.addMessage('assistant', `Error: ${error.message}`);
            this.setStatus('Error');
        }
    }

    addMessage(role, content, sources = []) {
        const container = document.getElementById('chatMessages');
        const template = document.getElementById('messageTemplate');

        // Remove welcome message
        const welcome = container.querySelector('.chat-welcome');
        if (welcome) welcome.remove();

        const clone = template.content.cloneNode(true);
        const message = clone.querySelector('.chat-message');

        message.dataset.role = role;

        // Render markdown for assistant messages
        const messageText = message.querySelector('.message-text');
        if (role === 'assistant') {
            messageText.innerHTML = marked.parse(content);
        } else {
            messageText.textContent = content;
        }

        // Add sources
        if (sources.length > 0) {
            const sourcesContainer = message.querySelector('.message-sources');
            sources.forEach(source => {
                const tag = document.createElement('span');
                tag.className = 'source-tag';
                tag.textContent = source.name || source.id;
                sourcesContainer.appendChild(tag);
            });
        }

        container.appendChild(clone);
        container.scrollTop = container.scrollHeight;
    }

    // UI Methods
    closeModals() {
        document.querySelectorAll('.modal').forEach(m => m.classList.remove('active'));
        document.getElementById('modalOverlay').classList.remove('active');
        this.hideLoading();
    }

    showLoading(text = 'Loading...') {
        document.getElementById('loadingText').textContent = text;
        document.getElementById('loadingOverlay').classList.add('active');
    }

    hideLoading() {
        document.getElementById('loadingOverlay').classList.remove('active');
    }

    setStatus(text) {
        document.getElementById('footerStatus').textContent = text;
    }

    switchTab(tab) {
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.toggle('active', btn.dataset.tab === tab);
        });
        document.querySelectorAll('.tab-content').forEach(content => {
            content.classList.toggle('active', content.id === `tab-${tab}`);
        });
    }

    showError(message) {
        this.setStatus(`Error: ${message}`);

        // Show toast
        const toast = document.createElement('div');
        toast.className = 'error-toast';
        toast.style.cssText = `
            position: fixed;
            bottom: 60px;
            right: 20px;
            padding: 12px 20px;
            background: var(--accent-red);
            color: white;
            font-family: var(--font-mono);
            font-size: 0.75rem;
            border-radius: 4px;
            box-shadow: var(--shadow-medium);
            animation: slideIn 0.3s ease;
            z-index: 3000;
        `;
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            toast.style.opacity = '0';
            setTimeout(() => toast.remove(), 300);
        }, 3000);
    }

    updateFooter() {
        const sourceCount = document.querySelectorAll('.source-card').length;
        const noteCount = document.querySelectorAll('.note-item').length;
        document.getElementById('footerStats').textContent = `${sourceCount} sources · ${noteCount} notes`;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diff = now - date;

        if (diff < 60000) return 'Just now';
        if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`;
        if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`;

        return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    window.app = new OpenNotebook();
});
