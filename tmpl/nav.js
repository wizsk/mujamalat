// nav.js
const selectedDictIdName = "sw-dict-item-selected";
for (let i = 0; i < dicts.length; i++) {
    dicts[i].onclick = async (e) => {
        e.preventDefault();

        const cur = e.target.getAttribute('data-dict-name');
        if (selectedDict === cur) return;
        document.getElementById(selectedDictIdName).id = "";
        e.target.id = selectedDictIdName;
        selectedDict = cur;
        selectedDictAr = e.target.getAttribute('data-dict-name-ar');

        if (cur === "ar_en") {
            ar_en_style.disabled = false;
            eng_style.disabled = true;
        } else if (cur === "hanswehr" || cur === "lanelexcon") {
            eng_style.disabled = false;
            ar_en_style.disabled = true;
        } else {
            eng_style.disabled = true;
            ar_en_style.disabled = true;
        }

        // {{if not .RDMode}}
        let p = "";
        if (preQuery !== "") {
            p = `?w=${preQuery}&idx=${queryIdx}`;
        }
        document.title = `${selectedDictAr}${currWord !== "" ? ": " + currWord : ""}`;
        window.history.replaceState(null, '', `/${selectedDict}${p}`);
        // {{end}}

        getResAndShow(currWord);
    }
}

/**
 *
 * @param {boolean} show
 * @param {boolean|null} force
 * @returns
 */
function showHideNav(show, force) {
    if (show) {
        if (!navHidden && !force) return;

        if (!readerMode) nav.style.transform = "";
        overlay.style.transform = "";
        navHidden = false;
    } else {
        if (navHidden && !force) return;

        if (!readerMode) {
            const s = querySelector.classList.contains('hidden') ? getFullHeight(nav)
                : getFullHeight(form) + getFullHeight(sw_dict);

            nav.style.transform = `translateY(-${s}px)`;
        }
        overlay.style.transform = `translateY(180px)`;
        navHidden = true;
    }
}

/** Set the div which will take space so, other elements don't do behind the nav */
function setNavHeight() { navSpace.style.height = `${nav.offsetHeight + 20}px` }

/**
 *
 * @param {HTMLButtonElement} el
 * @param {string} word
 * @param {number} idx
 */
async function changeQueryIdx(el, word, idx) {
    queryIdx = idx;
    // {{if not .RDMode}}
    document.title = `${selectedDictAr}: ${word}`;
    window.history.replaceState(null, '', `${window.location.pathname}?w=${preQuery}&idx=${queryIdx}`);
    // {{end}}

    const old = document.getElementById('querySelector-item-selected');
    if (old) old.id = "";

    el.id = 'querySelector-item-selected';
    getResAndShow(word);
}