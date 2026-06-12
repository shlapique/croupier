async function refresh() {
    const response = await fetch("/state");

    const state = await response.json();

    document.getElementById("state").textContent =
        JSON.stringify(state, null, 2);
}

async function next() {
    await fetch("/next", {
        method: "POST"
    });

    await refresh();
}

async function prev() {
    await fetch("/prev", {
        method: "POST"
    });

    await refresh();
}

document
    .getElementById("next")
    .addEventListener("click", next);

document
    .getElementById("prev")
    .addEventListener("click", prev);

refresh();
