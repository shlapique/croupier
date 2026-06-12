// ---------- STATE ----------

const state = {
    page: 0,
    items: [],
    selected: [],
    download: {
        running: false,
        currentFile: "",
        completed: 0,
        total: 0,
        progress: 0
    }
};


// ---------- POLLING ----------

let poller = null;

function startPolling() {
    if (poller) return;

    poller = setInterval(refresh, 2000);
}

function stopPolling() {
    if (!poller) return;

    clearInterval(poller);
    poller = null;
}


// ---------- HTTP ----------

async function api(url, options = {}) {
    const response = await fetch(url, {
        headers: {
            "Content-Type": "application/json"
        },
        ...options
    });

    if (!response.ok) {
        throw new Error(
            `HTTP ${response.status}`
        );
    }

    return response;
}


// ---------- REFRESH ----------

async function refresh() {
    try {
        const response =
            await api("/api/state");

        const newState =
            await response.json();

        Object.assign(state, newState);

        render();

        if (state.download.running) {
            startPolling();
        } else {
            stopPolling();
        }

    } catch (err) {
        console.error(
            "refresh failed",
            err
        );
    }
}


// ---------- RENDER ----------

function render() {
    renderPageInfo();
    renderFiles();
    renderSelected();
    renderDownload();
}

function renderPageInfo() {
    document.getElementById("pageInfo")
        .textContent =
        `Page ${state.page}`;
}

function renderFiles() {
    const root =
        document.getElementById("fileList");

    root.innerHTML = "";

    for (const item of state.items) {
        const row =
            document.createElement("div");

        row.className = "file";

        row.innerHTML = `
            <input
                type="checkbox"
                ${item.selected ? "checked" : ""}
            >
            <label>${item.name}</label>
        `;

        const checkbox =
            row.querySelector("input");

        checkbox.addEventListener(
            "change",
            async () => {
                await toggleItem(item.id);
            }
        );

        root.appendChild(row);
    }
}

function renderSelected() {
    const root =
        document.getElementById(
            "selectedList"
        );

    root.innerHTML = "";

    for (const item of state.selected) {
        const div =
            document.createElement("div");

        div.className =
            "selected-item";

        div.textContent =
            item.name;

        root.appendChild(div);
    }

    document.getElementById(
        "selectedCount"
    ).textContent =
        state.selected.length;

    document.getElementById(
        "downloadBtn"
    ).textContent =
        `Download (${state.selected.length})`;
}

function renderDownload() {
    const d = state.download;

    document.getElementById(
        "downloadFile"
    ).textContent =
        d.currentFile || "Idle";

    document.getElementById(
        "downloadText"
    ).textContent =
        `${d.completed} / ${d.total}`;

    document.getElementById(
        "progressBar"
    ).style.width =
        `${d.progress}%`;

    document.getElementById(
        "stopBtn"
    ).disabled =
        !d.running;
}


// ---------- ACTIONS ----------

async function nextPage() {
    await api(
        "/api/page/next",
        { method: "POST" }
    );

    await refresh();
}

async function prevPage() {
    await api(
        "/api/page/prev",
        { method: "POST" }
    );

    await refresh();
}

async function selectAll() {
    await api(
        "/api/select-all",
        { method: "POST" }
    );

    await refresh();
}

async function toggleItem(itemId) {
    await api(
        "/api/toggle-selection",
        {
            method: "POST",

            body: JSON.stringify({
                itemId
            })
        }
    );

    await refresh();
}

async function startDownload() {
    await api(
        "/api/download/start",
        {
            method: "POST"
        }
    );

    await refresh();
}

async function stopDownload() {
    await api(
        "/api/download/stop",
        {
            method: "POST"
        }
    );

    await refresh();
}


// ---------- EVENT WIRING ----------

function bindEvents() {
    document
        .getElementById("nextBtn")
        .addEventListener(
            "click",
            nextPage
        );

    document
        .getElementById("prevBtn")
        .addEventListener(
            "click",
            prevPage
        );

    document
        .getElementById("selectAllBtn")
        .addEventListener(
            "click",
            selectAll
        );

    document
        .getElementById("downloadBtn")
        .addEventListener(
            "click",
            startDownload
        );

    document
        .getElementById("stopBtn")
        .addEventListener(
            "click",
            stopDownload
        );
}


// ---------- BOOT ----------

window.addEventListener(
    "DOMContentLoaded",
    async () => {
        bindEvents();
        await refresh();
    }
);
