document.addEventListener('DOMContentLoaded', () => {
    const queryInput = document.getElementById('queryInput');
    const sendBtn = document.getElementById('sendBtn');
    const messagesContainer = document.getElementById('messages');
    const reportTab = document.getElementById('reportTab');
    const reportContent = document.getElementById('reportContent');
    const logsContainer = document.getElementById('logsContainer');
    const statusIndicator = document.getElementById('statusIndicator');
    const statusText = statusIndicator.querySelector('.text');
    const tabBtns = document.querySelectorAll('.tab-btn');
    const tabContents = document.querySelectorAll('.tab-content');
    const historyBtn = document.getElementById('historyBtn');
    const historyModal = document.getElementById('historyModal');
    const closeHistoryBtn = document.getElementById('closeHistoryBtn');
    const historyList = document.getElementById('historyList');

    // History Modal
    historyBtn.addEventListener('click', () => {
        loadHistory();
        historyModal.classList.add('active');
    });

    closeHistoryBtn.addEventListener('click', () => {
        historyModal.classList.remove('active');
    });

    historyModal.addEventListener('click', (e) => {
        if (e.target === historyModal) {
            historyModal.classList.remove('active');
        }
    });

    async function loadHistory() {
        try {
            const res = await fetch('/api/history');
            const history = await res.json();

            historyList.innerHTML = '';
            if (!history || history.length === 0) {
                historyList.innerHTML = '<div class="placeholder-text">历史请求为空</div>';
                return;
            }

            history.forEach(item => {
                const el = document.createElement('div');
                el.className = 'history-item';
                const date = new Date(item.timestamp).toLocaleString();
                el.innerHTML = `
                    <div class="history-query">${item.query}</div>
                    <div class="history-date">${date}</div>
                `;
                el.addEventListener('click', () => {
                    queryInput.value = item.query;
                    queryInput.style.height = 'auto';
                    queryInput.style.height = (queryInput.scrollHeight) + 'px';
                    sendBtn.disabled = false;
                    historyModal.classList.remove('active');
                    handleSearch();
                });
                historyList.appendChild(el);
            });
        } catch (err) {
            console.error('Failed to load history:', err);
            historyList.innerHTML = '<div class="placeholder-text">加载历史记录失败</div>';
        }
    }

    // Tab switching helper
    function switchTab(tabId) {
        // Update buttons
        tabBtns.forEach(b => {
            if (b.dataset.tab === tabId) b.classList.add('active');
            else b.classList.remove('active');
        });

        // Update content
        tabContents.forEach(c => c.classList.remove('active'));
        if (tabId === 'report') {
            reportTab.classList.add('active');
        } else if (tabId === 'podcast') {
            document.getElementById('podcastTab').classList.add('active');
        } else {
            document.getElementById('activitiesContent').classList.add('active');
        }
    }

    // Tab click handlers
    tabBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            switchTab(btn.dataset.tab);
        });
    });

    // Auto-resize textarea
    queryInput.addEventListener('input', function () {
        this.style.height = 'auto';
        this.style.height = (this.scrollHeight) + 'px';
        sendBtn.disabled = this.value.trim() === '';
    });

    // Handle send
    sendBtn.addEventListener('click', handleSearch);
    queryInput.addEventListener('keydown', (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            if (!sendBtn.disabled) handleSearch();
        }
    });

    async function handleSearch() {
        const query = queryInput.value.trim();
        if (!query) return;

        // Add user message
        addMessage(query, 'user');
        queryInput.value = '';
        queryInput.style.height = 'auto';
        sendBtn.disabled = true;

        // Set status
        setStatus('正在研究...', true);
        switchTab('activities'); // Switch to Activities tab
        reportContent.innerHTML = '<div class="placeholder-text">正在初始化研究代理...</div>';
        logsContainer.innerHTML = ''; // Clear previous logs

        try {
            // Start SSE connection
            const eventSource = new EventSource(`/api/run?query=${encodeURIComponent(query)}`);

            eventSource.onmessage = async (event) => {
                const data = JSON.parse(event.data);

                if (data.type === 'update') {
                    // Update status or partial content
                    if (data.step) {
                        setStatus(data.step, true);
                    }
                    if (data.log) {
                        // Optional: Add logs to a console or debug view
                        console.log(data.log);
                    }
                } else if (data.type === 'log') {
                    // Append log
                    const logEntry = document.createElement('div');
                    logEntry.className = 'log-entry';
                    logEntry.textContent = data.message;
                    logsContainer.appendChild(logEntry);
                    logsContainer.scrollTop = logsContainer.scrollHeight;
                } else if (data.type === 'result') {
                    // Final report
                    const report = data.report;
                    reportContent.innerHTML = report;

                    // Handle podcast
                    const podcastTabBtn = document.getElementById('podcastTabBtn');
                    const podcastContent = document.getElementById('podcastContent');

                    if (data.podcast_script) {
                        if (podcastTabBtn && podcastContent) {
                            podcastTabBtn.style.display = 'block'; // Show tab
                            // Convert newlines to <br> or wrap in <p>
                            const formattedScript = data.podcast_script.replace(/\n/g, '<br>');
                            podcastContent.innerHTML = `<div class="podcast-script" style="line-height: 1.6; font-size: 1.1em;">${formattedScript}</div>`;
                        }
                    } else {
                        // Hide if not present (for subsequent runs)
                        if (podcastTabBtn) podcastTabBtn.style.display = 'none';
                    }

                    renderMath();
                    highlightCode();
                    setStatus('已完成', false);
                    switchTab('report'); // Switch back to Report tab
                    // 处理 Mermaid 图表 - 添加这行
                    if (typeof window.processMermaidBlocks === 'function') {
                        setTimeout(window.processMermaidBlocks, 100);
                    }
                    eventSource.close();
                } else if (data.type === 'error') {
                    addMessage(`错误：${data.message}`, 'system');
                    setStatus('错误', false);
                    eventSource.close();
                }
            };

            eventSource.onerror = (err) => {
                console.error('EventSource failed:', err);
                setStatus('连接丢失', false);
                eventSource.close();
            };

        } catch (error) {
            console.error('Error:', error);
            addMessage('启动研究失败。', 'system');
            setStatus('错误', false);
        }
    }

    function addMessage(text, type) {
        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${type}`;

        let avatarSvg = '';
        if (type === 'user') {
            avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>`;
        } else {
            avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>`;
        }

        msgDiv.innerHTML = `
            <div class="avatar">${avatarSvg}</div>
            <div class="content"><p>${text}</p></div>
        `;
        messagesContainer.appendChild(msgDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }

    function setStatus(text, active) {
        statusText.textContent = text;
        if (active) {
            statusIndicator.classList.add('active');
        } else {
            statusIndicator.classList.remove('active');
        }
    }

    function renderMath() {
        if (!window.katex) return;

        // 1. Handle explicit code blocks (math, latex, tex)
        const mathBlocks = reportContent.querySelectorAll('code.language-math, code.language-latex, code.language-tex');
        mathBlocks.forEach(block => {
            const latex = block.textContent;
            const span = document.createElement('div');
            span.className = 'math-display';
            span.style.textAlign = 'center';
            span.style.margin = '1em 0';
            try {
                katex.render(latex, span, { displayMode: true, throwOnError: false });
                // Replace the parent <pre> if it exists, or just the code block
                if (block.parentElement && block.parentElement.tagName === 'PRE') {
                    block.parentElement.replaceWith(span);
                } else {
                    block.replaceWith(span);
                }
            } catch (e) {
                console.error('KaTeX error:', e);
            }
        });

        // 2. Auto-render inline and block math in the rest of the text
        if (window.renderMathInElement) {
            renderMathInElement(reportContent, {
                delimiters: [
                    { left: '$$', right: '$$', display: true },
                    { left: '\\[', right: '\\]', display: true },
                    { left: '$', right: '$', display: false },
                    { left: '\\(', right: '\\)', display: false }
                ],
                throwOnError: false
            });
        }
    }

    function highlightCode() {
        if (!window.hljs) return;
        reportContent.querySelectorAll('pre code').forEach((block) => {
            hljs.highlightElement(block);
        });
    }

    // Export Podcast JSON
    window.exportPodcastJson = function () {
        const jsonDiv = document.getElementById('podcastJsonData');
        if (!jsonDiv) {
            alert('无法找到播客数据');
            return;
        }
        try {
            const jsonText = jsonDiv.textContent;
            // Validate JSON
            JSON.parse(jsonText);

            const blob = new Blob([jsonText], { type: 'application/json' });
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = 'podcast_script.json';
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (e) {
            console.error('Export failed:', e);
            alert('导出失败：数据格式错误');
        }
    };
});

