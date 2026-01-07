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
    const chatContainer = document.getElementById('chatContainer');
    const resizer = document.getElementById('resizer');
    const collapseBtn = document.getElementById('collapseBtn');

    // Track which node types have been shown (only show each node type once)
    const shownNodes = new Set();

    // Node types to exclude from left panel (these appear in loops for each section)
    const excludedNodes = ['内容写作节点', '反思节点', '补充节点'];

    // Collapse functionality
    const collapsedIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="collapse-icon"><rect width="18" height="18" x="3" y="3" rx="2"></rect><path d="M15 3v18"></path></svg>`;
    const expandedIcon = `<svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="collapse-icon"><rect width="18" height="18" x="3" y="3" rx="2"></rect><path d="M9 3v18"></path></svg>`;

    // Store the width before collapsing
    let savedWidth = '40%';

    collapseBtn.addEventListener('click', () => {
        const isCurrentlyCollapsed = chatContainer.classList.contains('collapsed');

        if (!isCurrentlyCollapsed) {
            // About to collapse - save current width
            const currentWidth = chatContainer.style.width;
            if (currentWidth && currentWidth !== '50px') {
                savedWidth = currentWidth;
            }
        }

        chatContainer.classList.toggle('collapsed');
        const isCollapsed = chatContainer.classList.contains('collapsed');
        collapseBtn.title = isCollapsed ? '展开' : '收起';
        collapseBtn.innerHTML = isCollapsed ? collapsedIcon : expandedIcon;

        // When expanding, restore the saved width
        if (!isCollapsed) {
            chatContainer.style.width = savedWidth;
        }
    });

    // Resizer functionality
    let isResizing = false;

    resizer.addEventListener('mousedown', (e) => {
        isResizing = true;
        resizer.classList.add('active');
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';
    });

    document.addEventListener('mousemove', (e) => {
        if (!isResizing) return;

        const containerRect = document.querySelector('.app-main').getBoundingClientRect();
        const newWidth = ((e.clientX - containerRect.left) / containerRect.width) * 100;

        if (newWidth >= 20 && newWidth <= 80) {
            chatContainer.style.width = newWidth + '%';
        }
    });

    document.addEventListener('mouseup', () => {
        if (isResizing) {
            isResizing = false;
            resizer.classList.remove('active');
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        }
    });

    // Load history panel
    async function loadHistoryPanel() {
        const historyPanelContent = document.getElementById('historyPanelContent');
        try {
            const res = await fetch('/api/history');
            const history = await res.json();

            if (!history || history.length === 0) {
                historyPanelContent.innerHTML = '<div class="placeholder-text">历史会话为空</div>';
                return;
            }

            historyPanelContent.innerHTML = '<div class="history-grid"></div>';
            const grid = historyPanelContent.querySelector('.history-grid');

            history.forEach(item => {
                const card = document.createElement('div');
                card.className = 'history-card';

                const date = new Date(item.timestamp);
                const dateStr = date.toLocaleDateString('zh-CN', {
                    year: 'numeric',
                    month: '2-digit',
                    day: '2-digit'
                });
                const timeStr = date.toLocaleTimeString('zh-CN', {
                    hour: '2-digit',
                    minute: '2-digit'
                });

                card.innerHTML = `
                    <div class="history-card-title">${escapeHtml(item.query)}</div>
                    <div class="history-card-meta">
                        <span>
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
                                <line x1="16" y1="2" x2="16" y2="6"></line>
                                <line x1="8" y1="2" x2="8" y2="6"></line>
                                <line x1="3" y1="10" x2="21" y2="10"></line>
                            </svg>
                            ${dateStr}
                        </span>
                        <span>
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <circle cx="12" cy="12" r="10"></circle>
                                <polyline points="12 6 12 12 16 14"></polyline>
                            </svg>
                            ${timeStr}
                        </span>
                    </div>
                `;

                card.addEventListener('click', () => {
                    // Clear messages except the first one (system welcome message)
                    const firstMessage = messagesContainer.firstElementChild;
                    messagesContainer.innerHTML = '';
                    if (firstMessage) {
                        messagesContainer.appendChild(firstMessage);
                    }

                    queryInput.value = item.query;
                    queryInput.style.height = 'auto';
                    queryInput.style.height = (queryInput.scrollHeight) + 'px';
                    sendBtn.disabled = false;
                    switchTab('activities');
                    handleSearch();
                });

                grid.appendChild(card);
            });
        } catch (err) {
            console.error('Failed to load history panel:', err);
            historyPanelContent.innerHTML = '<div class="placeholder-text">加载历史记录失败</div>';
        }
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
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
        } else if (tabId === 'history') {
            document.getElementById('historyTab').classList.add('active');
            loadHistoryPanel();
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

    // Handle new chat
    document.getElementById('newChatBtn').addEventListener('click', () => {
        // Clear messages except the first one (system welcome message)
        const firstMessage = messagesContainer.firstElementChild;
        messagesContainer.innerHTML = '';
        if (firstMessage) {
            messagesContainer.appendChild(firstMessage);
        }

        // Clear input
        queryInput.value = '';
        queryInput.style.height = 'auto';
        sendBtn.disabled = true;

        // Clear report and logs
        reportContent.innerHTML = '<div class="placeholder-text">研究结果将显示在这里...</div>';
        logsContainer.innerHTML = '';
        shownNodes.clear();

        // Reset status
        setStatus('空闲', false);

        // Reset podcast tab
        const podcastTabBtn = document.getElementById('podcastTabBtn');
        const podcastContent = document.getElementById('podcastContent');
        if (podcastTabBtn) podcastTabBtn.style.display = 'none';
        if (podcastContent) podcastContent.innerHTML = '<div class="placeholder-text">播客脚本将显示在这里...</div>';

        // Hide share button
        const shareBtn = document.getElementById('shareBtn');
        if (shareBtn) {
            shareBtn.style.display = 'none';
            delete shareBtn.dataset.reportHtml;
        }

        // Switch to report tab
        switchTab('report');
    });

    // Handle share button
    document.getElementById('shareBtn').addEventListener('click', async () => {
        const shareBtn = document.getElementById('shareBtn');
        const reportHtml = shareBtn.dataset.reportHtml;

        if (!reportHtml) {
            alert('没有可分享的报告');
            return;
        }

        // Get the query from the last user message or use a default
        const userMessages = messagesContainer.querySelectorAll('.message.user');
        const queryText = userMessages.length > 0
            ? userMessages[userMessages.length - 1].querySelector('.content p').textContent
            : '研究报告';

        try {
            // Call the share API to create a shareable link
            const response = await fetch('/api/share', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    html: reportHtml,
                    query: queryText
                })
            });

            if (!response.ok) {
                throw new Error('分享失败');
            }

            const data = await response.json();

            // Open the share URL in a new tab
            window.open(data.url, '_blank');

        } catch (error) {
            console.error('Share error:', error);
            alert('分享失败，请稍后重试');
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
        shownNodes.clear(); // Clear shown nodes for new query

        try {
            // Start SSE connection
            const eventSource = new EventSource(`/api/run?query=${encodeURIComponent(query)}`);

            eventSource.onmessage = async (event) => {
                const data = JSON.parse(event.data);

                if (data.type === 'update') {
                    // Update status or partial content
                    if (data.step) {
                        setStatus(data.step, true);
                        // 同时在左边面板显示步骤进度（仅关键节点日志，每个节点类型只显示一次）
                        if (data.step.includes('---')) {
                            // Extract node type (e.g., "查询分析节点" from "--- 查询分析节点：...")
                            const nodeMatch = data.step.match(/---\s*(.+?)节点/);
                            if (nodeMatch) {
                                const nodeType = nodeMatch[1] + '节点';
                                // Skip if this node type is excluded
                                if (excludedNodes.includes(nodeType)) {
                                    return;
                                }
                                if (!shownNodes.has(nodeType)) {
                                    shownNodes.add(nodeType);
                                    addProgressMessage(data.step);
                                }
                            } else {
                                // Fallback: show if no node type pattern found
                                addProgressMessage(data.step);
                            }
                        }
                    }
                    if (data.log) {
                        // Optional: Add logs to a console or debug view
                        console.log(data.log);
                    }
                } else if (data.type === 'log') {
                    // Append log to activities panel
                    const logEntry = document.createElement('div');
                    logEntry.className = 'log-entry';
                    logEntry.textContent = data.message;
                    logsContainer.appendChild(logEntry);
                    logsContainer.scrollTop = logsContainer.scrollHeight;

                    // 同时在左边面板显示关键日志（只显示节点开始日志，每个节点类型只显示一次）
                    if (data.message.includes('---')) {
                        const nodeMatch = data.message.match(/---\s*(.+?)节点/);
                        if (nodeMatch) {
                            const nodeType = nodeMatch[1] + '节点';
                            // Skip if this node type is excluded
                            if (excludedNodes.includes(nodeType)) {
                                return;
                            }
                            if (!shownNodes.has(nodeType)) {
                                shownNodes.add(nodeType);
                                addProgressMessage(data.message);
                            }
                        } else {
                            // Fallback: show if no node type pattern found
                            addProgressMessage(data.message);
                        }
                    }
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

                    // Show share button when report is ready
                    const shareBtn = document.getElementById('shareBtn');
                    if (shareBtn) {
                        shareBtn.style.display = 'flex';
                        // Store report HTML for sharing
                        shareBtn.dataset.reportHtml = report;
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

    // 添加进度消息到左边面板
    function addProgressMessage(text) {
        const msgDiv = document.createElement('div');
        msgDiv.className = 'message system';

        const avatarSvg = `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"></path></svg>`;

        // Remove all decorative dashes from the message (both --- and ---)
        let cleanText = text
            .replace(/^---+\s*/, '')     // Remove leading "---" or "--- "
            .replace(/---+$/, '')        // Remove trailing "---"
            .replace(/---+\s*$/, '')     // Remove trailing "--- " or "---"
            .trim();

        msgDiv.innerHTML = `
            <div class="avatar">${avatarSvg}</div>
            <div class="content"><p>${cleanText}</p></div>
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

