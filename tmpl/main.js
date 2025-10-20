// don't remove this comment
let selectedDict = "{{.Curr}}";
let selectedDictAr = "{{index .DictsMap .Curr}}";
const contentHolder = document.getElementById("content");
const dicts = document.getElementsByClassName('sw-dict-item');
const changeDict = document.getElementById("change-dict");
const changeDictInpt = document.getElementById("change-dict-inpt");
const input = w;
let currWord = "{{if .Queries}}{{index .Queries .Idx}}{{end}}";
let preQuery = "{{.Query}}"; // this is current query belive it or not! lol
let queryIdx = 0;
try {
    queryIdx = parseInt("{{.Idx}}");
} catch (er) {
    console.log("warn:", er)
    queryIdx = 0;
}
let isChangeDictShwoing = false;


let resizeTimoutId;
window.addEventListener("resize", () => {
    clearInterval(resizeTimoutId);
    resizeTimoutId = setTimeout(() => {
        console.log("resized");
        navSpace.style.height = `${nav.offsetHeight + 20}px`;
    }, 100);
});

document.addEventListener("DOMContentLoaded", () => {
    setSavedFontSize();
    setNavHeight();

    const selected = document.getElementById('sw-dict-item-selected');
    if (selected && selected.scrollIntoView) {
        selected.scrollIntoView({
            // behavior: 'smooth',    // auto Or 'smooth' if you want animation
            block: 'nearest',
            inline: 'center'     // Center the element horizontally
        });
    }

    // {{if not .RDMode}} 
    if (w.value.length === 0) w.focus();
    let p = "";
    if (currWord !== "") {
        p = `?w=${preQuery}&idx=${queryIdx}`
    }
    window.history.replaceState(null, '', `${window.location.pathname}${p}`);
    // {{end}}
});

w.onfocus = () => showHideNav(true);

let lastScrollTop = 0;
let navHidden = true;
window.addEventListener("scroll", function () {
    let currentScroll = window.pageYOffset || document.documentElement.scrollTop;
    showHideNav(currentScroll < lastScrollTop);
    lastScrollTop = currentScroll <= 0 ? 0 : currentScroll; // Prevent negative scroll values
});

form.onsubmit = (e) => {
    e.preventDefault();
    window.location.href = `${window.location.pathname}?w=${w.value}&idx=${queryIdx}`;
}

let searhInvId;
w.oninput = () => {
    clearInterval(searhInvId);
    searhInvId = setTimeout(async () => {
        const query = w.value.trim().replace(/(\s+)/, " ");
        if (query === preQuery) return;
        if (query === "") {
            preQuery = "";
            return;
        }
        preQuery = query;

        const queryArr = query.split(" ");
        queryIdx = queryArr.length - 1;

        const word = queryArr[queryIdx];
        currWord = word;

        getResAndShow(word);


        querySelector.innerHTML = "";
        if (queryArr.length > 1) {
            for (let i = 0; i < queryArr.length; i++) {
                const v = queryArr[i];
                let idx = "";
                if (i < 10)
                    idx = `${i + 1}:`.replace(/[0-9]/g, (d) =>
                        String.fromCharCode(0x0660 + parseInt(d)));

                const c = document.createElement("button");
                c.onclick = (e) => changeQueryIdx(e, v, i);

                c.classList.add('querySelector-item')
                c.id = id = `${queryArr.length - 1 === i ? 'querySelector-item-selected' : ''}`;
                c.innerText = `${idx}${v}`;
                // b += `<button onclick="changeQueryIdx(this, ${JSON.stringify(v)}, ${i})"
                // class="querySelector-item" id="${queryArr.length - 1 === i ? 'querySelector-item-selected' : ''}">
                // ${idx}${v}</button>`
                querySelector.appendChild(c);
            }
            querySelector.classList.remove('hidden');
        } else {
            querySelector.innerHTML = "";
            querySelector.classList.add('hidden');
        }

        navSpace.style.height = `${nav.offsetHeight + 20}px`;

        // {{if not .RDMode}}
        document.title = `${selectedDictAr}: ${word}`;
        window.history.replaceState(null, '', `${window.location.pathname}?w=${word}&idx=${queryIdx}`);
        // {{end}}
    }, 250);
}

async function getResAndShow(word) {
    if (!word || word === "")
        contentHolder.innerHTML = "";

    console.log(`req: /content?dict=${selectedDict}&w=${word}`);
    const r = await fetch(`/content?dict=${selectedDict}&w=${word}`).catch((err) =>
        console.error(err)
    );

    if (r && r.ok) {
        const h = await r.text();
        contentHolder.innerHTML = h;
    } else {
        contentHolder.innerHTML =
            `<div style="direction: ltr; text-align: center;
                margin-top: 4rem; color: var(--alert);">
                    Cound't fetch results. Is the server running?
                </div>`;
    }
}

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

/**  @param {boolean} show */
function showHideNav(show) {
    if (show) {
        if (navHidden) return;
        nav.style.transform = "";
        overlay.style.transform = "";
        navHidden = true;
    } else {
        if (!navHidden) return;
        const s = querySelector.classList.contains('hidden') ? getFullHeight(nav)
            : getFullHeight(form) + getFullHeight(sw_dict);
        nav.style.transform = `translateY(-${s}px)`;
        overlay.style.transform = `translateY(100px)`;
        navHidden = false;
    }
}

/** Set the div which will take space so, other elements don't do behind the nav */
const setNavHeight = () => navSpace.style.height = `${nav.offsetHeight + 20}px`

function getFullHeight(element) {
    const rect = element.getBoundingClientRect();
    const style = window.getComputedStyle(element);

    const marginTop = parseFloat(style.marginTop);
    const marginBottom = parseFloat(style.marginBottom);
    const paddingTop = parseFloat(style.paddingTop);
    const paddingBottom = parseFloat(style.paddingBottom);
    const borderTop = parseFloat(style.borderTopWidth);
    const borderBottom = parseFloat(style.borderBottomWidth);

    // Full height = content height + padding + border + margin
    const fullHeight = rect.height + marginTop + marginBottom + paddingTop + paddingBottom + borderTop + borderBottom;
    return fullHeight;
}

// {{if .RDMode}}
function s(w) {
    console.log(w);
    preQuery = w;
    currWord = w;
    queryIdx = 0;
    input.value = w;
    getResAndShow(w);
    dict_container.classList.remove("hidden");
    setNavHeight();
}
// {{end}}