// btn.js

// -- font stuff
const fontDiffCVal = 2.0;

function saveFont(s) {
    window.localStorage.setItem("font-size-for-kamusssss", s)
}

function setSavedFontSize() {
    const s = window.localStorage.getItem("font-size-for-kamusssss");
    if (s) {
        document.body.style.fontSize = s;
        resetFont.classList.remove("hidden");
    }
}

function fontSizeInc() {
    const v = parseFloat(window.getComputedStyle(document.body, null).getPropertyValue("font-size"))
    const s = `${v + fontDiffCVal}px`;
    document.body.style.fontSize = s;
    saveFont(s);
    resetFont.classList.remove("hidden");
}

function fontSizeDec() {
    const v = parseFloat(window.getComputedStyle(document.body, null).getPropertyValue("font-size"))
    const s = `${v - fontDiffCVal}px`;
    document.body.style.fontSize = s;
    saveFont(s);
    resetFont.classList.remove("hidden");
}

function scroolToTop() { window.scrollTo({ top: 0, behavior: 'smooth' }); }

function resetFontSize() {
    console.log("fontsize reset-edd")
    localStorage.removeItem("font-size-for-kamusssss");
    document.body.style.fontSize = "";
    resetFont.classList.add("hidden");
}

plus.onclick = fontSizeInc;
minus.onclick = fontSizeDec;
up.onclick = scroolToTop;

resetFont.onclick = resetFontSize;

function toggleChangeDict() {
    if (changeDict.classList.contains("hidden")) {
        w.blur();
        // document.body.classList.add('no-scroll');
        changeDict.classList.remove("hidden");
        changeDictInpt.value = "";
        changeDictInpt.focus();
        isChangeDictShwoing = true;
    } else {
        changeDictInpt.blur();
        changeDict.classList.add("hidden");
        // document.body.classList.remove('no-scroll');
        isChangeDictShwoing = false;
    }
}

// if success then returns true
function selectDict(s, minus1) {
    let v = -1;

    if (typeof s === "string") {
        s = s.replace(/[\u0660-\u0669]/g, d => d.charCodeAt(0) - 0x0660);
        v = parseInt(s);
    } else if (typeof s === "number") {
        v = s;
    }

    if (minus1) v -= 1;

    if (v > -1 && v < dicts.length) {
        const n = dicts[v].getAttribute('data-dict-name');
        if (selectedDict !== n)
            dicts[v].click();
        toggleChangeDict();
        return true;
    }

    return false;
}

document.querySelectorAll(".change-dict-btn").forEach(e => {
    e.addEventListener('click', toggleChangeDict);
});