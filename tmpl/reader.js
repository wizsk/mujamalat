// don't remove
const popup = document.getElementById("popup"); // for chrome
const highlight = document.getElementById("highlight");
const openDictBtn = document.getElementById("openDictBtn");
const readerMenu = document.getElementById("readerMenu");
const readerMenuBtn = document.getElementById("readerMenuBtn");
const wordSpans = document.getElementsByClassName("rWord");
const vewingMode = document.getElementById("vewing-mode");
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


function openDictionay(w) {
    readerMode = false;
    querySelector.innerHTML = "";
    document.body.style.overflow = "hidden";
    dict_container_tougle.classList.remove('hidden');
    preQuery = w;
    currWord = w;
    queryIdx = 0;
    input.value = w;
    dict_container.classList.remove("hidden");
    setNavHeight();
    showHideNav(true);
    history.pushState({}, "", window.location.href);

    getResAndShow(w);
}


// Handle browser back/forward
window.addEventListener("popstate", (e) => {
    if (readerMenuOpen) {
        readerMenu.classList.add("hidden");
        readerMenuOpen = false;
    }
    else
        closeDictContainer();
});

let readerMenuOpen = false;
readerMenuBtn.onclick = () => {
    if (readerMenu.classList.toggle("hidden")) {
        document.body.style.overflow = "auto"
    }
    else {
        readerMenuOpen = true;
        document.body.style.overflow = "hidden"
        history.pushState({}, "", window.location.href);
    }
}

vewingMode.onchange = (e) => {
    const val = e.target.value;
    switch (val) {
        case "normal":
            poemStyle.disabled = true;
            break;
        case "poem":
            poemStyle.disabled = false;
            break;
    }
}


function closeDictContainer() {
    dict_container.classList.add('hidden')
    dict_container_tougle.classList.add('hidden')
    document.body.style.overflow = "auto";
    contentHolder.innerHTML = "";
    readerMode = true;
}
