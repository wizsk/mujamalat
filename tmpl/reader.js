// don't remove
const popup = document.getElementById("popup"); // for chrome
const highlight = document.getElementById("highlight");
const openDictBtn = document.getElementById("openDictBtn");

const wordSpans = document.getElementsByClassName("rWord");
for (let i = 0; i < wordSpans.length; i++) {
    const w = wordSpans[i].innerText;
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
        if (e.target.classList.contains("hi")) {
            e.target.classList.remove("hi");
            addOrRmHiClass(hWord, false);
            del = "&del=true"
        } else {
            e.target.classList.add("hi");
            addOrRmHiClass(hWord, true);
        }

        console.log(`/rd/high?w=${hWord}${del}`);
        const r = await fetch(`/rd/high?w=${hWord}${del}`, { method: "POST" })
            .catch(err => console.error(err));

        if (!r.ok) alert(`Couldn't save/del highlight: ${hWord}`);
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
    closeDictContainer();
});


function closeDictContainer() {
    dict_container.classList.add('hidden')
    dict_container_tougle.classList.add('hidden')
    document.body.style.overflow = "auto";
    contentHolder.innerHTML = "";
    readerMode = true;
}