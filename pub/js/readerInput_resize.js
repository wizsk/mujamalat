const ta = document.getElementById("txt");
const handle = document.getElementById("resizeHandle");

let startY = 0;
let startHeight = 0;
let dragging = false;

function startDrag(e) {
    dragging = true;
    handle.style.backgroundColor = "var(--alert)";

    startY = e.touches ? e.touches[0].clientY : e.clientY;
    startHeight = ta.offsetHeight;

    e.preventDefault();
}

function duringDrag(e) {
    if (!dragging) return;

    const y = e.touches ? e.touches[0].clientY : e.clientY;
    const dy = y - startY;

    let newHeight = startHeight + dy;
    const maxHeight = window.innerHeight * 0.9;

    newHeight = Math.max(80, Math.min(newHeight, maxHeight));
    ta.style.height = newHeight + "px";
}

function stopDrag() {
    dragging = false;
    handle.style.backgroundColor = "";
}

handle.addEventListener("mousedown", startDrag);
handle.addEventListener("touchstart", startDrag, { passive: false });

window.addEventListener("mousemove", duringDrag);
window.addEventListener("touchmove", duringDrag, { passive: false });

window.addEventListener("mouseup", stopDrag);
window.addEventListener("touchend", stopDrag);
