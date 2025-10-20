// don't remove this comment
let selectedDict = "{{.Curr}}";
// if false then the dictionary is open or reading :D
let readerMode = "{{if .RDMode}}y{{end}}" === "y";
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
    if (readerMode) return;
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
let navHidden = false;
window.addEventListener("scroll", function () {
    let currentScroll = window.pageYOffset || document.documentElement.scrollTop;
    showHideNav(currentScroll < lastScrollTop);
    lastScrollTop = currentScroll <= 0 ? 0 : currentScroll; // Prevent negative scroll values
});

// {{if .RDMode}}
let lastScrollTopDict = 0;
dict_container.addEventListener("scroll", function () {
    let currentScroll = dict_container.scrollTop;
    showHideNav(currentScroll < lastScrollTopDict);
    lastScrollTopDict = currentScroll <= 0 ? 0 : currentScroll; // Prevent negative scroll values
});
// {{end}}

form.onsubmit = (e) => {
    e.preventDefault();
    // {{if not .RDMode}}
    window.location.href = `${window.location.pathname}?w=${w.value}&idx=${queryIdx}`;
    // {{end}}
}

let searhInvId;
w.oninput = () => {
    clearInterval(searhInvId);
    searhInvId = setTimeout(async () => {
        const query = w.value.trim().replace(/(\s+)/, " ");
        if (query === preQuery) return;
        preQuery = query;
        const queryArr = query.split(" ");
        queryIdx = queryArr.length - 1;
        const word = queryArr[queryIdx];
        currWord = word;

        getResAndShow(word);


        querySelector.innerHTML = "";
        if (queryArr.length > 1) {
            let b = "";
            for (let i = 0; i < queryArr.length; i++) {
                const v = queryArr[i];
                let idx = "";
                if (i < 10)
                    idx = `${i + 1}:`.replace(/[0-9]/g, (d) =>
                        String.fromCharCode(0x0660 + parseInt(d)));

                b += `<button onclick='changeQueryIdx(this, ${JSON.stringify(v)}, ${i})'
                class="querySelector-item" id="${queryArr.length - 1 === i ? 'querySelector-item-selected' : ''}">
                ${idx}${v}</button>`
            }
            console.log(b)
            querySelector.innerHTML = b;
            querySelector.classList.remove('hidden');
        } else {
            querySelector.innerHTML = "";
            querySelector.classList.add('hidden');
        }

        navSpace.style.height = `${nav.offsetHeight + 20}px`;

        // {{if not .RDMode}}
        document.title = `${selectedDictAr}: ${word}`;
        window.history.replaceState(null, '', `${window.location.pathname}?w=${query}&idx=${queryIdx}`);
        // {{end}}
    }, 250);
}

async function getResAndShow(word) {
    if (!word || word === "")
        contentHolder.innerHTML = `{{template "not-found"}}`;

    contentHolder.innerHTML = `{{template "wait"}}`;

    console.log(`req: /content?dict=${selectedDict}&w=${word}`);
    const r = await fetch(`/content?dict=${selectedDict}&w=${word}`).catch((err) =>
        console.error(err)
    );

    if (r && r.ok) {
        const h = await r.text();
        contentHolder.innerHTML = h;
    } else {
        contentHolder.innerHTML = `{{template "server-issue"}}`;
    }
}

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
const wordSpans = reader.querySelectorAll("span");
for (let i = 0; i < wordSpans.length; i++) {
    const w = wordSpans[i].innerText;
    wordSpans[i].onclick = () => openDictionay(w);
}

function openDictionay(w) {
    readerMode = false;
    console.log(w);
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

// {{end}}