// don't remove
const wordSpans = document.getElementsByClassName("rWord");
for (let i = 0; i < wordSpans.length; i++) {
    const w = wordSpans[i].innerText;
    wordSpans[i].onclick = (e) => openOpopup(e, w);
}


function closePopup() {
    if (!lastClikedWord) return;
    lastClikedWord.classList.remove("clicked");
    lastClikedWord = null;
    popup.classList.add('hidden');
    hightligt.onclick = () => { };
    openDictBtn.onclick = () => { };
}

// let lastClikedWord = null;
// ontop of the main file
function openOpopup(e, w) {
    if (lastClikedWord) {
        if (lastClikedWord == e.target) {
            closePopup();
            return;
        }
        lastClikedWord.classList.remove("clicked");
    }

    hightligt.onclick = async () => {
        closePopup();
        let del = "";
        if (e.target.classList.contains("hi")) {
            e.target.classList.remove("hi");
            del = "&del=true"
        } else {
            e.target.classList.add("hi");
        }
        const hiw = e.target.innerText;
        console.log(`/rd/high?w=${hiw}${del}`);
        const r = await fetch(`/rd/high?w=${hiw}${del}`)
            .catch(err => console.error(err));

        if (!r.ok) alert(`Couldn't save/del highlight: ${hiw}`);
    }

    openDictBtn.onclick = () => {
        openDictionay(w);
        closePopup()
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

        if (rect.bottom + (popupHeight + 4) > screenH) {
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