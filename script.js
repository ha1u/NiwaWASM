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
    const searchExportArea = document.getElementById('searchExportArea');
    const exportSearchResultButton = document.getElementById('exportSearchResultButton');

    const exportBackupButton = document.getElementById('exportBackupButton'); // Renamed for clarity
    const importFile = document.getElementById('importFile');
    const exportFormatSelect = document.getElementById('exportFormatSelect');
    const exportFormattedDataButton = document.getElementById('exportFormattedDataButton');
    const showDataGuideButton = document.getElementById('showDataGuideButton');
    const dataGuideModal = document.getElementById('dataGuideModal');
    const closeDataGuideModal = document.getElementById('closeDataGuideModal');
    const understandDataGuideButton = document.getElementById('understandDataGuideButton');


    let currentView = 'record';
    let activeKebabMenu = null;
    let currentSearchResults = []; // 検索結果を保持する配列

    // --- Theme Management --- (変更なし)
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

    // --- Navigation --- (変更なし)
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

    // --- URL Auto-linking --- (変更なし)
    function autoLinkText(text) {
        const urlPattern = /(\b(https?|ftp|file):\/\/[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])|(\bwww\.[-A-Z0-9+&@#\/%?=~_|!:,.;]*[-A-Z0-9+&@#\/%=~_|])/ig;
        return text.replace(urlPattern, (url) => {
            const properUrl = url.startsWith('www.') ? 'http://' + url : url;
            const escapedUrlText = url.replace(/</g, "&lt;").replace(/>/g, "&gt;");
            return `<a href="${properUrl}" target="_blank" rel="noopener noreferrer">${escapedUrlText}</a>`;
        });
    }
    
    // --- Kebab Menu --- (変更なし)
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
            window.goCopyMemoContent(memo.id);
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
        displayAreaElement.innerHTML = ''; 
        if (!memos || memos.length === 0) return;

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
            kebabButton.innerHTML = '&#x22EE;';
            kebabButton.title = 'アクション';
            const kebabMenu = createKebabMenu(memo);
            item.appendChild(kebabMenu);
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

    window.jsDisplayMemos = (memos) => {
        renderMemos(memos, recordDisplayArea);
    };
    
    window.jsUpdateRecordListTitleVisibility = (visible) => {
        recordViewListTitle.style.display = visible ? 'block' : 'none';
        const messageElement = recordDisplayArea.querySelector('.message');
        if (messageElement) messageElement.remove();
        if (!visible && recordDisplayArea.innerHTML === '') {
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
    window.jsClearMemoInput = () => {
        memoContent.value = '';
        memoContent.focus();
    };
    window.jsCopyToClipboardAndFeedback = async (text, memoId) => {
        try {
            await navigator.clipboard.writeText(text);
            alert('内容をクリップボードにコピーしました。');
        } catch (err) {
            console.error('クリップボードへのコピーに失敗:', err);
            alert('コピーに失敗しました。ブラウザのコンソールで詳細を確認してください。');
        }
    };

    // --- Search View ---
    function performSearch() {
        const query = searchInput.value;
        window.goSearchMemos(query); // GoがjsDisplaySearchResultsを呼ぶ
    }
    searchButton.addEventListener('click', performSearch);
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') performSearch();
    });
    searchInput.addEventListener('input', () => {
        clearSearchButton.style.display = searchInput.value ? 'inline-block' : 'none';
        if (!searchInput.value) { // 検索語がクリアされたら結果も隠す
            searchResultsArea.innerHTML = '';
            searchExportArea.style.display = 'none';
            currentSearchResults = [];
        }
    });
    clearSearchButton.addEventListener('click', () => {
        searchInput.value = '';
        searchResultsArea.innerHTML = '';
        clearSearchButton.style.display = 'none';
        searchExportArea.style.display = 'none';
        currentSearchResults = [];
        searchInput.focus();
    });

    // goSearchMemosから呼び出される
    window.jsDisplaySearchResults = (resultsArray, message, resultsJsonString) => {
        searchResultsArea.innerHTML = ''; 
        if (message) {
            searchResultsArea.innerHTML = `<p class="message">${message}</p>`;
            searchExportArea.style.display = 'none';
            currentSearchResults = [];
        } else if (resultsArray && resultsArray.length > 0) {
            renderMemos(resultsArray, searchResultsArea);
            searchExportArea.style.display = 'block'; // エクスポートボタン表示
            try {
                 currentSearchResults = JSON.parse(resultsJsonString); // Goから渡されたJSON文字列をパースして保持
            } catch(e) {
                console.error("Failed to parse search results JSON:", e);
                currentSearchResults = []; // パース失敗時は空に
                 searchExportArea.style.display = 'none';
            }
        } else { // No message, no results (e.g. search cleared)
            searchExportArea.style.display = 'none';
            currentSearchResults = [];
        }
    };

    exportSearchResultButton.addEventListener('click', () => {
        if (currentSearchResults && currentSearchResults.length > 0) {
            const format = exportFormatSelect.value; // データ管理ビューの選択形式を利用
            const memosJson = JSON.stringify(currentSearchResults);
            window.goExportFormattedData(format, memosJson);
        } else {
            alert('エクスポートする検索結果がありません。');
        }
    });


    // --- Data Management View ---
    exportBackupButton.addEventListener('click', () => { // 旧 exportDataButton
        window.goExportBackupData(); // Gobバックアップ専用関数を呼ぶ
    });

    importFile.addEventListener('change', (event) => {
        const file = event.target.files[0];
        if (file) {
            if (!file.name.endsWith('.data')) {
                alert('無効なファイル形式です。.dataファイルを選択してください。');
                importFile.value = '';
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

    exportFormattedDataButton.addEventListener('click', () => {
        const format = exportFormatSelect.value;
        window.goExportFormattedData(format, "all"); // "all" を渡して全件エクスポート
    });
    
    // Data Guide Modal Logic
    showDataGuideButton.addEventListener('click', () => {
        dataGuideModal.style.display = 'block';
    });
    closeDataGuideModal.addEventListener('click', () => {
        dataGuideModal.style.display = 'none';
    });
    understandDataGuideButton.addEventListener('click', () => {
        dataGuideModal.style.display = 'none';
    });
    window.addEventListener('click', (event) => { // Modal外クリックで閉じる
        if (event.target == dataGuideModal) {
            dataGuideModal.style.display = 'none';
        }
    });
    
    // JS helper for Go to trigger downloads (変更なし)
    window.jsTriggerFileDownload = (filename, mimeType, base64Data) => {
        try {
            const byteCharacters = atob(base64Data); // Base64デコード
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
            // alert(`「${filename}」をエクスポートしました。`); // Go側でアラートを出す場合は不要
        } catch (e) {
            console.error("File download error:", e);
            alert("ファイルのエクスポートに失敗しました。");
        }
    };

    window.jsShowView = (viewName) => { showView(viewName); };

    // --- Initial App Load ---
    if (typeof window.goInitializeApp === "function") {
        window.goInitializeApp();
    } else {
        console.error("goInitializeApp is not defined. WASM might not be ready or failed to expose function.");
        // このアラートはWASMロード失敗時にも出る可能性がある
        // alert("アプリケーションの初期化に失敗しました。"); 
    }
    showView('record'); 
});