// don't remove
const reader = document.getElementById("reader");
const popup = document.getElementById("popup");
const highlight = document.getElementById("highlight");
const openDictBtn = document.getElementById("openDictBtn");
const readerMenu = document.getElementById("readerMenu");
const readerMenuBtn = document.getElementById("readerMenuBtn");
const wordSpans = document.getElementsByClassName("rWord");
const vewingMode = document.getElementById("vewing-mode");
const textAlign = document.getElementById("text-align");
const poemStyle = document.getElementById("poem");

for (let i = 0; i < wordSpans.length; i++) {
    if (wordSpans[i].dataset.oar !== "")
        wordSpans[i].onclick = openPopup;
}

function closePopup() {
    if (!lastClikedWord) return;
    lastClikedWord.classList.remove("clicked");
    lastClikedWord = null;
    popup.classList.add('hidden');
    highlight.onclick = () => { };
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

// let lastClikedWord = null;
// ontop of the main file
function openPopup(e) {
    if (lastClikedWord) {
        if (lastClikedWord == e.target) {
            closePopup();
            return;
        }
        lastClikedWord.classList.remove("clicked");
    }

    const hWord = e.target.dataset.oar;
    if (e.target.classList.contains("hi")) {
        highlight.classList.add("alert");
    } else {
        highlight.classList.remove("alert");
    }
    highlight.onclick = async () => {
        closePopup();

        let del = "";
        if (e.target.classList.contains("hi")) del = "&del=true";

        console.log(`/rd/high?w=${hWord}${del}`);
        fetch(`/rd/high?w=${hWord}${del}`, { method: "POST" })
            .then((res) => {
                if (res.status === 202) {
                    if (del !== "") {
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
        openDictionay(hWord);
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

const textJustifyClassName = 'text-justify';
textAlign.onchange = (e) => {
    const val = e.target.value;
    switch (val) {
        case "right":
            reader.classList.remove(textJustifyClassName);
            setPopUpPos();
            window.localStorage.removeItem(getTextAlignLSN());
            break;
        case "justify":
            reader.classList.add(textJustifyClassName);
            setPopUpPos();
            window.localStorage.setItem(getTextAlignLSN(), val);
            break;
    }
}

/** LSN = local storage name */
function getTextAlignLSN() {
    return `${window.location.pathname}-textAlign`
}

document.addEventListener('DOMContentLoaded', () => {
    if (window.localStorage.getItem(getVewingModeLSN()) === "poem") {
        poemStyle.disabled = false;
        vewingMode.value = "poem";
    } else if (window.localStorage.getItem(getTextAlignLSN()) === "justify") {
        reader.classList.add(textJustifyClassName);
        textAlign.value = "justify";
    }
});

const readerMenuAnkers = document.getElementsByClassName("readerMenuAnker");
for (let i = 0; i < readerMenuAnkers.length; i++) {
    readerMenuAnkers[i].addEventListener('click', (e) => {
        e.preventDefault();
        window.history.back();
        window.location.href = e.target.href;
    });
}

