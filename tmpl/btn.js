// btn.js

// -- font stuff
const resetFont = document.getElementById("resetFont");
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
    setTimeout(() => {
        setNavHeight();
    }, navSetHeightDelay);
    resetFont.classList.remove("hidden");
}

function fontSizeDec() {
    const v = parseFloat(window.getComputedStyle(document.body, null).getPropertyValue("font-size"))
    const s = `${v - fontDiffCVal}px`;
    document.body.style.fontSize = s;
    saveFont(s);
    setTimeout(() => {
        setNavHeight();
    }, navSetHeightDelay);
    resetFont.classList.remove("hidden");
}

function scroolToTop() {
  if (readerMode) {
    window.scrollTo({ top: 0, behavior: 'smooth' });
  } else {
    let scrlElm = window;

    // {{if not .RDMode}} this is for when just dict mode
    window.scrollTo({ top: 0, behavior: 'smooth' });
    // {{else}}
    dict_container.scrollTo({ top: 0, behavior: 'smooth' });
    scrlElm = dict_container;
    // {{end}}


    if (scrlElm.scrollTop != null&& scrlElm.scrollTop === 0 || scrlElm.scrollX === 0) {
      w.focus();
      w.setSelectionRange(w.value.length, w.value.length);
    }

    // else {
    //   scrlElm.addEventListener('scroll', function onScroll() {
    //     if (scrlElm.scrollTop === 0) {
    //       // Remove the event listener to avoid multiple triggers
    //       scrlElm.removeEventListener('scroll', onScroll);
    //       w.focus();
    //       w.setSelectionRange(w.value.length, w.value.length);
    //     }
    //   }, { passive: true });
    // }
  }
}

function resetFontSize() {
    console.log("fontsize reset-edd")
    localStorage.removeItem("font-size-for-kamusssss");
    document.body.style.fontSize = "";
    setTimeout(() => {
        setNavHeight();
    }, navSetHeightDelay);
    resetFont.classList.add("hidden");
}

document.getElementById("plus").onclick = fontSizeInc;
document.getElementById("minus").onclick = fontSizeDec;
document.getElementById("up").onclick = scroolToTop;

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
