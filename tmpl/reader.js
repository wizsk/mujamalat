// don't remove
// Create a reference for the Wake Lock.
let wakeLock = null;
const wakelockOptn = document.getElementById("wakelock");
let inactivityTimer = null;
const INACTIVITY_MINUTES = 7;
const reader = document.getElementById("reader");
const popup = document.getElementById("popup");
const highlight = document.getElementById("highlight");
const openDictBtn = document.getElementById("openDictBtn");
const readerMenu = document.getElementById("readerMenu");
const readerMenuBtn = document.getElementById("readerMenuBtn");
const wordSpans = document.querySelectorAll(".rWord");
const vewingMode = document.getElementById("vewing-mode");
const textAlign = document.getElementById("text-align");
const poemStyle = document.getElementById("poem");
const fontSelector = document.getElementById("font-selector");

for (let i = 0; i < wordSpans.length; i++) {
    if (wordSpans[i].dataset.oar !== "")
        wordSpans[i].onclick = openPopup;
}

function closePopup() {
    if (!lastClikedWord) return;
    lastClikedWord.classList.remove("clicked");
    lastClikedWord = null;
    popup.classList.add('hidden');
    openDictBtn.onclick = () => { };
}

async function addOrRmHiClass(word, add) {
    for (let i = 0; i < wordSpans.length; i++) {
        const w = wordSpans[i].dataset.oar;
        if (w === word) {
            if (add) {
                wordSpans[i].classList.add("hi");
            } else {
                wordSpans[i].classList.remove("hi");
            }
        }
    }
}


// ontop of the main file
function openPopup(e) {
    if (lastClikedWord) {
        if (lastClikedWord == e.target) {
            closePopup();
            return;
        }
        lastClikedWord.classList.remove("clicked");
    }

    const onlyAr = e.target.dataset.oar;

    if (e.target.classList.contains("hi")) {
        highlight.classList.add("alert");
    } else { highlight.classList.remove("alert");}

    highlight.onclick = async () => {
        closePopup();

        const hWord = onlyAr;
        let msg = "&add=true";
        const del = e.target.classList.contains("hi");
        if (del) msg = "&del=true";

        console.log(`/rd/high?w=${hWord}${msg}`);
        fetch(`/rd/high?w=${hWord}${msg}`, { method: "POST" })
            .then((res) => {
                if (res.status === 202) {
                    if (del) {
                        e.target.classList.remove("hi");
                        addOrRmHiClass(hWord, false);
                    } else {
                        e.target.classList.add("hi");
                        addOrRmHiClass(hWord, true);
                    }
                } else {
                    alert(`Couldn't save/del highlight: ${hWord}`);
                }
            })
            .catch((err) => {
                alert(`Couldn't save/del highlight: ${hWord}`);
                console.error(err)
            });
    }

    openDictBtn.onclick = () => {
        closePopup()
        openDictionay(onlyAr);
    }

    lastClikedWord = e.target;
    e.target.classList.add("clicked");
    popup.classList.remove('hidden');

    setPopUpPos();
}

function setPopUpPos() {
    if (!lastClikedWord) return;
    // Wait for popup to render so we can measure width
    requestAnimationFrame(() => {
        const rect = lastClikedWord.getBoundingClientRect();
        const popupWidth = popup.offsetWidth;
        const popupHeight = popup.offsetHeight;
        const screenW = window.innerWidth;
        const screenH = window.innerHeight;

        let top = rect.bottom + window.scrollY + 4;
        let left = (rect.left + (rect.width / 2) - (popup.offsetWidth / 2)) + window.scrollX;

        // If off right edge
        if (left + popupWidth > screenW) {
            left = window.scrollX + screenW - (popupWidth + 4);
        }

        // If off left edge
        if (left < 4) {
            left = 4;
        }

        // avoid buttom nav
        if (rect.bottom + (popupHeight + 4) + 100 > screenH) {
            top = rect.top + window.scrollY - 4 - popupHeight;
        }

        // Apply final position
        popup.style.top = `${top}px`;
        popup.style.left = `${left}px`;
    })
}


let dictContainerOpen = false;
function openDictionay(w) {
    dictContainerOpen = true;
    readerMode = false;
    querySelector.innerHTML = "";
    document.body.style.overflow = "hidden";
    dict_container_tougle.classList.remove('hidden');
    preQuery = w;
    currWord = w;
    queryIdx = 0;
    input.value = w;
    dict_container.classList.remove("hidden");
    showHideNav(true, true);
    setNavHeight();
    history.pushState({}, "", window.location.href);

    getResAndShow(w);
}


// Handle browser back/forward
window.addEventListener("popstate", (e) => {
    if (readerMenuOpen) {
        readerMenu.classList.add("hidden");
        document.body.style.overflow = "auto"
        readerMenuBtn.disabled = false;
        readerMenuOpen = false;
    } else if (dictContainerOpen) {
        dictContainerOpen = false;
        dict_container.classList.add('hidden')
        dict_container_tougle.classList.add('hidden')
        document.body.style.overflow = "auto";
        contentHolder.innerHTML = "";
        readerMode = true;
    }
});

let readerMenuOpen = false;
readerMenuBtn.onclick = () => {
    readerMenu.classList.remove("hidden");
    readerMenuBtn.disabled = true;
    readerMenuOpen = true;
    document.body.style.overflow = "hidden"
    history.pushState({}, "", window.location.href);
}


fontSelector.onchange = (e) => {
    const val = e.target.value;
    if (val === "font") {
        reader.style.fontFamily = "";
        setPopUpPos();
        window.localStorage.removeItem(getFontSelectorLSN());
        return;
    }
    reader.style.fontFamily = val;
    setPopUpPos();
    window.localStorage.setItem(getFontSelectorLSN(), val);
}

/** LSN = local storage name */
function getFontSelectorLSN() {
    return `${window.location.pathname}-seleted-font`
}

vewingMode.onchange = (e) => {
    const val = e.target.value;
    switch (val) {
        case "normal":
            poemStyle.disabled = true;
            setPopUpPos();
            window.localStorage.removeItem(getVewingModeLSN());
            break;
        case "poem":
            poemStyle.disabled = false;
            setPopUpPos();
            window.localStorage.setItem(getVewingModeLSN(), val);
            break;
    }
}

/** LSN = local storage name */
function getVewingModeLSN() {
    return `${window.location.pathname}-vewingMode`
}

const textRightClassName = 'text-right';
textAlign.onchange = (e) => {
    const val = e.target.value;
    switch (val) {
        case "right":
            reader.classList.add(textRightClassName);
            setPopUpPos();
            window.localStorage.setItem(getTextAlignLSN(), val);
            break;
        case "justify":
            reader.classList.remove(textRightClassName);
            setPopUpPos();
            window.localStorage.removeItem(getTextAlignLSN());
            break;
    }
}

/** LSN = local storage name */
function getTextAlignLSN() {
    return `${window.location.pathname}-textAlign`
}

document.addEventListener('DOMContentLoaded', () => {
    const font = window.localStorage.getItem(getFontSelectorLSN());
    if (font) {
        reader.style.fontFamily = font;
        fontSelector.value = font;
    }

    if (window.localStorage.getItem(getVewingModeLSN()) === "poem") {
        poemStyle.disabled = false;
        vewingMode.value = "poem";
    }

    if (window.localStorage.getItem(getTextAlignLSN()) === "right") {
        reader.classList.add(textRightClassName);
        textAlign.value = "right";
    }

});

if ("wakeLock" in navigator) {
    async function requestWakeLock() {
        try {
            wakeLock = await navigator.wakeLock.request("screen");

            wakeLock.addEventListener("release", async () => {
                wakelockOptn.value = "off";
                wakelockOptn.style.color = "var(--alert)";
                console.log("Wake Lock lost — attempting to restore");
                // Try restoring only if page is still visible
                if (document.visibilityState === "visible") {
                    await requestWakeLock();
                }
            });

            wakelockOptn.value = "on";
            wakelockOptn.style.color = "var(--ok)";
            console.log("Wake Lock active");
        } catch (err) {
            console.error("Wake Lock request failed:", err);
        }
    }

    function releaseWakeLock() {
        if (wakeLock) {
            wakeLock.release();
            wakeLock = null;
            console.log("Wake Lock manually released due to inactivity");
        }
    }

    document.addEventListener("visibilitychange", async () => {
        if (document.visibilityState === "visible") {
            // If wake lock is missing OR previously released → request again
            if (!wakeLock || wakeLock.released) {
                await requestWakeLock();
                resetInactivityTimer();
            }
        }
    });

    function resetInactivityTimer() {
        clearTimeout(inactivityTimer);
        inactivityTimer = setTimeout(() => {
            releaseWakeLock();
        }, INACTIVITY_MINUTES * 60 * 1000);
    }

    // Start lock when user interacts
    document.addEventListener("pointerdown", async () => {
        if (!wakeLock) {
            await requestWakeLock();
        }
        resetInactivityTimer();
    });

    // Also consider scroll & key activity
    ["scroll", "keydown", "touchstart"].forEach(event => {
        document.addEventListener(event, resetInactivityTimer, { passive: true });
    });
}

// lins in the reader menu..1stl close the reader menu
const readerMenuAnkers = document.getElementsByClassName("readerMenuAnker");
for (let i = 0; i < readerMenuAnkers.length; i++) {
    readerMenuAnkers[i].addEventListener('click', (e) => {
        e.preventDefault();
        window.history.back();
        window.location.href = e.target.href;
    });
}
