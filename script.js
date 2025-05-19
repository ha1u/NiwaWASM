// script.js
document.addEventListener('DOMContentLoaded', async () => {
    const go = new Go();
    let wasmInstance;

    try {
        const response = await fetch('main.wasm');
        if (!response.ok) {
            throw new Error(`Failed to fetch wasm: ${response.status} ${response.statusText}`);
        }
        const result = await WebAssembly.instantiateStreaming(response, go.importObject);
        wasmInstance = result.instance;
        go.run(wasmInstance); 
        console.log("WASM Loaded and Go main started.");
    } catch (err) {
        console.error('WASM Loading Error:', err);
        document.body.innerHTML = `<div style="padding: 20px; text-align: center; color: red;">アプリケーションの読み込みに失敗しました: ${err.message}<br>コンソールを確認してください。</div>`;
        return;
    }

    // DOM Elements
    const themeToggleIcon = document.getElementById('themeToggleIcon');
    const viewTitle = document.getElementById('viewTitle');

    const navRecordIcon = document.getElementById('navRecordIcon');
    const navSearchIcon = document.getElementById('navSearchIcon');
    const navDataIcon = document.getElementById('navDataIcon');
    const navIcons = [navRecordIcon, navSearchIcon, navDataIcon];

    const recordView = document.getElementById('recordView');
    const searchView = document.getElementById('searchView');
    const dataView = document.getElementById('dataView');
    const views = { record: recordView, search: searchView, data: dataView };
    const viewTitles = { record: '記録ビュー', search: '検索ビュー', data: 'データ管理ビュー' };

    const memoContent = document.getElementById('memoContent');
    const saveMemoButton = document.getElementById('saveMemoButton');
    const recordDisplayArea = document.getElementById('recordDisplayArea');
    const recordViewListTitle = document.getElementById('recordViewListTitle');

    const searchInput = document.getElementById('searchInput');
    const searchButton = document.getElementById('searchButton');
    const clearSearchButton = document.getElementById('clearSearchButton');
    const searchResultsArea = document.getElementById('searchResultsArea');

    const exportDataButton = document.getElementById('exportDataButton');
    const importFile = document.getElementById('importFile');

    let currentView = 'record';
    let activeKebabMenu = null;

    // --- Theme Management ---
    function applyTheme(theme) {
        document.body.classList.toggle('dark-mode', theme === 'dark');
        themeToggleIcon.textContent = theme === 'dark' ? '☀️' : '🌙';
        themeToggleIcon.title = theme === 'dark' ? 'ライトテーマに切り替え' : 'ダークテーマに切り替え';
    }

    function loadTheme() {
        const savedTheme = localStorage.getItem('niwaWASM_theme');
        if (savedTheme) {
            applyTheme(savedTheme);
        } else {
            const prefersDark = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
            applyTheme(prefersDark ? 'dark' : 'light');
        }
    }

    themeToggleIcon.addEventListener('click', () => {
        const newTheme = document.body.classList.contains('dark-mode') ? 'light' : 'dark';
        applyTheme(newTheme);
        localStorage.setItem('niwaWASM_theme', newTheme);
    });
    loadTheme();

    // --- Navigation ---
    function showView(viewName) {
        if (!views[viewName] || !viewTitles[viewName]) return;
        currentView = viewName;
        viewTitle.textContent = viewTitles[viewName];

        Object.values(views).forEach(view => view.classList.remove('active-view'));
        views[viewName].classList.add('active-view');

        navIcons.forEach(icon => icon.classList.remove('active'));
        const activeNavIcon = document.getElementById(`nav${viewName.charAt(0).toUpperCase() + viewName.slice(1)}Icon`);
        if(activeNavIcon) activeNavIcon.classList.add('active');
        
        if (viewName === 'search') searchInput.focus();
        closeAllKebabMenus();
    }

    navRecordIcon.addEventListener('click', () => showView('record'));
    navSearchIcon.addEventListener('click', () => showView('search'));
    navDataIcon.addEventListener('click', () => showView('data'));

    // --- URL Auto-linking ---
    function autoLinkText(text) {
        const urlPattern = /(\b(https?|ftp|file):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])|(\bwww\.[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])/ig;
        return text.replace(urlPattern, (url) => {
            const properUrl = url.startsWith('www.') ? 'http://' + url : url;
            // Escape HTML in URL text content to prevent XSS if URL itself contains HTML characters
            const escapedUrlText = url.replace(/</g, "&lt;").replace(/>/g, "&gt;");
            return `<a href="${properUrl}" target="_blank" rel="noopener noreferrer">${escapedUrlText}</a>`;
        });
    }
    
    // --- Kebab Menu ---
    function closeAllKebabMenus() {
        document.querySelectorAll('.kebab-menu').forEach(menu => menu.style.display = 'none');
        activeKebabMenu = null;
    }

    function createKebabMenu(memo) {
        const menuDiv = document.createElement('div');
        menuDiv.className = 'kebab-menu';
        menuDiv.dataset.memoId = memo.id;
        const ul = document.createElement('ul');

        const copyLi = document.createElement('li');
        const copyButton = document.createElement('button');
        copyButton.textContent = '内容コピー';
        copyButton.addEventListener('click', (e) => {
            e.stopPropagation();
            window.goCopyMemoContent(memo.id); // Call Go function
            closeAllKebabMenus();
        });
        copyLi.appendChild(copyButton);

        const deleteLi = document.createElement('li');
        const deleteButton = document.createElement('button');
        deleteButton.textContent = '削除';
        deleteButton.addEventListener('click', (e) => {
            e.stopPropagation();
            const shortContent = memo.content.length > 20 ? memo.content.substring(0, 20) + '...' : memo.content;
            if (confirm(`「${shortContent}」\nこの記録を削除しますか？`)) {
                window.goDeleteMemo(memo.id);
            }
            closeAllKebabMenus();
        });
        deleteLi.appendChild(deleteButton);

        ul.appendChild(copyLi);
        ul.appendChild(deleteLi);
        menuDiv.appendChild(ul);
        return menuDiv;
    }

    document.body.addEventListener('click', (event) => {
        if (activeKebabMenu && !activeKebabMenu.contains(event.target) && !event.target.closest('.kebab-menu-button')) {
            closeAllKebabMenus();
        }
    });
    
    // --- Record View & Search Results Display (Shared Logic) ---
    function formatDate(dateString) {
        if (!dateString) return '日時不明';
        const date = new Date(dateString);
        return date.toLocaleString('ja-JP', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' });
    }

    function renderMemos(memos, displayAreaElement) {
        displayAreaElement.innerHTML = ''; // Clear previous items
        if (!memos || memos.length === 0) {
            // Message for empty list is handled by jsUpdateRecordListTitleVisibility or jsDisplaySearchResults
            return;
        }

        memos.forEach(memo => {
            const item = document.createElement('div');
            item.className = 'memo-item';
            item.dataset.id = memo.id;

            const contentDiv = document.createElement('div');
            contentDiv.className = 'memo-item-content';
            contentDiv.innerHTML = autoLinkText(memo.content);

            const timestampDiv = document.createElement('div');
            timestampDiv.className = 'memo-item-timestamp';
            timestampDiv.textContent = `記録日時: ${formatDate(memo.createdAt)}`;
            
            const kebabButton = document.createElement('button');
            kebabButton.className = 'kebab-menu-button';
            kebabButton.innerHTML = '&#x22EE;'; // Vertical ellipsis (ケバブアイコン)
            kebabButton.title = 'アクション';
            
            const kebabMenu = createKebabMenu(memo);
            item.appendChild(kebabMenu); // Append menu first for positioning context

            kebabButton.addEventListener('click', (e) => {
                e.stopPropagation();
                if (activeKebabMenu === kebabMenu && kebabMenu.style.display === 'block') {
                    closeAllKebabMenus();
                } else {
                    closeAllKebabMenus();
                    kebabMenu.style.display = 'block';
                    activeKebabMenu = kebabMenu;
                }
            });

            item.appendChild(contentDiv);
            item.appendChild(timestampDiv);
            item.appendChild(kebabButton);
            displayAreaElement.appendChild(item);
        });
    }

    window.jsDisplayMemos = (memos) => { // Exposed to Go for Record View
        renderMemos(memos, recordDisplayArea);
    };
    
    window.jsUpdateRecordListTitleVisibility = (visible) => { // Exposed to Go
        recordViewListTitle.style.display = visible ? 'block' : 'none';
        const messageElement = recordDisplayArea.querySelector('.message');
        if (messageElement) messageElement.remove();

        if (!visible && recordDisplayArea.innerHTML === '') { // Check if area is truly empty
             recordDisplayArea.innerHTML = '<p class="message">記録はまだありません。</p>';
        }
    };

    saveMemoButton.addEventListener('click', () => {
        window.goSaveMemo(memoContent.value);
    });

    memoContent.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
            e.preventDefault();
            window.goSaveMemo(memoContent.value);
        }
    });

    window.jsClearMemoInput = () => { // Exposed to Go
        memoContent.value = '';
        memoContent.focus();
    };
    
    window.jsCopyToClipboardAndFeedback = async (text, memoId) => { // Exposed to Go
        try {
            await navigator.clipboard.writeText(text);
            alert('内容をクリップボードにコピーしました。');
            // Optional: Find the specific copy button and give visual feedback
            // const kebabMenu = document.querySelector(`.kebab-menu[data-memo-id="${memoId}"]`);
            // if (kebabMenu) { /* ... */ }
        } catch (err) {
            console.error('クリップボードへのコピーに失敗:', err);
            alert('コピーに失敗しました。ブラウザのコンソールで詳細を確認してください。');
        }
    };

    // --- Search View ---
    function performSearch() {
        const query = searchInput.value;
        window.goSearchMemos(query);
        clearSearchButton.style.display = query ? 'inline-block' : 'none';
    }

    searchButton.addEventListener('click', performSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') performSearch();
    });
    searchInput.addEventListener('input', () => {
        clearSearchButton.style.display = searchInput.value ? 'inline-block' : 'none';
    });

    clearSearchButton.addEventListener('click', () => {
        searchInput.value = '';
        searchResultsArea.innerHTML = ''; // Clear results area
        clearSearchButton.style.display = 'none';
        searchInput.focus();
    });

    window.jsDisplaySearchResults = (results, message) => { // Exposed to Go
        searchResultsArea.innerHTML = ''; // Clear previous results or messages
        if (message) {
            searchResultsArea.innerHTML = `<p class="message">${message}</p>`;
        } else if (results && results.length > 0) {
            renderMemos(results, searchResultsArea); // Reuse renderMemos
        }
        // If no message and no results, area remains empty (covered by Go's logic for "not found")
    };

    // --- Data Management View ---
    exportDataButton.addEventListener('click', () => {
        window.goExportData();
    });

    importFile.addEventListener('change', (event) => {
        const file = event.target.files[0];
        if (file) {
            if (!file.name.endsWith('.db')) {
                alert('無効なファイル形式です。.dbファイルを選択してください。');
                importFile.value = ''; // Reset file input
                return;
            }
            const reader = new FileReader();
            reader.onload = (e) => {
                try {
                    const base64Data = e.target.result.split(',')[1];
                    if (base64Data) {
                        window.goImportData(base64Data);
                    } else {
                        throw new Error('ファイルからデータを読み取れませんでした。');
                    }
                } catch (error) {
                     alert('ファイルの読み込みまたは処理に失敗しました: ' + error.message);
                } finally {
                    importFile.value = ''; 
                }
            };
            reader.onerror = () => {
                alert('ファイルの読み込み中にエラーが発生しました。');
                importFile.value = '';
            };
            reader.readAsDataURL(file);
        }
    });
    
    window.jsTriggerFileDownload = (filename, mimeType, base64Data) => { // Exposed to Go
        try {
            const byteCharacters = atob(base64Data);
            const byteNumbers = new Array(byteCharacters.length);
            for (let i = 0; i < byteCharacters.length; i++) {
                byteNumbers[i] = byteCharacters.charCodeAt(i);
            }
            const byteArray = new Uint8Array(byteNumbers);
            const blob = new Blob([byteArray], {type: mimeType});

            const link = document.createElement('a');
            link.href = URL.createObjectURL(blob);
            link.download = filename;
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
            URL.revokeObjectURL(link.href);
            alert(`「${filename}」をエクスポートしました。`);
        } catch (e) {
            console.error("File download error:", e);
            alert("ファイルのエクスポートに失敗しました。");
        }
    };

    // --- Global JS functions callable from Go (defined with window.) ---
    // window.jsAlert = (message) => alert(message); // Go can call global alert directly
    window.jsShowView = (viewName) => { showView(viewName); };

    // --- Initial App Load ---
    if (typeof window.goInitializeApp === "function") {
        window.goInitializeApp();
    } else {
        console.error("goInitializeApp is not defined. WASM might not be ready or failed to expose function.");
        alert("アプリケーションの初期化に失敗しました。");
    }
    showView('record'); // Set initial view and active nav icon
});