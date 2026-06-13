const STORAGE_KEY = "croupier.selected";

function loadSelected() {
    try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (raw) {
            return new Map(JSON.parse(raw));
        }
    } catch (e) {
        // Ignore corrupt/unavailable storage and start fresh.
    }
    return new Map();
}

function saveSelected() {
    localStorage.setItem(
        STORAGE_KEY,
        JSON.stringify(Array.from(selected.entries()))
    );
}

// Restored from the previous session, so checkboxes survive a reload.
const selected = loadSelected();

async function refresh() {
    const response = await fetch("/state");
    if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
    }
    const state = await response.json();
    renderFiles(state.Files);
    updateDownloadButton();
}

function renderFiles(files) {
    // A nil slice from the server serializes to JSON null (e.g. past the
    // first/last page), so fall back to an empty list instead of crashing.
    const list = Array.isArray(files) ? files : [];

    const root = document.getElementById("files");
    root.innerHTML = "";

    for (const file of list) {
        const row = document.createElement("div");
        row.className = "file";

        const checkbox = document.createElement("input");
        checkbox.type = "checkbox";
        // Match by Name: the server hands the same file a new Id after
        // navigation, but the filename stays stable.
        checkbox.checked = selected.has(file.name);

        checkbox.addEventListener("change", () => {
            if (checkbox.checked) {
                selected.set(file.name, file);
            } else {
                selected.delete(file.name);
            }
            saveSelected();
            updateDownloadButton();
        });

        const label = document.createElement("span");
        label.textContent = file.name;

        row.appendChild(checkbox);
        row.appendChild(label);
        root.appendChild(row);
    }
}

function updateDownloadButton() {
    document.getElementById("download").textContent =
        `Download (${selected.size})`;
}

function clearSelection() {
    selected.clear();
    saveSelected();
    document
        .querySelectorAll("#files input[type=checkbox]")
        .forEach((cb) => (cb.checked = false));
    updateDownloadButton();
}

function log(text) {
    const panel = document.getElementById("log");
    panel.textContent += text;
    panel.scrollTop = panel.scrollHeight;
}

// Serialize navigation: each next/prev fully finishes before the next runs,
// so rapid clicks can't race and render a stale page.
let nav = Promise.resolve();

function navigate(direction) {
    nav = nav
        .then(() => fetch("/" + direction, { method: "POST" }))
        .then(() => refresh())
        .catch((e) => log(`Error: ${e.message}\n`));
}

async function download() {
    const files = Array.from(selected.values());
    if (files.length === 0) {
        return;
    }

    // Reset selection + counter immediately on click.
    clearSelection();

    document.getElementById("cancel").style.display = "";

    log(`Sending ${files.length} file(s) to download...\n`);

    try {
        const response = await fetch("/download", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(files),
        });

        if (!response.ok) {
            log(`Error: HTTP ${response.status}\n`);
            return;
        }

        // Stream the server's log output and append it as it arrives.
        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        while (true) {
            const { value, done } = await reader.read();
            if (done) break;
            log(decoder.decode(value, { stream: true }));
        }
    } catch (e) {
        log(`Error: ${e.message}\n`);
    } finally {
        document.getElementById("cancel").style.display = "none";
    }
}

async function cancelDownload() {
    await fetch("/cancel", { method: "POST" });
}

document.getElementById("next")
    .addEventListener("click", () => navigate("next"));
document.getElementById("prev")
    .addEventListener("click", () => navigate("prev"));
document.getElementById("download").addEventListener("click", download);
document.getElementById("reset").addEventListener("click", clearSelection);
document.getElementById("cancel").addEventListener("click", cancelDownload);

refresh();
