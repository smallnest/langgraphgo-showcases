// ============================================
// OPEN NOTEBOOK - Application Logic
// ============================================

const translations = {
    en: {
        newNotebook: "New Notebook",
        notebooks: "NOTEBOOKS",
        noNotebooks: "No notebooks yet",
        createFirstNotebook: "Create your first notebook",
        sources: "SOURCES",
        addSourcesBegin: "Add sources to begin",
        supportedFormats: "PDF, TXT, MD, DOCX, HTML supported",
        notes: "NOTES",
        chat: "CHAT",
        transform: "TRANSFORM",
        noNotes: "No notes generated",
        useTransformations: "Use transformations to create notes from sources",
        chatWithSources: "Chat with your sources",
        askQuestions: "Ask questions about the content in your notebook",
        askQuestionPlaceholder: "Ask a question...",
        summary: "Summary",
        faq: "FAQ",
        studyGuide: "Study Guide",
        outline: "Outline",
        podcast: "Podcast",
        timeline: "Timeline",
        glossary: "Glossary",
        quiz: "Quiz",
        customTransformation: "Custom Transformation",
        describeGeneration: "Describe what you want to generate...",
        generate: "Generate",
        ready: "Ready",
        processing: "Processing...",
        name: "Name",
        researchNotes: "Research Notes",
        descriptionOptional: "Description (optional)",
        briefDescription: "A brief description...",
        cancel: "Cancel",
        createNotebook: "Create Notebook",
        addSource: "Add Source",
        uploadFile: "Upload File",
        pasteText: "Paste Text",
        url: "URL",
        dropFiles: "Drop files here or click to browse",
        noteTitle: "Note Title",
        content: "Content",
        pasteContent: "Paste or type your content here...",
        nameOptional: "Name (optional)",
        articleTitle: "Article Title",
        sourcesStats: "sources",
        notesStats: "notes",
        chunks: "chunks",
        generating: "Generating...",
        thinking: "Thinking...",
        error: "Error",
        copied: "Copied!",
        failedCopy: "Failed to copy",
        deleteNotebookConfirm: "Delete this notebook?",
        failedLoadNotebooks: "Failed to load notebooks",
        failedLoadSources: "Failed to load sources",
        failedLoadNotes: "Failed to load notes",
        failedLoadChat: "Failed to load chat sessions",
        pleaseSelectNotebook: "Please select a notebook first",
        pleaseAddSources: "Please add sources first",
        sourceText: "Text source"
    },
    zh: {
        newNotebook: "新建笔记本",
        notebooks: "笔记本",
        noNotebooks: "暂无笔记本",
        createFirstNotebook: "创建你的第一个笔记本",
        sources: "来源",
        addSourcesBegin: "添加来源",
        supportedFormats: "支持 PDF, TXT, MD, DOCX, HTML",
        notes: "笔记",
        chat: "对话",
        transform: "转换",
        noNotes: "暂无笔记",
        useTransformations: "使用转换从来源生成笔记",
        chatWithSources: "与来源对话",
        askQuestions: "询问关于笔记本内容的问题",
        askQuestionPlaceholder: "输入问题...",
        summary: "摘要",
        faq: "常见问题",
        studyGuide: "学习指南",
        outline: "大纲",
        podcast: "播客",
        timeline: "时间线",
        glossary: "术语表",
        quiz: "测验",
        customTransformation: "自定义转换",
        describeGeneration: "描述你想生成的内容...",
        generate: "生成",
        ready: "就绪",
        processing: "处理中...",
        name: "名称",
        researchNotes: "研究笔记",
        descriptionOptional: "描述 (可选)",
        briefDescription: "简要描述...",
        cancel: "取消",
        createNotebook: "创建笔记本",
        addSource: "添加来源",
        uploadFile: "上传文件",
        pasteText: "粘贴文本",
        url: "网址",
        dropFiles: "拖放文件到此处或点击浏览",
        noteTitle: "笔记标题",
        content: "内容",
        pasteContent: "在此粘贴或输入内容...",
        nameOptional: "名称 (可选)",
        articleTitle: "文章标题",
        sourcesStats: "来源",
        notesStats: "笔记",
        chunks: "块",
        generating: "生成中...",
        thinking: "思考中...",
        error: "错误",
        copied: "已复制!",
        failedCopy: "复制失败",
        deleteNotebookConfirm: "删除此笔记本？",
        failedLoadNotebooks: "加载笔记本失败",
        failedLoadSources: "加载来源失败",
        failedLoadNotes: "加载笔记失败",
        failedLoadChat: "加载对话失败",
        pleaseSelectNotebook: "请先选择一个笔记本",
        pleaseAddSources: "请先添加来源",
        sourceText: "文本来源"
    }
};

class OpenNotebook {
    constructor() {
        this.notebooks = [];
        this.currentNotebook = null;
        this.apiBase = '/api';
        this.currentChatSession = null;
        this.language = localStorage.getItem('language') || 'zh';

        this.init();
    }

    async init() {
        this.updateLanguage();
        this.bindEvents();
        await this.loadNotebooks();

        // Auto-select first notebook if available
        if (this.notebooks.length > 0) {
            this.selectNotebook(this.notebooks[0].id);
        }
    }

    bindEvents() {
        // Language toggle
        document.getElementById('btnLangToggle').addEventListener('click', () => {
            this.language = this.language === 'en' ? 'zh' : 'en';
            localStorage.setItem('language', this.language);
            this.updateLanguage();
            // Refresh content if needed
            if (this.currentNotebook) {
                this.renderNotebooks(); // Re-render to update counts text if strictly needed, mostly static
                this.loadSources(); // Re-render sources to update "chunks" text
                this.loadNotes();   // Re-render notes to update "sources" text
                this.updateFooter();
            }
        });

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

    updateLanguage() {
        const t = translations[this.language];

        // Update toggle button text to show what it will switch to
        document.getElementById('btnLangToggle').textContent = this.language === 'en' ? '中' : 'EN';

        // Update elements with data-i18n attribute
        document.querySelectorAll('[data-i18n]').forEach(el => {
            const key = el.getAttribute('data-i18n');
            if (t[key]) {
                el.textContent = t[key];
            }
        });

        // Update placeholders
        document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
            const key = el.getAttribute('data-i18n-placeholder');
            if (t[key]) {
                el.placeholder = t[key];
            }
        });
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
            this.showError(translations[this.language].failedLoadNotebooks);
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
                    <p>${translations[this.language].noNotebooks}</p>
                    <button id="btnCreateFirst" class="btn-primary">${translations[this.language].createFirstNotebook}</button>
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
                if (confirm(translations[this.language].deleteNotebookConfirm)) {
                    this.deleteNotebook(nb.id);
                }
            });

            container.appendChild(clone);
        });

        document.getElementById('notebookCount').textContent = this.notebooks.length;
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
            notebookCard.querySelector('.notebook-sources').textContent = `${sources.length} ${translations[this.language].sourcesStats}`;
            notebookCard.querySelector('.notebook-notes').textContent = `${notes.length} ${translations[this.language].notesStats}`;
        }
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
                        <p>${translations[this.language].addSourcesBegin}</p>
                        <p class="empty-hint">${translations[this.language].supportedFormats}</p>
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
                card.querySelector('.source-meta').textContent = this.formatFileSize(source.file_size) || translations[this.language].sourceText;
                card.querySelector('.chunk-count').textContent = source.chunk_count || 0;
                // Update "chunks" label manually as it is next to chunk-count
                card.querySelector('.chunk-count').nextElementSibling.textContent = ` ${translations[this.language].chunks}`;

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
            this.showError(translations[this.language].pleaseSelectNotebook);
            return;
        }
        document.getElementById('addSourceModal').classList.add('active');
        document.getElementById('modalOverlay').classList.add('active');
    }

    async handleFileUpload(e) {
        const files = e.target.files;
        if (!files.length) return;

        this.showLoading(translations[this.language].processing);

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
                this.showError(`${translations[this.language].error}: ${file.name}`);
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

        this.showLoading(translations[this.language].processing);

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

        this.showLoading(translations[this.language].processing);

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
            this.showError(translations[this.language].error);
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
                        <p>${translations[this.language].noNotes}</p>
                        <p class="empty-hint">${translations[this.language].useTransformations}</p>
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
                item.querySelector('.note-sources').textContent = `${note.source_ids?.length || 0} ${translations[this.language].sourcesStats}`;

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
                    <button class="btn-close">×</button>
                </div>
            </div>
            <div class="modal-body" style="max-height: calc(85vh - 120px); overflow-y: auto;">
                <div class="markdown-content">${renderedContent}</div>
            </div>
        `;

        // Add close button functionality
        const closeBtn = modal.querySelector('.btn-close');
        closeBtn.addEventListener('click', () => {
            const overlay = document.getElementById('modalOverlay');
            overlay.classList.remove('active');
            overlay.innerHTML = '';
        });

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
                this.setStatus(translations[this.language].copied);
            } catch (err) {
                this.showError(translations[this.language].failedCopy);
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
            this.showError(translations[this.language].error);
        }
    }

    // Transform Methods
    async handleTransform(type, element) {
        if (!this.currentNotebook) {
            this.showError(translations[this.language].pleaseSelectNotebook);
            return;
        }

        const sources = await this.api(`/notebooks/${this.currentNotebook.id}/sources`);
        if (sources.length === 0) {
            this.showError(translations[this.language].pleaseAddSources);
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
                element.querySelector('.transform-name').textContent = translations[this.language].generating;
            } else {
                element.textContent = translations[this.language].generating;
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
                    language: this.language,
                }),
            });

            if (element && element.classList.contains('transform-card')) {
                element.classList.remove('loading');
                element.disabled = false;
                element.querySelector('.transform-icon').innerHTML = originalIcon;
                element.querySelector('.transform-name').textContent = element.querySelector('.transform-name').textContent.replace(translations[this.language].generating, '');
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
                element.querySelector('.transform-name').textContent = element.querySelector('.transform-name').textContent.replace(translations[this.language].generating, '');
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
                    <h3>${translations[this.language].chatWithSources}</h3>
                    <p>${translations[this.language].askQuestions}</p>
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
            this.showError(translations[this.language].pleaseSelectNotebook);
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
            this.addMessage('assistant', translations[this.language].pleaseAddSources);
            return;
        }

        this.setStatus(translations[this.language].thinking);

        try {
            const response = await this.api(`/notebooks/${this.currentNotebook.id}/chat`, {
                method: 'POST',
                body: JSON.stringify({
                    message: message,
                    session_id: this.currentChatSession || undefined,
                    language: this.language,
                }),
            });

            this.addMessage('assistant', response.message, response.sources);
            this.currentChatSession = response.session_id;
            this.setStatus(translations[this.language].ready);
        } catch (error) {
            this.addMessage('assistant', `${translations[this.language].error}: ${error.message}`);
            this.setStatus(translations[this.language].error);
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

    showLoading(text) {
        document.getElementById('loadingText').textContent = text || translations[this.language].processing;
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
        this.setStatus(`${translations[this.language].error}: ${message}`);

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
        document.getElementById('footerStats').textContent = `${sourceCount} ${translations[this.language].sourcesStats} · ${noteCount} ${translations[this.language].notesStats}`;
    }

    formatDate(dateString) {
        const date = new Date(dateString);
        const now = new Date();
        const diff = now - date;

        if (this.language === 'zh') {
            if (diff < 60000) return '刚刚';
            if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`;
            if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`;
            return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' });
        }

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
